
# N4L - Notes for Learning
## A simple knowledge management language

*Notes for learning*<br>
*Narrative for loading*<br>
*Network for logical inference*<br>
*Nourishment for life*

N4L is an intentionally simple language, for keeping notes and
scanning into a structured format for uploading into a database or for
use with other tools.  The language is designed to encourage you to
think about how you express and structure notes. That, in turn,
encourages you to revisit, tidy and organize the notes again and again, while
being able to quickly turn them into a searchable graphical database, from which
and can reason through stories.

*One of the important ways we make notes is to draw pictures and place concepts
on maps, in which things are close together or laid out in a logical manner,
In the future, N4L should be able to support simple sketches too, but that's
for future development.*

## Why do we need a language?

These days there are too many software engineers and we tend to make
systems for them. So people are simply expected to learn how to use
computer code, and "APIs" do enter data.  This is not intuitive
(actually to anyone). Computers are a tool, and tools are supposed to
do the work for humans, not the other way around! So we want to try to make data entry easy.

The purpose of using a simple yet semi-formal language as a starting
point is to avoid the "information model trap" that befalls many data
representations, i.e. forcing users to put everything into a pre-approved model,
like filling out a rigid form. This makes it hard to back out of decisions
and change our minds. It makes modelling fragile and fraught with risk.

Without any structure, it's only guesswork to
understand intent. N4L is a compromise that allows you to use any kind of
familiar editor to write notes in pure text (Unicode).

## Command line tool

The N4L tool ingests a file of "notes" written in a simple language
and turns it into a machine representation in the form of a "Semantic Spacetime" graph
 (a form of knowledge graph). This format is only tangentially related to the
 usual Resource Description Framework (RDF)), so we shall not use of
 refer to RDF in what follows, except to occasionally clarify the distinction. 
The command options currently include:
<pre>
usage: N4L [-v] [-u] [-s] [file].dat
  -adj string
        a quoted, comma-separated list of short link names (default "none")
  -d    diagnostic mode
  -s    summary (node,links...)
  -u    upload
  -v    verbose
</pre>
For example, to parse and validate a file of notes, one can simply type:
<pre>
$ N4L chinese.in
$ N4L chinese.in Mary.in kubernetes.in
</pre>
Any errors will be flagged for correction. Using verbose mode gives extensive
commentary on the file, line by line:
<pre>
$ N4L -v chinese.in
</pre>
The final goal will normally be to upload the contents of the file to a database:
<pre>
$ N4L -u chinese.in
</pre>
However, before that, there are several operations than can be performed more efficiently
just from the command line for many data sets. This is because most knowledge input
is quite small in size, and quick feedback is very useful for ironing out flaws
and improving your source note material.

We can look at the subset of notes that are related by
a certain kind of relation, using abbreviated labels for relations. 
For example, to look for items linked by relation "(pe)" (which stands
for Pinyin to Hanzi translation) in a file of Chinese language, we could write:
<pre>
$ N4L -v -s -adj="pe" chinese.in
</pre>
We can add other kinds of relation too to expand the set:
<pre>
$ N4L -v -s -adj="pe,he" chinese.in
</pre>
This extracts a sub-graph from the total graph. It can be quite effective,
because most knowledge graphs are only sparsely linked (which is why logical
searches tend to yield nothing of interest).

In verbose mode, the standard output shows a summary of the text (events or items, etc)
and an excerpt of the adjacency matrix.
<pre>

$ N4L -v -s -adj="" Mary.in


------------------------------------
SUMMARIZE GRAPH.....

------------------------------------

0        Mary's mum

1        Nursery rhyme

0        SatNav invented later

0        Mary had a little lamb
         ... --( example of , 1 )-> Nursery rhyme [cutting edge high brow poem]
         ... --( written by , 1 )-> Mary's mum [poem cutting edge high brow _sequence_]
         ... --( then the next is , 1 )-> Whose fleece was white as snow [poem cutting edge high brow _sequence_]
         ... --( note/remark , 1 )-> Had means possessed not gave birth to [_sequence_ poem cutting edge high brow]

