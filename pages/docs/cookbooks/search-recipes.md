# Patterns: search recipes

> **Ten query shapes you'll reach for more than once. Copy, paste, adapt.**

The [Finding things](../searchN4L.md) page teaches the shape of a
question. This page is the cheat sheet — ten recipes that cover most
of what you'll want to do, each with a command, an expected-shape
output, and a one-line note on when to reach for it. The running corpus
is the reading list from [Your first story](../Tutorial.md); a couple of
recipes use `branches.n4l` or the chinese-notes corpus where the shape
of the example requires it.

!!! info "Shell escaping"
    `\` is an escape character in bash/zsh. Commands below use double
    backslashes (`\\notes`, `\\from`) or quote the whole command
    (`"\\notes brain"`). Both work.

## 1. Orbit around a topic

**When you want it:** you remember a topic and want to see everything
that points at it.

```
searchN4L "decision making"
```

```
0: "decision making"	in chapter: reading list

      -    (is the topic/theme of) - Thinking Fast and Slow
           -    (is a bibtex citation label for) - Superforecasting
      -    (is the topic/theme of) - Superforecasting
      -    (is the topic/theme of) - Thinking in Systems
```

The topic is the match; every book that pointed `(about)` at it is
listed below. Indented lines are one hop further out — the graph walks
a step beyond the immediate neighbours. Cap the walk with `\\limit N`.

## 2. Full orbit of one thing

**When you want it:** you remember a title and want everything you
ever wrote about it.

```
searchN4L notes about "Thinking Fast and Slow"
```

```
0: "Thinking Fast and Slow"	in chapter: reading list

      -    (is about topic/them) - decision making
      -    (is about topic/them) - heuristics
      -    (has author) - Daniel Kahneman
      -    (has bibtex citation) - Judgment under Uncertainty
           -    (has author) - Amos Tversky
           -    (note/remark) - the anchor is almost always the wrong number
      -    (note/remark) - two systems, one of them lazy, both of them you
      -    (is a bibtex citation label for) - Superforecasting
```

Every arrow in or out of the node, plus one hop further. When the node
has a lot of connections, this is how you see them all at once.

## 3. Author's footprint

**When you want it:** you remember a person and want to see what they
connect to across your notes.

```
searchN4L "Daniel Kahneman"
```

```
0: "Daniel Kahneman"	in chapter: reading list

      -    (is the author of) - Thinking Fast and Slow
           -    (is a bibtex citation label for) - Superforecasting
      -    (is the author of) - Judgment under Uncertainty
```

Two books by Kahneman you wrote down directly, plus the citation from
*Superforecasting* one hop further — a relationship between two books
that you never stated in a single note.

## 4. Path between two things

**When you want it:** you want the chain, not the neighbourhood. This
needs `(then)` / `(leads to)`-style arrows in your corpus — see
[Finding paths](../pathsolve.md) for why.

```
pathsolve -begin "once upon" -end "a little prince"
```

against `examples/branches.n4l`:

```
     - story path: 1 * once upon  -(then)->  a time  -(then)->  there was
      -(=>)->  a princess  -(=>)->  mischief!  -(=>)->  a little prince
```

The specific route through the branching story. On a reading-list-style
corpus (`(about)`, `(by)`, `(bib-cite)`), this will return "no paths
available" because those aren't the kind of arrows pathsolve follows.

## 5. Scope to one chapter

**When you want it:** the same word appears in several chapters and you
only want one.

```
searchN4L "bias" \\chapter "reading list"
```

The chapter filter narrows both the match and the walk. A reading-list
*bias* won't drag in a statistics-chapter *bias*.

## 6. Scope to a context

**When you want it:** the same word means different things in different
contexts, and you tagged your notes.

```
searchN4L "%%" \\context smalltalk brain wave \\limit 3
```

The `%%` is a wildcard name — match anything whose context includes
*smalltalk*, *brain*, or *wave*. Context is a set, not a conjunction:
any overlap counts. See [Context](../howdoescontextwork.md) for why.

## 7. Exact match for a short term

**When you want it:** your search term is a substring of too many other
nodes.

```
searchN4L "!A!"
```

`!A!` (or `|A|`) forces a whole-node match. Without the bangs, `A` is a
substring of thousands of strings in a big corpus. With them, only the
literal node named `A` comes back.

## 8. Browse a chapter in order

**When you want it:** you want to read your notes in the order you
wrote them, not in the order the graph returns them.

```
searchN4L "\\notes brain"
searchN4L "\\notes brain \\page 2"
```

This reflows the chapter from its original page order. Useful when the
graph view scrambles the narrative and you want the authored sequence
back.

## 9. Sequence browsing

**When you want it:** a chapter that was written as a sequence —
tagged `:: _sequence_ ::` — and you want the items in story order.

```
searchN4L "\\seq \"Mary had\""
```

```
  0. Mary had a little lamb
  1. Whose fleece was dull and grey
  2. And when it reached a certain age
  3. She'd serve it on a tray
```

`\seq`, `\sequence`, `\story`, and `\stories` are interchangeable
spellings of the same command.

## 10. Search without accents

**When you want it:** your keyboard can't produce the accented
characters in the source, but you know the word.

```
searchN4L "(fangzi)" \\chapter chinese
```

```
0: fángzi
      -    (pinyin has hanzi) - 房子
           -    (hanzi has english) - house  .. at home, domestic
```

Parenthesised search terms hit the unaccented copy the graph keeps for
exactly this purpose. `(fangzi)` finds `fángzǐ`, `fángzi`, and
variations. Without the parentheses, `fangzi` matches the accented
column and returns nothing.

## Bonus: chapter statistics

**When you want it:** a quick summary of a chapter's shape — how many
nodes, how many arrows, the dominant arrow types.

```
searchN4L "\\stats \\in brain"
```

For a deeper analysis (loops, sources, sinks, supernodes, centrality),
`graph_report` is the dedicated tool; see the developer docs for its
flags.

---

## Where to go next

- [Finding things](../searchN4L.md) — the shape of a question, in
  longer form.
- [Finding paths](../pathsolve.md) — when you want the chain, not
  the neighbourhood.
- [Context](../howdoescontextwork.md) — asking the same question
  different ways.
- The full query DSL lives in the repo at
  [`pages/docs/developers/searchN4L-flags.md`](https://github.com/markburgess/SSTorytime/blob/main/pages/docs/developers/searchN4L-flags.md),
  or try `searchN4L --help`.
