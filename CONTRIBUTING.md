# Contributing to Morpheus Lumerin Node

Thanks for contributing. This repo uses a **promote-up** branch model so changes land in CI environments in order. Following it keeps PRs reviewable and avoids accidental production deploys.

## Branch model (read this first)

| Branch | Role | What happens on push |
|--------|------|----------------------|
| **`dev`** | Integration / default PR target | Builds & checks; **not** a public release |
| **`test`** | Staging | Deploys test docs (nodedocs.dev), test TEE / nodes, pre-release artifacts |
| **`main`** | Production | Releases, mainnet TEE / nodes, production docs (nodedocs.mor.org) |

**Flow:** feature branch → PR into **`dev`** → maintainers promote **`dev` → `test`** → promote **`test` → `main`** for release.

```
you → [PR] → dev  →  [promote] → test  →  [promote] → main
```

### Default PR base: `dev`

When you open a pull request on GitHub, set the **base branch to `dev`** (not `main`).

- First-time and external contributors: always target **`dev`**.
- Only maintainers open promote PRs into `test` or `main` (e.g. `promote: … to test`).
- If GitHub defaults your PR to `main`, change it before requesting review (`Edit` → base → `dev`), or ask a maintainer to retarget.

## How to submit a change

1. Fork (or clone with write access) and create a branch from **current `dev`**:
   ```bash
   git fetch origin
   git checkout -b docs/your-topic origin/dev
   ```
2. Make a focused change (one concern per PR when practical).
3. Open a PR with **base = `dev`**. Fill in the PR template checklist.
4. Keep the branch up to date with `dev` if review takes a while:
   ```bash
   git fetch origin && git merge origin/dev
   ```
5. Maintainers merge to `dev`, then promote through `test` / `main` on their schedule.

### Commit / PR titles

Prefer conventional prefixes so history stays scannable:

- `feat(proxy-router): …` / `fix(…): …` / `docs(…): …` / `ci: …`
- Promote PRs: `promote: <summary> to test` (or `to main`)

## What lives where

| Path | Purpose |
|------|---------|
| `proxy-router/` | Go service (consumer/provider router, HTTP API, TEE attestation) |
| `ui-desktop/` | Electron consumer GUI (product name **MorpheusUI**) |
| `cli/` | Go CLI client |
| `docs/` | Mintlify site → [nodedocs.mor.org](https://nodedocs.mor.org) |
| `AGENTS.md` | Rules for AI coding assistants working in this repo |
| `proxy-router/docs/swagger.yaml` | Proxy-router API schema (not the hosted Inference API) |

## Docs changes (`docs/`)

- Pages are MDX with frontmatter (`title`, `description`, `audience`, `product`, `last_verified`, optional `source_url`).
- Add new pages to [`docs/docs.json`](docs/docs.json) navigation.
- Preview locally: `cd docs && mint dev` (requires [Mintlify CLI](https://mintlify.com/docs/installation)).
- Prefer linking to live status at [active.mor.org](https://active.mor.org) rather than inventing counts or prices.
- Do not confuse the **proxy-router** API (this repo) with the hosted **Morpheus Inference API** ([apidocs.mor.org](https://apidocs.mor.org)).

## Proxy-router changes

- Auth is HTTP Basic Auth; remote Morpheus models need a `session_id` header — see [API auth](https://nodedocs.mor.org/reference/api-auth).
- After API shape changes, regenerate swagger if your workflow normally does.
- Run focused tests from `proxy-router/`:
  ```bash
  cd proxy-router
  go test ./internal/... -count=1
  ```

## Review expectations

- Green Docs / lint checks on the PR when applicable.
- Clear summary + test plan (use the template).
- No secrets in the PR (`.env`, private keys, API keys, base64 secret blobs).

## Questions

- Product / node docs: [nodedocs.mor.org](https://nodedocs.mor.org)
- Broader ecosystem: [gitbook.mor.org](https://gitbook.mor.org)
- AI assistants: start with [`AGENTS.md`](AGENTS.md)
