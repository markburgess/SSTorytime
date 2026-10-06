//*****************************************************************
//
// markdown.go
//
// uses the goldmark lib
//
//*****************************************************************

package SSTorytime

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"

	// go get github.com/yuin/goldmark
	
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	_ "github.com/lib/pq"
)

//*****************************************************************
// Markdown reader
//*****************************************************************

type MDCell struct {
	Header string
	Value  string
}

type MDRow struct {
	Cells []MDCell
}

type MDTable struct {
	Headers []string
	Rows    []MDRow
	Context string
}

var DOC_DIRECTORY = make(map[string]string)

//*****************************************************************

func FractionateMarkdown(filename string) ([][]Sentence, int) {
	
	source,err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("Couldn't find or open",filename)
		os.Exit(-1)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
	)
	
	doc := md.Parser().Parse(text.NewReader(source))

	var sections []ExtSentence
	var current *ExtSentence
	var tables []MDTable
	var title string = ""
	var level int

	// collect together into current to assemble in an ExtSentence
	
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {

		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := n.(type) {

		case *extast.Table:
			tableData := ParseTable(n, source)
			tableData.Context = title
			tables = append(tables, tableData)

		case *ast.Heading:
			title = string(n.Text(source))
			level = n.Level

		case *ast.Paragraph:
			isnew, value, number := IsNewSection(string(n.Text(source)))

			if isnew {
				current = &ExtSentence{
					Level: level,
					Title: title,
					Number: "unnumbered",
					Content: value,
				}
			} else {
				current = &ExtSentence{
					Level: level,
					Title: title,
					Number: number,
					Content: value,
				}
			}

			sections = append(sections, *current)
		}

		return ast.WalkContinue, nil
	})

	// Now assemble normal text into pbsf format

	pbsf,count := AssemblePBSF(sections)

	for i, table := range tables {
		fmt.Printf("--- Table %d in context %s ---\n", i+1,table.Context)
		fmt.Printf("Headers (Ordered): %v\n\n", table.Headers)
		
		for j, row := range table.Rows {
			fmt.Printf("Row %d:\n", j+1)
			for _, cell := range row.Cells {
				fmt.Printf("  %-12s -> %s\n", cell.Header+":", cell.Value)
			}
			
			// Example: Lookup by header key still works
			if price, ok := row.GetByName("Price"); ok {
				fmt.Printf("  [Lookup Price]: %s\n", price)
			}
			fmt.Println()
		}
	}

	// Now split the parts

	for _,sec := range sections {
		fmt.Println("\nSEC",sec.Title,"at level",sec.Level,"with",sec.Content)
	}

	return pbsf,count
}


//*****************************************************************

func ParseTable(table *extast.Table, source []byte) MDTable {
	var headers []string
	var rows []MDRow

	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		switch section := child.(type) {
		case *extast.TableHeader:
			headers = ExtractValuesFromContainer(section, source)

		case *extast.TableRow:
			values := ExtractValuesFromContainer(section, source)
			if len(values) == 0 {
				continue
			}

			row := MDRow{
				Cells: make([]MDCell, len(values)),
			}

			for i, val := range values {
				headerName := fmt.Sprintf("col_%d", i+1)
				if i < len(headers) && headers[i] != "" {
					headerName = headers[i]
				}

				row.Cells[i] = MDCell{
					Header: headerName,
					Value:  val,
				}
			}

			rows = append(rows, row)
		}
	}

	return MDTable{
		Headers: headers,
		Rows:    rows,
	}
}

//*****************************************************************

func ExtractValuesFromContainer(container ast.Node, source []byte) []string {
	var values []string

	for child := container.FirstChild(); child != nil; child = child.NextSibling() {
		switch node := child.(type) {
		case *extast.TableCell:
			values = append(values, GetCellText(node, source))
		case *extast.TableRow:
			return ExtractValuesFromContainer(node, source)
		}
	}

	return values
}

//*****************************************************************

func GetCellText(n ast.Node, source []byte) string {
	var buf strings.Builder

	var walk func(ast.Node)
	walk = func(node ast.Node) {
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			if textNode, ok := c.(*ast.Text); ok {
				buf.Write(textNode.Segment.Value(source))
			}
			walk(c)
		}
	}

	walk(n)
	return strings.TrimSpace(buf.String())
}

//*****************************************************************

func (r MDRow) GetByName(header string) (string, bool) {
	for _, cell := range r.Cells {
		if strings.EqualFold(cell.Header, header) {
			return cell.Value, true
		}
	}
	return "", false
}


//
// markdown.go
//
