---
name: release
description: >-
  Prepare an auth (Go, supabase/auth fork) release from the develop-to-master
  delta. Inspects git history, writes the versioned changelog, checks
  openapi.yaml and the generated admin client, migrations (Go and Liquibase),
  README configuration and endpoint docs, and runs the Go version check, gofmt,
  vet, staticcheck, gosec and govulncheck (tests when asked). Use when cutting an
  auth release, writing release notes, checking release readiness, or asking
  what is unreleased between develop and master.
license: MIT
---

# Auth release workflow

`auth` is the Solomon fork of supabase/auth (GoTrue). It provides the identity layer for every Solomon product. A `v<version>` tag runs `.github/workflows/release.yaml`, which builds the binaries with `-X github.com/supabase/auth/internal/utilities.Version=<tag>`, builds the image, and creates the GitHub release from `changelogs/<tag>.md`. `deploy.yaml` rolls k8s. There is **no** version field in any manifest: the tag is the version.

Run commands from `auth/`. The default branch is **master**.

## 1. Branch analysis

Default range: **`master..develop`** (commits on `develop` not yet on `master`). If the user says to **omit uncommitted changes**, use only committed history.

1. `git fetch origin`, then `git log master..develop --format='%h %s' --reverse`, and read the commit bodies.
2. `git diff master..develop --stat`, then open: `internal/api/` (handlers, admin, SCIM, OAuth, SAML, passkeys, MFA), `internal/conf/` (config and env), `internal/models/`, `internal/tokens/`, `internal/ratelimit/`, `internal/security/`, `internal/hooks/`, `internal/mailer/`, `migrations/`, `liquibase_migrations/`, `openapi.yaml`, `client/admin/`, `k8s/`, `kustomize/`, `templates/`.
3. Group the findings into user-facing areas. Ignore upstream-merge noise unless it changes behavior.

**Changelog scope:** include the public and admin HTTP API, auth flows, configuration and secrets, migrations, security fixes, and the image and binaries. Leave out tooling, `.claude/` config, CI, and tests or refactors with no visible effect. If the delta is all internal, do not release unless asked. A summary-only request gets the range, the commit count, and the grouped summary, with no edits.

## 2. Version and changelog

- Use semver. The new version must be higher than the highest `changelogs/v*.md` (compare numerically, ignore `latest.md`). Use PATCH for fixes. Use MINOR for features, and on `0.x` also for breaking changes. Use MAJOR only on request.
- Mirror `changelogs/v0.1.0.md` (`# Changelog`, `## [<version>] - YYYY-MM-DD`, a summary paragraph, `**Section:**` headings with `*` bullets). Write `changelogs/v<version>.md` and an exact copy at `changelogs/latest.md`. The filename and tag have a `v`, and the heading does not. `release.yaml` reads `changelogs/<tag>.md`, so the tag **must** be `v<version>`.
- Never edit `CHANGELOG.md`. That file is the upstream release-please history.
- Typical sections: authentication and MFA, OAuth and SAML, admin and SCIM APIs, passkeys, configuration and secrets, security and rate limiting, migrations and deployment notes.

### Semantic versioning rules

These rules apply to every version this skill produces. The repo-specific bump guidance above decides which segment moves.

- **Format:** `MAJOR.MINOR.PATCH`, exactly three non-negative integers with no leading zeros (`0.13.0`, never `0.13`, `v0.13`, or `0.013.0`). Add a pre-release only when the user asks: `-rc.1` in npm and Cargo versions, the PEP 440 form (`0.2.0rc1`, `0.0.1b0`) in Python packages.
- **Ordering:** compare segment by segment as integers, never as strings (`0.10.0` > `0.9.0`, `0.2.10` > `0.2.9`). The current version is the highest released one, and the new one must be strictly greater.
- **Bumps:** MAJOR for incompatible changes to the public surface (API, wire or gRPC contract, CLI flags, env vars, config, schema, exported symbols). MINOR for backward-compatible features. PATCH for backward-compatible fixes. Increment exactly one segment and reset the lower ones to `0`.
- **`0.y.z`:** the public surface is not yet stable. Breaking changes bump MINOR. Never move to `1.0.0` or any new MAJOR without an explicit user request.
- **Immutability:** never reuse, decrease, or retag a published version. If a pushed tag is wrong, fix forward with a new PATCH.
- **No-op deltas:** if nothing consumer-visible changed, do not bump. Report that instead.
- **One version everywhere:** the manifest, the lockfile entry, the changelog filename and heading, README badges and mentions, any generated spec (`info.version`), and the git tag (with the `v` prefix where the workflow expects it) carry the same number.

## 3. Release readiness

### 3.1 API spec and generated admin client

- [ ] `openapi.yaml` documents every new or changed endpoint, parameter, and response (public and admin).
- [ ] `client/admin/client.go` is regenerated from it: `go generate ./client/admin` (oapi-codegen, configured in `client/admin/oapi-codegen.yaml`). `git diff --exit-code client/admin` must be clean after generation. Never hand-edit `client.go`.
- [ ] The `docs/*.go` swagger annotations match the handlers they describe.

