# N4L Plugin — Agent Instructions

This plugin provides 10+ skills for working with SSTorytime knowledge graphs. When assisting a user who is working with N4L files or SSTorytime, use these skills as composable building blocks.

## Available Skills

### Authoring & Import
| Skill | When to use |
|-------|-------------|
| `/n4l:scaffold` | User wants to create a new N4L file from a domain description |
| `/n4l:import` | User has a CSV file to convert to N4L |
| `/n4l:md-import` | User has a Markdown file to convert to N4L (headings → contexts, lists → edges, tables → row-per-node) |
| `/n4l:upload` | User wants to validate and upload an N4L file to the database |

### Querying & Exploration
| Skill | When to use |
|-------|-------------|
| `/n4l:search` | User has a specific query — translate to searchN4L/pathsolve and execute |
| `/n4l:explore` | User wants a broad overview of a topic across the graph |
| `/n4l:investigate` | User has a complex goal requiring multiple adaptive queries |
| `/n4l:interpret` | Search results contain arrow short codes that need decoding |

### Analysis & Presentation
| Skill | When to use |
|-------|-------------|
| `/n4l:narrate` | User wants paths between concepts presented as readable stories |
| `/n4l:hubs` | User wants to understand which concepts are structurally important and why |
| `/n4l:cone` | User wants to see causes and consequences of a concept |

### Learning
| Skill | When to use |
|-------|-------------|
| `/n4l:learn` | User is new to SSTorytime or asks "how do I..." questions |

## Composition Patterns

Skills are most powerful when chained together. Use these patterns when the user's goal spans multiple skills.

### Pattern: Deep Topic Analysis
**Goal:** "Tell me everything about X" / "What do I know about X?"
```
1. /n4l:explore X          → discover what exists (nodes, chapters, contexts)
2. /n4l:interpret           → decode any arrow short codes in the results
3. /n4l:hubs               → find which concepts are structural hubs
4. /n4l:cone X             → show causes and effects of key concepts
```

### Pattern: Story Discovery
**Goal:** "How does X connect to Y?" / "What's the path from X to Y?"
```
1. /n4l:search "paths from X to Y"  → find paths
2. /n4l:narrate X to Y              → present paths as readable narrative
3. /n4l:interpret                    → decode arrow codes in the story
```

### Pattern: Knowledge Audit
**Goal:** "How healthy is my graph?" / "What am I missing?"
```
1. /n4l:hubs                    → find bottlenecks, dead ends, unexplained origins
2. /n4l:cone [each bottleneck]  → understand the causal role of key nodes
3. /n4l:explore [dead ends]     → investigate what's missing
```

### Pattern: Import & Verify
**Goal:** "Import this data and make sure it's correct"
```
1. /n4l:import data.csv         → convert CSV to N4L
2. /n4l:upload output.n4l       → validate, fix errors, upload
3. /n4l:search [spot check]     → verify content is searchable
4. /n4l:interpret               → verify arrows make sense
```

### Pattern: Cross-Chapter Discovery
**Goal:** "Find surprising connections across my knowledge"
```
1. /n4l:hubs                    → identify semantic crossroads (NEAR hubs)
2. /n4l:search [crossroad]      → see what chapters it bridges
3. /n4l:narrate [chap1 concept] to [chap2 concept]  → narrate the bridge
```

### Pattern: Learning Workflow
**Goal:** User is new and wants to get started
```
1. /n4l:learn basics            → teach N4L fundamentals
2. /n4l:scaffold [their domain] → generate a starter file
3. /n4l:upload [their file]     → validate and upload
4. /n4l:explore [their topic]   → show them what they built
```

## Shared Output Conventions

When running skills in sequence, extract key findings to pass forward:

- **Node names** from search results can be used as input to `/n4l:cone`, `/n4l:narrate`, or further `/n4l:search` queries
- **Chapter names** from explore results can be used with `/n4l:search \\notes CHAPTERNAME`
- **Hub nodes** from `/n4l:hubs` are good candidates for `/n4l:cone` analysis
- **Arrow short codes** from any output should be passed to `/n4l:interpret` or looked up via `/n4l:search \\arrow CODE`
- **Dead ends and sources** from `/n4l:hubs` identify gaps worth exploring with `/n4l:explore`

## When to Compose vs. Use a Single Skill

**Single skill:** User asks a specific, bounded question ("find X", "upload this file", "what does (ph) mean?")

**Compose multiple skills:** User asks an open-ended question ("tell me about X", "how is my graph structured?", "find interesting connections"), or when the output of one skill raises a question that another skill can answer.

## SSTconfig/ Location

Multiple skills need SSTconfig/ for arrow definitions. The search order is consistent:
1. `SSTconfig/` in the current working directory
2. `$SSTORYTIME_HOME/SSTconfig/` environment variable
3. Prompt the user

Once found, reuse the same path for all subsequent skills in the conversation.

## Binary Dependencies

- `searchN4L` — needed by search, explore, investigate, interpret, narrate, cone
- `pathsolve` — needed by narrate, investigate (for deep path analysis)
- `graph_report` — needed by hubs
- `N4L` — needed by upload (validation and upload)

All binaries are built from the SSTorytime repo (`make` in the project root). They may be in `./` (CWD), `src/`, or on PATH.
