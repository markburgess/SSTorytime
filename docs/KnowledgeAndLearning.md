

# Knowledge and learning

You might think you know how to acquire knowledge. You learn stuff---how hard could it be?
You take a course, or read a book. But, of course, there is much more to it than that.

Knowing isn't just remembering. You might remember where the fire
extinguisher is, but when there's a fire, do you know what to do? You
might have seen people knit a sweater, but could you do it?  We need
to use knowledge to actually `know' it. It's like food. You can collect it,
but you don't know what it is until you actually taste it.

## Wikis and why documentation is so bad

There's an old joke: that Wikis are places where knowledge goes to
die. Well meaning individuals may invest hours of work to write
something down for others. But no one forms an intimate relationship
to what has been written, it is not knowledge. It's just a graveyard
of bits and bytes that means nothing to anyone except the person who
wrote it. The same is true of any book, even those written by a well
meaning author. Some books might be vanity projects, not meant to be
embraced, but teaching books always try to reach an audience in some
way. Success or failure depends on building a rapport with a certain reader.
You can't reach everyone, so you aim for a few.

## What is knowing, actually?

See also [the introductory article](https://medium.com/@mark-burgess-oslo-mb/from-cognition-to-understanding-677e3b7485de) on this topic.

Knowledge is more than memory. You can `learn' a page of text by
heart and still have no idea how to use it. As long as it remains a
lump of data, in your head or simply on paper, in a computer, or and the back of your hand,
it's of little use to you. Knowledge comes from knowing things deeply---by
forming a relationship to material. You know something when you know it like a friend.
You won't have go and look up details because access will be integrated into your
conscious experience and awareness of environments you know about. This is what it
means to have knowledge at your fingertips. We are designed to use our hands
and fingers. 

Writing stuff down is useless if no one reads it. This is why Wikis,
knowledge bases, and expert systems often fail. Most Wikis are
intended as 1:N communication.  Wikipedia can succeed due to scale:
it's N:N for large N, which means the information is passed through a
human brain frequently.

https://medium.com/@mark-burgess-oslo-mb/the-failure-of-knowledge-management-5d97bb748fc3

