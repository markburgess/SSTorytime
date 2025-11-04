
# Search examples

## Style

* Select a visual theme (web only)

<pre>
 \theme light
 \theme dark
 \theme slate
 \theme spaceblue
</pre>


## Word search

* Search for a precise word
<pre>
   !a1!
   "ephemeral or persistent"
</pre>

* Search for any combination of a set of words
<pre>
  word1 word2 word3
  recipe fish soup
</pre>

* Search for any combination of a set of words in a chapter
<pre>
  word1 word2 word3 \chapter dictionary
  recipe fish soup  \chapter "my recipes"
</pre>

* Search for any combination of a set of words in named context, any chapter
<pre>
  word1 word2 word3 \context "weird words"
  recipe fish soup  \chapter food
</pre>

* General word search

<pre>
  word1 word2 word3 \chapter "my chapter" yourchapter \context "weird words"
  recipe fish soup  \chapter food \context food, recipes, dishes
</pre>

## Table of contents
<pre>
\toc
\map
\chapters
\chapter mychapter
</pre>

## Notes

* Print original notes from a chapter
<pre>
\notes mychapter

</pre>

## Stories and sequences
<pre>
\story (wuya)
\story Mary
\sequence "create a pod"
\seq any \in \chapter kubernetes
\story any \chapter moon
</pre>

## Path solutions

<pre>
\paths \from start \to target
\from !a1! \to b6
</pre>

## Look for an arrow

\arrow succeed
\arrow 1
\arrow 232