1        Had means possessed not gave birth to

2        Whose fleece was white as snow
         ... --( then the next is , 1 )-> And everywhere that Mary went [poem cutting edge high brow _sequence_]

3        And everywhere that Mary went
         ... --( then the next is , 1 )-> The lamb was sure to go [cutting edge high brow _sequence_ poem]

4        The lamb was sure to go
         ... --( note/remark , 1 )-> SatNav invented later [cutting edge high brow _sequence_ poem]

-------------------------------------
Incidence summary of raw declarations
-------------------------------------
Total nodes 8
Total directed links of type Near 0
Total directed links of type LeadsTo 4
Total directed links of type Contains 1
Total directed links of type Express 2
Total links 7 sparseness (fraction of completeness) 0.125
    - row/col key [ 0 / 8 ] Had means possessed not gave birth to
    - row/col key [ 1 / 8 ] SatNav invented later
    - row/col key [ 2 / 8 ] The lamb was sure to go
    - row/col key [ 3 / 8 ] Mary had a little lamb
    - row/col key [ 4 / 8 ] Whose fleece was white as snow
    - row/col key [ 5 / 8 ] Nursery rhyme
    - row/col key [ 6 / 8 ] And everywhere that Mary went
    - row/col key [ 7 / 8 ] Mary's mum

 directed adjacency sub-matrix ...

     Had means posse .. (   0.0   0.0   0.0   0.0   0.0   0.0   0.0   0.0)
     SatNav invented .. (   0.0   0.0   0.0   0.0   0.0   0.0   0.0   0.0)
     The lamb was su .. (   0.0   1.0   0.0   0.0   0.0   0.0   0.0   0.0)
     Mary had a litt .. (   1.0   0.0   0.0   0.0   1.0   1.0   0.0   1.0)
     Whose fleece wa .. (   0.0   0.0   0.0   0.0   0.0   0.0   1.0   0.0)
       Nursery rhyme .. (   0.0   0.0   0.0   0.0   0.0   0.0   0.0   0.0)
     And everywhere  .. (   0.0   0.0   1.0   0.0   0.0   0.0   0.0   0.0)
          Mary's mum .. (   0.0   0.0   0.0   0.0   0.0   0.0   0.0   0.0)

 undirected adjacency sub-matrix ...

     Had means posse .. (   0.0   0.0   0.0   1.0   0.0   0.0   0.0   0.0)
     SatNav invented .. (   0.0   0.0   1.0   0.0   0.0   0.0   0.0   0.0)
     The lamb was su .. (   0.0   1.0   0.0   0.0   0.0   0.0   1.0   0.0)
     Mary had a litt .. (   1.0   0.0   0.0   0.0   1.0   1.0   0.0   1.0)
     Whose fleece wa .. (   0.0   0.0   0.0   1.0   0.0   0.0   1.0   0.0)
       Nursery rhyme .. (   0.0   0.0   0.0   1.0   0.0   0.0   0.0   0.0)
     And everywhere  .. (   0.0   0.0   1.0   0.0   1.0   0.0   0.0   0.0)
          Mary's mum .. (   0.0   0.0   0.0   1.0   0.0   0.0   0.0   0.0)

 Eigenvector centrality score for symmetrized graph ...

     Had means posse .. (   0.7)
     SatNav invented .. (   0.2)
     The lamb was su .. (   0.4)
     Mary had a litt .. (   0.9)
     Whose fleece wa .. (   1.0)
       Nursery rhyme .. (   0.7)
     And everywhere  .. (   0.5)
          Mary's mum .. (   0.7)

