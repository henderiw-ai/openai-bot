package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/henderiw-ai/openai-bot/apis/vector"
	"github.com/henderiw-ai/openai-bot/pkg/utils"
	"github.com/sashabaranov/go-openai"
)

const jsonExtension = ".json"

func main() {

	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)
	/*

		resp, err := client.CreateEmbeddings(
			context.Background(),
			openai.EmbeddingRequest{
				Input: []string{"i am wim", "i am mieke"},
				Model: openai.AdaEmbeddingV2,
				User:  "henderiw",
			},
		)
		if err != nil {
			fmt.Printf("Embeddings error: %v\n", err)
			return
		}

		//fmt.Println(resp)
		for _, data := range resp.Data {
			fmt.Printf("idx: %d, len floats: %d, object: %s\n", data.Index, len(data.Embedding), data.Object)
		}
	*/

	//fmt.Println("usage", resp.Usage)
	//fmt.Printf("data: %v, len: %d\n", resp.Data, len(resp.Data))
	//fmt.Println("model", resp.Model)

	files, err := utils.ReadFiles("../../data/output", jsonExtension, true)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fd, err := os.ReadFile(f)
		if err != nil {
			panic(err)
		}
		v := &vector.Vector{}
		if err := json.Unmarshal(fd, v); err != nil {
			panic(err)
		}
		fmt.Println()
		fmt.Println(filepath.Base(f))
		fmt.Printf("id: %s\ndata: %s\n", v.Id, v.Data)

		resp, err := client.CreateEmbeddings(
			context.Background(),
			openai.EmbeddingRequest{
				Input: []string{v.Data},
				Model: openai.AdaEmbeddingV2,
				User:  "henderiw",
			},
		)
		if err != nil {
			panic(err)
		}
		v.Values = resp.Data[0].Embedding

		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			panic(err)
		}
		os.WriteFile(filepath.Join("../../data/embeddings", filepath.Base(f)), b, 0644)
	}
}
