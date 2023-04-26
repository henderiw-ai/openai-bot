package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/henderiw-ai/openai-bot/apis/vector"
	"github.com/henderiw-ai/openai-bot/pkg/utils"
	"github.com/pkoukk/tiktoken-go"
)

const markdownExtension = ".md"

func main() {
	files, err := utils.ReadFiles("../../data", markdownExtension, true)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))

		extensions := parser.AutoHeadingIDs | parser.CommonExtensions
		p := parser.NewWithExtensions(extensions)
		doc := p.Parse(b)
		docs, hls, err := GenerateHeadingDocs(doc)
		if err != nil {
			panic(err)
		}
		fmt.Println(len(docs))
		for title, content := range docs {
			if len(content) == 0 {
				delete(docs, title)
			}
		}
		fmt.Println(len(docs))
		for title, content := range docs {
			fmt.Println()
			fmt.Println("title:", title)
			content := strings.ReplaceAll(string(content), "\n", "")
			content = strings.ReplaceAll(string(content), "\t", "")
			content = strings.ReplaceAll(string(content), "\r", "")
			content = strings.ReplaceAll(string(content), "  ", " ")
			content = strings.TrimSpace(string(content))
			fmt.Println(string(content))
			docs[title] = []byte(content)
		}
		fmt.Println()
		for _, hl := range hls {
			fmt.Println(hl)
		}

		tkm, err := tiktoken.EncodingForModel("text-embedding-ada-002")
		if err != nil {
			err = fmt.Errorf("EncodingForModel: %v", err)
			panic(err)
		}

		for title, content := range docs {
			fmt.Printf("title: %s, tokens: %d\n", title, len(tkm.Encode(string(content), nil, nil)))

			v := &vector.Vector{
				Id:     title,
				Tokens: len(tkm.Encode(string(content), nil, nil)),
				Data:   string(content),
			}

			b, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				panic(err)
			}
			os.WriteFile(filepath.Join("../../data/output", fmt.Sprintf("%s.json", title)), b, 0644)
		}
	}
}

type hlinks []string

func (r *hlinks) Add(hl string) {
	exists := false
	for _, l := range *r {
		if hl == l {
			exists = true
			break
		}
	}
	if !exists {
		*r = append(*r, hl)
	}
}

func GenerateHeadingDocs(n ast.Node) (map[string][]byte, []string, error) {
	docs := map[string][]byte{}
	hls := hlinks{}
	heading := 0
	headingID := ""

	ast.WalkFunc(n, func(node ast.Node, entering bool) ast.WalkStatus {
		switch x := node.(type) {
		case *ast.Heading:
			fmt.Println("#################")
			fmt.Printf("heading: idx: %d level: %d, id: %s, title: %t, special: %t\n", heading, x.Level, x.HeadingID, x.IsTitleblock, x.IsSpecial)
			if heading == 1 {
				headingID = x.HeadingID
				docs[headingID] = []byte{}
			}
			heading = (heading + 1) % 2
			fmt.Println("#################")
		case *ast.Link:
			fmt.Printf("link Destination: %s\n", string(x.Destination))
			hls.Add(string(x.Destination))
		default:
			// ignore the leaf/container if we are processing a header
			if heading == 0 {
				if l := node.AsLeaf(); l != nil {
					//fmt.Printf("leaf: content: %v, literal: %v\n", string(l.Content), string(l.Literal))
					if len(l.Literal) != 0 {
						docs[headingID] = append(docs[headingID], l.Literal...)
					}
				}
				//if c := node.AsContainer(); c != nil {
				//fmt.Printf("container: children %v\n", node.GetChildren())
				//}
			}
		}
		return ast.GoToNext
	})
	return docs, hls, nil
}
