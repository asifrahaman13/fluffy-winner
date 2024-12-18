package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/asifrahaman13/bhagabad_gita/internal/config"
	"github.com/asifrahaman13/bhagabad_gita/internal/helper"
	"github.com/gorilla/websocket"
	"github.com/qdrant/go-client/qdrant"
	"io"
	"net/http"
	"strings"
)

type ChatResponse struct {
	Response string `json:"response"`
}

type WebsocketMessage struct {
	ClientId  string `json:"clientId"`
	MessageId int    `json:"messageId"`
	Payload   string `json:"payload"`
	MsgType   string `json:"msgType"`
}
type VectorSearchResult struct {
	PageNum uint64 `json:"rank"`
	Content string `json:"content"`
}

const (
	OUTPUT_PATH          = "static/output.json"
	COLLECTION_NAME      = "test_collection"
	EMBEDDING_MODEL_NAME = "mxbai-embed-large"
	EMBEDDING_URL        = "http://localhost:11434/api/embeddings"
)

func getOllamaEmbedding(content string) ([]float32, error) {
	payload := map[string]string{
		"model":  EMBEDDING_MODEL_NAME,
		"prompt": fmt.Sprintf("Represent this sentence for searching relevant passages: %s", content),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %v", err)
	}
	resp, err := http.Post(EMBEDDING_URL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error making request to Ollama API: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from Ollama API: %s", string(body))
	}
	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error decoding Ollama API response: %v", err)
	}
	return result.Embedding, nil
}

func vectorSearch(query string) ([]VectorSearchResult, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		fmt.Println("Error creating client:", err)
		return []VectorSearchResult{}, err
	}
	embedding, err := getOllamaEmbedding(query)
	if err != nil {
		fmt.Println("Error getting embedding:", err)
		return []VectorSearchResult{}, err
	}
	limit := uint64(3)
	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: COLLECTION_NAME,
		Query:          qdrant.NewQuery(embedding...),
		Limit:          &limit,
		WithPayload:    qdrant.NewWithPayload(true),
		WithVectors:    qdrant.NewWithVectors(false),
	})
	if err != nil {
		panic(err)
	}
	result := []VectorSearchResult{}
	for _, res := range searchResult {
		pageContent := res.Payload["pageContent"]
		contentStr := pageContent.GetStringValue()
		result = append(result, VectorSearchResult{
			PageNum: res.Id.GetNum(),
			Content: contentStr,
		})
	}
	return result, nil
}

func main() {

	query := "brothers-in-law, grandfathers and so on. He was thinking in this way to satisfy\nhis bodily demands. Bhagavad-gītā was spoken by the Lord just to change this\n-DEMO-, and at the end Arjuna decides to fight under the directions of the Lord\nwhen he says, \"kariṣye vacanaṁ tava.\" \"I shall act according to Thy word.\"\n   In this world man is not meant to toil like hogs. He must be intelligent to\nrealize the importance of human life and refuse to act like an ordinary animal.\nA human being should realize the aim of his life, and this direction is given in\nall Vedic literatures, and the essence is given in Bhagavad-gītā. Vedic\nliterature is meant for human beings, not for animals. Animals can kill -DEMO-\nliving animals, and there is no question of sin on their part, but if a man kills\nan animal for the satisfaction of his uncontrolled taste, he must -DEMO- responsible\nfor breaking the laws of nature. In the Bhagavad-gītā it is clearly explained\nthat there are three kinds of activities according to the different modes of\nnature: the activities of goodness,"
	result, err := vectorSearch(query)
	if err != nil {
		fmt.Println("Error searching vectors:", err)
		return
	}
	for _, res := range result {
		fmt.Printf("Rank: %d\nContent: %s\n\n", res.PageNum, res.Content)
	}

}

func chatBotResponse(prompt WebsocketMessage, conn *websocket.Conn) {
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
		var chatResponse ChatResponse
		err = json.NewDecoder(res.Body).Decode(&chatResponse)
		if err != nil {
			fmt.Println("Error decoding response:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(""))
			return
		}
		buffer.WriteString(chatResponse.Response)
		if helper.IsSentenceEnd(*bytes.NewBufferString(buffer.String())) {
			textMessages := WebsocketMessage{
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

func HandleWebSocketConnection(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			conn.Close()
			break
		}
		var messageStruct WebsocketMessage
		err = json.Unmarshal(message, &messageStruct)
		if err != nil {
			fmt.Println("Error decoding message:", err)
		}
		result, err := vectorSearch(messageStruct.Payload)
		if err != nil {
			fmt.Println("Error searching vectors:", err)
			return
		}
		allContext := ""
		for _, res := range result {
			trimmedContent := strings.TrimSpace(res.Content)
			allContext += trimmedContent + "\n"
		}
		fmt.Println(allContext)
		allContext = strings.ReplaceAll(allContext, "\n", " ")
		messageStruct.Payload = fmt.Sprintf("You are an expert in spiritaul answers. User has the following query. Answer the query: %s . Also you have some additional context to give better ansser: %s", messageStruct.Payload, allContext)
		go chatBotResponse(messageStruct, conn)
	}
}
