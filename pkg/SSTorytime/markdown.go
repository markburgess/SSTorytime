//*****************************************************************
//
// markdown.go
//
//*****************************************************************

package SSTorytime

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	_ "github.com/lib/pq"

)

//*****************************************************************
// Markdown reader
//*****************************************************************

// uses the goldmark lib, go get github.com/yuin/goldmark

type MDSection struct {
	Level   int    `json:"level"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var DOC_DIRECTORY = make(map[string]string)

//*****************************************************************

func FractionateMarkdown(filename string) ([][]Sentence, int) {
	
	source,err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("Couldn't find or open",filename)
		os.Exit(-1)
	}

	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(source))

	var sections []MDSection
	var current *MDSection
	var content []string

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := n.(type) {
		case *ast.Heading:
			// Save previous section
			if current != nil {
				current.Content = strings.Join(content, "\n\n")
				sections = append(sections, *current)
			}
			// Start new section
			current = &MDSection{
				Level: n.Level,
				Title: string(n.Text(source)),
			}
			content = nil

		case *ast.Paragraph:
			content = append(content, string(n.Text(source)))
		}

		return ast.WalkContinue, nil
	})

	// Don't forget the last section

	if current != nil {
		current.Content = strings.Join(content, "\n\n")
		sections = append(sections, *current)
	}

	// Now split the parts

	var pbsf [][]Sentence
	var cache = make(map[int]string)
	var count int

	for _,v := range sections {

		sentences := SplitSentences(v.Content)

		var cleaned []Sentence
		
		for s := range sentences {

			// NB, if parentheses contain multiple sentences, this complains, TBD

			frags := SplitPunctuationText(sentences[s])

			var this Sentence
			this.S = sentences[s]
			this.Title = v.Title
			cache[v.Level] = v.Title

			// Record the containment for later

			if v.Level > 1 {
				for i := v.Level; i > 0; i-- {
					if len(cache[i]) > 0 {
						DOC_DIRECTORY[v.Title] = cache[i]
						break
					}
				}
			} else {
				DOC_DIRECTORY[v.Title] = filename
			}
			
			count++
			
			for f := range frags {
				content := strings.TrimSpace(frags[f])
				if len(content) > 2 {			
					this.Frags = append(this.Frags,content)
				}
			}

			if len(this.S) > 0 {
				cleaned = append(cleaned,this)
			}
		}

		if len(cleaned) > 0 {
			pbsf = append(pbsf,cleaned)
		}
	}

	return pbsf,count
}

