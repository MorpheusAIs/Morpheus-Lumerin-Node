# Morpheus Cowork

Morpheus Cowork is a local-first, project-scoped agent workspace inside
MorpheusUI. It can plan a task, read and update files in one folder selected by
the user, retain task history, request approval for consequential actions, and
run local schedules while the desktop app is available.

This is an original Morpheus implementation inspired by the workflow described
in the Cowork product guide. The guide is reference material, not executable
instructions or a specification that overrides this repository's security
rules.

Morpheus Cowork is **not a feature-identical copy of Anthropic Cowork**. In
particular, it does not contain Anthropic's proprietary models, cloud task
infrastructure, browser, computer-use bridge, connectors, plugin marketplace,
mobile dispatch, or isolated execution VM. The exact parity status is recorded
in [Feature parity](#feature-parity).

## Architecture

| Layer                                | Responsibility                                                                                                                                                                       |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| React route (`/cowork`)              | Three-pane projects/tasks UI, schedules, plans, artifacts, activity, approvals, and extension discovery                                                                              |
| Narrow preload API (`window.cowork`) | Fixed operations only; the renderer cannot choose an IPC channel or provide an arbitrary filesystem root                                                                             |
| Main-process Cowork IPC              | Validates the renderer origin and request values, opens the native folder picker, and removes private fields before returning records                                                |
| Project/task stores                  | Durable NeDB records under Electron's application user-data directory, separate from the disposable legacy cache                                                                     |
| Agent runner                         | Calls the exact active marketplace session through the local proxy-router and executes a bounded allowlist of Cowork tools                                                           |
| File tools                           | Enforce project-relative paths, size/count limits, secret-file exclusions, approvals, backups, and trash-based deletion                                                              |
| Research and artifact tools          | Fetch one approved public HTTPS page at a time, analyze bounded CSV/TSV data, safely extract text from existing PDF/DOCX/XLSX/PPTX files, and generate native files in those formats |
| Scheduler                            | Evaluates local IANA-time-zone recurrences and creates a new durable task for each claimed occurrence                                                                                |
| Extension catalog                    | Reads strictly validated project guidance, instruction-only skills, and inert remote MCP descriptors                                                                                 |

The runner is in the Electron main process. It has **no shell or VM**. It can
only invoke the fixed tools declared in `cowork-runner.ts`: planning; bounded
file listing, inspection, reading, searching, document-text extraction, and CSV
analysis; text and native document creation; directory creation;
copy/move/trash; approved public-HTTPS retrieval; read-only delegated analysis;
and task completion.

## Set up and use

From the repository root:

```bash
cp ui-desktop/.env.example ui-desktop/.env
cd ui-desktop
yarn install
yarn dev
```

Legacy installations may retain read-only chat history from the local TinyLlama
demonstration model. TinyLlama cannot start a new Chat or Cowork run and is not
representative of marketplace model quality.

To start a task:

1. Unlock or create the MorpheusUI wallet.
2. Open **Chat** and choose a marketplace LLM model.
3. Select the session duration, then either stake MOR or use one-off **Direct
   Pay** to open an on-chain peer-to-peer session. Wait for that exact session to
   appear as active before continuing.
4. Use the active session in normal **Chat**, or open **Cowork** and select that
   same session for agent work.
5. In Cowork, create a project and choose a dedicated working folder in the
   native folder picker. Avoid a home directory, wallet directory, or broad
   shared drive.
6. Add optional project instructions and start with **Manual** approval mode.
7. Describe a concrete deliverable and create the task with the already-active
   session.
8. Review the provider data-boundary notice and any approval card before allowing
   work to continue.
9. Follow the plan, activity, transcript, and artifact panels. A task can be
   paused, cancelled, resumed, or redirected with a follow-up instruction.

MorpheusUI has no subscriptions, plans, recurring billing, hosted entitlements,
or automatic renewals. Access comes from an on-chain P2P session with an
independent provider through the open-source Morpheus proxy/P2P flow, not a
hosted SaaS entitlement. For a stake-funded session, MOR is escrowed rather than
simply spent when the session opens, and unused stake is returned when the
session closes. **Direct Pay** is a one-off payment path for the selected
session duration; it is not a subscription and does not renew.

Chat owns model selection and the blockchain session lifecycle. Cowork neither
opens nor renews a session, stakes MOR, initiates Direct Pay, switches providers,
nor extends the expiry. If the exact selected session closes, expires, or
disappears, new Cowork work is blocked until the user explicitly opens and
selects another session in Chat. Close a stake-funded session in Chat to return
its unused stake; there is no separate `recover` RPC.

## Projects, tasks, and local data

A project stores:

- the canonical path selected through the native picker;
- its display name and user-entered instructions;
- an approval mode;
- enabled folder-guidance and skill IDs;
- durable task and schedule history.

The renderer receives the folder's display name, not its absolute path. Removing
a project from Cowork soft-archives its record, cancels active work, and pauses
its active schedules. It does not delete the connected folder or erase its task
history. Deleting an individual task removes its local record; it does not
delete artifacts already written to the project folder.

Tasks retain their visible messages, plan, activity, artifacts, model target,
and internal model conversation. Running work is marked paused if the desktop
process exits. Pending approvals remain reviewable. Recent project memory is a
bounded set of up to five completed-task summaries; it is model-written,
fallible context rather than a source of truth.

Resource limits keep persistent work bounded: a project can retain up to 500
tasks and 100 schedules; at most four tasks run across the app and two within
one project. Visible and model histories are capped by both entry count and
serialized size. The task rail loads lightweight summaries and fetches the full
model transcript only for the opened task.

Cowork databases are created with user-only permissions where the operating
system supports them:

```text
<Electron userData>/Cowork/projects.db
<Electron userData>/Cowork/tasks.db
<Electron userData>/Cowork/schedules.db
```

These files can contain prompts, model responses, task summaries, filenames,
and excerpts returned by tools. They are local but **not application-level
encrypted**. OS account security and full-disk encryption remain important.

## Performance and responsiveness

Cowork and every legacy top-level screen are route-split, so opening the app no
longer downloads and evaluates every screen up front. In the production build,
the initial renderer JavaScript fell from 6,314,805 bytes to 2,762,811 bytes
(56.25% smaller, 620,363 bytes gzip); Cowork is a separate 110,640-byte chunk.
Password-strength code loads only with onboarding.

Chat responses stream incrementally through bounded main-process IPC rather
than waiting for a complete response. Model downloads keep real progress and
true cancellation without exposing credentials or folder paths. Audio and
document attachments cross IPC as bounded binary buffers instead of amplified
base64 strings. Streaming chat updates are animation-frame batched and only
auto-scroll while the reader is near the bottom.

Cowork's rail queries only task summaries; the full transcript is fetched for
the selected task. Task events are animation-frame coalesced, keep only the
newest delta for each task, update the already-sorted rail in place, and
coalesce selected-task refreshes. Indexed task/project/schedule lookup keys keep
those queries bounded as histories grow. Recent memory uses a projected
five-record query, history sizes are tracked incrementally, transcript rows are
memoized, long rows use deferred rendering, and large task-store compaction is
spaced away from the active-run hot path. Chat and Cowork follow new output only
while the reader is near the bottom, and an unmounted Chat cancels its active
stream directly.

These are build/static and automated-behavior measurements, not Chrome trace
timings. Runtime Core Web Vitals were not measured because the required Chrome
trace integration was unavailable in this environment.

## File tools and approval modes

| Mode   | New files/directories/copies | Overwrites and moves                             | Deletion       |
| ------ | ---------------------------- | ------------------------------------------------ | -------------- |
| Manual | Ask every time               | Ask every time                                   | Ask every time |
| Auto   | Allowed without a prompt     | Ask before replacing an existing path; moves ask | Ask every time |
| Skip   | Allowed without a prompt     | Allowed without a prompt                         | Ask every time |

Skip is project-scoped consent for file changes, not unrestricted computer
access. It never bypasses deletion approval or the separate off-device data
approval. Auto is a deterministic file-policy mode; it does not yet run a
second model-based safety review.

Current file behavior includes:

- relative paths inside the connected project only;
- canonical-path and symlink escape checks;
- blocking hard-linked files and common credential/key paths;
- bounded UTF-8 reads, listings, and searches;
- bounded, inert text extraction from PDF, DOCX, XLSX, and PPTX sources, with
  encrypted documents, active OOXML content, external relationships, embedded
  objects, and suspicious archives rejected;
- text writes capped at 2 MiB;
- content-only professional DOCX, XLSX, PPTX, and PDF generation capped at 32
  MiB, with no macros, formulas, scripts, external relationships, images, or
  network-loaded resources;
- backups before replacing regular files, retained under
  `<Electron userData>/CoworkBackups/<project-id>`;
- backup retention capped at 100 files and 512 MiB per project, with individual
  backup/copy operations capped at 64 MiB;
- deletion through the operating system trash;
- an immutable maximum of 30 agent steps per run.

The denylist is defense in depth, not a secret scanner. Do not put private keys,
seed phrases, passwords, tokens, or sensitive account exports in a connected
project.

## Models, sessions, data boundaries, and vision labels

Chat is the only model-selection and session-opening surface. It lists
marketplace LLMs, lets the user select a duration and payment path, and opens the
on-chain P2P session. Cowork then lists only the wallet's active, unexpired
marketplace sessions and uses the exact chosen session ID and model. It cannot
use TinyLlama, another local model, or an arbitrary configured endpoint. Legacy
local-demo chat history may remain visible, but it is read-only.

The Chat model picker includes a **Cowork candidates** filter for marketplace
text/chat models. It excludes local, deleted, audio, and embedding models and
combines with model-name/tag search. Candidates are explicitly shown as
unverified: providers do not publish native-tool support, compatibility can be
checked only after a session opens, and models that reject native tools may use
the bounded compatibility protocol described below. The filter is omitted from
the dedicated Cowork setup picker because that view is already restricted to
text/chat candidates.

Every selected Cowork session routes inference to an independent Morpheus
provider. The provider-sharing approval covers the task instructions,
user/project guidance, enabled instruction-only skills, recent task summaries,
and project content returned by approved tools. Approval is bound to a
fingerprint of the exact marketplace session and model; a destination change
invalidates it and requires fresh consent. Scheduled off-device work creates a
waiting task rather than silently sending data.

Morpheus coordinates the marketplace; independent providers perform remote
inference. Review their privacy and operational properties before sharing
sensitive business material.

Models appear in three explicit groups: **Vision (declared)**, **Possible
vision (name match)**, and **Text / vision not declared**. A declaration comes
from recognized model metadata; a possible match is only a name heuristic.
Neither label is an end-to-end compatibility guarantee. Cowork does not yet
attach images to its task messages, so the categories improve discovery but do
not imply image input in Cowork.

The exact declared tags currently recognized are `vision`, `multimodal`,
`image`, and `vlm` (case-insensitive). The possible-vision name fragments are
`llava`, `vision`, `gpt-4o`, `gpt-4-turbo`, `claude-3`, `claude-4`,
`claude-sonnet`, `claude-opus`, `gemini`, `qwen-vl`, `qwen2-vl`,
`qwen2.5-vl`, `internvl`, `minicpm-v`, `pixtral`, `molmo`, `phi-3-vision`,
`phi-4-multimodal`, `idefics`, and `cogvlm`. The models shown inside each group
are the exact models attached to the wallet's active marketplace sessions; no
static list can represent that live inventory.

Native tool-calling support is not reliably declared in the marketplace model
record or provider health report. Cowork therefore starts with standard OpenAI
`tools` and omits the redundant optional `tool_choice` and
`parallel_tool_calls` fields. If—and only if—a bounded client-fault response
explicitly says that `tools` is unsupported, the rejected pre-action request is
retried once with the versioned `morpheus-cowork-v1` text tool protocol. That
protocol requires the model's entire response to be one exact JSON tool or final
envelope. Cowork generates the call ID, validates the fixed allowlist, argument
shape and byte limits, and applies the same folder, approval, network, and
mutation gates before executing anything. Prose, code fences, embedded JSON,
unknown tools, arrays, or extra envelope fields cannot authorize an action.

Compatibility mode is stored only for the exact session/model fingerprint and
is cleared when that destination changes. Persisted native tool history is
translated to plain assistant/user JSON envelopes so a backend that rejects the
OpenAI `tool` role is not sent one later. Generic `400` responses, authentication
or rate-limit errors, server failures, timeouts, and aborts never trigger an
automatic retry. A model that supports normal Chat but cannot follow either
native tools or the strict compatibility protocol remains unsuitable for
Cowork; the task fails without executing model-requested actions.

Every mutating file action is recorded durably as prepared before its side
effect starts, then updated with its succeeded or failed result. Reused tool-call
IDs replay that durable result without repeating the action; IDs reused with
different arguments are blocked. Arguments are restricted to the declared tool
fields and mutation identities use canonical hashes. If the app or a task save
fails after preparation but before the outcome is durable, runtime/startup
recovery marks the action ambiguous, closes unresolved tool calls, scrubs large
write payloads, and pauses the task. Recovery verifies the saved tool name and
arguments hash before replaying a result. A task with an ambiguous action keeps
every later mutation approval-gated, even in Skip mode. Recovery finishes
before IPC task actions or schedules can start.
Cowork never guesses that an ambiguous action failed and never silently replays
it.

## Project instructions and skills

Project instructions entered in the creation form are always included as
user-authored project context. A folder can additionally declare guidance under
`.morpheus/cowork/`; discovered folder guidance and skills are off until the
user enables them in the **Extensions** panel.

### Folder guidance

```text
<project>/.morpheus/cowork/instructions.md
```

The file must be valid UTF-8 and no larger than 64 KiB.

### Instruction-only skill

Each skill has a strictly validated manifest and a referenced instruction file:

```text
<project>/.morpheus/cowork/skills/research-memo/skill.json
<project>/.morpheus/cowork/skills/research-memo/SKILL.md
```

Example `skill.json`:

```json
{
  "schemaVersion": 1,
  "id": "research-memo",
  "name": "Research memo",
  "description": "Prepare a source-grounded research memo.",
  "instructionsFile": "SKILL.md",
  "capabilities": ["project-read", "artifact-create"],
  "enabled": false
}
```

Supported capability declarations are `project-read`, `project-write`,
`artifact-create`, `process-execution`, and `network-access`. They are risk
metadata only. A skill cannot add a tool, execute a script, invoke the fixed web
tool by itself, or override the system prompt and approval policy. Even
`enabled: true` in a
manifest is only an activation request visible in discovery; the user must
enable the skill in the project UI.

Catalog limits include 16 KiB per manifest, 64 KiB per instruction file, 64
skill-directory entries, and 512 KiB of discovered instruction text. A project
can enable at most eight skills and 128 KiB of combined folder/skill guidance
for a task. Enabling guidance stores SHA-256 hashes of the exact instruction
content reviewed by the user; modified guidance is disabled until it is
reviewed and enabled again.

## Public web retrieval and data analysis

`fetch_web_page` retrieves one user-approved public HTTPS document without
cookies, credentials, referrers, script execution, decompression, or browser
state. The approval shows the complete URL. Model-supplied query strings are
rejected, redirects cannot leave the approved origin, DNS is resolved and
pinned for every hop, and private/link-local/metadata destinations are blocked.
The total request is capped at 30 seconds, five same-origin redirects, 1 MiB of
response bytes, and 250,000 extracted text characters. Only HTML, plain text,
and JSON are accepted.

Web content is untrusted model input, not executable application content. The
tool does not provide search ranking, authenticated browsing, interactive
click/type automation, or automatic source citations. For defensible research,
give the task canonical source URLs and ask it to identify those URLs in the
artifact.

`analyze_csv` reads a bounded project CSV or TSV file and returns deterministic
column, type, missing-value, distinct-value, and numeric summary statistics.
It is analysis rather than a spreadsheet calculation engine: generated XLSX
files intentionally contain values only, not formulas or macros.

`read_document` extracts bounded text and structural metadata from existing
PDF, DOCX, XLSX, and PPTX files. Extracted text is treated as untrusted evidence,
never as executable instructions. Artifact previews use the same parser. This
is not a visual document renderer or a full-fidelity editing/versioning system.

## Remote MCP connector descriptors

The main process creates this application-owned directory:

```text
<Electron userData>/CoworkExtensions/
```

A user may place a `connectors.json` descriptor there:

```json
{
  "schemaVersion": 1,
  "connectors": [
    {
      "schemaVersion": 1,
      "id": "knowledge-service",
      "name": "Knowledge service",
      "description": "Company-approved remote MCP endpoint.",
      "transport": "streamable-http",
      "url": "https://mcp.example.com/mcp",
      "authentication": "oauth2",
      "capabilities": ["remote-read"],
      "enabled": false
    }
  ]
}
```

Transports are `streamable-http` or `sse`; authentication metadata is `none` or
`oauth2`; capabilities are `remote-read` and `remote-write`. The file is capped
at 128 KiB and 32 descriptors. Unknown fields—including embedded headers or
tokens—are rejected. Endpoints must use HTTPS, a DNS hostname, and no embedded
credentials, query, or fragment.

This is **discovery only**. Connector descriptors remain disabled and
unconnected. The app performs no MCP handshake, OAuth flow, token storage,
endpoint request, or connector action. Any future activation layer must add
DNS-resolution and redirect validation, authenticated secret brokering,
per-tool scopes, user consent, response limits, and audit logging.

## Local schedules

Cowork supports manual, hourly, daily, weekly, and weekday cadences with an
explicit IANA time zone. Schedules can be created, paused, resumed, deleted, or
run immediately. Daylight-saving gaps move to the first valid later minute;
ambiguous fall-back times select their first occurrence.

Each occurrence creates a separate durable Cowork task. The scheduler:

- runs only while MorpheusUI and its local services are running;
- checks for work in the Electron main process;
- collapses multiple missed occurrences into one bounded catch-up run;
- does not replay an occurrence that may have started before an app crash;
- requires the schedule's exact marketplace session to still be active at run
  time and pauses the schedule if it has closed, expired, or disappeared;
- never opens or renews a session, stakes MOR, initiates Direct Pay, changes
  models/providers, or extends an expiry;
- aborts a claimed occurrence when its schedule is paused or deleted;
- leaves off-device tasks waiting for explicit data-sharing approval.

The schedule's last-run record means the occurrence was claimed and a task was
created or start was requested. Review the linked task for actual completion,
approvals, and errors. There is no cloud scheduler, wake-from-sleep service, or
OS notification delivery.

## Security boundaries and remaining risks

Implemented hardening includes a fixed Cowork IPC surface, renderer-origin
checks, hidden absolute project paths, validated request fields, loopback binding
for the proxy-router admin API, guarded external navigation, and an
environment-specific Content Security Policy. DevTools no longer opens
automatically; development shortcuts and developer extensions remain available
in development builds.

The legacy renderer no longer receives the proxy-router Basic-auth credential.
Its proxy operations cross bounded main-process handlers, sensitive blockchain
actions receive native confirmation, wallet onboarding ignores renderer-supplied
destinations, and production CSP blocks direct renderer connections to loopback
services.

Important limitations:

- Electron renderer sandboxing, context isolation, and disabled Node integration
  are enabled. Cowork file enforcement is still an application-layer boundary,
  not an isolated per-task OS filesystem or code-execution VM.
- Model prompt-injection resistance is not a security boundary. Enforcement
  comes from the fixed tool allowlist, path checks, approvals, and data-boundary
  gates.
- The legacy non-Cowork renderer still uses an allowlisted request/response IPC
  bridge, while Cowork uses the narrower `window.cowork` API. Raw Electron event
  objects and proxy-router credentials are not exposed to either renderer API.
- Local databases are permission-restricted but not encrypted by the app.
- A model can make mistakes inside the permitted project scope. Preserve
  originals, review approvals, and keep independent backups.
- Path checks reject symlinks and hard-linked files, but Node's pathname-based
  filesystem API cannot provide an atomic OS sandbox. Another process running
  as the same OS user could race an ancestor-directory replacement between a
  check and an operation. Use a dedicated folder and do not run untrusted local
  software alongside Cowork for sensitive work.
- Task data leaves the computer after provider-sharing approval. An active
  marketplace session does not make its independent provider equivalent to
  on-device processing.
- Descriptor validation does not make an MCP endpoint trustworthy. Connector
  execution is intentionally absent.

## Feature parity

| Capability                                        | Status          | Current Morpheus behavior                                                                                                                                                                    |
| ------------------------------------------------- | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Project workspaces and connected-folder scope     | Implemented     | Native folder grant, durable projects, project instructions, archive behavior                                                                                                                |
| Persistent task history                           | Implemented     | Durable statuses, transcript, plan, activity, artifacts, pause/restart recovery                                                                                                              |
| Visible plans and progress                        | Implemented     | Model-managed plan steps and main-process activity records                                                                                                                                   |
| Task steering, pause, resume, cancel              | Implemented     | Follow-up instructions interrupt safely and re-enter shared start policy                                                                                                                     |
| File-action approval modes                        | Implemented     | Manual, Auto, and Skip with deletion always gated                                                                                                                                            |
| Off-device data consent                           | Implemented     | Consent is required before first model call and rebound if destination changes                                                                                                               |
| Scoped file tools                                 | Partial         | Bounded text/filesystem operations and safe document-text extraction; no arbitrary binary editing or full-fidelity document edit/render pipeline                                             |
| Professional DOCX, XLSX, PPTX, and PDF creation   | Implemented     | Strict content-only schemas, native binary generation, bounded output, and overwrite backups                                                                                                 |
| CSV/TSV analysis                                  | Implemented     | Deterministic bounded schema, quality, categorical, and numeric summaries; no formula engine or chart authoring                                                                              |
| Static artifacts                                  | Partial         | Text and safely extracted PDF/DOCX/XLSX/PPTX content are previewed in-app and files can be revealed in the OS; no visual document render, interactive live artifact, or version restore      |
| Local scheduled tasks                             | Partial         | Durable local cadences and run-now; requires the app to remain available and is not a cloud scheduler                                                                                        |
| Project memory                                    | Partial         | Up to five recent completed-task summaries; no semantic memory, global memory, or cross-device sync                                                                                          |
| Instruction-only project skills                   | Partial         | Strict discovery and explicit enablement; no scripts, packaged resources, automatic skill marketplace, or cross-surface distribution                                                         |
| Plugins, commands, and hooks                      | Not implemented | No installable workflow bundle, hook runtime, plugin marketplace, or organization-managed distribution                                                                                       |
| Remote MCP connector catalog                      | Partial         | Strict descriptor discovery and risk display only                                                                                                                                            |
| Authenticated MCP execution and connector actions | Not implemented | No OAuth flow, credentials, tool discovery, connector calls, writes, or sends                                                                                                                |
| Read-only delegated analysis                      | Partial         | Up to three bounded model workstreams; no persistent tool-using sub-agent processes                                                                                                          |
| Session-first model access                        | Implemented     | Chat selects the marketplace LLM, duration, and one-off funding path; Cowork uses only an exact active P2P session                                                                           |
| Native/non-native tool-call compatibility         | Implemented     | Standard OpenAI tools first; exact unsupported-tools rejection switches the same session to a strict, locally validated JSON protocol                                                        |
| Mutation replay protection                        | Implemented     | Durable prepared/result journal, canonical action identities, explicit approval after ambiguity, duplicate-ID rejection, and recovery-before-run ordering                                    |
| Subscriptions and hosted entitlements             | Not applicable  | No plans, subscriptions, recurring billing, hosted Cowork entitlement, or auto-renewal; access is through user-opened on-chain sessions                                                      |
| Vision-model categorization                       | Implemented     | Declared tags or a clearly marked name heuristic                                                                                                                                             |
| Cowork image or attachment inputs                 | Not implemented | Cowork can extract text from supported project documents but does not send image inputs or arbitrary chat attachments                                                                        |
| Public web research                               | Partial         | Approved one-page HTTPS retrieval with exact-target, DNS, same-origin redirect, and private-address checks; no search engine, automatic citations, signed-in session, or interactive browser |
| Computer use                                      | Not implemented | No screenshot/click/type automation or per-application permission system                                                                                                                     |
| Isolated shell, code execution, or VM             | Not implemented | No shell tool and no isolated execution runtime                                                                                                                                              |
| Cloud execution and schedules                     | Not implemented | Work depends on the local desktop and services                                                                                                                                               |
| Mobile dispatch and cross-device sync             | Not implemented | No mobile queue, cloud handoff, or synchronized projects/tasks                                                                                                                               |
| Live artifacts                                    | Not implemented | No refreshable data-backed UI artifact or sharing link                                                                                                                                       |
| Project/session sharing                           | Not implemented | Projects and task transcripts are local to this desktop profile                                                                                                                              |
| Enterprise compliance/admin controls              | Not implemented | No policy console, retention API, organization audit export, or group controls                                                                                                               |
| Anthropic proprietary infrastructure              | Not implemented | No claim is made to reproduce Anthropic's private services or product internals                                                                                                              |

## Verification and development notes

Run the repository verification from the repository root:

```bash
./verify.sh
```

The automated suite covers the proxy-router build, vet and tests; desktop unit
tests and TypeScript checks; the production Electron build; and unsigned package
creation. Before release, also manually verify:

- create/archive project and restart persistence;
- exact-active-session task, plan, file write, artifact preview, and follow-up
  steering;
- each approval mode, overwrite backup, and trash deletion;
- marketplace-provider data-sharing approval, including a changed, closed, or
  expired exact session;
- Chat-first model selection, duration selection, stake and one-off Direct Pay
  opening, then choosing normal Chat or Cowork without any automatic renewal;
- pause/cancel during a model request and restart recovery;
- schedule creation, pause/resume, run-now, DST behavior, and app restart;
- extension discovery, invalid manifests, explicit guidance enablement, and the
  inert connector display;
- packaged CSP, navigation guards, and no automatically opened DevTools;
- macOS, Windows, and Linux folder/path behavior.

Key implementation files are under `src/main/src/client/cowork-*`,
`src/preload/index.ts`, and `src/renderer/src/components/cowork/`.