This is one reason why current AI language models that seem to `know things'
are in fact as clueless about their subject matter as you are the day after reading
their results.

## Limits on learning

Even simply cramming facts by brute force memorization is hard. There
is a limit to the scaling of learning. Even machine learning can't
lead to endless improvements, because the cost-benefit of finding the
right data and automating learning rapidly becomes untenable. Adding
resources to capture every last variable takes too long and costs too
much energy. It's most effective for common short-term experiential
learning, like pattern recognition in the immediate environment of
sensory pre-processing. After that, a Monte Carlo approach to
guesswork or ``prediction'' would be much more energy efficient.

Of course, one might say that humans have developed vast learning
resources--far greater perhaps than could be accounted for by a strict
cost-benefit analysis.  Doesn't this indicate that learning still has
a benefit?  But evolution doesn't direct the course of change on an
actionable timescale, it merely expresses a lukewarm approval over
benefits still to be spat upon by competitive opportunism.  Adding new
data will at some point plateau in terms of semantics or meaningful
features.  What remains is quite battle tested over a wide range of
contexts but might be only very rarely useful.
What is rarely spoken about in machine learning is the importance of
forgetting data: of post selection to keep only the most useful
pattern memories.

We also need to expose contextual knowledge.  Brute force learning
also doesn't discriminate by context: there is a far greater number of
seemingly-infinite combinatoric contexts for relevant data. Each will
have its own learning plateau associated with its characteristic scales
(particularly timescales). In machine learning, transformer
architectures have discovered this, but they are still based mainly on
recall.


## Even Spock fell afoul of logic

Depending on your background in sciences or humanities, you will almost
certainly think very differently about how meaning arises. Those of us
in the natural sciences are trained to think "logically" or "rationally".
Those in humanities are apt to draw analogies and play loose and fast with
meanings. Both of these habits have their usage, but they are only strategies
for inference. Neither is right or wrong, and both can be misunderstood.

If we aim to write about universal truth for all humanity, we have
a communication problem of great delicacy to solve.
If, on the other hand, our goal in modelling is to remind ourselves
of how we think about something, to develop and evolve our own meaning,
then we have no responsibility to be accountable to others in our choice of
strategy. Indeed, we should be fairly suspicious of someone telling us how we *must*
do it.

Be clear: this is **not** an argument that right and wrong do not exist.
It's a statement that **language** is a utility that can and is used in
various ways. If we are flexible, we can learn from that. If we are inflexible,
we will simply be confused about the difference between intent and truth.

When we come to **the hard problem of context**, there are many more pitfalls
to modelling, so it's best not to make things harder than they need to be in the beginning.
The lesson, I believe, as a pedagog is to not allow perfect be a barrier to progress.

*You can and *should* revisit and modify your choices over and over again,
because it's exactly the process that contributes to learning, not the
putting of things in boxes for an archive you never revisit.*

## Examples and pitfalls in modelling

Not all relation types are as obvious as we may think:
Look at the example of friendship, which has inverse like this:
<pre>

 + has friend (fr) - is a friend of (isfriend)

</pre>
What type is this? Is friendship a mutual property (friends with) or is it a
personal judgement that might not be receiprocated (considers a friend)?
If we don't assume mutual friendship, we have a more powerful abiility to
encode individual beliefs:
<pre>
- properties   # NOT similarity/proximity

 + has friend (fr) - is a friend of (isfriend)

</pre>
If we want to enocde mutual friendship, we simply declare the relation
both ways, but we don't have to assume that:
<pre>

-friends

 John (wrote) Mary had a little lamb

 Mary (fr) Little Lamb

 Little Lamb (fr) Shawn
 Shawn Little (fr) Lamb

 Shawn (is a friend of) Team Wallace and Gromit  // use short/long as you think of it

 Team Wallace and Gromit (has member) Wallace
           "             (memb) Gromit

</pre>
If we parse this, we now see
<pre>
- including search pathway STtype Express -> has friend
   including inverse meaning is a friend of
    - row/col key [ 0 / 6 ] Shawn Little
    - row/col key [ 1 / 6 ] Little Lamb
    - row/col key [ 2 / 6 ] Mary
    - row/col key [ 3 / 6 ] Team Wallace and Gromit
    - row/col key [ 4 / 6 ] Shawn
    - row/col key [ 5 / 6 ] Lamb

 directed adjacency sub-matrix ...

        Shawn Little .. (   0.0   0.0   0.0   0.0   0.0   1.0)
         Little Lamb .. (   0.0   0.0   0.0   0.0   1.0   0.0)
                Mary .. (   0.0   1.0   0.0   0.0   0.0   0.0)
     Team Wallace an .. (   0.0   0.0   0.0   0.0   0.0   0.0)
               Shawn .. (   0.0   0.0   0.0   1.0   0.0   0.0)
                Lamb .. (   0.0   0.0   0.0   0.0   0.0   0.0)

 undirected adjacency sub-matrix ...

        Shawn Little .. (   0.0   0.0   0.0   0.0   0.0   1.0)
         Little Lamb .. (   0.0   0.0   1.0   0.0   1.0   0.0)
                Mary .. (   0.0   1.0   0.0   0.0   0.0   0.0)
     Team Wallace an .. (   0.0   0.0   0.0   0.0   1.0   0.0)
               Shawn .. (   0.0   1.0   0.0   1.0   0.0   0.0)
                Lamb .. (   1.0   0.0   0.0   0.0   0.0   0.0)

 Eigenvector centrality score for symmetrized graph ...

        Shawn Little .. (   0.1)
         Little Lamb .. (   1.0)
                Mary .. (   0.6)
     Team Wallace an .. (   0.6)
               Shawn .. (   1.0)
                Lamb .. (   0.1)

</pre>
By computing both directed and undirected matrices automatically, N4L allows us to
compare the effects of this modelling difference. In general, it's best not to assume
mutual relationships, as these can easily be symmetrized but undoing mutuality is hard.

When we say that A follows B, this may apply to things or actions.
* Space travel came after aircraft. 
* Shopping is done after work.
* Hammering is done after assembly.
Order applies to both processes and objects.

We could imagine a supply-chain worker noting:
<pre>
 delivery 123 (damaged) 2 boxes
</pre>
It's a fair thing to write in a moment of unexpected pressure. But which of the
four relations is this? That's the same as asking: what could we use this note
for later? The problem with it is that it's ambiguous.

The left hand side "delivery 123" is clear enough. It represents some shipment
and we could embellish this description like this
<pre>
 delivery 123 (contains) shoes
     "        (came from) Italy
     "        (received by) shift crew 12
</pre>
and so on. So no problem here. The relation "damaged" becomes an issue however
because it's referring to the condition or state of the delivery. 
A more flexible approach would be to rewrite this as
<pre>
 delivery 123 (condition) 2 boxes damaged
</pre>
because now
* condition is a generic and reusable relation, which is a propery attribute (type 3) of the delivery
* "2 boxes damaged" is an event that can be explained easily
For instance, now we can explain the event further:
<pre>
  2 boxes damaged (condition) water damage
         "        (contains) red stiletto box 1445
         "        (contains) black stiletto box 1446
</pre>





### Example: The "is a" fallacy

During the OO-movement to sanctify Object Orientation as a software modelling approach, 
Object Orientation rubber stamped
the idea that objects, i.e. "things" (rather than processes or activities) are the most important concept in a model, 
leaving *processes* asking: what am I 
then? (The answer was usually that processes should be thought of as methods that affect
objects, which is extremely limiting.)
Classification of objects into types was the goal of OO, because this is a way to simply
map ideas into first order logic, and that makes programming easy to understand.
Alas, squeezing processes into this isn't always easy.
The answer commonly associated with this was to use the "is a" or "is an instance of" relation
as the way of thinking about things.
<pre>
Object X is an instance of a class Square
A Square is a special case (inheriting) the class of Rectangle
etc.
</pre>
The trouble with this idea is that it attempts to assert an *static* or *invariant* truth
about the role of something (the square). But squares, indeed any properties or
roles, are typically context dependent. We use the same concept in different ways.
<pre>
In DIY: A hammer is a tool.
In music: A hammer is a musical instrument
In DIY: a drill is a tool for making holes.
In operations, a drill is a practice episode.
</pre>
If we insist of having different types for each of these cases (a type polymorphic approach),
we push the responsibilty of the technology back onto the person using it. Technology
is supposed to work for humans, not the other way around.

The example above of damaged delivery  is a good example of how this becomes
problematic. Suppose we introduce an object for a delivery, is that
"Delivery" or "Shoes"? Should we have a separate object for "Damaged delivery" or is
damage an attribute of the object. What could it mean? how would we explain it?

The virtue of a semantic language is that we never have to shoe-horn
(no pun intended) an idea into a rigid box, as we do when we try to
lock down data types. This is an affectation of logical reasoning,
but logic is highly restrictive (on purpose, as a matter of design).
That makes it precise, but also extremely fragile to variability.

### Example: Belonging

Some relationships can be tricky to fathom. The semantics of ownership,
for example, are not completely unambiguous. Suppose you want to say

<pre>
The bracelet "belongs to" Martin 
</pre> 

Is the bracelet a property of Martin or a part of him?  As an object,
we might choose to make this a part the "extended space of
martin". There is no right answer. You can choose what works for you.
The difference between the two is how they are searched.  If we
interpret the bracelet as "a part of" Martin then we can also say that
the bracelet contains a diamond and thus the diamond is also a part of
Martin, because "part of" is a transitive relationship. But if we say
that the bracelet is just something that characterizes him, it's not
clear that that is transitive because a bracelet may be characterized
by being golden but this does not imply that Martin is golden!

You might make the wrong choices about things initially, but it's easy to
change your decision because the definition of the relationship is
made independently of all the data where you use it. You'll figure out
the bugs in your wordings as you go, and it's precisely this reworking
that is learning.

The usefulness
of a language interface becomes clear now. It's much easier to edit your notes than to maintain
a database.

### Example: space or time?

Consider the use of a word in a sentence.
<pre>
It was a happy accident (???) happy
</pre>
What can we say about the relationship between these two?
* You could say that it is a property of the string (PROPERTY/ATTRIBUTE)
* Is it merely a part of the sentence (CONTAINS/PART OF).
* Is it a causal component that significantly influences the meaning? (LEADSTO/CAUSES)
Probably no one would think the left and the right hand side were similar to one another (SIMILAR/NEAR).

To say that happy is simply a property or attribute of the longer phrase is true, but it doesn't tell us whether
it contributes significantly to the meaning. To say that the longer phrase contains the word is also true, but
the same criticism applies. On the other hand, to say that happy leads to happy accidents is
unlikley though it could depend on the context.

If you're still trying to make an ontology of absolute truth, in the logical sense, you should
take a step back and rethink your model. When modelling, we fall into these traps because those of us
with mathematical background have been
taught to apply the discpline of logic when formulating structure. Philosophers and writers, on the
other hand, are taught to throw everything up in the air and consider every possibliity, none more
fundamental than the next. This can be liberating and infuriating in equal measure.

The important point is this: you can apply all of these possibilities and you would not wrong,
except in a specific context. So why not? the hard part of modelling should be limited to
understanding context. We should not try to limit the usage of language.

The fallacy of the logical truth/falsity approach is that meanings are not mutually exlcusive
ontological catgories, they are superpositions of meanings that remain in play until something
makes a definite selection. This is an evolutionary strategy (some might say it's a quantum-like
strategy--indeed, the mathematics of quantum `superposition and collapse' is a representation of
this kind of parallel hedging of bets. It's what software engineers sometimes call `lazy evaluation').




