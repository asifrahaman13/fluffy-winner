package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/asifrahaman13/bhagabad_gita/internal/config"
	"github.com/asifrahaman13/bhagabad_gita/internal/core/domain"
	"github.com/asifrahaman13/bhagabad_gita/internal/helper"
	"github.com/gorilla/websocket"
	"github.com/qdrant/go-client/qdrant"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	OUTPUT_PATH          = "static/output.json"
	COLLECTION_NAME      = "test_collection"
	EMBEDDING_MODEL_NAME = "mxbai-embed-large"
	EMBEDDING_URL        = "http://localhost:11434/api/embeddings"
)

type EmbeddingService struct {
	url string
}

type QdrantService struct {
	client *qdrant.Client
}

func (e *EmbeddingService) GetEmbedding(content string) ([]float32, error) {
	payload := map[string]string{
		"model":  EMBEDDING_MODEL_NAME,
		"prompt": fmt.Sprintf("Represent this sentence for searching relevant passages: %s", content),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %v", err)
	}
	resp, err := http.Post(e.url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error making request to embedding API: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from embedding API: %s", string(body))
	}
	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error decoding embedding API response: %v", err)
	}
	return result.Embedding, nil
}

func (q *QdrantService) VectorSearch(query string, embeddingService *EmbeddingService) ([]map[string]interface{}, error) {
	embedding, err := embeddingService.GetEmbedding(query)
	if err != nil {
		return nil, err
	}
	limit := uint64(3)
	searchResult, err := q.client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: COLLECTION_NAME,
		Query:          qdrant.NewQuery(embedding...),
		Limit:          &limit,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}
	var results []map[string]interface{}
	for _, res := range searchResult {
		results = append(results, map[string]interface{}{
			"pageNum": res.Id.GetNum(),
			"content": res.Payload["pageContent"].GetStringValue(),
		})
	}
	return results, nil
}

func chatBotResponse(prompt domain.WebsocketMessage, conn *websocket.Conn) {
	config, err := config.NewConfig()
	if err != nil {
		fmt.Println("Error getting config:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error getting config"))
		return
	}
	postUrl := config.LLamaUrl
	fmt.Printf("The message is from the client: %s and the client is: %s, message id is: %d, message type is: %s", prompt.Payload, prompt.ClientId, prompt.MessageId, prompt.MsgType)
	body := []byte(fmt.Sprintf(`{
		"model": "llama3.1",
		"stream": true,
		"prompt": "%s"
	}`, prompt.Payload))

	req, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error creating request:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error creating request"))
		return
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error making request"))
		return
	}
	defer res.Body.Close()

	var buffer strings.Builder
	for {
		var chatResponse domain.ChatResponse
		err = json.NewDecoder(res.Body).Decode(&chatResponse)
		if err != nil {
			fmt.Println("Error decoding response:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(""))
			return
		}
		buffer.WriteString(chatResponse.Response)
		if helper.IsSentenceEnd(*bytes.NewBufferString(buffer.String())) {
			textMessages := domain.WebsocketMessage{
				ClientId:  prompt.ClientId,
				MessageId: prompt.MessageId,
				Payload:   buffer.String(),
				MsgType:   "server",
			}

			jsonStringMessage, err := json.Marshal(textMessages)
			if err != nil {
				fmt.Println("Error marshaling message:", err)
				conn.WriteMessage(websocket.TextMessage, []byte("Error marshaling message"))
				return
			}

			err = conn.WriteMessage(websocket.TextMessage, jsonStringMessage)
			if err != nil {
				fmt.Println("Error sending message:", err)
				conn.Close()
				return
			}
			buffer.Reset()
		}
	}
}

func NewEmbeddingService(url string) *EmbeddingService {
	return &EmbeddingService{url: url}
}
func ErrorHandler(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func NewQdrantService(host string, port int) *QdrantService {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	ErrorHandler(err)
	return &QdrantService{client: client}
}

func HandleWebSocketConnection(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			conn.Close()
			break
		}
		var messageStruct domain.WebsocketMessage
		err = json.Unmarshal(message, &messageStruct)
		if err != nil {
			fmt.Println("Error decoding message:", err)
		}
		embeddingService := NewEmbeddingService(EMBEDDING_URL)
		qdrantService := NewQdrantService("localhost", 6334)
		result, err := qdrantService.VectorSearch(messageStruct.Payload, embeddingService)
		if err != nil {
			fmt.Println("Error searching vectors:", err)
			return
		}
		allContext := ""
		for _, res := range result {
			trimmedContent := strings.TrimSpace(res["content"].(string))
			allContext += trimmedContent + "\n"
		}
		fmt.Println(allContext)
		allContext = strings.ReplaceAll(allContext, "\n", " ")
		messageStruct.Payload = fmt.Sprintf("You are an expert in spiritaul answers. User has the following query. Answer the query: %s . Also you have some additional context to give better ansser: %s", messageStruct.Payload, allContext)
		go chatBotResponse(messageStruct, conn)
	}
}
