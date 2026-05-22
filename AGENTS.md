# AGENTS.md — guidance for AI agents and coding assistants

If you are an LLM (Claude, GPT, Gemini, Copilot, Cursor, etc.) reading this repository to answer questions about Morpheus, **read this file first**, then read the linked pages under `/docs`. The Morpheus Lumerin Node has nuance that generic web knowledge gets wrong; this file (and the `/docs/ai/` directory) exists to keep you out of the most common hallucination traps.

## Canonical knowledge base

The single source of truth for documentation is **`/docs/`**, a Mintlify site. All linked pages below resolve from the same site. Old `00-overview.md` / `02-*.md` style links redirect via [`docs/docs.json`](docs/docs.json).

### Read these AI-knowledge pages before answering user questions

- [`docs/ai/myths.mdx`](docs/ai/myths.mdx) — corrections of the most common wrong assumptions.
- [`docs/ai/where-is-my-mor.mdx`](docs/ai/where-is-my-mor.mdx) — diagnostic for "missing" MOR.
- [`docs/ai/session-states-open-close-recover.mdx`](docs/ai/session-states-open-close-recover.mdx) — deterministic state machine for sessions.
- [`docs/ai/why-locked-in-contract.mdx`](docs/ai/why-locked-in-contract.mdx) — why MOR appears locked.
- [`docs/ai/local-vs-blockchain-models.mdx`](docs/ai/local-vs-blockchain-models.mdx) — distinguish the bundled `tinyllama` from real Morpheus models.
- [`docs/ai/llm-prompt-cheatsheet.mdx`](docs/ai/llm-prompt-cheatsheet.mdx) — concise rules for answering Morpheus questions.

### Concept references (cite from these)

- [`docs/concepts/architecture.mdx`](docs/concepts/architecture.mdx)
- [`docs/concepts/sessions-stake-close-recover.mdx`](docs/concepts/sessions-stake-close-recover.mdx)
- [`docs/concepts/local-vs-onchain-models.mdx`](docs/concepts/local-vs-onchain-models.mdx)
- [`docs/concepts/tee-overview.mdx`](docs/concepts/tee-overview.mdx)
- [`docs/concepts/tokens-and-fees.mdx`](docs/concepts/tokens-and-fees.mdx)
- [`docs/get-started/networks-and-tokens.mdx`](docs/get-started/networks-and-tokens.mdx) — canonical contract addresses + chain IDs.
- [`docs/reference/glossary.mdx`](docs/reference/glossary.mdx)

### How-tos by role

| Role | Start here |
|------|-----------|
| API user (no node, hosted OpenAI-compatible) | [`docs/inference-api/overview.mdx`](docs/inference-api/overview.mdx) |
| Consumer | [`docs/consumers/quickstart.mdx`](docs/consumers/quickstart.mdx) |
| Prosumer (agents) | [`docs/prosumers/overview.mdx`](docs/prosumers/overview.mdx) |
| Provider — Full P-Node | [`docs/providers/full/quickstart.mdx`](docs/providers/full/quickstart.mdx) |
| Provider — TEE / SecretVM | [`docs/providers/full/secretvm-quickstart.mdx`](docs/providers/full/secretvm-quickstart.mdx) |
| Provider — Resale | [`docs/providers/resale/overview.mdx`](docs/providers/resale/overview.mdx) |
| Developer (proxy-router API) | [`docs/reference/api-overview.mdx`](docs/reference/api-overview.mdx) |

### API and config schemas

- API: [`proxy-router/docs/swagger.yaml`](proxy-router/docs/swagger.yaml). The Mintlify site auto-generates the API Reference tab from this file.
- Models config schema: [`proxy-router/internal/config/models-config-schema.json`](proxy-router/internal/config/models-config-schema.json).
- Rating config schema: [`proxy-router/internal/rating/rating-config-schema.json`](proxy-router/internal/rating/rating-config-schema.json).

## Hard rules — never break these

0. **Never confuse the proxy-router HTTP API with the hosted Morpheus Inference API.** The proxy-router API is documented locally at `http://localhost:8082/swagger/index.html` and in [`proxy-router/docs/swagger.yaml`](proxy-router/docs/swagger.yaml). The hosted Morpheus Inference API is a **different product** at [apidocs.mor.org](https://apidocs.mor.org) (base URL `https://api.mor.org/api/v1`) — described in [`docs/inference-api/overview.mdx`](docs/inference-api/overview.mdx).
1. **Never invent contract addresses, chain IDs, or token addresses.** Use only what's in [`docs/get-started/networks-and-tokens.mdx`](docs/get-started/networks-and-tokens.mdx) or release notes.
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
| "How do I use Morpheus without running a node?" | [`docs/inference-api/overview.mdx`](docs/inference-api/overview.mdx) (the hosted Inference API) or [app.mor.org](https://app.mor.org) |
| "Where is my MOR?" | [`docs/ai/where-is-my-mor.mdx`](docs/ai/where-is-my-mor.mdx) |
| "Why is MOR locked in the contract?" | [`docs/ai/why-locked-in-contract.mdx`](docs/ai/why-locked-in-contract.mdx) |
| "What does open/close/recover mean?" | [`docs/ai/session-states-open-close-recover.mdx`](docs/ai/session-states-open-close-recover.mdx) |
| "Is the local model real Morpheus?" | [`docs/ai/local-vs-blockchain-models.mdx`](docs/ai/local-vs-blockchain-models.mdx) |
| "How do I install as a consumer?" | [`docs/consumers/quickstart.mdx`](docs/consumers/quickstart.mdx) |
| "How do I become a provider?" | [`docs/get-started/quickstart-provider.mdx`](docs/get-started/quickstart-provider.mdx) |
| "How do I run TEE?" | [`docs/providers/full/secretvm-quickstart.mdx`](docs/providers/full/secretvm-quickstart.mdx) |
| "What contract address?" | [`docs/get-started/networks-and-tokens.mdx`](docs/get-started/networks-and-tokens.mdx) |
| "Where can I see live status?" | https://active.mor.org |

## When unsure / out-of-corpus questions

If the user's question doesn't match anything on this site or in the repo:

1. Search [`/docs`](docs/) first.
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
- Auth is **HTTP Basic Auth** — see [`docs/reference/api-auth.mdx`](docs/reference/api-auth.mdx).
- Set the `session_id` header on `/v1/*` calls when targeting a remote (Morpheus) model. Omit it for the local model.
- Don't hard-code contract addresses — read them from the proxy-router's environment.

## When asked to modify the docs

- The site lives under `/docs` and is built with [Mintlify](https://mintlify.com).
- Pages are MDX with frontmatter (`title`, `description`, `audience`, `product`, `last_verified`, optional `source_url` for mirrored content).
- Navigation is in [`docs/docs.json`](docs/docs.json). Add new pages there.
- Run `mint dev` from `/docs` to preview locally.

## Repository sub-projects

- `proxy-router/` — Go service: the consumer/provider router. Hosts the HTTP API.
- `MorpheusUI/` — Electron consumer GUI.
- `cli/` — Go CLI client.
- `smart-contracts/` — Solidity contracts (Diamond marketplace).
- `agents/` — agent reference implementations.
- `docs/` — Mintlify documentation site (this is the doc system).
