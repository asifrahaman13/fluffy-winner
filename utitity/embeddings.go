package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pdfcrowd/pdfcrowd-go"
	"github.com/qdrant/go-client/qdrant"
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

func embeddings() {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})

	if err != nil {
		fmt.Println("Error creating client:", err)
		return
	}
	// The Go client uses Qdrant's gRPC interface

	// client.CreateCollection(context.Background(), &qdrant.CreateCollection{
	// 	CollectionName: "test_collection",
	// 	VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
	// 		Size:     4,
	// 		Distance: qdrant.Distance_Cosine,
	// 	}),
	// })

	operationInfo, err := client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: "test_collection",
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewIDNum(1),
				Vectors: qdrant.NewVectors(0.05, 0.61, 0.76, 0.74),
				Payload: qdrant.NewValueMap(map[string]any{"pageContent": "Berlin", "pageNum": 1}),
			},
			{
				Id:      qdrant.NewIDNum(2),
				Vectors: qdrant.NewVectors(0.19, 0.81, 0.75, 0.11),
				Payload: qdrant.NewValueMap(map[string]any{"pageContent": "Berlin", "pageNum": 1}),
			},
			{
				Id:      qdrant.NewIDNum(3),
				Vectors: qdrant.NewVectors(0.36, 0.55, 0.47, 0.94),
				Payload: qdrant.NewValueMap(map[string]any{"pageContent": "Berlin", "pageNum": 1}),
			},
			// Truncated
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(operationInfo)
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
	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "test_collection",
		Query:          qdrant.NewQuery(0.2, 0.1, 0.9, 0.7),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(searchResult)
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
	err = os.WriteFile("output.json", jsonData, 0644)
	handleError(err)
	fmt.Println("Data successfully written to output.json")

}

func main() {
	pdfToJson()
	// embeddings()
	// vectorSearch()
}
