# AGENTS

Guidance for AI coding agents working in this repository.

## Goals

- Prefer clarity and simplicity over flexibility and extensibility.

## Quick Start

- Run all tests: `make tests`
- Recreate local test database: `make testdb`
- Install/update CLI used by setup scripts: `make cli`

## Repo Map

- Root package provides generic HTTP CRUD handlers over stores.
- [lyspg/](lyspg/) contains PostgreSQL CRUD primitives and SQL helpers.
- [lysclient/](lysclient/) provides client helpers used by tests and examples.
- [internal/cmd/lyscli/](internal/cmd/lyscli/) contains DB setup tooling.
- [internal/sql/ddl/](internal/sql/ddl/) contains SQL used for local DB/user setup.

## Improvement Suggestion Policy

When suggesting or making improvements, prioritize in this order:

1. Correctness and regression risk.
2. Error handling and data safety.
3. Naming errors or inconsistencies.
4. Simplicity and readability.
5. Missing or weak tests.
6. Performance.

Unless absolutely necessary, do not suggest changes that add abstraction layers, increase flexibility, or introduce multiple ways to do the same thing. Prefer one clear, obvious approach over configurable or extensible designs.

## Change Guardrails

- Do not alter public request/response behavior silently.
- Keep SQL access patterns aligned with existing `lyspg` usage where practical.

## Testing Expectations

- Prefer targeted tests first, then full suite when risk is broader.
- Keep table-driven style where already used.
- If DB behavior changes, validate against the PostgreSQL-backed tests in this repo.

## Environment Notes

- Tests rely on a real PostgreSQL setup. See [CONTRIBUTING.md](CONTRIBUTING.md).

## Source Docs

- Project overview and usage patterns: [README.md](README.md)
- Development setup and constraints: [CONTRIBUTING.md](CONTRIBUTING.md)
- Store/route patterns and feature docs: [project wiki](https://github.com/loveyourstack/lys/wiki)