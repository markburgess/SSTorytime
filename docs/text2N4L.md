
# text2N4L

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