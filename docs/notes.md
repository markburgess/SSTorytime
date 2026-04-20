
# The notes tool

The `notes` tool is the simplest way of retrieving what you wrote in your notes.N4L file. It just outputs what you entered in
roughly the same order as your original input, page by page. This is useful when reading things back as you wrote them.

Often we remember that we wrote something in a certain place, but we don't remember the details. This tool helps you to
see how you intentionally wrote the notes, but without comments and variables.

`notes` works page by page.

<pre>
$ src/notes fox and crow


Title: chinese story about fox and crow
Context: 

Wūyā Hé Húli (pinyin for hanzi) 乌鸦和狐狸 (hanzi for english) The Crow and the Fox 

Title: chinese story about fox and crow
Context: _sequence_ 

Húli zài shùlín lĭ zhăo chī de.  Tā lái dào yì kē dà shù xià, 
狐狸   在   树林   里  找   吃  的。  他  来  到  一 棵 大  树  下, (pinyin for english) The fox was in the woods looking for food. He came to a tree, 

...

</pre>
This take only a page number as an argument for controlling long note sets:
<pre>
$ src/notes -page 2 brain

</pre>

## Command-line flags

```
notes [-page N] <chapter or section>
```

Declaration: [`src/notes/notes.go:61`](https://github.com/markburgess/SSTorytime/blob/main/src/notes/notes.go#L61).

- `-page N` (default `1`) — which page of the chapter to show. Pages are 1-indexed.

The remaining positional arguments are joined with spaces to form the **chapter or section** to browse. For example `notes fox and crow` looks up the chapter matching the substring `"fox and crow"`.

## Pagination semantics

`notes` calls `GetDBPageMap(sst, chapter, context, page)` under the hood (see [`src/notes/notes.go:87`](https://github.com/markburgess/SSTorytime/blob/main/src/notes/notes.go#L87)). Pagination is driven by the `PageMap` table — the narrative ordering preserved at N4L-upload time.

- **Page size** is determined by the PageMap rows for the chapter. A long chapter that has, say, 47 rows will naturally split into several pages.
- **Requesting `-page 2` on a chapter with only 1 page** returns no output — there are simply no rows. The tool still exits `0`; it does not treat this as an error.
- **`-page 0` or negative** — the query returns empty. Prefer `-page 1` as the minimum.
- **Pages retain context grouping.** The tool reprints the `Title:` and `Context:` header every time the chapter or context changes within a page, so cross-context navigation stays readable.

## Invalid chapter behavior

If you pass a chapter substring that matches nothing:

- `notes` silently produces no output beyond a trailing newline.
- Exit code is `0`.

If you pass **no** chapter argument at all, the tool prints `Enter a chapter to browse` and exits `-1` (see [`src/notes/notes.go:70-73`](https://github.com/markburgess/SSTorytime/blob/main/src/notes/notes.go#L70-L73)).

!!! tip "Find valid chapter names first"
    If you do not remember what chapters are in the graph, use
    `searchN4L \\chapter any` or `searchN4L \\toc any` to list them, then feed a matching
    substring to `notes`. Chapter matching is substring-based, so `notes chin` will match
    `"notes on chinese"`.

## Web version

The web browser has an equivalent to the notes command line tool. Enter the relevant chapter into the chapter field and
press `browse`, then use the `next` and `previous` page buttons to move through the pages.

![Equivalent in web browser](figs/notes.png 'notes search')

## Exit codes & environment

- **Exit `0`** — success (including zero-result queries).
- **Exit `-1`** — no chapter argument provided; also any library/DB error.
- **Exit `2`** — invalid flag (Go `flag` package default).

Environment variables:

- `POSTGRESQL_URI` — overrides the DSN in [`pkg/SSTorytime/session.go:41`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L41).
- `SST_CONFIG_PATH` — not directly used by `notes`; arrows are read from the database.

If the database is unreachable, `notes` exits with a connection error before producing any page output.
