
<!--
 SSTorytime - a ChiTek-i project by Mark Burgess

 Semantic Spacetime Story graph database library over postgresql (SSTorytime)
 This is an NLnet sponsored project, See https://nlnet.nl/project/SmartSemanticDataLookup/

-->


> [!NOTE]
> Commands now use \ in front as in other tools, e.g. `\from a1 \to s2` (see [searchN4L](docs/searchN4L.md))
> A Python interface is coming.

# SSTorytime

## Graphs are the language of spacetime process

Graphs are popular once again, but the technologies for dealing with them are clunky and designed by
technologists rather than scientists. This project makes working with graphs simple. 

<div>
<a href="https://www.amazon.com/dp/1797773704/ref=sr_1_fkmrnull_1?keywords=smart+space+time+burgess&qid=1551778123&s=gateway&sr=8-1-fkmrnull"><img width="180" align="right" src="https://github.com/markburgess/SSTorytime/blob/main/docs/figs/SmartSpaceTime.jpg"></a>

Graphs may be used:

<ul>
<li> As visualization of processes.
<li> As a map of space and time.
<li> As a map of a process, like Gant charts and path integrals.
<li> As computational device (a multi-matrix algebra representation). 
<li> - e.g. social networks with centralities and flow patterns, link weight as contact frequencies...
<li> As a distributed index over semantic relationships.
<li> And more ...
</ul>
</div>

## Take part in this Semantic Spacetime Project (SST)

 Keywords, tags: Open Source Smart Graph Database API for Postgres, Go(lang) API, Explainability of Knowledge Representation

