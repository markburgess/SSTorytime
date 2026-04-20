---
name: SSTorytime
description: read files of N4L notes and create a semantic knowledge store with graph methods based on semantic spacetime and promise theory
allowed-tools: bash, read
---

# SSTorytime

A semantic knowledge store for humans and AI -- compile notes and turn them into a searchable graph

## Prerequisites

- go programming language
- latest postgres
- postgres contrib package
- Full functionality requites the webserver http_server to be running

## Install

If Linux
- install postgres and postgres contrib packages
- install go language
- install SSTorytime respository
```bash
git clone git@github.com:markburgess/SSTorytime.git
make
```


## Concepts
- A graph consists of nodes joined by arrows
- Arrows are directed edges in the graph
- Every arrow has an inverse, pointing in the opposite direction
- There are four type of arrow:
  - LT (leads to),
  - CN (contains),
  - EP (express property),
  - NR (near to)
- Each type of arrow may have many different names, which describe exact meaning
- Search commands are denoted by \command, e.g. \path, \sequence, \chapter


## Procedure
- Step 1: A user writes notes in N4L. The compiler validates file.N4L based on arrows defined in SSTconfig/ directory
- Step 2: User compiles notes with the bin/N4L tool and they become a graph.
- Step 3: Search the graph, edit notes, and repeat
- Advanced: the user can define new arrows by editing files in SSTconfig/



## Tools
- Tools are compiled into src/bin of the SSTorytime repositor

- To validate N4L file of notes
```bash
src/bin/N4L FILE.N4L
```
 - To upload N4L notes:
```bash
- src/bin/N4L -u *.N4L
```

- To wipe entire existing database:
```bash
- src/bin/N4L -wipe
```

- To wipe all data and upload files to replace existing database:

```bash
- src/bin/N4L -wipe -u *.N4L
```

- To remove a single chapter from the database
```bash
- src/bin/removeN4L

```

- To convert an ASCII text file into a single chapter file of notes:
```bash
- src/bin/text2N4L FILE.txt
```

- To start a web interface to the database
```bash
- src/bin/http_server
```

- To obtain a command line text report about graph data.
```bash
- src/bin/graph_report -chapter CHAPTER
```



## Usage

- A user writes notes in files called SOMETHING.N4L
- Use the compiler src/bin/N4L SOMETHING.N4L to validate
- Upload data to search and visualize connections

## Example data

To start with example data from the package
- cd examples; make

To add a single file of notes

## N4L usage

- Create a text file containing lines with fields: first items (arrow) second item (arrow) third item, etc
- Lines may be labelled with @token to refer later as $token.1, $token.2, etc
- The ditto character " may be used to copy the previous line's field
- The $PREV.1, $PREV.2 may also be used to refer to the previous line's fields

## Search Usage

To search existing knowledge, post a string, with form key "name" to the /searchN4L endpoint of http_server

A table of contents listing chapters and search contexts can be obtained with command:
- \toc - list full contents
- \toc PATTERN - list chapters matching PATTERN


- There are several independent kinds of search:

  - Random access lookup, match a search pattern

    - unicode text match, using postgres ts_vector or SQL string search
    - Submitting a quoted string to preserve spaces in a search pattern
    - Strings that may contain unicode accents can enclosed in parentheses () to unaccent characters, e.g. "nǐhǎo" can be matched by (nihao)

    - default matches substring of node description
       - use quotes to include spaces in search pattern, e.g. "several word sequence"
    - match whole node description
       - To use postgres ts_vector (okapi BM25 ranked) patterns, e.g. to join neighbouring lexemes separated by space
```
       strange<->kind<->of<->woman
```

to skip 2 lexemes

```
       strange<2>woman
```

- Use | character to delimint: e.g.

```
|a1|, "|deep purple|"
```

- Default searches match against all data in the whole graph
  - To limit any search to a particular chapter, PATTERN \chapter CHAPTER
  - To limit any search to a particular context keywords, PATTERN \context CHAPTER

- \notes PATTERN search reproduces the compiled notes in the exactly as they appear in .N4L files

- \from START-NODE \to END-END solves for connected paths from start to finish
- \from START-NODE finds an expandng cone of paths leaving from a node
- \to END-NODE finds a converging cone of paths arriving at a node

- \story search looks for paths connected by


## Special searches

- \arrow PATTERN/NUMBER retrieves an arrow definition
- (nclass,ncptr) 