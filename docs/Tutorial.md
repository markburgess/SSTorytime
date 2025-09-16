
# Tutorial on N4L and SSTorytime

This is a project in the age of so-called Artificial Intelligence. But
it's not about machines.  It's a project about knowledge--*your*
knowledge, human knowledge, but not the abstract knowledge of Mankind
that people talk about for AI. Learning is difficult, and remembering
is difficult, but we can make tools that help. That's what this is
about.

SSTorytime takes information that you enter in the form of

* Personal notes (written in a simple format called N4L).
* Data entered by a computer program (using the API for creating Nodes and Links).

It turns these data into a graph, which you can then browse and query with the ready-made tools here, or by writing your own programs using the API. 

## The tools

Once you've got a bunch of notes, you can upload them into a
searchable format and analyze them for patterns and hidden
connections. Alternatively, you can upload structured data from some
source to form a graph, using the API. This requires programming
knowledge, and a clear idea about how to use a graph in the first
place.  Many articles and books about graph data give bad modelling
advice, so read on first!

You can use the note taking language for writing reports (incident reports, forensic details, patient plans, or just notes about your favourite movies.)

* You can use tools like `graph_report` to study large data. For small data, *you* are the
most important part of the model.

* You can use `searchN4L` to query the graph database.

* Use `pathsolve` to find possible paths between certain items.

* Use `notes` to read back the notes in the page order in which you wrote them.

There are also tools for helping you to get started making notes about a text document.

* Use `text2N4L` to read a text file and select parts to produce an editable file in N4L that you can use as a starting point for your own notes.

## The data model

The knowledge database is divided into

* **Chapter**s: each file starting or section with a ` - ` mark is a chapter section.
* **Context**s: inside each chapter, you can tag subsections with ` :: context, tags, words...` that may be used to limit or enable search criteria.
* **Nodes** and **Links**: each item is a node, and each connection between items is a link. Links have types that you define for yourself, but they must all belong to one of the four meta-types: `leads to`, `contains`, `expresses property`, or `is near or similar to`.

The purpose of chapters and contexts is to enable a separation of
concerns. You could throw everything into one giant chapter with no contexts, e.g. if
you have a single homogeneous graph like a social network, but for most semantic graphs you
will want to separate your notes. e.g. you might have chapters on learning French or Chinese, and you could have contexts like "at the bank" or "ordering at the restaurant".

## Some terminology

When you search for a topic you will match a node in the graph. Relating to that starting point, there are different views:

* The **orbit** of a node is all the nodes connected to it, to some depth or radius (number of link hops).
When you search, you typically get to see matching nodes and their orbits up to **radius** 2 hops.

* A set of **path**s is set of connected links starting from a set of FROM nodes and ending with a set of TO nodes.

* A **pagemap** is a rendering of a page of notes that you typed as N4L, in the same order.

## What use is this?

How how can you compete with AI and with other people in the information age?
It's a bit like going to the gym to get fit. No one can do it for you, but there
are tools to help you. So here is a methodology with tools, to improve the user experience 
of learning.

If you're here, you're likely a programmer or an IT person, so you're looking
for simple answers in code, as software. Software is much more than code, of course.
On one level, there is a graph database here, but that's missing the point.
There are plenty of graph databases, but people use them poorly. This is not that.

## The thinking behind it...

It's the usual story: garbage in, garbage out. Putting data and
information into boxes is easy, but knowing how to find it again is
hard.  But, surely that's why we have databases!? Before search
engines studied search more carefully, people thought they could just
use logic to find data by Random Access. That only works when you know
what you're looking for, but we're become stuck with that model.
That's partly because we are cavalier about stuffing things into data
models we feel are orderly and logical. Yet our thinking is far from
orderly and logical later on when we're trying to find things again.

To make a good way of encapsulating stuff, we need to understand the
process of thinking. Instead of trying to order information (with
hopelessly ambitious ontologies), we need to think about how to
connect the dots of our thinking for later retrieval. This is what
authors and teachers have to think about when producing
material. Everyone knows the difference between a good and a bad
teacher.

* How we think is quite personal. If we try to make the ultimate
database of knowledge, it won't suit everyone. No one can feed
knowledge into you. It's more like tending a garden of your own
thinking.

The usual way of working is this: stuff everything into a database as
quickly as possible and then search everything from the database. The
source data are quickly thrown away, and we rely on an often hastily
thrown together archive. We even call these data warehouses or
lakes--landfill.  The user has to know what field to search in the database, an
often how to write queries in a special language. It's quite far from how
we think in the moment.