* This is a work in progress during 2025, as part of an [NLnet project](https://nlnet.nl/project/SmartSemanticDataLookup/). It's currently in an R&D phase, so comments are welcome but there is much to be done. This is not an RDF project. <br><br> [TO-DO List and HOW **YOU** CAN CONTRIBUTE!](docs/howtocontribute.md)

* See these Medium articles for a conceptual introduction
* * [Why Semantic Spacetime (SST) is the answer to rescue property graphs](https://medium.com/@mark-burgess-oslo-mb/why-semantic-spacetime-sst-is-the-answer-to-rescue-property-graphs-2c004fe705b2)
* * [From cognition to understanding](https://medium.com/@mark-burgess-oslo-mb/from-cognition-to-understanding-677e3b7485de): 
* * [Searching in Graphs, Artificial Reasoning, and Quantum Loop Corrections with Semantics Spacetime](https://medium.com/@mark-burgess-oslo-mb/searching-in-graphs-artificial-reasoning-and-quantum-loop-corrections-with-semantics-spacetime-ea8df54ba1c5)
* * [The Shape of Knowledge part 1](https://medium.com/@mark-burgess-oslo-mb/semantic-spacetime-1-the-shape-of-knowledge-86daced424a5)
* * [The Shape of Knowledge part 2](https://medium.com/@mark-burgess-oslo-mb/semantic-spacetime-2-why-you-still-cant-find-what-you-re-looking-for-922d113177e7)
* * [Why are we so bad at knowledge graphs?](https://medium.com/@mark-burgess-oslo-mb/why-are-we-so-bad-at-knowledge-graphs-55be5aba6df5)
* * [The Role of Intent and Context Knowledge Graphs With Cognitive Agents](https://medium.com/@mark-burgess-oslo-mb/the-role-of-intent-and-context-knowledge-graphs-with-cognitive-agents-fb45d8dfb34d)
* * [Designing Nodes and Arrows in Knowledge Graphs with Semantic Spacetime](https://medium.com/@mark-burgess-oslo-mb/designing-nodes-and-arrows-in-knowledge-graphs-with-semantic-spacetime-0992b9cae595)
* * [Avoiding the Ontology Trap: How biotech shows us how to link knowledge spaces](https://medium.com/@mark-burgess-oslo-mb/avoiding-the-ontology-trap-how-biotech-shows-us-how-to-link-knowledge-spaces-654bcbb9122a)
* * [Using Knowledge Maps for Learning Comprehension](https://mark-burgess-oslo-mb.medium.com/using-knowledge-maps-for-learning-comprehension-15e162a251cd)

This project aims to turn intentionally created data (like written
notes or snippets cut and pasted into a file) into linked and
searchable knowledge maps, tracing the stories that we call reasoning,
and solving puzzles by connecting the dots between bits of information
you curate.

## For instance...

You might want to look up something ad hoc:

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/front.png 'Testing a web interface')

You might want to be prompted for reminders:

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/reminder.png 'Testing a web interface')

You might be surveying the breadth of your knowledge:

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/TOC.png 'Testing a web interface')

Or just curious about something:

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/sailing.png 'Testing a web interface')

## The concept

Knowledge maps are graph (network) structures that link together
events, things, and ideas into a web of relationships. They have enjoyed renewed
interest in recent years, because of "AI" -- but, since the 1990s, people have largely been
doing them wrong--trying to model things and ideas instead of processes.

Graphs are great ways to find out where processes start and stop, who
is most important in their execution, the provenance of transacted
things or ideas, and their rate of spreading, etc etc.  The pathways
through such a web form journeys, histories, or stories, planning
itineraries or processes, depending on your point of view.  We can
interpret graphs in many ways. Your imagination is the limit,

Stories are one of the most important forms of information, whether they
describe happenings, calculations, tales of provenance, system audits... Stories
underpin everything that happens.

Getting data into story form isn't as easy as it sounds, so we start
by introducing a simple language "N4L" to make data entry as painless
as possible.  Then we add tools for browsing, visializing, analysing
the resulting graph, solving for paths, and divining storylines
through the data. The aim is to support human learning, and to assist
human perception--though the results may be used together with "AI" in the future.
Finally, there will be an API for programmers to incorporate these methods
into their own explorations, either in Go or in Python. As a sort of "better, faster Python",
Go is recommended for power scripting.

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/graph.png 'Testing a web interface')

Note-taking may be an intuitive but semi-formal approach to
getting facts for reasoning, for knowledge capture, querying, and
dissemination of individual thinking easy, for humans and general
use. (AI can only capture knowledge from humans, so even if we want to
use AI, we'd better get the knowledge representations right.)  Whether
we are analysing forensic evidence, looking for criminal behaviour,
learning a foreign language, or taking notes in school for an exam.

Today, computer tools ask people to enter data through APIs by programming,
or by typing into special forms that are stressful and unnatural. We can do better,
just as we can do better at retrieving the information and searching it.

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/last.png 'Testing a web interface')
![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp1.png 'Testing a web interface')
![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp6.png 'Testing a web interface')


*Imagine being able to take notes easily, work with them, and later be
able to "rummage around" in everything to understand what you were
thinking, and how it all fits together.  In other words, remaining in
control of what you see and ask, rather than handing over to a batch
job summary by an LLM text tool in which you get an answer `take it or leave it'.*

* [Getting started](docs/README.md)
* [The Mission of SSTorytelling](docs/Storytelling.md)
* [A quick tutorial](docs/Tutorial.md)
* [An example / case study](docs/example.md)
* [Basics of Knowledge Engineering](docs/KnowledgeAndLearning.md)
* [How does context work?](docs/howdoescontextwork.md)
* [N4L - Notes For Learning/Loading](docs/N4L.md)
* [searchN4L - preliminary search/testing tool](docs/searchN4L.md)
* [pathsolve - preliminary path solving tool](docs/pathsolve.md)
* [Related work and links](docs/outreach.md)
* [API](docs/API.md) - An overview of the golang programmers API.
* [FAQ](docs/FAQ.md)

## History

The roots of this project go back almost 20 years for me, when I was working in configuration
management (with the CFEngine project) and realized that the main problem there was not
fixing machinery but rather understanding the monster you've created! Knowledge Management
was built into CFEngine 3, but later removed again when `the market wasn't ready'. Over those
20 years, I've studied and learned how to approach the problem in better ways. I've implemented
the concepts using a variety of technologies, including my own. In this latest version, I'm
combining those lessons to make a version that builds on standard Postgres.


![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp4.png 'Testing a web interface')
![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/Alice.png 'Testing a web interface')

While the actual graphs in knowledge might not turn out to be anything like the beautiful
fabrications one sees in fancy online graph engines, the information is still important.
Most knowledge is patchy and disconnected.
Knowledge capture requires tools for collecting factual notes, data
relationships, and structures for representing and organizing them, so
that they can be found easily. Many mistakes have been made around
this in the past, trying to force discipline onto people at the wrong
moment and neglecting to do so when it matters. As a lifelong teacher
of ideas, I've studied this problem and think we can do better than
what we have today.

One of the goals of this project is to better understand what we call "reasoning".
One used to think of reasoning, philosophically, as logical argumentation. As computers
entered society we replaced this idea with actual first order logic. But, if you ask
a teacher (and if we've learned anything from the Artificial Intelligence journey)
then we realize that the way humans arrive at conclusions has a more complicated
relationship to logic. We first decide emotionally, narrowly or expansively depending
on our context, and then we try to formulate a "logical" story to support that.
This is why we strive to study the role of stories in learning and understanding for this project.

## The tools

The tool-set consistent of several components, starting with:

* [N4L](docs/N4L.md) - The N4L compiler (This is now merged with N4L-db)

* [searchN4L](docs/searchN4L.md) - a simple and experimental command line tool for testing the graph database

* [text2N4L](docs/text2N4L.md) - scan a text file and turn it into a set of notes in N4L file for further editing

* [removeN4L](docs/removeN4L.md) - remove an uploaded chapter from the database

* [notes](docs/notes.md) - a simple command line browser of notes in page view layout

* [pathsolve](docs/pathsolve.md) - a simple and experimental command line tool for testing the graph database

* [graph_report](docs/graph_report.md) - a simple and experimental command line tool for reporting on graph data, detecting loops, sources, sinks, etc, symmetrizing on different links and finding eigenvector centrality.

* [http_server](docs/Tutorial.md) - a prototype webserver providing the SSTorytime browsing service

* [API_EXAMPLE_1](src/API_EXAMPLE_1.go) - a simple store and retrieve example of the graph database.

* [API_EXAMPLE_2](src/API_EXAMPLE_2.go) - multi/hyperlink example, joining several nodes through a central hub.

* [API_EXAMPLE_3](src/API_EXAMPLE_3.go) - a maze solving example, showing higher functions.

* [API_EXAMPLE_4](src/API_EXAMPLE_4.go) - a path solving example, with loop corrections (quantum style).

* [python_integration_example.py](src/python_integration_example.py) - a basic Python example

* [SSTorytime.py](src/SSTorytime.py) - Includable Python interface for SSTorytime, basic functions (TBD)



