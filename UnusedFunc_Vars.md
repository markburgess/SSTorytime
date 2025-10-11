# SSTorytime.go Comprehensive Unused Code Analysis

**Generated:** October 11, 2025  
**File:** `pkg/SSTorytime/SSTorytime.go`  
**Total Lines:** 9,309

## Summary

| Category  | Total   | Used    | Unused | Unused %  |
| --------- | ------- | ------- | ------ | --------- |
| Functions | 224     | 210     | 14     | 6.25%     |
| Constants | 9       | 9       | 0      | 0%        |
| Variables | 8       | 7       | 1      | 12.5%     |
| **TOTAL** | **241** | **226** | **15** | **6.22%** |

## Unused Functions (14 items)

All unused functions have been marked with `// UNUSED:` comments in the source code.

### 1. GetNodeContext (Line 838)

```go
func GetNodeContext(ctx PoSST, node Node) []string
```

**Purpose:** Retrieves node context as a string slice  
**Category:** Context Management  
**Replacement:** Use `GetNodeContextString()` instead

### 2. ScoreContext (Line 4027)

```go
func ScoreContext(i, j int) bool
```

**Purpose:** Intended to score context relevance (placeholder - always returns true)  
**Category:** Scoring/Ranking  
**Status:** Incomplete implementation

### 3. GetSparseOccupancy (Line 5300)

```go
func GetSparseOccupancy(m [][]float32, dim int) []int
```

**Purpose:** Calculates sparse matrix occupancy counts  
**Category:** Matrix Operations  
**Status:** Mathematical utility not currently needed

### 4. TransposeMatrix (Line 5340)

```go
func TransposeMatrix(m [][]float32) [][]float32
```

**Purpose:** Transposes a 2D float32 matrix  
**Category:** Matrix Operations  
**Status:** Standard matrix operation, might be useful to keep

### 5. NextLinkArrow (Line 5550)

```go
func NextLinkArrow(ctx PoSST, path []Link, arrows []ArrowPtr) string
```

**Purpose:** Finds the next arrow in a link path matching given arrow types  
**Category:** Path Analysis  
**Status:** Superseded by other path analysis methods

### 6. ContextInterferometry (Line 6048)

```go
func ContextInterferometry(now_ctx string)
```

**Purpose:** Unknown - marked as "deleted" in comments  
**Category:** Context Processing  
**Status:** Explicitly removed, empty implementation

### 7. GetUnixTimeKey (Line 7449)

```go
func GetUnixTimeKey(now int64) string
```

**Purpose:** Generates a database-suitable key from Unix timestamp  
**Category:** Time Utilities  
**Status:** Superseded by other time functions

### 8. NewNgramMap (Line 7536)

```go
func NewNgramMap() [N_GRAM_MAX]map[string]float64
```

**Purpose:** Creates a new n-gram frequency map array  
**Category:** Text Processing  
**Status:** N-gram functionality handled differently

### 9. SplitCommandText (Line 7661)

```go
func SplitCommandText(s string) []string
```

**Purpose:** Splits command text using punctuation rules  
**Category:** Text Processing  
**Status:** Specialized for unused command parsing

### 10. Array2Str (Line 8480)

```go
func Array2Str(arr []string) string
```

**Purpose:** Converts string array to comma-separated string  
**Category:** String Utilities  
**Status:** Superseded by standard library functions

### 11. Str2Array (Line 8495)

```go
func Str2Array(s string) ([]string, int)
```

**Purpose:** Converts string to string array  
**Category:** String Utilities  
**Status:** Superseded by standard library functions

### 12. RunErr (Line 9009)

```go
func RunErr(message string)
```

**Purpose:** Prints colored error messages  
**Category:** Error Handling  
**Status:** Simple utility easily replaced

### 13. ContextString (Line 9038)

```go
func ContextString(context []string) string
```

**Purpose:** Concatenates context strings with spaces  
**Category:** String Utilities  
**Status:** Simple utility superseded by standard library

### 14. Already (Line 9174)

```go
func Already(s string, cone map[int][]string) bool
```

**Purpose:** Checks if a string already exists in a cone data structure  
**Category:** Data Structure Utilities  
**Status:** Simple search function easily replaced

## Unused Variables (1 item)

### 1. GR_DAY_TEXT (Line 7346)

```go
var GR_DAY_TEXT = []string{
    "Monday", "Tuesday", "Wednesday", "Thursday",
    "Friday", "Saturday", "Sunday",
}
```

**Purpose:** Array of day names for date/time processing  
**Category:** Time Constants  
**Status:** Related functionality uses other day representations  
**Note:** Marked with `// UNUSED:` comment

## Unused Constants (0 items)

All constants in the file are currently being used.

## Categorized Analysis

### By Function Category:

- **Matrix Operations:** 2 functions (GetSparseOccupancy, TransposeMatrix)
- **String Utilities:** 4 functions (Array2Str, Str2Array, ContextString, SplitCommandText)
- **Context Management:** 2 functions (GetNodeContext, ContextInterferometry)
- **Path/Graph Analysis:** 2 functions (NextLinkArrow, Already)
- **Time Utilities:** 1 function (GetUnixTimeKey)
- **Text Processing:** 1 function (NewNgramMap)
- **Scoring/Ranking:** 1 function (ScoreContext)
- **Error Handling:** 1 function (RunErr)

### By Removal Priority:

#### High Priority (Safe to Remove):

- `ContextInterferometry` - Already marked as deleted
- `ScoreContext` - Incomplete placeholder
- `Array2Str` / `Str2Array` - Superseded by stdlib
- `RunErr` - Simple utility
- `Already` - Simple search utility
- `ContextString` - Simple string concatenation
- `GR_DAY_TEXT` - Unused time constant

#### Medium Priority (Consider Keeping):

- `TransposeMatrix` - Standard matrix operation
- `GetSparseOccupancy` - Useful matrix analysis
- `NewNgramMap` - Might be needed for text processing

#### Low Priority (Possibly Useful):

- `GetNodeContext` - Alternative to existing function
- `NextLinkArrow` - Specialized path analysis
- `GetUnixTimeKey` - Time utility
- `SplitCommandText` - Specialized text processing

## File Statistics

- **Total Lines:** 9,309
- **Function Density:** 2.4% of lines contain function definitions
- **Code Quality:** 93.78% of defined functions are actively used
- **Maintenance:** Low unused code ratio indicates good code hygiene

## Analysis Methodology

1. **Function Extraction:** Used regex to find all public function definitions
2. **Usage Search:** Searched across 65 Go files for function calls
3. **Pattern Matching:** Looked for `SST.FunctionName`, `SSTorytime.FunctionName`, and direct calls
4. **Internal Usage:** Checked for usage within SSTorytime.go itself
5. **Manual Verification:** Cross-referenced results for accuracy

## Recommendations

### Immediate Actions:

1. **Remove High Priority Items:** Safe to delete without impact
2. **Document Medium Priority Items:** Add better documentation if keeping
3. **Archive Functionality:** Consider moving matrix operations to separate utility package

### Code Maintenance:

1. **Regular Audits:** Perform similar analysis quarterly
2. **Documentation:** Better document functions intended for future use
3. **Testing:** Add tests for functions intended to be kept

### Future Considerations:

- Some functions may be called via reflection (not detected)
- Functions may be intended for planned features
- Consider the maintenance cost vs. potential future value

## Notes

- Analysis performed using automated scripts with manual verification
- Some dynamic or reflection-based usage may not be detected
- Functions marked as unused should be reviewed before removal
- Constants and variables show better usage discipline than functions

