# Why Semantic Spacetime?

Knowledge does not scale the way most tools pretend it does. A shoebox of
notes is easy: you read it through in an evening and roughly remember what
is in it. A filing cabinet already needs an index. A wiki needs a tended
front page, or it rots. By the time the collection is genuinely useful it
has grown large enough that no single person holds it in their head — which
is the moment we reach for a database and quietly hope the database will
hold it *for* us.

It will not. A database is a warehouse with very good shelving; the knowing
still has to happen in a human. SSTorytime is built around that fact rather
than against it. It borrows the *smart spacetime* work of Mark Burgess — the
observation that relationships between ideas, like relationships between
events in physics, have only a handful of genuinely distinct shapes — and
uses it to keep the human in the driver's seat while the machine does the
bookkeeping.

## The tending metaphor is not decoration

Think of a garden. You discover plants, you plant new ones, you weed, you
notice where the light falls. Over time you learn the garden. None of this
can be delegated; if someone else tends it for you, you have a pleasant
place to walk but you do not *know* it. The bigger the garden, the more
tending it takes. This is the shape of knowledge work, and the reason most
knowledge-management projects quietly fail is that they try to replace the
tending rather than make it cheaper.

## What goes wrong with RDF, OWL, and topic maps

The standard response to "I have a lot of notes" has, for twenty-five years,
been: put them in a triple store and write an ontology. The trouble is that
writing an ontology is the work of already understanding the material. It
is the finish line dressed up as the starting pistol.

Three failures compound.

**Schema-first rigidity.** You cannot write down what you do not yet know
how to classify. The model punishes the fuzzy first draft, which is exactly
the shape real thinking has when it begins.

**Closed-world assumption.** If it isn't in the graph, it doesn't exist.
Fine for a booking system, disastrous for a notebook where the unknown is
the whole point.

**The ontology trap.** The more effort invested in a schema, the higher the
switching cost when it turns out to have been wrong — and it always turns
out to have been wrong, because classification is itself an act of
learning. Ontologies become organisational weapons rather than descriptive
tools. Mark has
[written about this at length](https://medium.com/@mark-burgess-oslo-mb/avoiding-the-ontology-trap-how-biotech-shows-us-how-to-link-knowledge-spaces-654bcbb9122a).
The duck-billed platypus is the stock example: a warm-blooded, egg-laying,
milk-producing animal that refuses to sit in one of biology's largest
boxes. Real knowledge is full of platypuses.

## The four relations — semantics, not an ontology

Across every domain — biology, code, history, groceries — the relationships
between things fall into four broad shapes.

- **NEAR** — similarity or proximity. *A is like B.* Symmetric.
- **LEADSTO** — causation or sequence. *A happens before B, or brings B about.*
- **CONTAINS** — composition. *A is part of B.*
- **EXPRESSES** — property or attribute. *A has the quality B.*

These are not *an ontology*. They are the semantics of spacetime — the four
patterns a physicist would need to describe any set of events and their
connections. Every relation you might write down ("is the capital of",
"rhymes with", "was painted by") is a flavour of one of the four. You
compose named arrows — `about`, `by`, `bib-cite`, `is-the-capital-of` —
and declare them in the project's `SSTconfig/` files against one of the
four meta-types. The *names* are yours; the *physics* underneath is shared.
The ontology trap tells you *what things are*; Semantic Spacetime asks
only *how they relate*, and leaves the naming to the author.

## Context, and the role of intent

Four kinds of link are not yet enough. The same relationship means different
things in different settings: Alice trusts Bob on a rope pitch is not Alice
trusts Bob with the company credit card. Every link therefore carries a
pointer into a shared [context](glossary.md#context) directory,
and context itself is split in two.

*Ambient* context is the scene the note was written inside — chapter, place,
time-of-day, the `:: tags ::` that precede a block. *Intentional* context
is what the reader brings: what they are looking for, and why. At query
time the two are compared and the overlap is the answer.

That split matters because **intent is where knowledge begins**. What we
miss when we try to cram knowledge is a sense of intent — of connecting
what we are actually trying to do to what we can remember (see
[Knowledge and learning](../KnowledgeAndLearning.md) for the longer
argument). `text2N4L` can generate ambient context automatically from the
source, but it cannot guess intent; only you know what you are looking
for. [How does context work?](../howdoescontextwork.md) walks this through
in examples.

## Stories are the whole point

Knowledge is not a bag of facts; it is the walk that leads from one
understanding to the next. A fact is recalled; a story is *told*.
SSTorytime preserves narrative order explicitly — through
[PageMap](glossary.md#pagemap) and
[sequence mode](glossary.md#sequence-mode) — so a concept comes back to
you as a route rather than a flat list of attributes. If your notes do
not form an obvious story, they are incomplete, and the hunt for what is
missing is itself the learning. Mark's essay
[Storytelling](../Storytelling.md) makes the case in full.

## Knowing as a relationship, not a fetch

You know something when you know it like a friend. You tip your hat,
then say good morning, then stop for a chat. Only after
long familiarity can you claim to know. The machine can make the revisiting
cheap; it cannot do the revisiting for you. You write the notes. You decide
the arrows. The system remembers, relates, and surfaces; it does not invent.
Answers come back with a path you can follow and, if you disagree, edit.
There is no oracle.

## When to reach for SST

Semantic Spacetime is the right tool when your source material arrives as
**narrative** — a notebook, a course, a research thread, an incident
postmortem — and the structure should emerge from the writing rather than
precede it; when you need to ask **contextual** questions (what does this
person do at work, as distinct from at home) rather than merely categorical
ones; when you care about **paths** — the chain of reasoning between ideas,
not just the leaves; or when you plan to feed the result to an LLM and want
a graph to traverse rather than a pile of embeddings.

It is *not* the right tool when your data is strictly tabular and you really
just want SQL with joins; when a regulator or a standards body requires OWL
or SPARQL; or when you are building something transactional where
closed-world assumptions are load-bearing.

For everything else — notes, teaching material, biographies, software
architectures, research corpora, any domain meant to be learned rather than
merely looked up — the SST model was built for it.

## Where to go next

- [How does context work?](../howdoescontextwork.md) — the
  ambient-vs-intentional split, worked through examples.
- [Thinking in arrows](../arrows.md) — the four types in detail, with
  worked examples of when each is the right choice.
- [Writing N4L by hand](../N4L.md) — the notation itself.
- [Knowledge and learning](../KnowledgeAndLearning.md) and
  [Storytelling](../Storytelling.md) — Mark's longer essays, for readers
  who want the full framing.
