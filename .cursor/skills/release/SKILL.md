---
name: release
description: >-
  Prepare auth releases from develop-to-master branch deltas or release context.
  Inspects git history, writes versioned changelogs, updates README, and runs Go
  release validation. Use when cutting an auth release, bumping versions, generating
  changelog or release notes, or asking what is unreleased between develop and master.
license: MIT
---

# Auth release workflow

Prepare releases for **auth** (Go authentication server). Run all commands from the `auth/` directory (submodule root).

## Branch analysis

Default compare range: `develop..master` (`target..source`).

If the user says to **omit uncommitted changes**, base analysis only on the committed branch range (`git log target..source`, `git diff target..source`).

Do not stop at `git log --oneline`.

1. List commits chronologically: `git log target..source --format='%h %s' --reverse`.
2. Read commit bodies when present.
3. Review stats: `git diff target..source --stat`.
4. Open behavior-heavy diffs for API, persistence, runtime, config, migrations, and operator docs.
5. Group related commits into user-facing areas.

**Exclude from the changelog** anything that does not affect consumers of the product — runtime behavior, public HTTP API, admin API, operator-facing configuration, or published container/binary artifacts. Do not mention:

- Removal or addition of internal dev tooling, editor config, or CI-only changes with no user impact
- Refactors, tests, or docs that only support development workflow unless they change documented public behavior

If a commit mixes product and non-product changes, include only the product-facing parts. When the branch delta is entirely non-product work, do not cut a release unless the user explicitly asks.

For summary-only requests, report the branch range, commit count, grouped summary, and optional commit-by-commit rundown. Skip release file edits.

## Changelog format

Mirror recent files such as `changelogs/v0.1.0.md`.

```markdown
# Changelog

## [<version>] - YYYY-MM-DD

<One short active-voice summary paragraph.>

**Section title:**

* Bullet describing what changed and why it matters.
```

Rules:

- Create `changelogs/v<version>.md` and update `changelogs/latest.md` to match **exactly**.
- Use a `v` prefix in the **filename** only (e.g. `changelogs/v0.2.0.md`); the heading uses semver **without** `v` (e.g. `## [0.2.0] - YYYY-MM-DD`).
- Git release tags must use the `v` prefix so CI resolves `body_path` correctly (`v0.2.0` → `changelogs/v0.2.0.md`).
- Document **product changes only** — see branch analysis exclusions above.
- User-facing, active voice; no commit hashes or PR links unless the user asks.
- End every edited file with exactly one trailing newline.
- Do **not** update [CHANGELOG.md](../../CHANGELOG.md) (upstream release-please history).

Typical section themes: authentication and MFA, OAuth and SAML, admin and SCIM APIs, passkeys, configuration and secrets, security and rate limiting, migrations and deployment notes.

## Version selection

Use [semver](https://semver.org/). The new version must exceed the highest `changelogs/v*.md` (ignore `latest.md`). Use today's date in `## [<version>] - YYYY-MM-DD`.

Infer bump from the unreleased delta unless the user gives an explicit version:

| Bump | When |
|---|---|
| **PATCH** | Bug fixes, docs, internal-only changes |
| **MINOR** | New features without breaking changes |
| **MAJOR** | Breaking changes or removals |

On `0.x`, prefer **MINOR** for large additive changes unless the user requests a major bump.

There is no `pyproject.toml` or `Cargo.toml` version field. The release version is recorded in changelogs and applied at build time via the git tag (`-X github.com/supabase/auth/internal/utilities.Version`).

## Release artifacts

- `changelogs/v<version>.md` and `changelogs/latest.md`
- `README.md` — release/version notes, operator-facing config or API changes (do not paste the full changelog)

### README consistency

1. Read the full `README.md`.
2. Reconcile with the changelog; extend existing sections, do not paste the changelog.
3. Document new endpoints, configuration, breaking changes, or migration steps operators need.

## Validation

Respect user instructions (e.g. skip tests).

**Do not** run `make test` or `make docker-test` during routine release preparation — the suite is slow and needs Postgres. Run them only when the user explicitly asks (e.g. "run tests before release"). `make release-test` runs lint plus the full test suite when a hard gate is needed.

From `auth/`:

```bash
make check-go-version
make check-format
make lint
```

Fix formatting, vet, staticcheck, gosec, or vulncheck failures before finishing.

## Deliverable

Report briefly:

- Compared branch range and commit count
- Grouped summary of unreleased changes
- New version and date
- Files created or updated
- README changes (or note if none)
- Validation results (note skipped checks if applicable)

Do not commit unless the user asks.
