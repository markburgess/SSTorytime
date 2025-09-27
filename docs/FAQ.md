
## Frequently Asked Questions

* **Can I use URLs in N4L?**

Yes, you should enclose them in quotes because they usually include the substring "//", which is also a comment designation.

* **Why does it take so long to upload data?**

Uploading to a database is a slow process compared to retrieving as there are many checks that have to happen. try to debug your data as far as possible using the text interface in N4L before actually committing to the database.

In addition, *Unicode* decoding is a very slow process so long files seem to take forever to read, never mind the actual database uploading. I don't know of any way to speed this up presently. Unless we know that a file is simple
ASCII encoding, it's easy to get bad character conversion without using this longwinded decoding.

* **Why are there relationships that I didn't intend when I browse the data?**

Be careful to ensure that you haven't accidentally used any of the annotation markers (e.g. +,-,=) without surroundings spaces in your text, as these will be interpreted as annotations. Use the verbose mode in N4L to debug.

* **Why do I see chapters that don't seem to be relevant?**

This is probably a result of certain words and phrases belonging to more than one chapter, and thus bridging chapters that you didn't intend. This bridging is intentional, as it allows >"lateral thinking", which is an important source of discovery.

* **Why doesn't pathsolve understan my search on the command line**

Shell characters interfere with the syntax. We need to escape characters, e.g. using single quotes to avoid expansion:
<pre>
$ ./pathsolve -begin '!a1!' -end s1
$ ./searchN4L \\from '!a1!'
</pre>

* **Why doesn't a path solution work?**

Path solving is a potentially exponential process. Without some constraints it could take a very long time. You can restrict the time significantly by specifying precise start and end nodes, e.g. write `from !a1! to !b6!` to match the precise a1 (not a substring of many possibilities. You can also use a context `from a1 to b4 context connection`. See also 'Why are the results different each time?'

By default, SSTorytime will also try to search all possible path types. Narrative links are neary always arrow type 1 (leads to), so you can try to limit by arrow too `from door arrow 1`.

* **Why are the results different each time?**

Lookup in a database is not a deterministic process. The database may select different values on each search and return them in a different order. The default number of data returned is 10 items. If there are many possible matches, the probability of getting the same 10 will decrease with more possibilities. You can also increase the number of matches `mysearch limit 20`. The more you constrain your search the more likely you are to get the same answer each time. 