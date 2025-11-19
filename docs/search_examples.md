
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

The strings you type are normally treated as potential substrings to match within words.
If you want to insist a precise word match with nothing else included, i.e. the search term is
the entire node string,then you can usethe pling/bang/exclamation character !on both sides of the term,
or the vertical bar "pipe" symbol (which is not confused with the ! not operator):
<pre>
   !a1!
   |a1|
   "ephemeral or persistent"
</pre>

* Search with spaces in the string

If your search term contains spaces, exclose them in double quotes:
<pre>
  "fish soup"
</pre>
This treats "fish soup" as a single string,rather than as "fish" OR "soup".

* Search for any "OR" combination of a set of words

If your words are implicitly ORed together, then just separate by spaces.
<pre>
  word1 word2 word3
  recipe fish soup
</pre>
Conversely, words separated by spaces are ORed together.

* Search with logical expressions

You can use `& = AND`, `! = NOT`, `| = OR` in expressions, i.e. [postgres ts_vector search logic](https://www.postgresql.org/docs/current/textsearch-controls.html) in search terms, if you place them in quotes. This is very powerful. Notice that the !character is also used for hard-delimiting of strings. You might need to enclose your expression in quotes to keep it together.
<pre>
 a1&!b6
 "a1 & !b6"
 brain&!notes
</pre>
Note that, without the quotes, the latter string would  be understood as `a1 OR & OR NOT b6`.

Postgres' `ts_vector()` functionalityis very powerful, but it relies on a dictionary. Currently only English language is supported. Hopefully this will change in the future.

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