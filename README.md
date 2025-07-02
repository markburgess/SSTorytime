<!--
 SSTorytime - a ChiTek-i project by Mark Burgess

 Semantic Spacetime Story graph database library over postgresql (SSTorytime)
 This is an NLnet sponsored project, See https://nlnet.nl/project/SmartSemanticDataLookup/

-->

# SSTorytime

 Keywords, tags: Open Source Smart Graph Database API for Postgres, Go(lang) API, Explainability of Knowledge Representation

* This is a work in progress during 2025, as part of an [NLnet project](https://nlnet.nl/project/SmartSemanticDataLookup/). It's currently in an R&D phase, so comments are welcome but there is much to be done. This is not an RDF project. <br><br> [HOW **YOU** CAN CONTRIBUTE!](docs/howtocontribute.md)

* See these Medium articles for a conceptual introduction
* * [From cognition to understing](https://medium.com/@mark-burgess-oslo-mb/from-cognition-to-understanding-677e3b7485de): 
* * [The Shape of Knowledge](https://medium.com/@mark-burgess-oslo-mb/semantic-spacetime-1-the-shape-of-knowledge-86daced424a5)
* * [Why you still can’t find what you’re looking for…](https://medium.com/p/922d113177e7)

This project aims to turn intentionally created data (like written
notes or snippets cut and pasted into a file) into linked and
searchable knowledge maps, tracing the stories that we call reasoning,
and solving puzzles by connecting the dots between bits of information
you curate.

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/front.png 'Testing a web interface')

Knowledge maps are graph (network) structures that link together
events, things, and ideas into a web of relationships. They are great
ways to find out where processes start and stop, who is most important
in their execution, the provenance of transacted things or ideas, and
their rate of spreading, etc etc.  The pathways through such a web
form journeys, histories, or stories, planning itineraries or
processes, depending on your point of view.  We can interpret graphs
in many ways. Your imagination is the limit,

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
![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp6.png 'Testing a web interface')
![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/webapp1.png 'Testing a web interface')

*Imagine being able to take notes easily, work with them, and later be
able to "rummage around" in everything to understand what you were
thinking, and how it all fits together.  In other words, remaining in
control of what you see and ask, rather than handing over to a batch
job summary by an LLM text tool in which you get an answer `take it or leave it'.*

* [Getting started](docs/README.md)
* [A quick tutorial](docs/Tutorial.md)
* [The Mission and Approach](docs/approach.md)
* [Basics of Knowledge Engineering](docs/KnowledgeAndLearning.md)
* [SSTorytelling](docs/Storytelling.md)
* [N4L - Notes For Learning/Loading](docs/N4L.md)
* [searchN4L - preliminary search/testing tool](docs/searchN4L.md)
* [pathsolve - preliminary path solving tool](docs/pathsolve.md)
* [Related work and links](docs/outreach.md)
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

* [N4L](docs/N4L.md) - a standalone Unicode text based note taking language for jotting down notes in a way
        that can be parsed and loaded into a semantic database. 
        N = note, 4 = 4 semantic type, L = language
        N4L = notes for loading

* [N4L-db](docs/N4L.md) - a version of N4L that depends on the Golang package SSToryline in /pkg and uploads to a postgres database. This version is a compatible superset of N4L which prepares a database for searchN4L.

* [searchN4L](docs/searchN4L.md) - a simple and experimental command line tool for testing the graph database

* [notes](docs/notes.md) - a simple command line browser of notes in page view layout

* [pathsolve](docs/pathsolve.md) - a simple and experimental command line tool for testing the graph database

* [graph_report](docs/graph_report.md) - a simple and experimental command line tool for reporting on graph data, detecting loops, sources, sinks, etc, symmetrizing on different links and finding eigenvector centrality.

* [http_server](docs/Tutorial.md) - a prototype webserver providing the SSTorytime browsing service

* [API](docs/API.md) - An overview of the golang programmers API.

* [API_EXAMPLE_1](src/API_EXAMPLE_1.go) - a simple store and retrieve example of the graph database.