</pre>
A useful ranking of nodes (known as EVC, or Eigenvector Centrality, which is something like Google's PageRank) 
can be calculated from the weighted graph matrix (see below). The higher the score number, the more
interconnected or "important" a term of text is, e.g.
<pre>
$ ../src/N4L -v -s -adj="" chinese.in

  ...

 Eigenvector centrality score for symmetrized graph ...

            Fángjiān .. (   0.3)
             jiàoshì .. (   0.8)
              Kètáng .. (   0.2)
                   教室 .(   0.2)
     place/area/dist .. (   0.1)
                   表现 .(   0.7)
            Biǎoxiàn .. (  0.6)
                   课堂 .(   0.8)
         performance .. (   0.5)
                   房间 .(   0.3)
                   地方 .(   0.2)
                   表演 .(   0.1)
           classroom .. (   1.0)
                room .. (   0.2)
              Dìfāng .. (   0.3)

</pre>

## Language syntax

The N4L language has only a small number of features. It's power hopefully lies in its simplicity.
It consists of text, small or larger (but pragmatically not huge), and relationships between them
(in parentheses). The vocabulary of parenthetic relations is defined separately in a set of
files in the configuration directory called `SSTconfig/` 
(see below).

<pre>

#  a comment for the rest of the line
// also a comment for the rest of the line

-section/chapter                 # declare section/chapter as the subject

: list, context, words :         # context (persistent) set
::  list, context, words ::      # any number of :: is ok

+:: extend-list, context, words :: # extend the existing context set
-:: delete, words :                # prune the existing context set

A                                # Item
Any text not including a "("     # Item
"A..."                           # Quoted item
'also "quoted" item'             # Useful if item contains double quotes
A (relation) B                   # Relationship
A (relation) B (relation) C      # Chain relationship
" (relation) D                   # Continuination of chain from previous single item
$1 (relation) D                  # Continuination of chain from previous first item
$2 (relation) E                  # Continuation from second previous

@myalias                         # alias this line for easy reference
$myalias.1                       # a reference to the aliased line for easy reference

NOTE TO SELF ALLCAPS             # picked up as a "to do" item, not actual knowledge

"paragraph =specialword paragraph paragraph paragraph paragraph
 paragraph paragraph paragraph paragraph paragraph
  paragraph paragraph =specialword *paragraph paragraph paragraph
paragraph paragraph paragraph paragraphparagraph"

where [=,*,..]A                        # implicit relation marker

</pre>
Here A,B,C,D,E stand for unicode strings. Reserved symbols:
<pre>
(), +, -, @, $, and # 
</pre>
Literal parentheses can be quoted. There should be no whitespace after the initial quote
of a quoted string.

## Reserved relation names

For the purpose of automating sequence capture and rendering of multimedia objects,
the following relation types are reserved:

<pre>
   leadsto:
      + then (then) - prior

   properties::
      + has url   (url) - is a URL for (isurl)
      + has image (img) - is an image for (isimg)

</pre>

## Sequence mode

Sometimes it's useful to link items together into a chain or sequence.
By adding the sequence directive to a context. From the example of the Mary had a little lamb above:
<pre>

$ more Mary.in

-poetry

 :: cutting edge, high brow ::

 +:: _sequence_ , poem ::   // starting sequence mode

@title Mary had a little lamb  (note) Had means possessed not gave birth to
              "                (written by) Mary's mum

       Whose fleece was white as snow
       And everywhere that Mary went

       // no need to be contiguous

       The lamb was sure to go        (note) SatNav invented later

 -:: _sequence_ ::          // ending sequence mode

 $title.1 (example of) Nursery rhyme

</pre>
This results is a sequence of lines linked by `then' arrows, until the context is removed.
<pre>
Mary had a little lamb (then) Whose fleece was white as snow (then) ...
</pre>
Then is a pre-defined and effectively reserved association.

* Only the first items on a line are linked. 
* Only new items are linked, so the use of a " or variable reference will not trigger a new item.

## Example

Assocations have explanatory power, so we want to take advantage of that.
On the other hand, we don't want to type a lot when making notes, so
it's sensible to make extensive use of abbreviations.

<pre>
-chinese notes

::food::

  meat    (is english for the pinyin) ròu
   "      (is english for the chinese or hanzi)  肉

  # more realistic with abbreviations ...

 菜 (hp) Cài (pe) vegetable 
 meat (eh) 肉 (hp) Ròu
 beef  (eh) 牛肉  (hp) Niúròu
 lamb  (eh) 羊肉  (hp) Yángròu
 chicken (eh)  鸡肉 (hp)  Jīròu

:: phrases, in the hotel ::

@robot I'm waiting for some food from the robot (eh) 我在等机器人送来的食物 (hp) Wǒ zài děng jīqìrén sòng lái de shíwù

:: technology ::

jīqìrén (pe) robot (example) $robot.1

</pre>

Notice how the implicit "arrows" in relations like 
<pre>(is english for the pinyin)</pre> or its short form
<pre>(pe)</pre> effectively define the `types' of thing they are 
attached to at either end. So we don't need to define the ontology for things
because it emerges automatically from the names
we've given to relationships.

Semantic reasoning can make use of both the precision and the fuzziness of associative types
when reasoning. This is a powerful feature that enables automated
inference with lateral thinking, just as humans do. Languages that use
logic to define ontologies are greatly over-constrained and make
reasoning precise but trivial, because they can only retrieve exactly
what you typed into the model.

## A hint for learning

When you are interested in learning something, say a foreign language, put the thing you are'trying to
learn first on the line, because your eye will tend to favour the first thing on the line, and your
brain will immediately kick in and try to process it, whereas it will easily skip over the later things
on the line. So, for example, if you're trying to learn Chinese, as in the example above, don't write
<pre>
 meat (eh) 肉 (hp) Ròu
 beef  (eh) 牛肉  (hp) Niúròu
 lamb  (eh) 羊肉  (hp) Yángròu
 chicken (eh)  鸡肉 (hp)  Jīròu
</pre>
Write this:
<pre>
 THING TO LEARN  (relation)   WHAT IT MEANS

 ròu     (ph)  肉   (he)  meat
 niúròu  (ph)  牛肉  (he)  beef 
 yángròu (ph)  羊肉  (he)  lamb  
 jīròu   (ph)  鸡肉  (he)  chicken
</pre>
Now you will immediately see the left hand words that you're trying to learn, and it takes an effort to
look to the right when you've tried to remember and want the answer.

In this order, you can also easily add other annotations like examples and related notes:

<pre>
-notes on chinese

:: please, thank you, thankyou ::

 thankyou (eh) 谢谢 (hp) xièxiè

 qǐng (ph) 请 (he) please
   " (e.g.) qǐng zài zhèlǐ děng  (ph) 请在这里等 (he) please wait here
   " (e.g.) qǐng jìn             (ph) 请进    (he) please enter/come in
   " (e.g.) qǐng kàn yīxià       (ph) 请看一下 (he) please take a look
   " (e.g.) qǐng děng yīxià      (ph) 请等一下 (he) please wait a bit
   " (e.g.) qǐng zuò             (ph) 请坐    (he) please sit

:: yes, no, negative :: 

 yes       (eh) 是的 (hp) Shì de
  "        (note) There is no simple one to one translation for yes and no
 yes/right (eh) 对   (hp) Duì

</pre>

## How relationships work

A piece of text can be thought of as an item or an event.
Relationships between items or events are written inside parentheses, as in the
examples above. They are designed to be highly
abbreviated for note taking. 

As written, the examples above look a bit like any old RDF (Resource
Description Framework) triplets. However, with the underlying
assumptions of the language that we'll lay out below, they are much
more powerful than the ad hoc references in RDF, because RDF
relationships are just strings without any semantics.

In order for references to be used for reasoning (and effective
semantic search), they need some basic properties. The simplest thing
we can do is to classify each relationship as though it were a special
case of one of four basic types, depending on how you want to
interpret it. This might be tricky in the beginning, so you can stick
to some predefined relation.

It turns out that every relationship basically falls into one of
four basic types that help you to imagine sketching the items on a map.
Here are the four types:
* 0 **similarity / near, alike** something is close to something else (proximity,closeness)
* 1 **leadsto    / affects, causes** one thing follows from the other (sequences)
* 2 **contains   / contains** something is a part of something else (boxes in boxes)
* 3 **properties / express** something just has a name or an attribute (descriptive)

For example:
* 0 - A **(sounds like)** B, or B **(sounds like)** A
* 1 - A **(causes)** B , or B **(is caused by)** A
* 2 - A **(is the boss of)** B, or B **(has boss)** A
* 3 - A **(has a degree in)** B, B **(is a qualification of)** A

*(Technical note the use of integers allows us to use signs for orientation.
Similarity is directionless 0 = -0; for the others there is a difference between
positive and negative inverses.)*

** When naming relations, stick to one tense (e.g. present tense) and for the short form use
something that would easiy match any occurrence of the meaning of the arrow. This will make
debugging and automation of suggestion easier later. In other words, try to be unambiguous and consistent
in your use of language.**

These four classes of association can be literal or metaphorical (all language
is an outgrowth of [metaphors for space and time](https://www.amazon.com/Smart-Spacetime-information-challenges-process/dp/B091F18Q8K/ref=tmm_hrd_swatch_0)).
behave like placing items around
each other in a mind-map on paper. Things that belong close together
because they are aliases for one another are *similar*.  If one thing
leads to another, e.g. because it causes it or because it precedes it
in a history then we use *leadsto*. Some items are parts of other items,
so we use *contains*. Finally, something that's purely descriptive
or is expressed by an item, e.g. "is blue" or 

Some authors who write about semantic networks have suggested that the
way to think about arrows and nodes is as "nouns" (things) and "verbs"
(actions). This is a simple idea, but it's not quite right. The catch lies
in the way language semantics rely almost entirely on metaphors to express
ideas. We frequently speak of "nouning verbs" and "verbing nouns", e.g.
in Silicon Valley speak:
<pre>
 The company's spend is ...   (vs)    I need to spend .. an expenditure
 I have a big ask ...         (vs)    I need to ask you .. a question

 I question your use of language ... with a question
 I expensed by trip ... as an expense
</pre>
Spend is a verb (expenditure or budget are nouns. Ask is a verb, question is
a noun, but we now use both for both!
We see that language is used and abused in fluid ways, so we need more
discipline in thinking about what the functions of terms are.


## Context - what is it?

You add "context" to knowledge by adding keywords and phrases to describe the circumstances
in which your notes apply. Think of these like the "tags" that you are often asked to
add to articles and posts on social media. But context can be much extensive than keywords; and you
are not limited to five items!

When searching for knowledge later, you will typically start by entering a context: what are
you looking for. Context can be a subject heading, a topic, etc. The items under this heading
are related to that, but might not actually contain that keyword. For example, if you are looking
for phrasse in a foreign language that have to do with a restaurant visit, you would arrange to
organize and tag certain phrases with `restaurant, eating, cafe, pay the bill', etc.
The way context is used is still an area of development, but there are two things to remember:

* The keywords are something like a sensory stream, describing what might be
going on in the mind of the user when they are looking for the relevant information: is it hot, cold, are
you busy, relaxed, angry, in a restaurant, on the bus, etc. You imagine classifying things you want to remember or know about
according to these `states of being'.
* Contexts are 'lookup' keys, acting like an index or table of contents.
* Although we will later show how to apply logical thinking to focus and sharpen searches, you should
not think of context as logical (Boolean) variables.

That said, you are free to write collections of contexts either with commas or "OR" bars, as you like:
<pre>

:: position, location , directions | orientation | configuration ::

 compass (has direction) north
   "     (has direction) south
</pre>
*Technical note: N4L's context model is based on the contextual decision-making from the software called CFEngine,
which is an agent based language for describing maintenance policy in computers.
If you know CFEngine, you might be confused about how to use context in N4L--that's because it's logically
'backwards' compared to the CFEngine policy language. In CFEngine, the sensory feed from a computer comes
from the agents that observe and inspect the state of the computer, and the context class expressions
in the CFEngine language are effectively search criteria to select when to activate, given the set of
states or classes observed. In N4L the computer under observation is the set of notes you read into it.
So the contexts are terms that provide the sensory data, not the selection criteria. The user will later
be the `policy engine', deciding what is relevant. So, you will never need to type logical expressions in
your notes, except for highly skilled and specialized notes that we'll come back to later.*

## The `SSTconfig` directory

**NB: this configuration layout has changed **

Arrows for use in notes are defined in a number of files under the `SSTconfig/` directory.
The N4L compilers will look for such a directory under `./` and `../` etc, or you can set an environment
variable
<pre>
setenv SST_CONFIG_PATH mypath
export SST_CONFIG_PATH = my_path
</pre>
There is now a separate file for each of the arrow STtypes:
<pre>
arrows-LT-1.sst
arrows-NR-0.sst
arrows-CN-2.sst
arrows-EP-3.sst
annotations.sst
</pre>
so that everyone can share a set of standard definitions. Since it can be difficult to figure
out how to register arrows, it seems a more sustainable way of proceeding than expecting everyone
to define their own arrows.

The structure of this file is similar to the basic language, but the sections
are used to define the four types of arrows and their meanings.
The syntax takes the following form for the first three kinds of arrow:
<pre>
- [leadsto | contains | properties ]

    + forward reading (forward alias) - reverse reading (backward alias)
    ...
</pre>
For the fourth or zeroth type, there is only one direction for the meaning,
since the arrow reads the same both forwards and backwards. Note, this does not
mean the arrow is directionless, only that the reading of the arrow against its flow
has the same meaning! 

### Leads to arrows (causality and order)

Arrows that express relationships putting items in a certain order
are called "leads to" arrows:
<pre>
- leadsto

 	# Define arrow causal directions ... left to right
 	# what does A -----> B mean, and what is its opposite?

        + leads to (lt) - arriving from (af)

 	# causal order, preconditions, succession

 	+ forwards (fwd) - backwards (bwd)              # A (forwards)	B,  B (backwards) A
 	+ affects  (aff)  - affected by (baff)  	# A (affects) 	B,  B (affected by) A
 	+ causes   (cf)  - is caused by (cb)
 	+ used for (for)  - is a possible use of (useof)
 	+ generates (cf)  - is generated by (gen)
 	+ determines (det)  - is determined by (detby)

        // Flow charts / FSMs etc

	+ next if yes (ifyes) - is a positive outcome of (bifyes)
	+ next if no (if no)  - is anegitive outcome of (bifno)

        + intends (intt)    - is the intent of (iof)
        + proposed (prop)    - proposed by (propby)
        + decided (decide)    - decided by (decidby)
        + spoke to (spoke)    - was spoken to by (talked)
        + implements (impl) - was implemented by (implorg)        
        + named after (named) - inspired the name (inspname)

 	# these next two are mutually complementary interpretations
 	+ succeeded by (succ) - preceded by (pre)
 	+ comes before (bfr)  - comes after (aft)

        ## other meanings

        + wrote (wrote) - written by (written)
        + invented (invent) - invented by (inventby)

        # Numbers can be interpreted either as set order (value) 
        # or by set containment (count), so be careful with semantics!

        # succeeds is more accurate in terms of order
        #
	# + is less than (lth) - is greater than (gth)

     :: construction, building, industry ::

     + supplies (supply) - is supplied by (supplyby)
     + delivered to (delivaddr) - delivery address for (delivgoods)

     + handles (handles) - is handled by (handleyby)      // has dual meanings!
     + coordinates (coord) - is coordinated by (coordby)

     :: chinese language ::

 	+ english for pinyin (ep) - pinyin for english (pe)  
 	+ pinyin for hanzi (ph) - hanzi for pinyin (hp)
 	+ hanzi for english (he) - english for hanzi (eh)
 	+ english for Norwegian (en) - Norwegian for english (ne)
 	+ english to norsk (en) - norsk to english (ne)

</pre>

### Contains arrows (membership)

Belonging to a group or a container is also a directed relationship
so we read it differently in either direction.
<pre>
- contains

 + has component (has) - is component of (part)
 + contains (c)        - is within (in)
 + is a set of (setof) - is part of set (inset) // designations can be multi-valued
 + contains (cont)     - is an element of (el)
 + subsumes (sub)      - is subsumed by (subby)
 + swallows (sw)       - is swallowed by (swby)
 + consists of (cons)  - is part of (pt)
 + makes use of (uses) - occurs in (occurs)
 + has aspect (aspect) - is an aspect of (aspect of)
 + has key issue (key issue) - is a key issue of (is key)
 + generalizes (general) - is a special case of (special)
 + includes (includes) - is a kind of of (kind of)
 + has example (hasex) - example of (ex)
 + has member (memb) - belongs to (belong)

 + has friend (fr) - is considered a friend of (isfrof)
 + discusses (disc) - is discussed in (isdisc)
 + obeys the rule (rule) - is a rule for (rule4)

 + owns (owns) - is owned by (ownby)
 + rents (rents) - is rented by (rentby)
 + employer of (employs) - is employed at (workat)
 + based in (org) - is the home of (home)

 + right word? (word?) - right usage? (usage?)

</pre>


### Properties arrows (attributes)

<pre>
- properties

 # properties are more type-centric in a logical sense
 # because they are ontological

 #  A (expresses) B, B (is expressed by) A

 + has resource/reference  (resource) - is a resource for (isresource)
 + NOT correct about (wrongabt) - is NOT a case of (wrong)
 + expresses (expr)          - is expressed by  (exprby)
 + has property (prop)       - is a property of (propof)
 + means (means)             - is meant by (meansb)
 + is pronounced as (pronas) - is the pronunciation for (pronof)
 + is short for (short for)  - can be shortened to (shorter)
 + has friend (friend) - is a friend of (isfriend)

 # Here we see type classifiers in action

 + is called (called) - name of (pname)

 + may have state (instate) - is a possible state of (stateof)
 + may have value (hasX)    - is a possible value for (isX)

 + note/remark (note) - is a note or remark about (isnotefor)
 + added remark (remark) - is a remark concerning (concerns)
 + please note! (NB) - is an important remark concerning (important regarding)

 + stands for (sfor)  - is a case of (case of)
 + refers to (ref)    - may be referred to as (refas)
 + has role (role)    - is the role of (isrole)
 + has employment status (emplstatus) - employment status of (emplstatof)

 + likes (lk) - is liked by (lkby)

</pre>

### Similarity or proximty

How similar or close together are things? Simiarity or equivalence is
not a directed relation in its meaning, but it can be directed in its
applicability. For example, A is next to B implies that B is next to A,
there is no other reading of it. However A next to B doesn't mean that
B must be next to A: what if there is the equivalent a one-way street or one-way glass
connecting things.
<pre>
- similarity # nearness, proximity

 looks like         	(ll)
 sounds like        	(sl)

 equals             	(eq)
 same as                (=)
 is not the same as 	(!eq)
 is not                 (not)

 similar to         	(sim)
 associated with    	(ass)  # (loose coupling)
 see also               (see)
 near               	(nr)

 met with               (met) // a mutual coincidence

 comes together with    (and)
</pre>

### Annotations

Annotations are special characters used to mark up a longer text, i.e. to pick out
certain words within a body of text. A word that is prefixed by such a character will
be linked to the whole text using the relationship declared in this list, e.g.
<pre>
  in a sentence %specialword can be marked ...
</pre>
The `% sign` generates an implicit link:
<pre>
  in a sentence +specialword can be marked ...   (discusses) specialword
</pre>
*Languages that do not use spaces are not supported here, so one must introduce
an artificial space separator in those cases.*

The declarations are as follows:
<pre>
 - annotations

 // for marking up a text body: body (relation) annotation
 // hyphen is illegal, as it's common in text and ambiguous to section grammar

 % (discusses)
 = (depends on)
 * (is a special case of)
 > (has subject)
</pre>
the symbols + and - are reserved.

















