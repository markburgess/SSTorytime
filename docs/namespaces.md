
# Namespace concept in SSTorytime: "Many Worlds" Knowledge Graphs

The principles for SSTorytime are that knowledge is a personal
(individual) creation, a personal journey--but sharing is caring, and we can be inspired by another's 
work (see diagram).

![Alpha interface](https://github.com/markburgess/SSTorytime/blob/main/docs/figs/namespaces.png 'namespaces')

## The principles

* Knowledge is personal, but it is designed for sharing. Access control may be applied to it, but the requirements for access control are as yet unclear.

* N4L is a tool for keeping personal notes. Notes can be shared as text or as an http web service point.

* SQL tables from postgres can be shared by a sharing ring (work with Andras Gerlits) based on permission models already handled by postgres, with TLS privacy. 

* Each user's interaction with the system is individual, so progress tracking and configuration settings need to be personalized. 

* * Each user has a personal progress tracker in every http-db point.
* * Each http-db point has its own private tables of extended node or link information, by using local NPtrs as indexes. This needn't have anything to do with the graph, except as a linked directory service.

* Learning benefits from exposure to alternative viewpoints (even opposites), so there is a need to relate different users' notes in some way. This is primariy a configuration issue: an association `USER-N(db,NPtr)`.

* Different style schemes for different areas are also desirable, based on colours and fonts, so there must be individual condigurable "skins". 

*The tech around discriminating user spaces and login issues will not be considered in the first iteration of the technology as these are trivial but complicating. Rather, it's important to develop the primary issues that concern learning so that users can get to work as quickly as possible.*

