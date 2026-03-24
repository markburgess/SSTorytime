
//******************************************************************
//
// CSS template expander, because no variables in CSS
//
//******************************************************************

package main

import (
	"fmt"
	"strings"
	"os"
	"bufio"
)

//******************************************************************

func main() {

var theme = make(map[string]map[string]string)

        theme["dark"] = make(map[string]string)

	theme["dark"]["PAGE_BACKGROUND_COLOUR"] = "var(--gray-12)"
	theme["dark"]["LINK_ITEM_BACKGROUND_COLOUR"] = "var(--gray-2)"
	theme["dark"]["LINK_SECONDARY_COLOUR"] = "var(--gray-5)"
	theme["dark"]["FULL_TEXT_COLOUR"] = "var(--choco-0)"
	theme["dark"]["CARD_BORDER_COLOUR"] = "var(--green-0)"
	theme["dark"]["TITLE_SUMMARY_COLOUR"] = "var(--choco-2)"
	theme["dark"]["TABLE_BACKGROUND_COLOUR"] = "var(--gray-12)"
	theme["dark"]["SPAN_COLOUR"] = "var(--choco-0)"
	theme["dark"]["HOVER_COLOUR"] = "var(--indigo-2)"
	theme["dark"]["ERROR_COLOUR"] = "var(--stone-10)"
	theme["dark"]["CURRENT_TIME_COLOUR"] = "var(--yellow-8)"
	theme["dark"]["DETAIL_BACKGROUND_COLOUR"] = "slategray"
	theme["dark"]["DETAIL_SUMARY_BACKGROUND_COLOUR"] = "darkslategray"
	theme["dark"]["DETAIL_FORE_COLOUR"] = "tomato"
	theme["dark"]["MAIN_CONTENT_PANEL_COLOUR"] = "var(--choco-2)"
	theme["dark"]["MATH_COLOUR"] = "orange"
	theme["dark"]["SST_ARROW_0_COLOUR"] = "#C9A227"
	theme["dark"]["SST_ARROW_1_COLOUR"] = "#2E6F95"
	theme["dark"]["SST_ARROW_2_COLOUR"] = "#C62828"
	theme["dark"]["SST_ARROW_3_COLOUR"] = "#4C8C4A" // NEAR
	theme["dark"]["SST_ARROW_4_COLOUR"] = "#C62828"
	theme["dark"]["SST_ARROW_5_COLOUR"] = "#2E6F95"
	theme["dark"]["SST_ARROW_6_COLOUR"] = "#C9A227"
	theme["dark"]["TOC_COLOUR"] = "var(--pink-4)"
	theme["dark"]["TOC_SINGLE_COLOUR"] = "var(--orange-4)"
	theme["dark"]["STATUS_INDICATOR_COLOUR"] = "var(--gray-5)"
	theme["dark"]["STATUS_OK_COLOUR"] = "var(--green-5)"
	theme["dark"]["STATUS_NOT_OK_COLOUR"] = "var(--red-6)"
	theme["dark"]["LINE_NUM_COLOUR"] = "var(--yellow-4)"

	theme["slate"] = make(map[string]string)
	theme["slate"]["PAGE_BACKGROUND_COLOUR"] = "var(--gray-8)"
	theme["slate"]["LINK_ITEM_BACKGROUND_COLOUR"] = "var(--gray-2)"
	theme["slate"]["LINK_SECONDARY_COLOUR"] = "var(--gray-5)"
	theme["slate"]["FULL_TEXT_COLOUR"] = "var(--green-0)"
	theme["slate"]["CARD_BORDER_COLOUR"] = "var(--stone-4)"
	theme["slate"]["TITLE_SUMMARY_COLOUR"] = "var(--choco-2)"
	theme["slate"]["TEXT_CONTENT_COLOUR"] = "var(--green-2)"
	theme["slate"]["TABLE_BACKGROUND_COLOUR"] = "var(--stone-12)"
	theme["slate"]["SPAN_COLOUR"] = "var(--choco-0)"
	theme["slate"]["HOVER_COLOUR"] = "var(--indigo-2)"
	theme["slate"]["ERROR_COLOUR"] = "var(--stone-10)"
	theme["slate"]["CURRENT_TIME_COLOUR"] = "var(--yellow-8)"
	theme["slate"]["DETAIL_BACKGROUND_COLOUR"] = "darkslategray"
	theme["slate"]["DETAIL_SUMARY_BACKGROUND_COLOUR"] = "lightslategray"
	theme["slate"]["MAIN_CONTENT_PANEL_COLOUR"] = "var(--choco-2)"
	theme["slate"]["MATH_COLOUR"] = "var(--green-2)"
	theme["slate"]["SST_ARROW_0_COLOUR"] = "#C9A227"
	theme["slate"]["SST_ARROW_1_COLOUR"] = "#2E6F95"
	theme["slate"]["SST_ARROW_2_COLOUR"] = "#C62828"
	theme["slate"]["SST_ARROW_3_COLOUR"] = "#4C8C4A" // NEAR
	theme["slate"]["SST_ARROW_4_COLOUR"] = "#C62828"
	theme["slate"]["SST_ARROW_5_COLOUR"] = "#2E6F95"
	theme["slate"]["SST_ARROW_6_COLOUR"] = "#C9A227"
	theme["slate"]["TOC_COLOUR"] = "var(--pink-4)"
	theme["slate"]["TOC_SINGLE_COLOUR"] = "var(--orange-4)"
	theme["slate"]["STATUS_INDICATOR_COLOUR"] = "var(--gray-5)"
	theme["slate"]["STATUS_OK_COLOUR"] = "var(--green-5)"
	theme["slate"]["STATUS_NOT_OK_COLOUR"] = "var(--red-6)"
	theme["slate"]["LINE_NUM_COLOUR"] = "var(--choco-0)"

	theme["spaceblue"] = make(map[string]string)
	theme["spaceblue"]["PAGE_BACKGROUND_COLOUR"] = "var(--blue-12)"
	theme["spaceblue"]["LINK_ITEM_BACKGROUND_COLOUR"] = "var(--blue-2)"
	theme["spaceblue"]["LINK_SECONDARY_COLOUR"] = "var(--blue-5)"
	theme["spaceblue"]["FULL_TEXT_COLOUR"] = "var(--cyan-0)"
	theme["spaceblue"]["CARD_BORDER_COLOUR"] = "var(--stone-4)"
	theme["spaceblue"]["TITLE_SUMMARY_COLOUR"] = "var(--cyan-2)"
	theme["spaceblue"]["TEXT_CONTENT_COLOUR"] = "var(--green-2)"
	theme["spaceblue"]["TABLE_BACKGROUND_COLOUR"] = "var(--blue-12)"
	theme["spaceblue"]["SPAN_COLOUR"] = "var(--cyan-0)"
	theme["spaceblue"]["HOVER_COLOUR"] = "var(--indigo-2)"
	theme["spaceblue"]["ERROR_COLOUR"] = "var(--stone-10)"
	theme["spaceblue"]["CURRENT_TIME_COLOUR"] = "var(--yellow-8)"
	theme["spaceblue"]["DETAIL_BACKGROUND_COLOUR"] = "navy"
	theme["spaceblue"]["DETAIL_SUMARY_BACKGROUND_COLOUR"] = "#6495ED"
	theme["spaceblue"]["MAIN_CONTENT_PANEL_COLOUR"] = "var(--cyan-2)"
	theme["spaceblue"]["MATH_COLOUR"] = "lightblue"
	theme["spaceblue"]["SST_ARROW_0_COLOUR"] = "#C9A227"
	theme["spaceblue"]["SST_ARROW_1_COLOUR"] = "#2E6F95"
	theme["spaceblue"]["SST_ARROW_2_COLOUR"] = "#C62828"
	theme["spaceblue"]["SST_ARROW_3_COLOUR"] = "#4C8C4A" // NEAR
	theme["spaceblue"]["SST_ARROW_4_COLOUR"] = "#C62828"
	theme["spaceblue"]["SST_ARROW_5_COLOUR"] = "#2E6F95"
	theme["spaceblue"]["SST_ARROW_6_COLOUR"] = "#C9A227"
	theme["spaceblue"]["TOC_COLOUR"] = "var(--pink-4)"
	theme["spaceblue"]["TOC_SINGLE_COLOUR"] = "var(--orange-4)"
	theme["spaceblue"]["STATUS_INDICATOR_COLOUR"] = "var(--gray-5)"
	theme["spaceblue"]["STATUS_OK_COLOUR"] = "var(--green-5)"
	theme["spaceblue"]["STATUS_NOT_OK_COLOUR"] = "var(--red-6)"
	theme["spaceblue"]["LINE_NUM_COLOUR"] = "var(--cyan-0)"

	theme["red"] = make(map[string]string)
	theme["red"]["PAGE_BACKGROUND_COLOUR"] = "var(--pink-12)"
	theme["red"]["LINK_ITEM_BACKGROUND_COLOUR"] = "var(--orange-2)"
	theme["red"]["LINK_SECONDARY_COLOUR"] = "var(--orange-5)"
	theme["red"]["FULL_TEXT_COLOUR"] = "var(--yellow-0)"
	theme["red"]["CARD_BORDER_COLOUR"] = "var(--stone-4)"
	theme["red"]["TITLE_SUMMARY_COLOUR"] = "var(--red-5)"
	theme["red"]["TEXT_CONTENT_COLOUR"] = "var(--yellow-7)"
	theme["red"]["TABLE_BACKGROUND_COLOUR"] = "var(--red-12)"
	theme["red"]["SPAN_COLOUR"] = "var(--yellow-0)"
	theme["red"]["HOVER_COLOUR"] = "var(--indigo-2)"
	theme["red"]["ERROR_COLOUR"] = "var(--stone-10)"
	theme["red"]["CURRENT_TIME_COLOUR"] = "var(--yellow-8)"
	theme["red"]["DETAIL_BACKGROUND_COLOUR"] = "brown"
	theme["red"]["DETAIL_SUMARY_BACKGROUND_COLOUR"] = "#B87333"
	theme["red"]["MAIN_CONTENT_PANEL_COLOUR"] = "var(--red-5)"
	theme["red"]["MATH_COLOUR"] = "moccasin"
	theme["red"]["SST_ARROW_0_COLOUR"] = "#C9A227"
	theme["red"]["SST_ARROW_1_COLOUR"] = "#2E6F95"
	theme["red"]["SST_ARROW_2_COLOUR"] = "#C62828"
	theme["red"]["SST_ARROW_3_COLOUR"] = "#4C8C4A" // NEAR
	theme["red"]["SST_ARROW_4_COLOUR"] = "#C62828"
	theme["red"]["SST_ARROW_5_COLOUR"] = "#2E6F95"
	theme["red"]["SST_ARROW_6_COLOUR"] = "#C9A227"
	theme["red"]["TOC_COLOUR"] = "var(--pink-1)"
	theme["red"]["TOC_SINGLE_COLOUR"] = "var(--yellow-5)"
	theme["red"]["STATUS_INDICATOR_COLOUR"] = "var(--gray-5)"
	theme["red"]["STATUS_OK_COLOUR"] = "var(--green-5)"
	theme["red"]["STATUS_NOT_OK_COLOUR"] = "var(--red-6)"
	theme["red"]["LINE_NUM_COLOUR"] = "var(--yellow-4)"

	theme["style"] = make(map[string]string)
	theme["style"]["PAGE_BACKGROUND_COLOUR"] = "var(--brown-0)" // "var(--sand-0)"
	theme["style"]["LINK_ITEM_BACKGROUND_COLOUR"] = "var(--blue-8)"
	theme["style"]["LINK_SECONDARY_COLOUR"] = "var(--blue-11)"
	theme["style"]["FULL_TEXT_COLOUR"] = "var(--teal-12)"
	theme["style"]["CARD_BORDER_COLOUR"] = "var(--stone-4)"
	theme["style"]["TITLE_SUMMARY_COLOUR"] = "var(--blue-10)"
	theme["style"]["TEXT_CONTENT_COLOUR"] = "var(--gray-9)"
	theme["style"]["TABLE_BACKGROUND_COLOUR"] = "var(--stone-2)"
	theme["style"]["SPAN_COLOUR"] = "var(--gray-11)"
	theme["style"]["HOVER_COLOUR"] = "var(--indigo-8)"
	theme["style"]["ERROR_COLOUR"] = "var(--stone-10)"
	theme["style"]["CURRENT_TIME_COLOUR"] = "var(--yellow-8)"
	theme["style"]["DETAIL_BACKGROUND_COLOUR"] = "#eeeeee"
	theme["style"]["DETAIL_SUMARY_BACKGROUND_COLOUR"] = "var(--sand-1)"
	theme["style"]["MAIN_CONTENT_PANEL_COLOUR"] = "var(--stone-12)"
	theme["style"]["MATH_COLOUR"] = "var(--gray-9)"
	theme["style"]["SST_ARROW_0_COLOUR"] = "#C9A227"
	theme["style"]["SST_ARROW_1_COLOUR"] = "#2E6F95"
	theme["style"]["SST_ARROW_2_COLOUR"] = "#C62828"
	theme["style"]["SST_ARROW_3_COLOUR"] = "#4C8C4A" // NEAR
	theme["style"]["SST_ARROW_4_COLOUR"] = "#C62828"
	theme["style"]["SST_ARROW_5_COLOUR"] = "#2E6F95"
	theme["style"]["SST_ARROW_6_COLOUR"] = "#C9A227"
	theme["style"]["TOC_COLOUR"] = "purple"
	theme["style"]["TOC_SINGLE_COLOUR"] = "var(--orange-10)"
	theme["style"]["STATUS_INDICATOR_COLOUR"] = "var(--gray-5)"
	theme["style"]["STATUS_OK_COLOUR"] = "var(--green-5)"
	theme["style"]["STATUS_NOT_OK_COLOUR"] = "var(--red-6)"
	theme["style"]["LINE_NUM_COLOUR"] = "var(--stone-7)"


	GenerateFile("dark",theme)
	GenerateFile("slate",theme)
	GenerateFile("spaceblue",theme)
	GenerateFile("red",theme)
	GenerateFile("style",theme)


}

//******************************************************************

func GenerateFile(title string,theme map[string]map[string]string) {

	// INPUT

	filename := "public/style.css.in"

	file, err := os.Open(filename)

	if err != nil {
		fmt.Println("error",err)
		os.Exit(-1)
	}

	defer file.Close()

	// OUTPUT

	outputfile := fmt.Sprintf("public/%s.css",title)

	fp, err := os.Create(outputfile)

	if err != nil {
		fmt.Println("Failed to open file for writing: ",outputfile)
		os.Exit(-1)
	}

	defer fp.Close()

	//fmt.Fprintln(fp," // DON'T WRITE TO THIS FILE, master is at",filename)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		for css,colour := range theme[title] {
			if strings.Contains(line,css) {
				line = strings.Replace(line,css,colour,1)
			}
		}
		// Process or print the line
		fmt.Fprintln(fp,line)
	}
}





























