# Why Semantic Spacetime?

Knowledge graphs are fashionable again. Every LLM vendor wants one bolted to
its side; every enterprise platform claims to have one under the hood. Users
reach for them in the hope that a picture of connected ideas might finally be
the antidote to folder-tree memory. Then they try to build one and find they
need a doctorate in ontology engineering just to describe their grocery list.

SSTorytime is a rejection of that gatekeeping — not of graphs, but of the
idea that graphs must be preceded by schemas. The project borrows its
foundations from work on *smart spacetime* by Mark Burgess; it keeps what is
useful about graph data structures and discards what has made them hostile to
ordinary humans.

## What goes wrong with RDF, OWL, and topic maps

RDF and its descendants ask you to commit to a worldview before you type your
first triple. You must design classes. You must choose an ontology — ideally
one someone else has already published — and map your particular data onto it.
If your data doesn't fit, you extend the ontology. If it still doesn't fit,
you argue with the ontology's maintainers. If you are a working researcher
with a deadline this week, you lose interest.

Three failures compound:

**Schema-first rigidity.** You cannot write down what you don't already know
how to classify. The model punishes fuzzy first drafts, which is exactly how
real thinking starts.

**Closed-world assumption.** If it isn't in the graph, it doesn't exist. This
is fine for an airline booking system and disastrous for a learner's notebook,
where unknowns are the whole point.

**The ontology trap.** The more effort you invest in schema, the higher the
switching cost when the schema turns out to be wrong. Ontologies become
organizational weapons rather than descriptive tools. [Mark has written about
this at length](https://medium.com/@mark-burgess-oslo-mb/avoiding-the-ontology-trap-how-biotech-shows-us-how-to-link-knowledge-spaces-654bcbb9122a).

## The SST answer

Semantic Spacetime starts from an empirical claim: across every domain humans
try to describe — biology, code, history, groceries — relationships between
things collapse into **four universal patterns**.

- **NEAR** — similarity. *A is like B.* Symmetric.
- **LEADSTO** — causation or sequence. *A happened before B, or caused B.*
- **CONTAINS** — composition. *A is part of B.*
- **EXPRESSES** — property. *A has attribute B.*

That is the whole ontology. Everything else is vocabulary layered on top: you
can define "is the capital of" as a flavour of CONTAINS and "precedes in the
scale" as a flavour of LEADSTO, but the *physics* of those relations is
already known. You do not have to design it.

Built on that base, SST makes four further commitments:

**Context is first-class.** Every link carries a compact pointer into a shared
context directory, so "Alice trusts Bob *in the context of mountaineering*" is
a different edge from "Alice trusts Bob *in the context of finance*." Context
is split into *ambient* (what the author was writing about) and *intentional*
(what the searcher is asking about), and the two are compared at query time.
This is why SSTorytime can answer questions that strict triple stores cannot.

**Stories are paths.** Knowledge is not a bag of facts; it is a walk — the
sequence that leads from one understanding to the next. SSTorytime preserves
that ordering explicitly (through **PageMap** and **Sequence mode**) and
retrieves it as a path rather than an aggregate. When you ask about a concept
you get back not a list of attributes but a *route*.

**Open-world assumption.** If it isn't in the graph, maybe nobody wrote it
down yet. Absence is not falsehood. The system is built to absorb messy,
half-finished notes and improve them incrementally rather than demanding
completeness up front.

**Notation-first.** The authoring medium is a DSL called **N4L** — Notes for
Learning — that looks like indented bullet points. You can learn it in ten
minutes. You can revise it in any text editor. Only when you're ready do you
upload it, at which point the parser builds the graph for you. The graph is
a consequence of the notes, not a prerequisite for them.

## Cyborg enhancement, not replacement

The project's stated aim is not to be an Artificial Intelligence. It is,
in the project's own words, a **cyborg enhancement** — a tool that helps you
*know your own thinking*. The distinction matters. An AI that answers for you
still leaves you ignorant; only slightly less ignorant than before, because
you now have a false fluency. SSTorytime goes the other way: it makes your
own scattered notes searchable, pathable, and story-shaped so that *you* build
knowledge by tending it like a garden (see
[Storytelling](../Storytelling.md) and
[Knowledge & Learning](../KnowledgeAndLearning.md) for Mark's longer essays
on this).

The practical consequence: every feature in SSTorytime is designed to keep
the human in the driver's seat. You write the notes. You decide the arrows.
The system remembers, relates, and surfaces — it does not invent.

## When to reach for SST

Semantic Spacetime is the right tool when:

- Your source material arrives as **narrative** — a notebook, a course, a
  research thread, an incident postmortem, an ongoing investigation — and you
  want the structure to emerge from the writing, not precede it.
- You need to ask **contextual** questions ("what does this person do when
  they're at work" vs. "what do they do at home") rather than merely
  categorical ones.
- You care about **paths**: the chain of reasoning that connects one idea to
  another, not just the leaves.
- You plan to feed the result to an LLM and want the model to have a real
  graph to traverse rather than a pile of embeddings.

It is *not* the right tool when:

- Your data is strictly tabular, every row has the same shape, and you really
  just want SQL with joins.
- You need OWL reasoning or SPARQL compatibility because a regulator or a
  standards body says so.
- You are building something transactional where closed-world assumptions are
  load-bearing — a ledger, a booking system, a permission matrix.

For everything else — notes, teaching material, biographies, software
architectures, research corpora, domain knowledge meant to be taught and
learned — the SST model was *built* for it.

## Where to go next

- [How Context Works](../howdoescontextwork.md) — the ambient-vs-intentional
  split, worked through examples.
- [Arrows & Relationships](../arrows.md) — the four types in detail, with the
  7-channel storage encoding.
- [N4L Reference](../N4L.md) — the authoring DSL.
- [Architecture](architecture.md) — how the pieces fit together in code.
- [Knowledge & Learning](../KnowledgeAndLearning.md) and
  [Storytelling](../Storytelling.md) — Mark's manifesto essays, for readers
  who want the full philosophical framing.