## Context: where, when, and strategy. A Scene Description Language?

Consciousness is not the hard problem of knowledge: context is.

When we describe the context in which we learned something we are describing the key
for finding it later.

We use the term context in several ways.

* **Strategic context***: When taking notes, we use headers and headlines to descibe a context of "aboutness":
keywords, perhaps in a phrase, that describe what we'll find in the passage.
This is a strategioc use of context. It says: my strategy was to put this here so you would
find it, if you looked up these keywords.

* **Scene description**: thanks to our senses, the context in which we think of something
may depend on a complex web of happenings (of causality) that's ultimately summarized by a sort of 
snapshot of *the state the scene*
that we think of as context.
The fullest possible description of context is thus a background story for the moment.
Think of forensic investigators solving a mystery. They assemble context as factual descriptors,
causal motives and how all of the above come together. 

We might imagine a Scene Description Language
to be the ultimate goal of context. Our hypothesis here is that this imaginary Scene Description Language
is just a form of the semantic spacetime that we are developing under N4L.

In IT knowledge systems, the concept of a formal "ontology" has found popularity.
This is the idea that there is a kind of spanning tree of correct categorical meanings
for knowledge. Unfortunately, this idea has been shown to be flawed, if not merely false,
many times. There are always many possible spanning trees, or interpretations that classify
meaning. When we tell a story we might start at the beginning and follow chronology, or we might start
with the outcome and work backwards. We might hop back and forth atemporally in order
to discuss the relevance of the parts to the whole. The context is not a spanning tree.

