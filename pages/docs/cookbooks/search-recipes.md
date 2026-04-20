# 10 useful queries

Ten ready-to-run `searchN4L` recipes, each with annotated expected output and a short note on when to reach for it. All examples assume you are in the repo root with `src/bin/searchN4L` compiled and data loaded (the built-in `examples/` corpora are used for several recipes).

!!! info "Shell escaping"
    `\` is an escape character in bash/zsh, so commands in this cookbook use double
    backslashes (`\\notes`, `\\from`) or quote the command (`"\\notes brain"`). Both work.

## 1. Orbit around a topic

**Use when**: you want to see everything linked to a concept, ranked by the graph's own sense of relevance.

```bash
./src/bin/searchN4L brain
```

Expected output: a list of nodes whose text contains `brain` as a substring, each followed by its immediate neighbours grouped by [arrow](../concepts/glossary.md#arrow-sttype) type.

```
0: neuroscience brain
      -    (has aspect) - waves
      -    (is discussed in) - smalltalk chapter
      -    (see also) - electroencephalogram
```

Indentation encodes distance. Use `\\limit N` to cap the result count.

## 2. Path between two nodes

**Use when**: you want to find *how* two concepts connect, not just that they do.

```bash
./src/bin/searchN4L "\\from start \\to \"target 1\""
```

Expected output:

```
     - story path:  start  -(leads to)->  door  -(leads to)->  passage  -(debug)->  target 1
     -  [ Link STTypes: -(+leads to)->  -(+leads to)->  -(+leads to)-> . ]
```

Default path depth is `5`. Use `\\depth N` for longer searches and `\\min N` to skip trivial loops.

## 3. Context-scoped search

**Use when**: the same word means different things in different chapters.

```bash
./src/bin/searchN4L "%%" "\\context smalltalk brain wave" "\\limit 3"
```

The `%%` is a wildcard name — match anything whose context includes `smalltalk`, `brain`, or `wave`. Context is a set, not a conjunction; any overlap counts.

```
0: what's up?   in chapter: notes on chinese
      -    (english has hanzi) - 什么事
      -    (hanzi has pinyin) - shénme shì     .. smalltalk, questions
```

## 4. Browse a chapter page by page

**Use when**: you want to read notes in the order you wrote them.

```bash
./src/bin/searchN4L "\\notes brain"
./src/bin/searchN4L "\\notes brain \\page 2"
```

Output is a reflowed rendering of the chapter's `PageMap` rows. See [notes.md](../notes.md#pagination-semantics) for pagination behaviour. The standalone `notes` CLI is equivalent:

```bash
./src/bin/notes -page 2 brain
```

## 5. Sequence browsing with `\seq`

**Use when**: chapters marked with `+:: _sequence_ ::` have narrative order you want to follow.

```bash
./src/bin/searchN4L "\\seq \"Mary had\""
```

```
  0. Mary had a little lamb
  1. Whose fleece was dull and grey
  2. And when it reached a certain age
  3. She'd serve it on a tray
```

The parser accepts `\seq`, `\sequence`, `\story`, or `\stories` interchangeably — all set the `param.Sequence` flag (see [service_search_cmd.go:407-409](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go#L407-L409)).

## 6. Exact match with `!term!`

**Use when**: a short search term is a substring of many other strings and you want the literal.

```bash
./src/bin/searchN4L "!A!"
```

In a graph where `A`, `AB`, `ABC`, and many other strings exist, `!A!` matches only the literal `A`. Equivalent: `|A|`. This is a substitute for direct `NodePtr` lookup when you do not know the node's address.

## 7. Chapter-scoped full-text

**Use when**: you know the answer is in a specific chapter and want to reduce noise.

```bash
./src/bin/searchN4L "bjorvika" "\\chapter oslo"
```

The `\\chapter <string>` scope filters both the substring match and the neighbour traversal to nodes whose `Chap` column contains `oslo`. Use `\\chapter any` for an explicit wildcard.

## 8. Direct NodePtr lookup

**Use when**: you have a node address from an earlier search and want to jump straight there.

```bash
./src/bin/searchN4L "(1,1)"
```

A [NodePtr](../concepts/glossary.md#nodeptr) is a tuple `(Class, CPtr)` where `Class` is the size bucket (1 for single-word ngrams, 2 for two-word, … up to 6 for `>1KB`) and `CPtr` is the index within that bucket. See [service_search_cmd.go](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/service_search_cmd.go) and `IsLiteralNptr`. Output is the node's orbit at radius 1.

## 9. Arrows introspection with `\arrow`

**Use when**: you want to see which relationship short codes are defined, or look up an arrow by its STtype.

```bash
./src/bin/searchN4L "\\arrow ph,pe"
./src/bin/searchN4L -- "\\arrow -2"
```

!!! note "Why `--` before `\arrow -2`"
    Go's `flag` package reads any token starting with `-` as a potential flag, so
    `./src/bin/searchN4L "\\arrow -2"` is rejected with "flag provided but not
    defined: -2". The `--` sentinel tells the flag parser "positional arguments
    only from here on" and the `-2` reaches the query DSL intact.

```
192. (3) ph -> pinyin has hanzi
190. (3) pe -> pinyin has english

  9. (-2) in -> is in
 11. (-2) is an emphatic proto-concept in -> is emph in
```

The columns are `ArrPtr`, `(STtype)`, `short -> long`. STtype codes are explained in [graph_report.md](../graph_report.md#st-type-codes).

## 10. Accented/unaccented search with `"(parenthesized)"`

**Use when**: your keyboard cannot produce the accented characters in the source corpus.

```bash
./src/bin/searchN4L "(fangzi)" "\\chapter chinese"
```

Parenthesized strings match the **unaccented** tsvector column (`Node.UnSearch`). So `(fangzi)` finds `fángzi`, `fángzǐ`, etc. Without parentheses, the search is against the accent-preserving column.

## Bonus: graph statistics per chapter

**Use when**: you want a one-shot summary of a chapter's shape.

```bash
./src/bin/searchN4L "\\stats \\in brain"
```

Or the dedicated CLI for richer analysis:

```bash
./src/bin/graph_report -chapter brain -sttype L,C -depth 6
```

The `graph_report` tool reports loops, sources, sinks, supernodes, and eigenvector centrality — more expensive than `\\stats` but also more informative. See [graph_report.md](../graph_report.md) for the full interpretation.

## Where to go next

- The full command reference lives at [searchN4L.md#the-query-dsl](../searchN4L.md#the-query-dsl).
- For path-specific work, `pathsolve` is the dedicated tool — see [pathsolve.md](../pathsolve.md).
- Every recipe above also works through the HTTPS API at `:8443/searchN4L` — see [WebAPI.md](../WebAPI.md).
