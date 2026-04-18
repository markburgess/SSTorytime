# N4L Syntax Reference

N4L (Notes for Learning) is a plain-text language for creating knowledge graphs in SSTorytime.

## Comments

```
# This is a comment
// This is also a comment
```

## Chapters

Chapters are top-level organizational containers. Start with a dash:

```
- my_chapter_name
```

## Context Tags

Context tags filter searches and disambiguate nodes. They persist until changed.

```
:: restaurant, food, dining ::          # Set context
+:: breakfast ::                        # Add to existing context
-:: dining ::                           # Remove from context
```

Any number of colons works: `: tag :` or `::: tag :::`.

## Nodes and Relationships

A node is any text that doesn't contain unquoted parentheses:

```
coffee                                  # Simple node
coffee (contains) caffeine              # Relationship: coffee CONTAINS caffeine
A (causes) B (leads to) C              # Chained relationships
```

Quote text with special characters:

```
"text with (parens)"
'text with "quotes"'
```

## Arrow Syntax

Arrows go in parentheses between nodes:

```
Node A (arrow_name) Node B
```

Use short codes from SSTconfig/ arrow definitions:

```
beef (eh) 牛肉                          # english has hanzi
```

## Ditto Marks

`"` at the start of a line refers to the first item of the previous line:

```
meat (means) flesh
"    (e.g.) beef                        # = meat (e.g.) beef
"    (e.g.) pork                        # = meat (e.g.) pork
```

## References and Aliases

```
@label Node text                        # Create a named alias
$label.1 (arrow) Other                  # Reference first item of aliased line

$1 (arrow) Something                    # Reference first item of previous line
$2 (arrow) Something                    # Reference first item of 2 lines ago
```

## Sequence Mode

Automatically chains consecutive items with "(then)" arrows:

```
+:: _sequence_ ::
Step one
Step two                                # Auto-creates: Step one (then) Step two
Step three                              # Auto-creates: Step two (then) Step three
-:: _sequence_ ::
```

## Annotations

Mark words within text bodies to create implicit relationships:

```
# In annotations.sst, define annotation markers:
% (discusses)                           # %word creates (discusses) link
= (depends on)                         # =word creates (depends on) link
* (is a special case of)               # *word creates (is a special case of) link
> (has subject)                         # >word creates (has subject) link
```

## Complete Example

```
- italian_cooking

:: food, cooking, italian ::

# Ingredients and their categories
pasta (contains) flour
"     (contains) eggs
"     (contains) water

# Types of pasta
+:: types ::
spaghetti (is an example of) pasta
penne     (is an example of) pasta
-:: types ::

# A simple recipe
+:: recipe, sequence ::
+:: _sequence_ ::
boil water
add salt to water
add pasta to pot
cook for 8 minutes
drain and serve
-:: _sequence_ ::
```