All this commodity thinking leads to a canned soup knowledge
cuisine. No wonder we end up with a culture of soundbites and
hearsay. If you want fresh organic knowledge, served up just as you like it, you have to put in the
work of tending your crop yourself. Knowledge, after all, isn't knowledge if you don't know it.
No one can know it for you, so it's up to you to curate it and organize it to suit yourself--perhaps
collaborating with friends or colleagues (but only in small groups).

*If you subscribe to the vision of replacing humans with "AGI" (Artificially Gathered Information),
you'll be shocked and disappointed by this project. If you're a teacher or a writer, you might
quite like it.*

## How to start

With SSTorytime, the source files are your main focus, and the database is just a convenient
aid to remembering, because retrieval sometimes needs help. You will spend most of your time
writing and editing your notes, written in N4L. You adapt the language to suit yourself, with
a couple of simple principles to follow. Then you regularly upload your notes into the database
and see how it looks when it comes out.

You start with a simple text file, in your favourite editor. Somewhere you like to jot down notes, but
as plain text (not a special format like Word or Open Office).
<pre>

- my notes     # you give it a title
               # and you can leave comments to yourself.

IF YOU WRITE IN ALL CAPS, YOU WILL BE REMINDED OF THE NOTE LATER!


 Mostly you just write notes

    "  (e.g.)  This is a simple example that illustrates the line above

 The >ditto symbol of inverted commas has a special meaning

 Other symbols can be defined with your own meanings, like >"special meanings"

</pre>
You can also refer to the previous line
<pre>

@mylabel foot (note) important concept!  # will refer to this label below, defined with @

  # english  to  hanzi   to   pinyin  & back to (english)

    hand    (eh)   手    (hp)  shǒu     (pe)     $THIS.1  

  # references are referred to with $name.position

  $PREV.3 (e.g.) nǐ de zuǒ shǒu  (hp) 你的左手 (he) your left hand

  $mylabel.1  (eh) 脚 (hp) jiǎo (e.g.) nǐ de yòu jiǎo  (ph) 你的右脚 (he) your right foot 
</pre>
You can save this as a text file. It's helpful, but not necessary, to use a suffix `.n4l`.
This file is already available in the distribution:
<pre>
$ cd SSTorytime
$ make
$ cd example
$ ../src/N4L tutorial.n4l
</pre>
When you run this, you'll see something like this:

![A Flow Chart is a knowledge representation](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/nooptions.png 'Without options, you only see your note to self.')

If you choose verbose output, you see more of what's going on:

![A Flow Chart is a knowledge representation](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/verbose.png 'Verbose output')

* First N4L reads a number of configuration files in `SSTconfig/*`. These contain arrow definitions.
* Then it reads your file and chops it into parts that are related.
* N4L thinks that each line is an event, or an item.
* If you out something in parentheses, it treats it as a relationship or an "arrow" that points from one item to another. You can define your own arrows, and the idea is to use them to find things more easily.
* If you use the "ditto" inverted commas under an item, you don't have to type it again.
* You can define special symbols like = >, etc in the configuration to automatically annotate words inside a longer piece of text.

That already covers a lot of possibilities!

## Uploading to the database

To upload notes to the database, you use the `N4L` version of the tool. To get
started, you can try some of the examples:
<pre>
$ cd examples
$ make 
</pre>
You see this runs the following command:
<pre>
../src/N4L -u -wipe doors.n4l Mary.n4l chinese*n4l branches.n4l doubleslit.n4l ConstructionProcesses.n4l wardleymap.n4l 
brains.n4l kubernetes.n4l SSTorytime.n4l integral.n4l reasoning.n4l
</pre>
The `-u` option tells the program to upload to the database. The `-wipe` option, tells it to override
whatever is already in the database and start again. If you want to append new data, you can simply
omit the `-wipe` option:
<pre>
$ ../src/N4L -u LoopyLoo.n4l
</pre>
The examples in the documentation assume you set up from the start with these uploads:
<pre>
$ cd examples
$ make 
$ ../src/N4L -u LoopyLoo.n4l
</pre>

## Searching

You can search the SSTorytime knowledge graph with (hopefully intuitive) phrases like these:
<pre>
from a1
notes about chinese context restaurant
notes chapter brain
please in chinese
paths from a1 to s1
</pre>
You can either use the `searchN4L` tool, or the web browser by running the web server `http_server` and connecting to `localhost:8080` or an address, e.g.`192.168.0.5:8080`.

On the command line:
<pre>

