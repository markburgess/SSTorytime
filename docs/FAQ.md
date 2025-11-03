
## Frequently Asked Questions

### General

* **What does it do?**

SSTorytime tools allow you capture information that you would like to know, make notes on, edit, and learn
over time, to yield a searchable database of "knowable" information: but, of course, it's not magic--it isn't
knowledge unless you actually know it yourself. It let's you see your own thinking in the mirror!

It turns written notes (or data input through an API) into a map or
"Knowledge Graphs", to show you how you think about things. It creates
your personal view, rather than someone else's imposed point of view.
It encourages you to engage in the on-going process of learning by
making notes, editing, compiling into a graph/map and iterating so
that you learn as part of that process. If you choose to share the result, others can browse your view
and see an explanation your terms, in your language.

Thinking is easier in free text, rather than trying to force feed your thoughts into a scheme for searching.

Using the additional tools `text2N4L`, you can scan an existing text and create an initial set of notes
automatically to make browsing a book easy, searchable, and informative.

* **What would I use it for?**

* * Do you have HR processes that track staff capablities?
* * Are you designing team topologies without noting down your thinkning?
* * Are you designing a process, a user interface, user journeys through a website or organization?
* * Are you mapping supply chains or tracing work execution?
* * Are you creating service directories, tracing dependencies?
* * Do you want to annotate your music collection or photos, tracing attribution, comments, similarities, etc

* **What is the process?**

The five step programme is this:
* * JOT IT DOWN WHEN YOU THINK OF IT. . .
* * TYPE IT INTO N4L AS SOON AS YOU CAN. . .
* * ORGANIZE AND TIDY YOUR N4L NOTES EVERY DAY. . .
* * UPLOAD AND BROWSE THEM ONLINE. . .
* * REMEMBER, IT ISN'T KNOWLEDGE IF YOU DON'T ACTUALLY KNOW IT !!

Look at your thinking in the mirror.

* **Why does it take so long to upload data?**

Uploading to a database is a slow process compared to retrieving as there are many checks that have to happen. try to debug your data as far as possible using the text interface in N4L before actually committing to the database.

In addition, *Unicode* decoding is a very slow process so long files seem to take forever to read, never mind the actual database uploading. I don't know of any way to speed this up presently. Unless we know that a file is simple
ASCII encoding, it's easy to get bad character conversion without using this longwinded decoding.

* **Why do some searches take a long time?**
Immediately after uploading the data, the database will be building its indices. Indexing takes time, then the return time should stabilize and be faster. After that, the scope of the search determines the speed of a query. If you are searching for a lot of data, it takes time to assemble into a graphable structure. 

### Writing notes in N4L

* **Can I use URLs in N4L?**

Yes, you should enclose them in quotes because they usually include the substring "//", which is also a comment designation.


* **Why are there relationships that I didn't intend when I browse the data?**

Be careful to ensure that you haven't accidentally used any of the annotation markers (e.g. +,-,=) without surroundings spaces in your text, as these will be interpreted as annotations. Use the verbose mode in N4L to debug.

* **Why do I see chapters that don't seem to be relevant in search results?**

SSTorytime does some "lateral thinking". If you don't explicitly restrict to a particular chapter, it will
take examples from everywhere.
Seeing unexpected results is probably a result of certain words and phrases belonging to more than one chapter, and thus bridging chapters that you didn't intend. This bridging is intentional, as it allows >"lateral thinking", which is an important source of discovery.

* **Why doesn't pathsolve understand my search on the command line**

Shell characters interfere with the syntax. We need to escape characters, e.g. using single quotes to avoid expansion:
<pre>
$ ./pathsolve -begin '!a1!' -end s1
$ ./searchN4L \\from '!a1!'
</pre>

* **Why doesn't a path solution work?**

Path solving is a potentially exponential process. Without some constraints it could take a very long time. You can restrict the time significantly by specifying precise start and end nodes, e.g. write `from !a1! to !b6!` to match the precise a1 (not a substring of many possibilities. You can also use a context `from a1 to b4 context connection`. See also 'Why are the results different each time?'

By default, SSTorytime will also try to search all possible path types. Narrative links are neary always arrow type 1 (leads to), so you can try to limit by arrow too `from door arrow 1`.

* **Why does a path search take so long?
Path searches grow exponentially with the length of the path, so they get slower and slower as the distance between nodes
increases. If you know the type of arrow along the whole path, you can speed up the search by specifying the arrow types, or the sttypes, e.g. using the STtypes:
<pre>
./searchN4L -v \\from \!gun\! \\to scarlet \\arrow +3,-3,0
</pre>
And using the arrows:
<pre>
./searchN4L -v \\from \!a1\! \\to b6 \\arrow 20,21
</pre>
Remember to always give pairs of arrow,inverse since the FROM and the TO match opposite arrow directions.

* **Why are the results different each time?**

Lookup in a database is not a deterministic process. The database may select different values on each search and return them in a different order. The default number of data returned is 10 items. If there are many possible matches, the probability of getting the same 10 will decrease with more possibilities. You can also increase the number of matches `mysearch limit 20`. The more you constrain your search the more likely you are to get the same answer each time. 