### 3.2 Migrations

- [ ] New Go migrations in `migrations/` are additive, use timestamped names, and run cleanly with `make migrate_dev` on a fresh DB.
- [ ] The Solomon-specific Liquibase changelogs in `liquibase_migrations/` (`changelog-auth.sql`, `-rls-auth`, `-index-auth`, `-grants-auth`, `-procedures-auth.xml`, `-public`) have unique changeset ids and `--rollback`s.
- [ ] The solomon root repo keeps its own copies in `solomon/migrations/` (they currently differ). When an auth schema change must reach solomon, zion, or the Rust services' test schemas, list that follow-up.

### 3.3 Docs and config

- [ ] `README.md`: the Configuration sections (Top-Level, API, Database, Logging, Observability, JWT, External providers, Email, Phone, CAPTCHA, …) list every new or renamed `GOTRUE_*` / `API_*` / `DB_*` env var from `internal/conf/`, with defaults. The Endpoints section covers new or changed routes. The fork-specific notes (Backward compatibility, Inherited features) are current.
- [ ] `example.env`, `example.docker.env`, `k8s/`, `kustomize/overlays/*`: new env vars and secrets are wired in.
- [ ] `docs/saml_key_rotation.md` and the other operator docs, when the matching area changed.

### 3.4 Lint and security gate

```bash
make check-go-version   # Go version pinned consistently in go.mod, Dockerfiles, submodules
make check-format       # gofmt (run `make format` to fix)
make lint               # check-go-version + vet + staticcheck + gosec + govulncheck
make build              # binaries build
```

Every command must pass. Fix vet, staticcheck, gosec, and govulncheck findings, and bump vulnerable modules (`go get <mod>@<fixed>` then `go mod tidy`). Do **not** run `make test` or `make docker-test` during routine prep, because they are slow and need Postgres (`docker-test` starts it on `:5437` and migrates). Run them when the user asks. `make release-test` is lint + test.

## 4. Deliverable

Report the range and commit count, the grouped summary, the version and date, the files changed, the README, docs, and openapi updates, the gate results (tests skipped unless requested), and the follow-ups (solomon records the new submodule SHA, and migration copies in other repos). Do not commit, tag, or push unless asked. When the user does ask, tag `v<version>` on **master**.

### Pull request title and summary

Always finish the deliverable with a ready-to-paste text for the release pull request (the source branch of the range from section 1 into its target). The user copies it into GitHub, so do not open the PR yourself. Give it as two separate fenced `markdown` blocks, one labelled **PR title** and one labelled **PR summary**, so each copies in one click.

- **PR title:** one line, 70 characters or fewer, in the form `Release v<version>: <the one or two headline changes>`. No trailing period.
- **PR summary:** written for reviewers, from the changelog and the actual diff, not from memory:
  - `## Summary`: two or three sentences on what ships and why it matters.
  - `## Changes`: the changelog bullets, grouped by area, consumer-facing items only.
  - `## Breaking changes / migration`: what downstream repos or operators must do. Omit the section when there is nothing.
  - `## Release checklist`: the version, the files that carry it, and each readiness check marked pass, fixed, or skipped (with the reason). Never mark a check as passed that was not run.
  - `## Follow-ups`: the steps that happen after the merge (tag, downstream pin bumps, submodule SHAs).
- Keep the title and summary consistent with `changelogs/v<version>.md`: same version, same facts. If there is no release (a no-op delta), say so instead and give no PR text.

## Release rules

Release-specific rules the user has set for this repository. They override the steps above. Use the format `- YYYY-MM-DD: <rule>. <reason, if given>`.

_None recorded yet._

## Keeping this skill current

This skill is versioned with the repository, so it must evolve with it. Update it **in the same change**, not as a follow-up, whenever:

- **Something changes:** layout, commands or Make targets, env vars, ports, generated artifacts, CI workflows, cross-repo contracts, or dependency pins. Also when you find a statement here that is wrong or stale.
- **The user sets a rule:** a preference, convention, or prohibition ("never use X", "always do Y", "don't touch Z", a frontend pattern to avoid). Record it under **Release rules** (in this file), with the date (YYYY-MM-DD) and the reason if given. Rules about how code is written go in the `development` skill. Rules about how versions ship go in the `release` skill.
- **You learn something reusable:** a pitfall you hit, a mistake a line here would have prevented, or a non-obvious procedure you worked out. Add it to the section it belongs to.
- **A task repeats:** if it is not covered by `development` or `release`, create a new skill at `.claude/skills/<kebab-name>/SKILL.md` (frontmatter `name` = the folder name, plus a `description` that says what it does and when to use it), and list it under "Related skills" in the `development` skill.

How to edit:

- Check every path, command, and target against the repo before writing it down.
- Keep it short. Edit the existing section instead of appending duplicates, and delete what is no longer true.
- Never record secrets, tokens, personal data, or one-off details of a single task.
- An explicit instruction from the user in the current conversation wins over a recorded rule. If they change a rule, update the entry.
- Tell the user in your final message which skill you changed and why. The skill file is part of this repository (often a git submodule), so include it when they ask you to commit.
