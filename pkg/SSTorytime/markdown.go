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

	var current *ExtSentence
	var title string = "introductory qprologue"
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
			TABLES = append(TABLES, tableData)

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

			SECTIONS = append(SECTIONS, *current)
		}

		return ast.WalkContinue, nil
	})

	// Now assemble normal text into pbsf format

	pbsf,count := AssemblePBSF(SECTIONS)

	return pbsf,count
}

//*****************************************************************
// Build ancillary structures
//*****************************************************************

func CompileTOC(fp *os.File, filealias string, sections []ExtSentence) {

	// map to n4l

	fmt.Fprintf(fp,"\n # BEG table of contents ############################\n")
	fmt.Fprintf(fp,"\n :: _sequence_, table of contents :: \n")
	
	for _,sec := range sections {
		fmt.Fprintf(fp,"\n%s (%s) Document part in %s\n",sec.Title,EXPR_TAB_HEADER_L,filealias)
		fmt.Fprintf(fp,"\n%s (%s) Document part in %s\n",sec.Title,EXPR_TABNAME_S,sec.Content)
	}

	fmt.Fprintf(fp,"\n -:: _sequence_ :: \n")
	fmt.Fprintf(fp,"\n # END table of contents ############################\n")
}

//*****************************************************************

func CompileTabular(fp *os.File, filealias string, tables []MDTable) {

	// Map to n4l

	for _, table := range tables {

		fmt.Fprintf(fp," :: _sequence_, %s :: \n",table.Context)

		for _,h := range table.Headers {
			if len(h) > 0 {
				fmt.Fprintf(fp,"\n # table %s  ############## \n",table.Context)
				fmt.Fprintf(fp," %s  (%s) %s\n",h,INV_EXPR_TAB_HEADER_L,table.Context)	
			}
		}

		for j, row := range table.Rows {

			last := ""
			
			for _, cell := range row.Cells {
				fmt.Fprintf(fp," # table row\n")
				fmt.Fprintf(fp,"\n Associative row %d (%s) %s\n",j+1,EXPR_TABNAME_S,cell.Value)
				fmt.Fprintf(fp," %s (%s) %s \n", table.Context,CONT_TABLE_NAME_L,cell.Value)
				if !strings.HasPrefix(cell.Header,"col_") {
					fmt.Fprintf(fp," %s (%s) %s \n", cell.Header,INV_EXPR_TAB_HEADER_L, cell.Value)
				}
				if last != "" {
					fmt.Fprintf(fp," %s (%s) %s \n", last,EXPR_TABNAME_S, cell.Value)
					last = cell.Value
				}
			}
		}

		fmt.Fprintf(fp," -:: _sequence_ :: \n")
	}

}

//*****************************************************************
// Tools
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
					headerName = CleanText(headers[i])
				}

				row.Cells[i] = MDCell{
					Header: headerName,
					Value:  CleanText(val),
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
