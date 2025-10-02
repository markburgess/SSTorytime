
# How you can contribute to SSTorytime

SSTorytime and the Semantic Spacetime model of knowledge is a unique project based on leading edge research. You'll recognize some of the ideas that are old, but there are some new concepts to know too!

* The best way to help is to try it!

The hardest part at the moment is installing the software. It's easiest of you have a Linux computer. Just follow the destructions on [getting started](README.md). It's too early to be packaging the software for Windows or MacIntosh at the moment, but you can still try writing notes in the N4L language, which doesn't require any tools!

* Are you a teacher? Get your students to try it, make a class project!

* Use it as much as possible! Taking notes is about enhancing your **human** abilities rather than replacing yourself with an AI chat! When you choose your own words, you understand more powerfully than when someone puts words in your mouth! Start small and expand as you feel inspired.

* Get your friends and kids involved. Anyone can use it. It's ideal for school projects. You can write in any language.

* Even before you try to set up a database, you can use the note-taking language N4L to get started.
[See the tutorial!](Tutorial.md)

* Please *read* about the project to make sure you understand how it is special, don't assume it's like something else you already know!

* Create examples of your own notes to try it out and to give everyone ideas. 

* If you want to think about how to visualize information in a knowledge base in new and exciting way, then give it a try! If you have specific expertise in web programming, get in touch!

* As the documentation becomes more complete, you'll be able to write your own programs to query data. Later there will be a Python interface too.


## The TO-DO list

There are many things left to implement in SST. If you want to make
suggestions, please get in touch and offer some feedback.

* The user interface is only provisional, so there are many things to do to improve style and colouration.
The focus has been on obtaining a "responsive" design that will work on all devices, large and small. This is obviously both a challenge and a matter of individual taste.

* * User-selected colours and styles

* * Using Images and other Media Types more seamlessly as nodes in a graph is desirable. Presently, one can give a URL to an image or resource, but the browsing experience is incomplete.

* * Reading text out loud (in multiple languages) is desirable.

* Basic solving for paths in a graph has been implemented in a generic way, but there cases (such as for criminal investigation or research trial mapping) where more advanced solutions involving types and weights are needed. This remains to be done, because some effort first has to be made to understand the scaling and speed of lookup for different kinds of data. Path searching is computationally intensive and lends itself to a batch-style job (e.g. `graph_report`) rather than an interactive search.
