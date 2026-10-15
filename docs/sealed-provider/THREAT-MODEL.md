# Threat model

Draft design. The promise below is conditional, and several of its conditions are not yet met. Read [STATUS-AND-KNOWN-ISSUES.md](STATUS-AND-KNOWN-ISSUES.md) together with this file.

## Adversary

The primary adversary is the provider's own operator: a person with administrator and root rights on the Mac that serves the model, who wants to read consumers' prompts and answers. This includes any process running as the engine's own user, because on a Mac the operator and the engine normally share one user.

The adversary may do the following: run any process as root or as the engine's user; list a hardened process's file descriptors and their paths; send it signals; read, create and replace files in directories the engine's user can write; connect to sockets in those directories; race the engine's spawn sequence; observe all network traffic; install system software that is allowed on a Mac at Full Security with SIP on.

Relays and network observers are a second adversary class. They see ciphertext, request sizes and timing.

Not assumed of the adversary: a kernel vulnerability, a break in Apple's platform, or physical attacks on the hardware. Kernel memory reads are excluded by assumption A-8 (known interfaces measured; unknown ones accepted as trust in Apple). Physical attacks are not analyzed in the sources this RFC draws on.

## What is promised

Under the conditions below, a consumer's prompt and the model's answer are readable only inside code whose hashes the consumer pinned. Specifically, plaintext exists only in the user memory of the pinned engine, the user memory of its pinned children (GPU-visible buffers are pages of the child's own address space on Apple silicon, which is assumption A-16), and the kernel buffers of the local sockets between them. It never reaches a file, a log, a crash report, or a socket another process can reach.

The conditions are: SIP enabled; Full Security boot; a (Mac model, macOS build) pair on which the assumptions were measured to hold; and the pinned children free of exploitable memory-safety bugs. On a pair that has not been measured, the specification offers serving only under an "untested" label and makes no promise.

The answer comes from weights whose hash the model's on-chain listing declares, and the consumer receives that hash inside the attested session. For app-verified models the hash also equals one compiled into the app. This shows the weights match what the model's owner declared. It does not show they are a particular model the consumer reviewed.

The consumer-facing label "TEE" here names attested, pinned-process isolation (App Attest, the hardened runtime and R1-R11). It does not name a hardware enclave that holds the plaintext.

## What is not promised (stated residuals)

- **One child serves many sessions.** A consumer who exploits a memory-safety bug in a pinned child through crafted input runs code in a process that serves other consumers' sessions of the same listing. The specification chose this over one child per consumer, which was rejected for its memory cost. The consumer-visible session description states it.
- **Timing and size are visible.** The provider operator and relays see request sizes and timing by observing traffic. The engine and children emit no size, length or timing record of their own.
- **Memory-safety of the pinned children.** The promise for any session depends on the pinned code. Splash and mlx-serve are third-party programs of substantial size and have not been audited for memory safety by this work.
- **The code signer is currently the operator.** The engine is signed by the party that builds it. Until reproducible builds (R11) let a second party rebuild and match the published binary, the pin is trust in the pin publisher (assumption A-10). Each child pin also rests on the operator's own build until a second party reproduces it.
- **Plaintext in local kernel socket buffers.** Engine-to-child traffic is not encrypted (decision D1). This rests on assumption A-8 and A-4 and is revisited if a way to read local socket contents without reading process memory is found.
- **Effective context length is not attested.** Splash's memory budget is computed from the machine, so a consumer cannot verify it. Weights and confidentiality are unaffected.
- **Unknown kernel read paths** are accepted as trust in Apple (A-7, A-8).

## Assumptions A-1 to A-21

Status uses two words. "Measured" means a measurement exists and its scope is stated. "Open" means relied on without a measurement that has a working control, or not yet run. Where the specification says an open assumption "blocks", the release does not ship, or a pair cannot be labelled tested, until it is measured. At the time of writing, no (Mac model, macOS build) pair has had every blocking assumption measured, so no pair is labelled tested.

| # | Assumption | Status |
|---|---|---|
| A-1 | Root cannot get a task port to a hardened process without `get-task-allow`, except through Apple-signed tools holding a task-port entitlement | Open: relied on since the first generation; per-build re-check pending |
| A-2 | dtrace offers no read path under SIP | Measured for attach (three root runs on the test Mac, 2026-09-29, with an unhardened control that attaches); read control pending, so not counted as passed |
| A-3 | Swap is encrypted under a key root cannot use and cannot be switched off while running | Measured for the first part (`vm.swapusage` reports encrypted); the other parts are open |
| A-4 | Local socket contents are not exposed to Endpoint Security, content filters, BPF, kdebug or `proc_info` | Open: believed; blocks the tested label; load-bearing for D1 |
| A-5 | No unencrypted hibernation image holds memory | Open: unmeasured; blocks the tested label |
| A-6 | The Metal compiler service cannot be shadowed through the bootstrap namespace | Open: unsettled; blocks shipping any child that compiles Metal at runtime |
| A-7 | Apple's system services, system-volume libraries, Secure Enclave and build toolchain are out of scope | Trust anchor, accepted rather than measured |
| A-8 | Root cannot read kernel memory with SIP on and a Full Security boot through any known interface | Open: relied on; known interfaces checked through dtrace (read control pending); unknown interfaces accepted as trust |
| A-9 | A kernel panic writes no user memory or in-flight socket buffers to disk in a form root can use | Open: unmeasured; blocks the tested label |
| A-10 | The pinned engine hash is a build of the audited engine source | Open: trust in the pin publisher; closed only by R11 second-party reproduction |
| A-11 | SIP is on and the Mac booted with Full Security | Partly measured: App Attest refused non-Full boots 4 of 4 on the first-generation engine; per-model re-measure and the engine's SIP check are open |
| A-12 | Withdrawn in a spec revision (fallback for A-13) | Withdrawn |
| A-13 | A kernel-enforced launch requirement can name an ad-hoc code-directory hash and the kernel refuses any other image | Measured on the test Mac (macOS 27.0), 2026-09-30: another hash and a file swapped after hashing were refused; the matching hash ran. Re-measure per pair; relies on an undeclared system-library function |
| A-14 | Local-peer audit token semantics, stale-token refusal, spawn-time exception ports and descriptor listing behave as the identity checks need | Partly measured: token read at read time, stale-token refusal and descriptor listing measured 2026-09-30; a dead peer yields no token (measured 2026-10-01); the rest relied on |
| A-15 | A self launch constraint makes the kernel refuse any launch other than by launchd in a user's GUI domain | Open: unmeasured; blocks shipping |
| A-16 | Root cannot read or write another process's GPU-visible memory through IOKit or accelerator clients without an Apple-granted entitlement | Open: unmeasured; blocks the tested label |
| A-17 | Root cannot loosen a running process's sandbox through extensions or launchd grants | Open: unmeasured; blocks shipping |
| A-18 | Only a model's owner or an authorized delegate can set its listing's content identifier | Relied on; read from contract source; re-check after every contract upgrade |
| A-19 | Every App Attest assertion carries the OS-measured code-directory hash of the producing binary, and Apple refuses non-Full boots | Relied on; re-check quarterly and after every macOS update |
| A-20 | A child's GPU faults and resource kills produce no report carrying prompt-derived bytes | Open: unmeasured; blocks the tested label |
| A-21 | macOS runs on bare metal and the sealed system volume cannot be replaced by root at Full Security | Relied on; engine refuses under a hypervisor (specified, not built) |

The specification limits what a provider's own checks prove: a provider-run check protects an honest provider from mistakes and proves nothing to a consumer. What a consumer relies on is what the pinned engine enforces itself plus what the design team measured per Mac model and macOS build.

## Same-user attacker: arms measured so far

The adversary arms written so far test the spawn module against a process with the engine's own uid. The results are in [STATUS-AND-KNOWN-ISSUES.md](STATUS-AND-KNOWN-ISSUES.md). They have not yet been run as root, so the root half of the adversary is untested.
