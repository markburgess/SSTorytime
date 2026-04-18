# SST Arrow Types Reference

Every relationship in SSTorytime belongs to one of four Semantic Spacetime (SST) meta-types.

## The Four Meta-Types

### Type 0: NEAR / Similarity

Things that are close, alike, or equivalent. Bidirectional (no forward/reverse distinction).

**When to use:** Synonyms, translations, comparisons, associations, "see also" links.

**Common arrows:**
| Arrow | Short Code |
|-------|-----------|
| looks like | (ll) |
| sounds like | (sl) |
| synonym | (syn) |
| see also | (see) |
| similar to | (sim) |
| equals | (eq) |
| same as | (=) |
| associated with | (ass) |
| compare to | (compare) |
| is a variant of | (variant) |

### Type 1: LEADSTO / Causality

One thing follows from, causes, or leads to another. Directional.

**When to use:** Sequences, processes, cause-effect, temporal ordering, prerequisites.

**Common arrows:**
| Forward | Code | Reverse | Code |
|---------|------|---------|------|
| leads to | (fwd) | comes from | (bwd) |
| causes | (cause) | is caused by | (cause-by) |
| then / next | (then) | prior | (prior) |
| affects | (aff) | is affected by | (aff-by) |
| results in | (result) | was a result of | (result-of) |
| enables | (enables) | is enabled by | (enabled-by) |
| creates | (cr) | is created by | (crtd) |
| replaces | (repl) | is replaced by | (repl-by) |

### Type 2: CONTAINS / Membership

Something is part of something else. Directional.

**When to use:** Part-whole, sets, groups, ingredients, components, membership.

**Common arrows:**
| Forward | Code | Reverse | Code |
|---------|------|---------|------|
| contains | (contain) | belongs to | (belong) |
| has component | (has-cmpt) | is component of | (cmpt-of) |
| has ingredient | (ingred) | is an ingredient of | (ingred-of) |
| consists of | (consists) | makes up part of | (mkpt) |
| has member | (has-memb) | is a member of | (is-memb) |
| is a set of | (setof) | is part of the set | (in-set) |
| has a part | (has-pt) | is a part of | (pt-of) |
| occurs within | (occurs) | situates occurrences of | (occurence-of) |

### Type 3: EXPRESS / Properties

Something has an attribute, name, description, or annotation. Directional.

**When to use:** Names, descriptions, notes, examples, attributes, metadata, annotations.

**Common arrows:**
| Forward | Code | Reverse | Code |
|---------|------|---------|------|
| has example | (e.g.) | is an example of | (ex) |
| note/remark | (note) | is a note about | (isnotefor) |
| has name | (name) | is the name of | (nom-de) |
| means | (means) | is meant by | (meansb) |
| describes | (describe) | is described by | (descr-by) |
| has version | (version) | is a version of | (is-version) |
| has status | (status) | is the status of | (status-of) |
| to be decided | (tbd) | to be figured out | (tbf) |

## Arrow Definition File Format

Arrow definitions live in `SSTconfig/` in four files:

- `arrows-NR-0.sst` — Similarity (Type 0)
- `arrows-LT-1.sst` — LeadsTo (Type 1)
- `arrows-CN-2.sst` — Contains (Type 2)
- `arrows-EP-3.sst` — Express/Properties (Type 3)

### Directional arrows (Types 1, 2, 3)

```
- leadsto
  :: context section ::
  + forward reading (short_code) - reverse reading (reverse_code)
```

Example: `+ causes (cause) - is caused by (cause-by)`

### Symmetric arrows (Type 0 only)

```
- similarity
  :: context section ::
  arrow_name    (short_code)
```

Example: `similar to (sim)`

NEAR arrows have no `+`/`-` prefix and no inverse — they are bidirectional.

Context sections (`:: ::`) within files are organizational groupings and can be ignored when extracting arrow names.

## Choosing the Right Arrow Type

| Your relationship is about... | Use Type |
|-------------------------------|----------|
| A is similar to / like / equivalent to B | NEAR (0) |
| A causes / leads to / happens before B | LEADSTO (1) |
| A contains / has part / includes B | CONTAINS (2) |
| A has property / is described as / means B | EXPRESS (3) |

**Tip:** If unsure, use `(tbd)` as a placeholder and refine later. The docs encourage this.
