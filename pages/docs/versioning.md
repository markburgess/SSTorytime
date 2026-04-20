# Versioning

This page describes both the **current state** of release management in
SSTorytime and a **proposed direction** for hardening it. Today the
project has no formal versioning discipline; this is normal for
research-driven software at this stage, and the lack of a changelog
doesn't make the code worse — but users who want to pin a known-good
revision, or contributors who want to reason about breaking changes,
need a shared framework to point at.

Consider this page aspirational as much as descriptive: it documents
what exists now and what we'd like to adopt as the project matures.

## Current state

As of this writing the repository has:

- **No SemVer tags.** `git tag` on the repo shows nothing — every
  commit on `main` is the current "latest", and there is no `v0.x.y`
  lineage.
- **No `CHANGELOG.md`.** User-visible changes are discoverable only
  through `git log`.
- **No release artefacts.** There are no pre-built binaries on the
  GitHub Releases page; every user builds from source via `make all`
  (see [Build system](build-system.md)).
- **No deprecation markers.** When an API changes, the old form is
  either kept indefinitely or removed silently.
- **No CI-enforced compatibility.** The
  [`.github/workflows/build.yml`](https://github.com/markburgess/SSTorytime/tree/main/.github/workflows)
  workflow runs `cd src && make` (build-only) and `go vet`. It does
  not run the test harness — `tests/run_tests` silently passes on
  database failures, so running it without a Postgres service would
  produce green CI that hides real errors. See
  [Testing → CI posture](testing.md) for the rationale and
  [Observability → gaps worth naming](operations/observability.md#4-gaps-worth-naming)
  for the broader story. No exported-symbol diffing or API compatibility
  check runs either.

For a pre-1.0 research codebase this is reasonable. It does mean that:

- Downstream users should vendor or pin by commit SHA, not by tag.
- Any database-schema change can land in a commit without a loud
  signal — users should be cautious about upgrading without reading
  diffs of `pkg/SSTorytime/postgres_types_functions.go`.
- Contributors shouldn't assume that "the API" is stable; refactors
  happen.

## Proposed adoption

The recommended direction, once the API surface stabilises, is
straightforward:

### 1. Adopt SemVer 2.0.0

Follow [semver.org](https://semver.org/) conventions:

- `MAJOR.MINOR.PATCH`
- **MAJOR** bump for any breaking change to a public Go export, CLI
  flag, database schema, or HTTP API.
- **MINOR** bump for backwards-compatible additions.
- **PATCH** bump for backwards-compatible bug fixes.

Tag releases:

```bash
git tag -a v0.1.0 -m "Initial tagged release"
git push --tags
```

For pre-1.0 releases, MINOR bumps may include breaking changes — this
signals "the API is still forming". Commit to strict SemVer after the
first `v1.0.0`.

### 2. Maintain a `CHANGELOG.md`

Follow the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
format. One section per version, with subsections for `Added`,
`Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`. Example:

```markdown
# Changelog

## [Unreleased]
### Added
- docs(testing): three-tier test harness documentation

## [0.1.0] - 2026-04-20
### Added
- MkDocs Material site scaffolding
- Visual layer (~16 mermaid diagrams + 18 AI-generated images)
### Fixed
- `make db` target invocation
```

Update the changelog in the same PR that lands the change — don't
defer to "release time".

### 3. GitHub Releases with binaries

Once a tag lands, create a GitHub Release:

1. Go to the Releases page → *Draft a new release*.
2. Pick the tag you just pushed.
3. Title: `vX.Y.Z`. Body: copy the CHANGELOG section for this version.
4. Attach pre-built binaries for the common platforms (linux-amd64,
   linux-arm64, darwin-amd64, darwin-arm64). A simple script:
   ```bash
   for pair in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
       GOOS=${pair%/*} GOARCH=${pair#*/} \
         go build -o dist/N4L-$GOOS-$GOARCH ./src/N4L
   done
   ```

This gives users who don't want to install Go a way to try the tools,
and gives automation a stable download URL.

## Deprecation policy

When a public identifier is going to change or go away:

1. Mark it in the GoDoc comment with a leading `Deprecated:` line, as
   per [Go's convention](https://go.dev/wiki/Deprecated):
   ```go
   // OldFunc does X.
   //
   // Deprecated: use NewFunc instead; OldFunc will be removed in v2.0.0.
   func OldFunc(...) { ... }
   ```
2. Add a `[DEPRECATED]` tag to any documentation page that references
   the old form, linking to the replacement.
3. Keep the old form working through at least one minor release. Remove
   it no sooner than the next major bump.
4. In `CHANGELOG.md`, list it under `Deprecated` at deprecation time
   and again under `Removed` when it's pulled.

This gives users a clear signal window: "this is going away, here's
what to do instead, and you have at least one release cycle to migrate."

## Release process (proposed)

A lightweight checklist for cutting a release:

- [ ] All tests pass (`cd tests && ./run_tests`).
- [ ] `mkdocs build --strict` passes.
- [ ] `CHANGELOG.md` has an entry for this version with today's date.
- [ ] `main` has the commit you want to tag.
- [ ] Create the tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`.
- [ ] Push: `git push && git push --tags`.
- [ ] Build release binaries (script above).
- [ ] Draft a GitHub Release, attach binaries, paste the changelog
      section.
- [ ] Announce on the LinkedIn group with a one-paragraph summary.

## Known limitations

Even with all of the above adopted, a few gaps remain worth naming:

- **No CI/CD pipeline.** Releases would still be hand-cut. Phase 8
  adds a build-validation workflow but not a release-automation one.
- **No binary distribution today.** Until the release process is
  formalised, users build from source. This is a friction point for
  casual try-kickers.
- **No database migration framework.** Schema changes in
  `pkg/SSTorytime/postgres_types_functions.go` happen without
  versioned migration files. Users upgrading an existing database may
  need to drop and re-create (`make ramdb` + re-load examples). A
  migration tool (e.g. [`golang-migrate`](https://github.com/golang-migrate/migrate))
  would be a natural future investment.
- **No compatibility guarantees between N4L versions.** N4L source
  files written for today's parser may or may not parse cleanly under
  a future version. Keeping the parser-test corpus in
  [`tests/`](https://github.com/markburgess/SSTorytime/tree/main/tests)
  healthy is the primary backstop here.

## Why this matters

Versioning isn't bureaucracy — it's a promise to users. "If you pin
`v0.3.x`, you can upgrade to `v0.3.y` without reading the diff" is a
valuable property, and the cost of providing it (tags, a changelog,
occasional deprecation notes) is low compared to the frustration of
unannounced breakage.

This page is the place to start that conversation. It is not yet
project policy; it is a proposal that the documentation plan invites
maintainers and contributors to discuss.