By following trails of thought, we are assembling a trail of prerequisites that
our brains, evolved for navigation in a landscape, can understand. If we 
fall prey to the conceits of perfect logic we will tend to over-constrain information
so that it becomes impossible to find without the precise criteria used to store
it. This probably the most common mistake in using tools like RDF with OWL (the web ontology language)
as these as based on first order logic.

This ends up working against us, because we need to reproduce the precise lookup key
to find what we're looking for---and we might not even be clear about what we're
looking for! The main benefit of memory is creative composition of ideas, by "mixing it up".

Think of learning a language. You can quickly the vocabulary but you
can't recall it on demand. It takes several years of constant use to
develop the recall methods.


In this project, we look at trying to separate the processes that
manage different scales of learning to establish efficient recall structures
after post-selection.

## Completeness and scale

When parsing new information, word by word, we have no idea what a
complete sentence is going to mean or be about.  We have to wait for
the finish line to go back over the full expression.  We are
constantly buffering information in chunks to ascribe meaning to it.
The same applies to a paragraph. One sentence may say little, but a
full paragraph can make a point.  We can never be certain when a stream
of information is finished or complete in some sense. There can be diversions
to order, parentheses that form the grammar of the telling of the story, and so on.

Think about trying to learn a language. First we start by repetition
of words and phrases. We can use our basic cognitive skills to cram
words like Random Access Memory.  This leaves us only with parroting
skills. We don't end up with any knowledge of the grammar unless we go
back over many sentences and begin to search for patterns.  We have to
work analytically to learn more. We have to be active ourselves.  It's
not a job we can outsource to someone else. If you want to learn how
to make the tea and polish shoes, don't hire a butler. If you want to
understand how to arrange and organize learn history, don't delegate
to an assistant (especially and AI chat).

