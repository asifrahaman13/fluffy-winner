package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pdfcrowd/pdfcrowd-go"
	"github.com/qdrant/go-client/qdrant"
	"io"
	"log"
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

func handleError(err error) {
	if err != nil {
		why, ok := err.(pdfcrowd.Error)
		if ok {
			os.Stderr.WriteString(fmt.Sprintf("Pdfcrowd Error: %s\n", why))
		} else {
			os.Stderr.WriteString(fmt.Sprintf("Generic Error: %s\n", err))
		}
		panic(err.Error())
	}
}

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

func embeddings() {
	jsonFile, err := os.ReadFile(OUTPUT_PATH)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}
	var pageData []map[string]interface{}
	err = json.Unmarshal(jsonFile, &pageData)
	if err != nil {
		fmt.Println("Error parsing JSON data:", err)
		return
	}
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		fmt.Println("Error creating Qdrant client:", err)
		return
	}
	client.CreateCollection(context.Background(), &qdrant.CreateCollection{
		CollectionName: COLLECTION_NAME,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     1024,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	var points []*qdrant.PointStruct
	for _, page := range pageData {
		pageContent := page["pageContent"].(string)
		pageNum := page["pageNum"].(float64)
		embedding, err := getOllamaEmbedding(pageContent)
		if err != nil {
			fmt.Printf("Error getting embedding for page %v: %v\n", pageNum, err)
			continue
		}
		fmt.Println("The dimension of the embedding is: ", len(embedding))
		payload := qdrant.NewValueMap(map[string]any{
			"pageContent": pageContent,
			"pageNum":     pageNum,
		})
		point := &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(uint64(pageNum)),
			Vectors: qdrant.NewVectors(embedding...),
			Payload: payload,
		}
		points = append(points, point)
	}
	operationInfo, err := client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: COLLECTION_NAME,
		Points:         points,
	})
	if err != nil {
		fmt.Println("Error upserting points:", err)
		return
	}
	fmt.Println("Upsert operation successful:", operationInfo)
}

func vectorSearch() {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		fmt.Println("Error creating client:", err)
		return
	}
	embedding, err := getOllamaEmbedding("brothers-in-law, grandfathers and so on. He was thinking in this way to satisfy\nhis bodily demands. Bhagavad-gītā was spoken by the Lord just to change this\n-DEMO-, and at the end Arjuna decides to fight under the directions of the Lord\nwhen he says, \"kariṣye vacanaṁ tava.\" \"I shall act according to Thy word.\"\n   In this world man is not meant to toil like hogs. He must be intelligent to\nrealize the importance of human life and refuse to act like an ordinary animal.\nA human being should realize the aim of his life, and this direction is given in\nall Vedic literatures, and the essence is given in Bhagavad-gītā. Vedic\nliterature is meant for human beings, not for animals. Animals can kill -DEMO-\nliving animals, and there is no question of sin on their part, but if a man kills\nan animal for the satisfaction of his uncontrolled taste, he must -DEMO- responsible\nfor breaking the laws of nature. In the Bhagavad-gītā it is clearly explained\nthat there are three kinds of activities according to the different modes of\nnature: the activities of goodness,")
	if err != nil {
		fmt.Println("Error getting embedding:", err)
		return
	}
	limit := uint64(3)
	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: COLLECTION_NAME,
		Query:          qdrant.NewQuery(embedding...),
		Limit:          &limit,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(searchResult)
	searchResultJSON, err := json.MarshalIndent(searchResult, "", "  ")
	if err != nil {
		log.Println("Error marshalling search results:", err)
		return
	}
	fmt.Println(string(searchResultJSON))
}

func pdfToJson() {
	client := pdfcrowd.NewPdfToTextClient("demo", "ce544b6ea52a5621fb9d55f8b542d14d")
	client.SetPageBreakMode("custom")
	client.SetCustomPageBreak("\n---PAGE_BREAK---\n")
	txt, err := client.ConvertFile("static/gita.pdf")
	handleError(err)
	pages := strings.Split(string(txt), "\n---PAGE_BREAK---\n")
	var pageData []map[string]interface{}
	for i, pageContent := range pages {
		pageData = append(pageData, map[string]interface{}{
			"pageNum":     i + 1,
			"pageContent": strings.TrimSpace(pageContent),
		})
	}
	jsonData, err := json.MarshalIndent(pageData, "", "  ")
	handleError(err)
	err = os.WriteFile(OUTPUT_PATH, jsonData, 0644)
	handleError(err)
	fmt.Println("Data successfully written to output.json")
}

func main() {
	// pdfToJson()
	// embeddings()
	vectorSearch()
}
