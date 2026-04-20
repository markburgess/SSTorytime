
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

## Web version

The web browser has an equivalent to the notes command line tool. Enter the relevant chapter into the chapter field and
press `browse`, then use the `next` and `previous` page buttons to move through the pages.

![Equivalent in web browser](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/notes.png 'notes search')


 