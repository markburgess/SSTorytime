// **************************************************************************
//
// json_marshalling.go
//
// **************************************************************************

package SSTorytime

import (
	"fmt"
	_ "github.com/lib/pq"

)

//**************************************************************
// Json parsing
//**************************************************************

type PathValue struct {
	Path  string
	Value interface{}
}

//**************************************************************

func FindJsonStructureAsPaths(prefix string, value interface{}, list *[]PathValue) bool {

	var result bool

	switch v := value.(type) {

	case map[string]interface{}:

		for key, child := range v {
			newPath := key
			if prefix != "" {
				newPath = prefix + "." + key
			}
			result = FindJsonStructureAsPaths(newPath, child, list)
		}

	case []interface{}:

		for i, child := range v {
			newPath := fmt.Sprintf("%s.%d", prefix, i)
			result = FindJsonStructureAsPaths(newPath, child, list)
		}

	default: // Reached a leaf node (string, float64, bool, nil)

		if result, ok := v.(float64); ok {
			*list = append(*list, PathValue{
				Path:  prefix,
				Value: fmt.Sprintf("%d",int(result)),
			})

			return true
		}
		
		if result, ok := v.(string); ok {
			*list = append(*list, PathValue{
				Path:  prefix,
				Value: result,
			})

			return true
		}
	}

	return result
}

