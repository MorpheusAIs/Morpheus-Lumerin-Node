# Credits

This proposal composes work that other people built. Each entry names the project, links it, and says what its authors made that this design depends on. The contribution here is to combine these pieces so that a Mac provider can serve open-weight models inside an attested, hardened process, and to do the engineering and measurement that makes that combination checkable. It does not replace any of them.

## Morpheus-Lumerin-Node

[github.com/MorpheusAIs/Morpheus-Lumerin-Node](https://github.com/MorpheusAIs/Morpheus-Lumerin-Node) is the project this RFC is addressed to. Its contributors built the desktop app and its provider wizard, the proxy-router that opens and serves sessions, the on-chain provider, model and bid registry and session contracts, and the payment and staking model. The sealed lane described here is an addition to that router, and the one-click flow is an addition to that wizard. The project's own TEE work for servers (Secret VM on AMD SEV) established the idea of attested providing on this network; this proposal is its counterpart for Apple silicon, and the consumer-side verification here follows the same shape of challenge, attestation and pinned identity.

## MLX

[MLX](https://github.com/ml-explore/mlx) was built by Apple's machine learning research team and the ml-explore contributors. It is the array framework and Metal backend that the Apple-silicon inference engines used here run on. Related projects from the same group are used directly: [MLX C](https://github.com/ml-explore/mlx-c) (the C interface mlx-serve calls), [MLX Swift LM](https://github.com/ml-explore/mlx-swift-lm) (the model and generation library the first-generation engine is built on), and [MLX LM](https://github.com/ml-explore/mlx-lm).

## Splash

[Splash](https://github.com/incoai/splash) (Apache-2.0) was built by Inco ([inco.ai](https://inco.ai)). It is a local inference engine for Apple silicon specialized per model, with a model-specific draft for speculative decoding, kernels compiled for the model's exact shapes, and a memory plan computed for the machine. This design uses its native engine mode, whose protocol carries token IDs over standard input and output with no network code and no disk writes by default. That property is what lets Splash run as a sealed child with no change to its code. Its Metal library is embedded in the executable, which is what the integrity requirement relies on. The Qwen3.8-27B listing in this design is served by Splash.

## mlx-serve

[mlx-serve](https://github.com/ddalcu/mlx-serve) (MIT) was built by David Dalcu and contributors. It is a native Zig inference server for Apple silicon that serves MLX-format models behind OpenAI-compatible and Anthropic-compatible HTTP APIs. The Qwen Flash-Next listing is served by it. This proposal adds a 53-line patch that accepts each request as a passed file descriptor instead of a listening socket; we intend to offer it to the project as an optional mode. mlx-serve's own NOTICE file lists the upstream work it incorporates (including MTPLX, dflash-mlx and Splash kernels, jinja.cpp, and llama.cpp), and those credits belong with it.

## Qwen models

The Qwen models are released by the Qwen team at Alibaba Cloud. Both listings in this design are Qwen-family open-weight models. Model weights keep their own licenses.

## Hugging Face swift-transformers

[swift-transformers](https://github.com/huggingface/swift-transformers) provides the tokenizer and hub code the first-generation engine uses and the planned engine-side Splash front end would use.

## Apple platform features

The isolation rests on features that Apple built and documents (several through headers rather than guides). Our work is to combine them, measure how they behave together on current macOS, and record where they do not behave as assumed.

- **App Attest** (DeviceCheck): hardware-backed keys and assertions, with the code-directory-hash opt-in that makes every assertion name the binary that produced it, and the refusal to attest below Full Security.
- **Hardened runtime** and code signing: no task-port access without `get-task-allow`, library validation, and the code-signing status word the engine reads for each child.
- **Launch constraints**: the kernel enforcement that lets a spawner require a specific code-directory hash and refuses any other image.
- **System Integrity Protection, Full Security boot, sealed system volume, encrypted swap**: the platform properties the threat model counts on (A-3, A-8, A-11, A-21).
- **Mach exception ports, `posix_spawn` attributes, audit tokens and `LOCAL_PEERTOKEN`**: the process-identity and fault-handling primitives behind R1 to R3.
- **Metal** and the app sandbox facility: the GPU path children use and the sandbox each child applies to itself.

## Review and measurement

The design specification went through repeated adversarial review, and the spawn module through an independent multi-family review; their findings, including the five defects still open, are in [STATUS-AND-KNOWN-ISSUES.md](STATUS-AND-KNOWN-ISSUES.md). Defects reported there are in our code unless stated otherwise.
