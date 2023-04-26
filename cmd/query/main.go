package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/henderiw-ai/openai-bot/apis/vector"
	"github.com/henderiw-ai/openai-bot/pkg/utils"
	"github.com/sashabaranov/go-openai"
)

const jsonExtension = ".json"

type Result struct {
	Id     string
	Score  float64
	Tokens int
	Data   string
}

type Bot struct {
	client *openai.Client
	vdb    map[string]vector.Vector
}

func (r *Bot) Query(q string) ([]Result, error) {
	resp, err := r.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{q},
			Model: openai.AdaEmbeddingV2,
			User:  "henderiw",
		},
	)
	if err != nil {
		return nil, err
	}

	results := []Result{}
	for id, v := range r.vdb {
		r := Result{
			Id:     id,
			Score:  cosineDistance(resp.Data[0].Embedding, v.Values),
			Tokens: v.Tokens,
			Data:   v.Data,
		}
		results = append(results, r)
	}
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score < results[j].Score
	})
	return results, nil
}

func (r *Bot) Complete(q string, results []Result) (*openai.ChatCompletionResponse, error) {
	var sb strings.Builder
	sb.WriteString("The answers should be based on the following input\n")

	tokens := 0
	for _, result := range results {
		fmt.Println("tokens: ", tokens)
		tokens += result.Tokens
		if tokens > 2000 {
			break
		}
		sb.WriteString(result.Data)
	}

	fmt.Println()
	fmt.Println(sb.String())
	fmt.Println()

	resp, err := r.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistent answering questions about condition kpt sdk, an sdk for kpt packages",
					Name:    "henderiw",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "keep the answers short",
					Name:    "henderiw",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: q,
					Name:    "henderiw",
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: sb.String(),
					Name:    "henderiw",
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func main() {
	files, err := utils.ReadFiles("../../data/embeddings", jsonExtension, true)
	if err != nil {
		panic(err)
	}
	vdb := map[string]vector.Vector{}
	for _, f := range files {
		fd, err := os.ReadFile(f)
		if err != nil {
			panic(err)
		}
		v := &vector.Vector{}
		if err := json.Unmarshal(fd, v); err != nil {
			panic(err)
		}

		split := strings.Split(filepath.Base(f), ".")
		vdb[split[0]] = *v
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)

	bot := Bot{
		client: client,
		vdb:    vdb,
	}

	for {
		q := StringPrompt("Ask a Question?")
		results, err := bot.Query(q)
		if err != nil {
			panic(err)
		}
		fmt.Println()
		for _, result := range results {
			fmt.Printf("result id %s score: %f\n", result.Id, result.Score)
		}

		resp, err := bot.Complete(q, results)
		if err != nil {
			panic(err)
		}
		fmt.Println()
		fmt.Println(resp)
	}

}

func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func dot(v1, v2 []float32) float64 {
	var sum float64
	for i, f := range v1 {
		sum += float64(f) * float64(v2[i])
	}
	return sum
}

func norm(v []float32, pow float64) float64 {
	var s float64
	for _, xval := range v {
		s += float64(xval) * float64(xval)
	}
	if s == 0 {
		return 0
	}
	return math.Pow(s, 1/pow)
}

func cosine(v1, v2 []float32) float64 {
	v1Norm := norm(v1, 2)
	v2Norm := norm(v2, 2)
	dot := dot(v1, v2)
	return (dot / (v1Norm * v2Norm))
}

func cosineDistance(v1, v2 []float32) float64 {
	return 1 - cosine(v1, v2)
}
