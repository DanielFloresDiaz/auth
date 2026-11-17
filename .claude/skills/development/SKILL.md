---
name: development
description: >-
  How to develop the auth Go server (supabase/auth fork): layout, config, the
  two migration systems (Go migrations and Solomon Liquibase changelogs),
  openapi.yaml and the generated admin client, the dev/test Postgres, git hooks,
  and which checks to run. Use when changing handlers, auth flows, config,
  models, migrations, or the admin API in auth. Also covers validating every
  change (lint, typecheck, tests) and the user's project rules for this repo.
license: MIT
---

# Auth development guide

## Layout

| Path | Role |
|---|---|
| `main.go`, `cmd/` | CLI: `serve`, `migrate`, `admin`, `version` |
| `internal/api/` | HTTP handlers (signup, token, verify, OTP, OAuth, SAML, MFA, passkeys, admin, SCIM) |
| `internal/conf/` | Configuration parsed from env (`GOTRUE_*` etc.) |
| `internal/models/`, `storage/` | DB models and access |
| `internal/tokens/`, `security/`, `ratelimit/`, `hooks/`, `mailer/`, `observability/` | Cross-cutting services |
| `migrations/` | Upstream-style Go migrations (`*.up.sql`, applied by `make migrate_dev` / `hack/migrate.sh`) |
| `liquibase_migrations/` | Solomon-specific schema, roles, RLS, grants, procedures (Liquibase) |
| `openapi.yaml` | API spec. Source for `client/admin` |
| `client/admin/` | **Generated** Go admin client (`go generate ./client/admin`, oapi-codegen) |
| `docs/` | Swagger annotation files and operator docs |
| `hack/` | Scripts (migrate, Go version check, coverage, init SQL) |
| `k8s/`, `kustomize/`, `dockerfiles/` | Deployment |

## Setup

```bash
make deps
make hooks                   # lefthook: gofmt check + lint on commit (needs lefthook installed)
cp example.env .env
make dev                     # docker compose dev stack (Postgres on :5437, Liquibase)
make migrate_dev
```

The Go version is pinned in `go.mod` and must match the Dockerfiles. `make check-go-version` verifies this. Change all of them together.

## Changing the API

1. Implement the handler in `internal/api/` and add its config to `internal/conf/` if needed.
2. Document it in `openapi.yaml`, update `docs/*.go` and the README Endpoints section.
3. Regenerate the admin client: `go generate ./client/admin`.
4. Add tests next to the handler. They need Postgres (`make docker-test`, or `make dev` + `make migrate_test` + `make test`).

## Migrations

- Changes to the upstream auth schema go in `migrations/` as a new timestamped file. Never edit an applied migration.
- Solomon roles, RLS, grants, and procedures go in `liquibase_migrations/` as Liquibase changesets with `--rollback`.
- The solomon root (`solomon/migrations/`) and the Rust services (rate-limiter-rs and request-scheduler-rs test schemas) carry their own copies of some `auth` tables (`auth.projects`, `auth.api_keys`). Propagate column changes they depend on.

## Checks

| Change | Run |
|---|---|
| Go code | `make check-format` (or `make format`), `make lint` (vet, staticcheck, gosec, govulncheck), then the related tests |
| API surface | The above + `go generate ./client/admin` + `openapi.yaml` + README |
| Migrations | `make docker-test` (fresh DB, migrate, full test run) |
| Dependencies | `make vulncheck`, `go mod tidy` |

Never edit `CHANGELOG.md` (upstream history). Use the `release` skill for fork releases.

## Related skills

- `release`

## Project rules

Rules the user has set for this repository. They override the defaults above. Add an entry whenever the user states one (see "Keeping this skill current"), using the format `- YYYY-MM-DD: <rule>. <reason, if given>`.

_None recorded yet._

## Keeping this skill current

This skill is versioned with the repository, so it must evolve with it. Update it **in the same change**, not as a follow-up, whenever:

- **Something changes:** layout, commands or Make targets, env vars, ports, generated artifacts, CI workflows, cross-repo contracts, or dependency pins. Also when you find a statement here that is wrong or stale.
- **The user sets a rule:** a preference, convention, or prohibition ("never use X", "always do Y", "don't touch Z", a frontend pattern to avoid). Record it under **Project rules** (in this file), with the date (YYYY-MM-DD) and the reason if given. Rules about how code is written go in the `development` skill. Rules about how versions ship go in the `release` skill.
- **You learn something reusable:** a pitfall you hit, a mistake a line here would have prevented, or a non-obvious procedure you worked out. Add it to the section it belongs to.
- **A task repeats:** if it is not covered by `development` or `release`, create a new skill at `.claude/skills/<kebab-name>/SKILL.md` (frontmatter `name` = the folder name, plus a `description` that says what it does and when to use it), and list it under "Related skills" in the `development` skill.

How to edit:

- Check every path, command, and target against the repo before writing it down.
- Keep it short. Edit the existing section instead of appending duplicates, and delete what is no longer true.
- Never record secrets, tokens, personal data, or one-off details of a single task.
- An explicit instruction from the user in the current conversation wins over a recorded rule. If they change a rule, update the entry.
- Tell the user in your final message which skill you changed and why. The skill file is part of this repository (often a git submodule), so include it when they ask you to commit.
