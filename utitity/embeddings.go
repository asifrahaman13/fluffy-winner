package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"github.com/google/uuid"
	"github.com/pdfcrowd/pdfcrowd-go"
	"github.com/qdrant/go-client/qdrant"
)

const (
	OutputPath         = "static/output.json"
	FinalOutputPath    = "static/result.json"
	CollectionName     = "test_collection"
	EmbeddingModelName = "mxbai-embed-large"
	EmbeddingURL       = "http://localhost:11434/api/embeddings"
)

// ErrorHandler handles errors
func ErrorHandler(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// PDFProcessor handles PDF to JSON conversion
type PDFProcessor struct {
	client pdfcrowd.PdfToTextClient
}

func NewPDFProcessor() *PDFProcessor {
	client := pdfcrowd.NewPdfToTextClient("demo", "ce544b6ea52a5621fb9d55f8b542d14d")
	client.SetPageBreakMode("custom")
	client.SetCustomPageBreak("\n---PAGE_BREAK---\n")
	return &PDFProcessor{client: client}
}

func (p *PDFProcessor) ConvertToJSON(inputPath string) {
	txt, err := p.client.ConvertFile(inputPath)
	ErrorHandler(err)
	pages := strings.Split(string(txt), "\n---PAGE_BREAK---\n")
	var pageData []map[string]interface{}
	for i, pageContent := range pages {
		pageData = append(pageData, map[string]interface{}{
			"pageNum":     i + 1,
			"pageContent": strings.TrimSpace(pageContent),
		})
	}
	jsonData, err := json.MarshalIndent(pageData, "", "  ")
	ErrorHandler(err)
	err = os.WriteFile(OutputPath, jsonData, 0644)
	ErrorHandler(err)
	fmt.Println("Data successfully written to output.json")
}

// EmbeddingService handles embedding requests
type EmbeddingService struct {
	url string
}

func NewEmbeddingService(url string) *EmbeddingService {
	return &EmbeddingService{url: url}
}

func (e *EmbeddingService) GetEmbedding(content string) ([]float32, error) {
	payload := map[string]string{
		"model":  EmbeddingModelName,
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

// QdrantService handles interactions with Qdrant
type QdrantService struct {
	client *qdrant.Client
}

func NewQdrantService(host string, port int) *QdrantService {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	ErrorHandler(err)
	return &QdrantService{client: client}
}

func (q *QdrantService) UpsertEmbeddings(pageData []map[string]interface{}, embeddingService *EmbeddingService) {
	q.client.CreateCollection(context.Background(), &qdrant.CreateCollection{
		CollectionName: CollectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     1024,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	var points []*qdrant.PointStruct
	for _, page := range pageData {
		pageContent := page["pageContent"].(string)
		pageNum := page["pageNum"].(float64)
		embedding, err := embeddingService.GetEmbedding(pageContent)
		if err != nil {
			fmt.Printf("Error getting embedding for page %v: %v\n", pageNum, err)
			continue
		}
		payload := qdrant.NewValueMap(map[string]any{
			"pageContent": pageContent,
			"pageNum":     pageNum,
		})
		uuid := uuid.New().String()
		point := &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(uuid),
			Vectors: qdrant.NewVectors(embedding...),
			Payload: payload,
		}
		points = append(points, point)
	}
	operationInfo, err := q.client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: CollectionName,
		Points:         points,
	})
	ErrorHandler(err)
	fmt.Println("Upsert operation successful:", operationInfo)
}

func (q *QdrantService) VectorSearch(query string, embeddingService *EmbeddingService) ([]map[string]interface{}, error) {
	embedding, err := embeddingService.GetEmbedding(query)
	if err != nil {
		return nil, err
	}
	limit := uint64(3)
	searchResult, err := q.client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: CollectionName,
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

type Page struct {
	PageContent string `json:"pageContent"`
	PageNum     int    `json:"pageNum"`
	PageIdx     int    `json:"pageIdx"`
}

func splitIntoChunks(content string, maxWords int) []string {
	words := strings.Fields(content)
	var chunks []string

	for i := 0; i < len(words); i += maxWords {
		end := i + maxWords
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)
	}

	return chunks
}

func split() {
	// Input and output file paths
	inputFile := "static/output.json"
	outputFile := "static/result.json"

	// Read the input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		return
	}

	// Parse the input JSON
	var pages []Page
	err = json.Unmarshal(data, &pages)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	var processedPages []Page

	// Process each page
	for _, page := range pages {
		if len(strings.Fields(page.PageContent)) > 100 {
			chunks := splitIntoChunks(page.PageContent, 100)
			for idx, chunk := range chunks {
				processedPages = append(processedPages, Page{
					PageContent: chunk,
					PageNum:     page.PageNum,
					PageIdx:     idx,
				})
				// Optional: Adjust page numbers if required
			}
		} else {
			processedPages = append(processedPages, page)
		}
	}

	// Write the output file
	outputData, err := json.MarshalIndent(processedPages, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	err = os.WriteFile(outputFile, outputData, 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		return
	}

	fmt.Println("Processing complete. Output written to", outputFile)
}

func main() {
	// split()
	// pdfProcessor := NewPDFProcessor()
	embeddingService := NewEmbeddingService(EmbeddingURL)
	qdrantService := NewQdrantService("localhost", 6334)

	// // Step 1: Convert PDF to JSON
	// pdfProcessor.ConvertToJSON("static/gita.pdf")
	// // split()

	// // Step 2: Load JSON and Upsert Embeddings
	// jsonFile, err := os.ReadFile(FinalOutputPath)
	// ErrorHandler(err)
	// var pageData []map[string]interface{}
	// err = json.Unmarshal(jsonFile, &pageData)
	// ErrorHandler(err)
	// qdrantService.UpsertEmbeddings(pageData, embeddingService)

	// Step 3: Perform Vector Search
	query := "how a man should treat a wife"
	results, err := qdrantService.VectorSearch(query, embeddingService)
	ErrorHandler(err)
	for _, res := range results {
		fmt.Printf("PageNum: %v\nContent: %v\n\n", res["pageNum"], res["content"])
	}

}