$ ./searchN4L please in chinese
$ ./searchN4L "(zai chengshi)"
$ ./searchN4L from a1
$ ./searchN4L notes about chinese context restaurant
$ ./searchN4L notes about brain context waves
$ ./searchN4L chapter brain
$ ./searchN4L from a1
$ ./searchN4L from a1 to b5
$ ./searchN4L to "target 3"
</pre>
Notice that, if you want to match accented characters using unaccented substitutes (like 'o' for 'ø'), then you write the word in parentheses).

You can also types these directly into the web browser:

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp11.png 'Testing a web interface')
![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp12.png 'Testing a web interface')
![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp13.png 'Testing a web interface')


![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp2.png 'Testing a web interface')


![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp3.png 'Testing a web interface')

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp4.png 'Testing a web interface')

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp5.png 'Testing a web interface')

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp6.png 'Testing a web interface')



## Browsing the results

Eventually, there will be tools for scripting the search in simple
ways, because the most powerful ways to search any structure are to
use a programming language that allows you to express your own
intent. You can see examples in the demos and proof of concept
directory under src/demo_poc.  But as the project progresses, you can
use the `notes` and `searchN4L` tool to play around with the result.
The simplest way to see what you entered (which is like a cleaned up version of `more`)
is to use:
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



## What's the point?

When you make notes, you should think about what you want to see when you look back at your notes.
For example, suppose you are learning French. 

<pre>
- French phrases

 petit-déjeuner (means) breakfast

    "  (e.g.) Je voudrais commander le petit-déjeuner (means) I would like to order breakfast
    "  (note) Don't forget to say please!
</pre>

* Notice that you can use accents and Unicode characters freely. 
* Notice that you can make intuitive short names for arrows like (e.g.). You can define what these mean in the configuration. More on that later.
* Notice you can define many different kind of arrows with different meanings, e.g. (e.g.), (note).

You start to see a pattern in the notes: usually, if you're trying to
remember something, you want to see the raw thing, like the word for
breakfast. You also want to remember how to use it, so you naturally
add a couple of examples just after the item. N4L will connect these
dots to show you related things later. But, more importantly, you
don't event have to do anything with N4L except write stuff
down. These notes are already your potential knowledge in the
making--and this simple structure helps you to be systematic in
writing things down. You will spend a lot of time just curating these
notes, altering, editing, improving, and most of the value is actually
there.

You don't learn French by putting it in a database. You learn by revisiting it, and by remembering
relevance and context. Just writing the notes is 80 percent of the job.
* The N4L compiler can help you to find errors and make a good structure.
* When your notes become long, it's hard to keep an good overview.
* Once inside the database, you can present the information in different ways.
When you upload it to a database,
you can still find things quickly, even when you're not sitting in front of your text editor--perhaps using your phone.

From here, it's up to you how you want to proceed. If you're feeling perverse, you could add
more languages:

<pre>
- French phrases, and other languages

 petit-déjeuner (means) breakfast

    "  (e.g.) Je voudrais commander le petit-déjeuner (means) I would like to order breakfast
    "  (note) Don't forget to say please!

 I would like to order breakfast
      # let's add Norwegian..
    "  (betyr på norsk) Jeg vil bestille frokost

      # let's add Mandarin
    "  (中文意思是) 我想订早餐 

</pre>



## It's not rocket science, unless ...

Writing notes isn't all that easy. It takes a certain self-discipline, but it gets easier over time.
Forcing yourself to start is often the biggest hurdle. The news is that you can drip new notes into your
working files occasionally over a long time. You don't have to sit down and study for hours at a time.
On the other hand, it's only when you do make time to sit and study that you actually learn.

Once again, the message is: writing it down is nice, putting it into a database is cool, but it's all
wasted effort if you don't look at it yourself regularly. No one learned French by writing in their school
book, or even by cramming for the exam. You only learn by using knowledge. It isn't knowledge if you don't know it. 

It's not rocket science, unless of course it is rocket science.
<pre>

--rocket science

 rockets are powered projectiles

 Rocket science in finance (wikipedia) "https://en.wikipedia.org/wiki/Rocket_science_(finance)"

 HOW TO SPELL VONBRAUNS NAME???

 Werner Von Braun (developed) V2 aircraft
         "        (developed) NASA early rockets  

 ASK FRIEND AT NASA...

 Apollo Program
 Mercury Program
 Gemini Program ...

 Space Camp movie ..

</pre>

## Tracking Your Progress

Currently under development is a new feature.
As you start to use the tool more, you can use the checkboxes to click and leave a trace of what you've worked
on. This will be traced and give feedback on your learning goals.

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/progresstracker.png 'Testing a web interface')


