# Sharing a graph

![Pen-and-ink drawing of a wooden filing cabinet labelled N4L FILES on the left — drawer half-open, stacks of papers inside — with a single arrow pointing to a small CRT computer terminal labelled CACHE on the right. The cabinet is heavily shaded and solid; the terminal is drawn with thinner lines, suggesting it could be replaced at any time.](../figs/source_of_truth.jpg){ align=center }

> **Share the files, not the database. Your N4L files are the graph; the
> database is a cache anyone can rebuild from them in seconds.**

A collaborator does not need your PostgreSQL, your credentials, or a
snapshot. They need the text files you wrote. Given those, one command on
their machine produces a graph identical to yours.

---

## The short version

Give them a directory of `.n4l` files — over a git remote, a zip, a flash
drive, anything that moves text around. On their end:

```
./src/bin/N4L -u *.n4l
```

Same nodes, same arrows, same query answers. The database on their machine
is new; the graph in it is yours.

---

## A concrete example

The repo's reading-list example works as a demo. Clone or copy the folder
that contains `examples/reading-list.n4l`, hand it to a colleague, and
they run:

```
./src/bin/N4L -u examples/reading-list.n4l
./src/bin/searchN4L "decision making"
```

They see the same three books you see. Nothing else was transferred.

---

## What not to share

- **`~/.SSTorytime`.** It holds a plaintext password. It is theirs to
  create on their own machine, with their own database credentials, and
  to keep `chmod 600`.
- **The PostgreSQL data directory.** It is a cache of what is already in
  the `.n4l` files. Copying it wastes bytes and ties the recipient to
  your exact PostgreSQL version.

---

## What you might share as well

If you have defined your own arrow vocabulary — shorthand names beyond
what ships in the project — the arrow declarations live in `SSTconfig/`.
Hand that directory over along with the `.n4l` files so their ingest
recognises the same arrows yours does.

If you have not touched `SSTconfig/`, nothing extra is needed.

---

## Across machines, across people

There is no special protocol and no server in the middle. Sharing a graph
is sharing files; the graph reappears wherever those files are ingested.
Version-control them and the sharing story becomes the same one you
already use for code.
