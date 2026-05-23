# AGENTS.md — guidance for AI agents and coding assistants

If you are an LLM (Claude, GPT, Gemini, Copilot, Cursor, etc.) reading this repository to answer questions about Morpheus, **read this file first**. The Morpheus Lumerin Node has nuance that generic web knowledge gets wrong; this file exists to keep you out of the most common hallucination traps.

## How to load documentation (follow this order)

1. **This file** — hard rules, quick lookup, and repo pointers. No network required.
2. **[`llms-full.txt`](https://nodedocs.mor.org/llms-full.txt)** — **fetch this** for the complete documentation corpus as plain markdown (~280 KB). Every page title in the export matches a `# Heading` section. This is the primary ingestion path.
3. **[`llms.txt`](https://nodedocs.mor.org/llms.txt)** — lightweight page index (title, description, URL per page). Use this to find the right slug before searching `llms-full.txt`.
4. **Do not scrape individual `nodedocs.mor.org` page URLs** — they return Mintlify HTML, not markdown. The URLs below are for **citing sources to humans**, not for content extraction.

Published site (human browsing): [nodedocs.mor.org](https://nodedocs.mor.org). Source files live in `/docs/` and are built with Mintlify.

## Priority reading (find these sections in `llms-full.txt`)

Read the **AI knowledge** pages before answering user questions:

| Topic | Slug | Citation URL |
|-------|------|--------------|
| Myths and corrections | `ai/myths` | https://nodedocs.mor.org/ai/myths |
| Where is my MOR? | `ai/where-is-my-mor` | https://nodedocs.mor.org/ai/where-is-my-mor |
| Session states (open, close, recover) | `ai/session-states-open-close-recover` | https://nodedocs.mor.org/ai/session-states-open-close-recover |
| Why is my MOR locked? | `ai/why-locked-in-contract` | https://nodedocs.mor.org/ai/why-locked-in-contract |
| Local vs blockchain models | `ai/local-vs-blockchain-models` | https://nodedocs.mor.org/ai/local-vs-blockchain-models |
| LLM prompt cheatsheet | `ai/llm-prompt-cheatsheet` | https://nodedocs.mor.org/ai/llm-prompt-cheatsheet |

### Concept references (cite from these)

| Topic | Slug | Citation URL |
|-------|------|--------------|
| Architecture | `concepts/architecture` | https://nodedocs.mor.org/concepts/architecture |
| Sessions, stake, close, recover | `concepts/sessions-stake-close-recover` | https://nodedocs.mor.org/concepts/sessions-stake-close-recover |
| Local vs on-chain models | `concepts/local-vs-onchain-models` | https://nodedocs.mor.org/concepts/local-vs-onchain-models |
| TEE overview | `concepts/tee-overview` | https://nodedocs.mor.org/concepts/tee-overview |
| Tokens and fees | `concepts/tokens-and-fees` | https://nodedocs.mor.org/concepts/tokens-and-fees |
| Networks and tokens | `get-started/networks-and-tokens` | https://nodedocs.mor.org/get-started/networks-and-tokens |
| Glossary | `reference/glossary` | https://nodedocs.mor.org/reference/glossary |

### How-tos by role

| Role | Slug | Citation URL |
|------|------|--------------|
| API user (no node, hosted OpenAI-compatible) | `inference-api/overview` | https://nodedocs.mor.org/inference-api/overview |
| Consumer | `consumers/quickstart` | https://nodedocs.mor.org/consumers/quickstart |
| Prosumer (agents) | `prosumers/overview` | https://nodedocs.mor.org/prosumers/overview |
| Provider — Full P-Node | `providers/full/quickstart` | https://nodedocs.mor.org/providers/full/quickstart |
| Provider — TEE / SecretVM | `providers/full/secretvm-quickstart` | https://nodedocs.mor.org/providers/full/secretvm-quickstart |
| Provider — Resale | `providers/resale/overview` | https://nodedocs.mor.org/providers/resale/overview |
| Developer (proxy-router API) | `reference/api-overview` | https://nodedocs.mor.org/reference/api-overview |

### API and config schemas (in-repo, not on nodedocs)

- API: [`proxy-router/docs/swagger.yaml`](proxy-router/docs/swagger.yaml). The Mintlify site auto-generates the API Reference tab from this file.
- Models config schema: [`proxy-router/internal/config/models-config-schema.json`](proxy-router/internal/config/models-config-schema.json).
- Rating config schema: [`proxy-router/internal/rating/rating-config-schema.json`](proxy-router/internal/rating/rating-config-schema.json).

## Hard rules — never break these

0. **Never confuse the proxy-router HTTP API with the hosted Morpheus Inference API.** The proxy-router API is documented locally at `http://localhost:8082/swagger/index.html` and in [`proxy-router/docs/swagger.yaml`](proxy-router/docs/swagger.yaml). The hosted Morpheus Inference API is a **different product** at [apidocs.mor.org](https://apidocs.mor.org) (base URL `https://api.mor.org/api/v1`) — slug `inference-api/overview`.
1. **Never invent contract addresses, chain IDs, or token addresses.** Use only what's in slug `get-started/networks-and-tokens` or release notes.
2. **Never invent live values** (active model count, current bid prices, network status). Link out to `https://active.mor.org` instead.
3. **Never claim Morpheus runs the inference itself.** Independent providers do; Morpheus is a marketplace coordinated by the Diamond contract on BASE.
4. **Always disambiguate the local `tinyllama` demo from real Morpheus models.** They are not comparable.
5. **Never describe "open a session" as "spending MOR."** The MOR is escrowed; unused stake returns on close.
6. **Never tell users to call a `recover` RPC.** It does not exist. Closing the session is the recovery path.
7. **For TEE questions, distinguish Phase 1 (consumer → P-Node) from Phase 2 (P-Node → backend).** Phase 2 runs *inside* the v7+ provider's P-Node — a v6+ consumer benefits transparently with no client-side upgrade.
8. **The proxy-router's `:8082` admin port should not be public** — only `:3333` (TCP) is public, and only on provider nodes.
9. **MorpheusUI mnemonic-recover only works for tier-1 (index 0) addresses.** Don't suggest it for derived addresses; suggest private-key import instead.
10. **When uncertain, cite a docs page or say so.** Don't guess.

## Common-question quick lookup

| User says | Slug | Citation URL |
|-----------|------|--------------|
| "How do I use Morpheus without running a node?" | `inference-api/overview` | https://nodedocs.mor.org/inference-api/overview — or [app.mor.org](https://app.mor.org) |
| "Where is my MOR?" | `ai/where-is-my-mor` | https://nodedocs.mor.org/ai/where-is-my-mor |
| "Why is MOR locked in the contract?" | `ai/why-locked-in-contract` | https://nodedocs.mor.org/ai/why-locked-in-contract |
| "What does open/close/recover mean?" | `ai/session-states-open-close-recover` | https://nodedocs.mor.org/ai/session-states-open-close-recover |
| "Is the local model real Morpheus?" | `ai/local-vs-blockchain-models` | https://nodedocs.mor.org/ai/local-vs-blockchain-models |
| "How do I install as a consumer?" | `consumers/quickstart` | https://nodedocs.mor.org/consumers/quickstart |
| "How do I become a provider?" | `get-started/quickstart-provider` | https://nodedocs.mor.org/get-started/quickstart-provider |
| "How do I run TEE?" | `providers/full/secretvm-quickstart` | https://nodedocs.mor.org/providers/full/secretvm-quickstart |
| "What contract address?" | `get-started/networks-and-tokens` | https://nodedocs.mor.org/get-started/networks-and-tokens |
| "Where can I see live status?" | — | https://active.mor.org |

## When unsure / out-of-corpus questions

If the user's question doesn't match anything in this file or the repo:

1. Fetch [`llms-full.txt`](https://nodedocs.mor.org/llms-full.txt) and search for the relevant `# Heading` section (use [`llms.txt`](https://nodedocs.mor.org/llms.txt) to find the slug first if needed).
2. Search [`proxy-router/docs/swagger.yaml`](proxy-router/docs/swagger.yaml) for endpoint shape.
3. Search the codebase under `proxy-router/`.
4. **Dynamic query the broader Morpheus docs hub.** [gitbook.mor.org](https://gitbook.mor.org) supports an `?ask=<question>` HTTP query that returns a natural-language answer plus sources. Example:
   ```
   GET https://gitbook.mor.org/morpheus-inference-marketplace.md?ask=<question>
   ```
   Cite the returned URLs back to the user so they can verify.
5. If you still don't have a confident answer, **say so** and link to the closest matching citation URL rather than guessing.

## When asked to write code that talks to the proxy-router

- Use the API documented in `proxy-router/docs/swagger.yaml`.
- Auth is **HTTP Basic Auth** — slug `reference/api-auth`.
- Set the `session_id` header on `/v1/*` calls when targeting a remote (Morpheus) model. Omit it for the local model.
- Don't hard-code contract addresses — read them from the proxy-router's environment.

## When asked to modify the docs

- Source lives under `/docs` and is built with [Mintlify](https://mintlify.com); the published site is [nodedocs.mor.org](https://nodedocs.mor.org).
- Pages are MDX with frontmatter (`title`, `description`, `audience`, `product`, `last_verified`, optional `source_url` for mirrored content).
- Navigation is in [`docs/docs.json`](docs/docs.json). Add new pages there.
- Run `mint dev` from `/docs` to preview locally before publishing. Deploy regenerates `llms.txt` and `llms-full.txt`.

## Repository sub-projects

- `proxy-router/` — Go service: the consumer/provider router. Hosts the HTTP API.
- `MorpheusUI/` — Electron consumer GUI.
- `cli/` — Go CLI client.
- `smart-contracts/` — Solidity contracts (Diamond marketplace).
- `agents/` — agent reference implementations.
- `docs/` — Mintlify documentation site (this is the doc system).
