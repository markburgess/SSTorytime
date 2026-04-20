
# text2N4L

![Six tools arranged on a worktable — sieve, builder's square, magnifying glass, open book, twin flashlights, microscope — representing the family of CLI utilities in the SSTorytime toolkit.](figs/tools_hero.jpg){ align=center }

Sometimes you want to make notes on a text that's already written in natural language,
and it might be quite long. Reworking the text in note form would take a long time and might
be difficult.

The `text2N4L` command reads a plain text `filename.txt` like the examples in `examples/example_data`
and turns it into a prototype N4Lfile automatically, based on a model of deconstructing narrative
language (a Tiny Language Model). Nothing is uploaded into the database. You can use `N4L` to do that
later. This give you the opportunity to edit and rework, add to and delete from the proposal.

**Note**: while this sounds like a nice idea, it can be quite expensive in terms of memory. Scanning
even a fraction of a book can produce a lot of text and cross referencing, so unicode encoding time combined
with the the upload time to the database diverges quite quickly. A book, like *Moby Dick* or Darwin's *Origin of Species*
will likely take several hours to upload.

By default, the tool selects only a 50% fraction of the sentences that have been measired for their
significance or their level of `intent'. 
<pre>
$ text2N4L ../examples/example_data/promisetheory1.dat 

Wrote file ../examples/example_data/promisetheory1.dat_edit_me.n4l
Final fraction 62.18 of requested 50.00 sampled

</pre>
You can change the fraction sampled
<pre>
$ text2N4L -% 77 ../examples/example_data/MobyDick.dat 
</pre>
Because there is uncertainty in how to select the relevant parts,
`text2N4L` will oversample, especially for low percentages. As you reach
100%, there is no ambiguity.

The generated file takes sentences from the source document and prefixes them with labels:
<pre>
@sen9471   Towards thee I roll, thou all-destroying but unconquering whale, to the last I grapple with thee, from hell’s heart I stab at thee, f
or hate’s sake I spit my last breath at thee.
              " (is in) part 210 of ../examples/example_data/MobyDick.dat

@sen9473   and since neither can be mine, let me then tow to pieces, while still chasing thee, though tied to thee, thou damned whale!
              " (is in) part 210 of ../examples/example_data/MobyDick.dat

@sen9475   The harpoon was darted, the stricken whale flew forward, with igniting velocity the line ran through the grooves, ran foul.
              " (is in) part 210 of ../examples/example_data/MobyDick.dat

@sen9476   Ahab stooped to clear it, he did clear it, but the flying turn caught him round the neck, and voicelessly as Turkish mutes bowstring 
their victim, he was shot out of the boat, ere the crew knew he was gone.
              " (is in) part 210 of ../examples/example_data/MobyDick.dat

</pre>
You can add you own notes, say at the end of the file:

<pre>

$sen9471.1  (note) This line was immortalized in the movie Star Trek: Wrath of Khan by Khan himself.

</pre>

## The `-%` flag in detail

`text2N4L` has exactly one flag:

```
text2N4L [-%  percent] filename
```

Declaration: [`src/text2N4L/text2N4L.go:41`](https://github.com/markburgess/SSTorytime/blob/main/src/text2N4L/text2N4L.go#L41).

### Percentage semantics

The number you pass is the **approximate target fraction** of source sentences to keep in the generated `.n4l` file. The selection runs in two passes (`SelectByRunningIntent` and `SelectByStaticIntent`) and merges the results, so the actual output is usually **slightly higher** than requested:

| You ask for | You typically get |
|---|---|
| `-% 10` | 15–25 % (oversampling is worst at low percentages) |
| `-% 30` | 40–50 % |
| `-% 50` (default) | 60–65 % |
| `-% 77` | 85–90 % |
| `-% 100` | 100 % — **no selection ambiguity**; every sentence is kept |

The skew toward oversampling is intentional: two heuristic rankings (intent-based and statistical) can each nominate a sentence, and the merged set is the union. At 100 % the ambiguity collapses and you get a faithful 1:1 conversion.

### What happens at each boundary

- `-% 0` or omitted — defaults to `50`. See the flag default.
- `-% 100` — deterministic: every sentence is converted to a `@senN ... (is in) partN` line.
- Negative or non-numeric values — Go's `flag` package rejects with a usage error; exit code `2`.

### File-size considerations

`text2N4L` loads the whole source file into memory, builds an `N-gram` frequency table (1- to 4-grams), and emits one `.n4l` sentence per selected input sentence **plus a context-dictionary** of intentional phrases.

Rough scaling, measured on a 2024-era laptop:

| Source size | RAM | Time | Output size |
|---|---|---|---|
| 100 KB (~1,000 sentences) | ~100 MB | seconds | ~200 KB at 50 % |
| 1 MB (~10,000 sentences) | ~500 MB | minutes | ~2 MB at 50 % |
| 10 MB (e.g. *Moby-Dick*) | multi-GB | hours | ~15–20 MB at 50 % |

!!! warning "Large documents are slow to upload, not just to fractionate"
    The hard cost isn't `text2N4L` — it's the downstream `N4L -u` step. Once you have a
    large `.n4l` file, every cross-reference becomes a Postgres round-trip. Budget hours
    for book-sized corpora. Consider chunking the source text into chapters and running
    `text2N4L` per chapter.

### Exit codes & environment

- **Exit `0`** — success (output file written).
- **Exit `-1`** — file system error creating the output `.n4l` (see [`src/text2N4L/text2N4L.go:107-110`](https://github.com/markburgess/SSTorytime/blob/main/src/text2N4L/text2N4L.go#L107-L110)).
- **Exit `-2`** — missing filename argument (see [`src/text2N4L/text2N4L.go:48-51`](https://github.com/markburgess/SSTorytime/blob/main/src/text2N4L/text2N4L.go#L48-L51)).
- **Exit `2`** — invalid flag (Go `flag` package default).

`text2N4L` does **not** connect to PostgreSQL — it produces a file and stops — so `POSTGRESQL_URI` is not consulted. `SST_CONFIG_PATH` is also unused.