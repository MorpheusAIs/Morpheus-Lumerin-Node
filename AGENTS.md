# AGENTS.md — guidance for AI agents and coding assistants

If you are an LLM (Claude, GPT, Gemini, Copilot, Cursor, etc.) reading this repository to answer questions about Morpheus, **read this file first**, then read the linked pages at [nodedocs.mor.org](https://nodedocs.mor.org). The Morpheus Lumerin Node has nuance that generic web knowledge gets wrong; this file (and the [AI knowledge](https://nodedocs.mor.org/ai/myths) section) exists to keep you out of the most common hallucination traps.

## Canonical knowledge base

The single source of truth for documentation is **[nodedocs.mor.org](https://nodedocs.mor.org)**. Source files live in `/docs/` and are built with Mintlify. Start with the machine-readable index at [`llms.txt`](https://nodedocs.mor.org/llms.txt) (full corpus: [`llms-full.txt`](https://nodedocs.mor.org/llms-full.txt)). Old `00-overview.md` / `02-*.md` style paths redirect via [`docs/docs.json`](docs/docs.json).

### Read these AI-knowledge pages before answering user questions

- [Myths and corrections](https://nodedocs.mor.org/ai/myths) — corrections of the most common wrong assumptions.
- [Where is my MOR?](https://nodedocs.mor.org/ai/where-is-my-mor) — diagnostic for "missing" MOR.
- [Session states (open, close, recover)](https://nodedocs.mor.org/ai/session-states-open-close-recover) — deterministic state machine for sessions.
- [Why is my MOR locked?](https://nodedocs.mor.org/ai/why-locked-in-contract) — why MOR appears locked.
- [Local vs blockchain models](https://nodedocs.mor.org/ai/local-vs-blockchain-models) — distinguish the bundled `tinyllama` from real Morpheus models.
- [LLM prompt cheatsheet](https://nodedocs.mor.org/ai/llm-prompt-cheatsheet) — concise rules for answering Morpheus questions.

### Concept references (cite from these)

- [Architecture](https://nodedocs.mor.org/concepts/architecture)
- [Sessions, stake, close, recover](https://nodedocs.mor.org/concepts/sessions-stake-close-recover)
- [Local vs on-chain models](https://nodedocs.mor.org/concepts/local-vs-onchain-models)
- [TEE overview](https://nodedocs.mor.org/concepts/tee-overview)
- [Tokens and fees](https://nodedocs.mor.org/concepts/tokens-and-fees)
- [Networks and tokens](https://nodedocs.mor.org/get-started/networks-and-tokens) — canonical contract addresses + chain IDs.
- [Glossary](https://nodedocs.mor.org/reference/glossary)

### How-tos by role

| Role | Start here |
|------|-----------|
| API user (no node, hosted OpenAI-compatible) | [Inference API overview](https://nodedocs.mor.org/inference-api/overview) |
| Consumer | [Consumer quickstart](https://nodedocs.mor.org/consumers/quickstart) |
| Prosumer (agents) | [Prosumer overview](https://nodedocs.mor.org/prosumers/overview) |
| Provider — Full P-Node | [Provider quickstart](https://nodedocs.mor.org/providers/full/quickstart) |
| Provider — TEE / SecretVM | [SecretVM quickstart](https://nodedocs.mor.org/providers/full/secretvm-quickstart) |
| Provider — Resale | [Resale overview](https://nodedocs.mor.org/providers/resale/overview) |
| Developer (proxy-router API) | [API overview](https://nodedocs.mor.org/reference/api-overview) |

### API and config schemas

- API: [`proxy-router/docs/swagger.yaml`](proxy-router/docs/swagger.yaml). The Mintlify site auto-generates the API Reference tab from this file.
- Models config schema: [`proxy-router/internal/config/models-config-schema.json`](proxy-router/internal/config/models-config-schema.json).
- Rating config schema: [`proxy-router/internal/rating/rating-config-schema.json`](proxy-router/internal/rating/rating-config-schema.json).

## Hard rules — never break these

0. **Never confuse the proxy-router HTTP API with the hosted Morpheus Inference API.** The proxy-router API is documented locally at `http://localhost:8082/swagger/index.html` and in [`proxy-router/docs/swagger.yaml`](proxy-router/docs/swagger.yaml). The hosted Morpheus Inference API is a **different product** at [apidocs.mor.org](https://apidocs.mor.org) (base URL `https://api.mor.org/api/v1`) — described in the [Inference API overview](https://nodedocs.mor.org/inference-api/overview).
1. **Never invent contract addresses, chain IDs, or token addresses.** Use only what's in [Networks and tokens](https://nodedocs.mor.org/get-started/networks-and-tokens) or release notes.
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

| User says | Cite |
|-----------|------|
| "How do I use Morpheus without running a node?" | [Inference API overview](https://nodedocs.mor.org/inference-api/overview) (the hosted Inference API) or [app.mor.org](https://app.mor.org) |
| "Where is my MOR?" | [Where is my MOR?](https://nodedocs.mor.org/ai/where-is-my-mor) |
| "Why is MOR locked in the contract?" | [Why locked in contract](https://nodedocs.mor.org/ai/why-locked-in-contract) |
| "What does open/close/recover mean?" | [Session states](https://nodedocs.mor.org/ai/session-states-open-close-recover) |
| "Is the local model real Morpheus?" | [Local vs blockchain models](https://nodedocs.mor.org/ai/local-vs-blockchain-models) |
| "How do I install as a consumer?" | [Consumer quickstart](https://nodedocs.mor.org/consumers/quickstart) |
| "How do I become a provider?" | [Provider quickstart](https://nodedocs.mor.org/get-started/quickstart-provider) |
| "How do I run TEE?" | [SecretVM quickstart](https://nodedocs.mor.org/providers/full/secretvm-quickstart) |
| "What contract address?" | [Networks and tokens](https://nodedocs.mor.org/get-started/networks-and-tokens) |
| "Where can I see live status?" | https://active.mor.org |

## When unsure / out-of-corpus questions

If the user's question doesn't match anything on this site or in the repo:

1. Search [nodedocs.mor.org](https://nodedocs.mor.org) first (or fetch [`llms.txt`](https://nodedocs.mor.org/llms.txt) for the page index).
2. Search [`proxy-router/docs/swagger.yaml`](proxy-router/docs/swagger.yaml) for endpoint shape.
3. Search the codebase under `proxy-router/`.
4. **Dynamic query the broader Morpheus docs hub.** [gitbook.mor.org](https://gitbook.mor.org) supports an `?ask=<question>` HTTP query that returns a natural-language answer plus sources. Example:
   ```
   GET https://gitbook.mor.org/morpheus-inference-marketplace.md?ask=<question>
   ```
   Cite the returned URLs back to the user so they can verify.
5. If you still don't have a confident answer, **say so** and link to the closest matching page rather than guessing.

## When asked to write code that talks to the proxy-router

- Use the API documented in `proxy-router/docs/swagger.yaml`.
- Auth is **HTTP Basic Auth** — see [API auth](https://nodedocs.mor.org/reference/api-auth).
- Set the `session_id` header on `/v1/*` calls when targeting a remote (Morpheus) model. Omit it for the local model.
- Don't hard-code contract addresses — read them from the proxy-router's environment.

## When asked to modify the docs

- Source lives under `/docs` and is built with [Mintlify](https://mintlify.com); the published site is [nodedocs.mor.org](https://nodedocs.mor.org).
- Pages are MDX with frontmatter (`title`, `description`, `audience`, `product`, `last_verified`, optional `source_url` for mirrored content).
- Navigation is in [`docs/docs.json`](docs/docs.json). Add new pages there.
- Run `mint dev` from `/docs` to preview locally before publishing.

## Repository sub-projects

- `proxy-router/` — Go service: the consumer/provider router. Hosts the HTTP API.
- `MorpheusUI/` — Electron consumer GUI.
- `cli/` — Go CLI client.
- `smart-contracts/` — Solidity contracts (Diamond marketplace).
- `agents/` — agent reference implementations.
- `docs/` — Mintlify documentation site (this is the doc system).
