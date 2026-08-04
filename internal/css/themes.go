// Package css holds theme token maps and the CSS template used by go:generate
// (internal/libexec/cssgen). Generated files live in internal/server/public/
// and are embedded with the rest of the static assets.
package css

// ThemeNames are written as {name}.css under internal/server/public/.
var ThemeNames = []string{"style", "dark", "slate", "spaceblue", "red"}

// ThemeVars returns placeholder → value maps for each theme name.
func ThemeVars() map[string]map[string]string {
	// Port of the former css-builder theme maps.
	dark := map[string]string{
		"PAGE_BACKGROUND_COLOUR":          "var(--gray-12)",
		"LINK_ITEM_BACKGROUND_COLOUR":     "var(--gray-2)",
		"LINK_SECONDARY_COLOUR":           "var(--gray-5)",
		"FULL_TEXT_COLOUR":                "var(--choco-0)",
		"CARD_BORDER_COLOUR":              "var(--green-0)",
		"TITLE_SUMMARY_COLOUR":            "var(--choco-2)",
		"TABLE_BACKGROUND_COLOUR":         "var(--gray-12)",
		"SPAN_COLOUR":                     "var(--choco-0)",
		"HOVER_COLOUR":                    "var(--indigo-2)",
		"ERROR_COLOUR":                    "var(--stone-10)",
		"CURRENT_TIME_COLOUR":             "var(--yellow-8)",
		"DETAIL_BACKGROUND_COLOUR":        "slategray",
		"DETAIL_SUMARY_BACKGROUND_COLOUR": "darkslategray",
		"DETAIL_FORE_COLOUR":              "tomato",
		"MAIN_CONTENT_PANEL_COLOUR":       "var(--choco-2)",
		"MATH_COLOUR":                     "orange",
		"SST_ARROW_0_COLOUR":              "#C9A227",
		"SST_ARROW_1_COLOUR":              "#2E6F95",
		"SST_ARROW_2_COLOUR":              "#C62828",
		"SST_ARROW_3_COLOUR":              "#4C8C4A",
		"SST_ARROW_4_COLOUR":              "#C62828",
		"SST_ARROW_5_COLOUR":              "#2E6F95",
		"SST_ARROW_6_COLOUR":              "#C9A227",
		"TOC_COLOUR":                      "var(--pink-4)",
		"TOC_SINGLE_COLOUR":               "var(--orange-4)",
		"STATUS_INDICATOR_COLOUR":         "var(--gray-5)",
		"STATUS_OK_COLOUR":                "var(--green-5)",
		"STATUS_NOT_OK_COLOUR":            "var(--red-6)",
		"LINE_NUM_COLOUR":                 "var(--yellow-4)",
		"TEXT_CONTENT_COLOUR":             "var(--choco-0)",
	}

	style := map[string]string{
		"PAGE_BACKGROUND_COLOUR":          "var(--gray-1)",
		"LINK_ITEM_BACKGROUND_COLOUR":     "var(--gray-2)",
		"LINK_SECONDARY_COLOUR":           "var(--gray-5)",
		"FULL_TEXT_COLOUR":                "var(--gray-12)",
		"CARD_BORDER_COLOUR":              "var(--stone-4)",
		"TITLE_SUMMARY_COLOUR":            "var(--choco-8)",
		"TABLE_BACKGROUND_COLOUR":         "var(--gray-0)",
		"SPAN_COLOUR":                     "var(--choco-8)",
		"HOVER_COLOUR":                    "var(--indigo-6)",
		"ERROR_COLOUR":                    "var(--red-8)",
		"CURRENT_TIME_COLOUR":             "var(--yellow-8)",
		"DETAIL_BACKGROUND_COLOUR":        "var(--gray-2)",
		"DETAIL_SUMARY_BACKGROUND_COLOUR": "var(--gray-3)",
		"DETAIL_FORE_COLOUR":              "var(--red-8)",
		"MAIN_CONTENT_PANEL_COLOUR":       "var(--choco-8)",
		"MATH_COLOUR":                     "var(--green-8)",
		"SST_ARROW_0_COLOUR":              "#C9A227",
		"SST_ARROW_1_COLOUR":              "#2E6F95",
		"SST_ARROW_2_COLOUR":              "#C62828",
		"SST_ARROW_3_COLOUR":              "#4C8C4A",
		"SST_ARROW_4_COLOUR":              "#C62828",
		"SST_ARROW_5_COLOUR":              "#2E6F95",
		"SST_ARROW_6_COLOUR":              "#C9A227",
		"TOC_COLOUR":                      "var(--pink-8)",
		"TOC_SINGLE_COLOUR":               "var(--orange-8)",
		"STATUS_INDICATOR_COLOUR":         "var(--gray-6)",
		"STATUS_OK_COLOUR":                "var(--green-7)",
		"STATUS_NOT_OK_COLOUR":            "var(--red-7)",
		"LINE_NUM_COLOUR":                 "var(--yellow-8)",
		"TEXT_CONTENT_COLOUR":             "var(--gray-12)",
	}

	slate := clone(dark)
	slate["PAGE_BACKGROUND_COLOUR"] = "var(--gray-8)"
	slate["FULL_TEXT_COLOUR"] = "var(--green-0)"
	slate["CARD_BORDER_COLOUR"] = "var(--stone-4)"
	slate["TABLE_BACKGROUND_COLOUR"] = "var(--stone-12)"
	slate["DETAIL_BACKGROUND_COLOUR"] = "darkslategray"
	slate["DETAIL_SUMARY_BACKGROUND_COLOUR"] = "lightslategray"
	slate["MATH_COLOUR"] = "var(--green-2)"
	slate["TEXT_CONTENT_COLOUR"] = "var(--green-2)"

	spaceblue := clone(dark)
	spaceblue["PAGE_BACKGROUND_COLOUR"] = "#0b1220"
	spaceblue["LINK_ITEM_BACKGROUND_COLOUR"] = "#152238"
	spaceblue["FULL_TEXT_COLOUR"] = "#e8eef8"
	spaceblue["MAIN_CONTENT_PANEL_COLOUR"] = "#c5d4f0"

	red := clone(dark)
	red["PAGE_BACKGROUND_COLOUR"] = "#1a0a0a"
	red["CARD_BORDER_COLOUR"] = "var(--red-6)"
	red["HOVER_COLOUR"] = "var(--red-4)"

	return map[string]map[string]string{
		"style":     style,
		"dark":      dark,
		"slate":     slate,
		"spaceblue": spaceblue,
		"red":       red,
	}
}

func clone(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