What we miss when trying to cram knowledge is a sense of `intent', i.e.
connecting what we're actually trying to do to what we can remember.
You miss what you actually want to say, and how to compose that
freely, without being constrained by a series of logical barriers or
brute for memory acts.
A single experience of being able to use a word in practice will burn itself
into memory more effectively than a hundred hours of cramming.

Some of the answer to this lies in grammar: the rules of composition, but
it's more than that. Ideation is the process of finding words that
you've never seen together before and placing that new combination
and its meaning into a new context.


The old jibe `those who cannot do teach' is about this mismatch
between crammed knowledge and applied knowledge. Of course, the best
teachers have already done things and try to pass on that experience of
transmuting facts into action, though students won't necessarily have
the context or experience to follow them. A good teacher creates that
context. This is why we are asked to do exercises, not just reading.

As a student, my own first step in processing the new information
was to rewrite it in my own words; create my own story from the one
given to me. How would I explain it to someone else? I can't just be actor
memorizing lines, I need to have tried and failed.

So how can *assist* humans to arrive at that understanding, especially if they
are challenged in memory ability or have cognitive difficulties?
They need to recall a basic vocabulary of things and activities (nouns and verbs).
When they are missing a word, how can we prompt something appropriate?

The bottom line is: knowledge is like knowing a friend. It's the same brain,
and the same skill we use for both. You start by tipping your hat when you
recognize someone, then you might say good morning. You'll stop for a chat,
maybe have a coffee and have your first date, but little of this
will stay with you unless you start to care. Only when you long for their
return can you cay that you know something about them.

## Phases

* You take notes, in little patches of things to remember, hoping to trigger a larger memory.

* When you have enough notes, you start to put them in some kind of order, so that things that belong
together can be found together. Neurons that fire together, wire together! 

* We find out quickly that there can't be a box for everything. If we try to be too precise, we'd
package everything in its own box, because every situation is unique. Instead, we realize that
knowledge is about we approximate ideas by looking for what's similar rather than what's different.
There's a time for collation and a time for discrimination. There's a Chinese proverb that things group into
categories, but people divide into groups. It suggests that people are contentious, while everything else
can be treated with a more welcoming approach. Knowledge is also about being open minded!

* Writing down errors and misunderstandings, and resolving them helps us to overcome barriers
to seeing patterns. Some of the best knowledge comes from overcoming a misunderstand---from asking the right question.

* It takes work (revising and rewriting, re-ordering, tidying, copying and eliminating, etc)
to become familiar with materials. It's the start of a relationship, to making a new friend.

