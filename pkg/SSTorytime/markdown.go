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
var SECTIONS []ExtSentence
var TABLES []MDTable

//*****************************************************************

func FractionateMarkdown(dir,filename string) ([][]Sentence, int) {
	
	source,err := ioutil.ReadFile(dir+filename)

	if err != nil {
		fmt.Println("Couldn't find or open",dir+filename)
		os.Exit(-1)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
	)
	
	doc := md.Parser().Parse(text.NewReader(source))

	var current *ExtSentence
	var title string = "introductory qprologue"
	var level int

	// strip extension
	filename,_ = MakeTitleFromFileName(filename)
	
	// collect together into current to assemble in an ExtSentence
	
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {

		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := n.(type) {

		case *extast.Table:
			tableData := ParseTable(n,source,filename)
			tableData.Context = title
			TABLES = append(TABLES, tableData)

		case *ast.Heading:
			_,value,_ := IsNewSection(string(n.Text(source)),filename)
			title = value
			level = n.Level

		case *ast.Paragraph:
			isnew, value, number := IsNewSection(string(n.Text(source)),filename)
			
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

			SECTIONS = append(SECTIONS, *current)
		}

		return ast.WalkContinue, nil
	})

	// Now assemble normal text into pbsf format

	pbsf,count := AssemblePBSF(SECTIONS)

	return pbsf,count
}

//*****************************************************************

func FractionateSmallMarkdown(source []byte,name string) ([][]Sentence, int) {

	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var current *ExtSentence
	var title string = "introductory qprologue"
	var level int
	
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {

		if !entering {
			return ast.WalkContinue, nil
		}
		
		switch node := n.(type) {
		case *ast.Heading:
			_,value,_ := IsNewSection(GetNodeText(node,source),name)
			title = value
			level = node.Level
			
		case *ast.Paragraph:
			
			isnew, value, number := IsNewSection(GetNodeText(node,source),name)
			
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
			
			SECTIONS = append(SECTIONS, *current)
		}
		
		return ast.WalkContinue, nil
	})
	
	
	// Now assemble normal text into pbsf format
	
	pbsf,count := AssemblePBSF(SECTIONS)
	
	return pbsf,count
}
	
//*****************************************************************
// Tools
//*****************************************************************

func ParseTable(table *extast.Table, source []byte,filename string) MDTable {
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
					headerName = CleanText(filename+" "+headers[i])
				}

				_,name := DeNumberItem(val)
				
				row.Cells[i] = MDCell{
					Header: headerName,
					Value:  CleanText(name),
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

//*****************************************************************

func GetNodeText(n ast.Node, source []byte) string {
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

//
// markdown.go
//