* Rhymes, slogans, and catch phrases are useful for remembering ideas, because they draw on
our affinity for pattern and motor skills.

## The goal

We're looking to be able to recall information easily, and to understand what it means by building
stories around the nuggets we collect. 

We already have plain text search, as well as the trusty "index" to
help us with random recall. An index offers a way into a body of tidy
knowledge from a simple term. Good indices are made by people who try to think of
all the things people might use to look for information, 
not just the words they used in their own writing. We can do the same.

In a way, N4L is just a fancy index generator.
How could a more sophisticated
data structure help us to find what we need?  In the past people have
tried before with taxonomies, Topic Maps, and Resource Description
Frameworks based on technical Ontologies. These are all useful, but
only if users find their own methods for using them.  It's possible to
abuse them because they don't guide users towards good habits: have no
constraints.

Bureaucracy itself is an early attempt to manage knowledge. If we put things into an archive,
alphabetically, or by subject, etc, this is supposed to make it easier to find.
The mistake bureaucrats make is the same one logicians make: they force the
work of deciding where to put something onto the user
at the wrong moment, i.e. at a time when they are most familiar with it. Finding
it later, in a completely different state of mind, is a much more difficult task. 
Forcing users to work too much at the beginning to make everything fit into a perfect box
of someone else's making acts as a disincentive rather than an encouragement. So we want to take away the barriers
of documentation, letting users write things down in scraps and notes as they like. Then we want to
incentivize them to come back and make that knowledge a work of beauty in their own minds---something they
can fall in love with.


## Alphabets and simple classifications

Alphabets are collections of standard symbols. The Greek alphabet (from the first
two symbols A,B) were a turning point for script.
What an alphabet does is to offer up a smallish `menu' of items for
reconstructing intentional words. The words have meanings, and we
remember the sounds well.  

We can try to do the same thing for meaning. By creating a smaller palette
of words and phrases (as one does in domain specific languages, or specialized
jargons) we simplify the amount of language we need to remember. This works
well until it doesn't: at some point, we all run out of words and need new ones.
So expressivity makes extensibility important. If we are too restrictive,
it works against meaning. If we have too many meanings, nothing has a certain meaning.
There's a balance to be found between these cases.

Words are also associated with sounds and other sensory experiences. 
Seeing the sounds (unless
you're deaf) is possible with a phonetic encoding, and this is what an
alphabet enables with only a small number of symbols. In symbolic
languages, like Chinese and ancient Egyptian, symbolic were also tied
visually rather than phonetically to meanings, i.e. independently of
the sounds.  Later, the need for phonetics was overcome by using
phonetic transcriptions based on a small number of well known
symbols. They had to invent a surrogate alphabet for names and foreign
words, in particular.

The way we use language is creative and diverse. There are no rules
when you are taking notes.


## Classification and misclassification

As we write down our notes, we can't possibly ensure that every item
is correctly classified into the right chapter, or in the right
order. We still need to be able to find that misplaced book in the
library, or a missing person amongst the busy goings-on of the
world. It's an unavoidable issue, but it can be remedied by regular
tidying, rehearsing procedures, maintenance of materials, and so on.
It can be remedied, but never solved once and for all. As soon as we
stop, the knowledge dies. What we realize, then, is that the need
for `searching' cannot be eliminated as long as knowledge is alive.
Indexing is the tool that makes that possible, alongside the storyline
of tables of contents and order.

Over-constraint, or over-doing classification leads to ideas like data normalization,
or the normal forms that were invented for databases when humans were supposed
to type in the data by hand. Consistency errors were an issue, so a scheme of normalized
forms was invented to reduce the possibility of error.

If we over-constrain data by placing them perfectly so that they can only be found with
the precise (highly precise) key, then we reduce the chances of finding the information
to only perfectly asked questions. That's unhelpful. The goal is to make things easy to find,
and to encourage browsing and stumbling across possibly relevant connections.

