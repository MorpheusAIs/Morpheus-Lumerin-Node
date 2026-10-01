// SealedEngine — attested inference engine for Apple Silicon providers.
//
// P2 skeleton: per-boot App Attest key + attestation (with the CD-hash opt-in,
// so every attestation and assertion names this exact binary), and per-session
// assertions binding a consumer challenge to an ephemeral X25519 key. Nothing is
// written to disk; keys and sessions live only in this process.
import CoreServices
import CryptoKit
import DeviceCheck
import Foundation
import MLXLMCommon

// 1. Fatal signals exit immediately so the crash reporter never captures
//    register state (measured leak: root `kill -SEGV` → .ips with registers).
Signals.installExitHandlers()
setvbuf(stdout, nil, _IOLBF, 0)

// 2. Environment variables can redirect or instrument MLX/Metal, so only an
//    allowlist is accepted; and MLX's GPU kernels come from bytes inside this
//    signed executable (covered by the attested code hash), never from disk.
//    Metal's on-disk cache of compiled GPU code is switched off for the same
//    reason: everything the GPU runs is compiled in this process from attested bytes.
//    Metal would also take headers for runtime-compiled kernels from the
//    bundle's Resources folder, so there must be none (see RuntimeGuard).
//    The standard-library module those compiles use is pinned to Apple's
//    prebuilt copy on the sealed system volume, never a writable cache file.
RuntimeGuard.enforceEnvironment()
RuntimeGuard.disableMetalShaderCache()
RuntimeGuard.guardBundleResources()
RuntimeGuard.pinMetalStandardLibraryModule()
RuntimeGuard.installEmbeddedMetalLibrary()

// 3. App Attest refuses bundles LaunchServices does not know about.
LSRegisterURL(Bundle.main.bundleURL as CFURL, true)

let port = UInt16(ProcessInfo.processInfo.environment["SEALED_ENGINE_PORT"] ?? "") ?? 29451
let engine = Engine()
var server: HTTPServer?  // retained for the process lifetime

engine.bootstrap { error in
    if let error {
        FileHandle.standardError.write(Data("attestation failed: \(error)\n".utf8))
        exit(2)
    }
    do {
        server = try HTTPServer(port: port) { request, respond in
            engine.handle(request, respond: respond)
        }
        server?.start()
        print("sealed-engine listening on 127.0.0.1:\(port) keyId=\(engine.keyIdBase64)")
    } catch {
        FileHandle.standardError.write(Data("listen failed: \(error)\n".utf8))
        exit(3)
    }
}
RunLoop.main.run()  // keep the main run loop alive for system frameworks

final class Engine {
    private let service = DCAppAttestService.shared
    private let queue = DispatchQueue(label: "sealed-engine.state")
    private var keyId = ""
    private var attestation = Data()
    private var attestClientData = Data()
    private var sessions: [String: Session] = [:]
    private var activeGenerations = 0
    private static let maxGenerations = 2

    private static let maxSessions = 256
    private static let sessionTTL: TimeInterval = 3600

    /// One consumer session. Keys are derived at /session time from the client
    /// key the consumer sent (and the assertion covers), so no later frame can
    /// rebind the session. Counters enforce strict ordering; sessions expire
    /// after an hour or when the cap evicts the oldest.
    final class Session {
        let keys: EnvelopeKeys
        let created = Date()
        var recvSeq: UInt64 = 0
        var sendSeq: UInt64 = 0
        var generating = false
        init(keys: EnvelopeKeys) { self.keys = keys }
    }
    /// Weights manifest hash bound into every session and probe; all-zero when
    /// no model is loaded (echo mode, used by the envelope tests).
    private var weightsHash = Data(count: 32)
    /// In-process model (P4); nil = echo mode.
    private var model: ModelHost?
    /// SHA256("morpheus-sealed/provider/v1" ‖ lowercase provider address from
    /// SEALED_PROVIDER_ADDR; empty if unset — an unbound engine).
    private let providerTag: Data = {
        let addr = (ProcessInfo.processInfo.environment["SEALED_PROVIDER_ADDR"] ?? "").lowercased()
        return Data(SHA256.hash(data: Data("morpheus-sealed/provider/v1".utf8) + Data(addr.utf8)))
    }()

    var keyIdBase64: String { queue.sync { keyId } }
    fileprivate func sessionsFor(_ id: String) -> Session? { sessions[id] }

    /// Must be called on `queue`. Drops expired sessions, then the oldest over the cap.
    private func evictSessions() {
        let now = Date()
        sessions = sessions.filter { now.timeIntervalSince($0.value.created) < Self.sessionTTL }
        while sessions.count >= Self.maxSessions,
              let oldest = sessions.min(by: { $0.value.created < $1.value.created })?.key {
            sessions[oldest] = nil
        }
    }

    /// Generates this boot's key and fetches its Apple attestation.
    /// Attestation freshness comes from per-session assertions: keys die at
    /// every restart, so a replayed old attestation names a key that cannot sign.
    func bootstrap(_ done: @escaping (Error?) -> Void) {
        guard service.isSupported else { return done(EngineError.unsupported) }
        // Load (and hash) the model BEFORE attesting: the engine only ever serves
        // with the weights its sessions are bound to already in memory.
        if let dir = ProcessInfo.processInfo.environment["SEALED_MODEL_DIR"], !dir.isEmpty {
            Task {
                do {
                    let host = try await ModelHost.load(directory: URL(fileURLWithPath: dir))
                    self.queue.sync {
                        self.model = host
                        self.weightsHash = host.weightsHash
                    }
                    print("model loaded: weights \(host.weightsHash.map { String(format: "%02x", $0) }.joined())")
                    self.attestKey(done)
                } catch {
                    done(error)
                }
            }
            return
        }
        attestKey(done)
    }

    private func attestKey(_ done: @escaping (Error?) -> Void) {
        service.generateKey { keyId, error in
            guard let keyId, error == nil else { return done(error ?? EngineError.keygen) }
            var nonce = Data(count: 32)
            _ = nonce.withUnsafeMutableBytes { SecRandomCopyBytes(kSecRandomDefault, 32, $0.baseAddress!) }
            let clientData = Data("sealed-engine/attest/v1".utf8) + nonce
            self.service.attestKey(keyId, clientDataHash: Data(SHA256.hash(data: clientData))) { att, error in
                guard let att, error == nil else { return done(error ?? EngineError.attest) }
                self.queue.sync {
                    self.keyId = keyId
                    self.attestation = att
                    self.attestClientData = clientData
                }
                done(nil)
            }
        }
    }

    func handle(_ req: HTTPRequest, respond: @escaping (Int, [String: Any]) -> Void) {
        switch (req.method, req.path) {
        case ("GET", "/attest"):
            queue.sync {
                respond(200, [
                    "keyId": keyId,
                    "attestation": attestation.base64EncodedString(),
                    "clientData": attestClientData.base64EncodedString(),
                ])
            }
        case ("POST", "/session"):
            guard let body = try? JSONSerialization.jsonObject(with: req.body) as? [String: Any],
                  let b64 = body["challenge"] as? String,
                  let challenge = Data(base64Encoded: b64), challenge.count == 32,
                  let clientPub = (body["clientPub"] as? String).flatMap({ Data(base64Encoded: $0) }), clientPub.count == 32
            else { return respond(400, ["error": "challenge and clientPub must be 32 bytes, base64"]) }
            let session = Curve25519.KeyAgreement.PrivateKey()
            guard let keys = try? EnvelopeKeys(enginePriv: session, clientPub: clientPub, challenge: challenge) else {
                return respond(400, ["error": "invalid client key"])
            }
            let sessionId = UUID().uuidString
            // clientData = challenge ‖ client X25519 ‖ engine X25519 ‖ weights hash (128 bytes).
            let clientData = challenge + clientPub + session.publicKey.rawRepresentation + queue.sync { weightsHash }
            let keyId = queue.sync { self.keyId }
            service.generateAssertion(keyId, clientDataHash: Data(SHA256.hash(data: clientData))) { assertion, error in
                guard let assertion, error == nil else {
                    return respond(503, ["error": "assertion refused: \(error.map { "\($0)" } ?? "unknown")"])
                }
                self.queue.sync {
                    self.evictSessions()
                    self.sessions[sessionId] = Session(keys: keys)
                }
                respond(200, [
                    "sessionId": sessionId,
                    "clientData": clientData.base64EncodedString(),
                    "assertion": assertion.base64EncodedString(),
                ])
            }
        case ("POST", "/probe"):
            // Public liveness proof for third-party verifiers: an assertion over
            // challenge ‖ providerTag ‖ 32 zero bytes ‖ weights hash, where
            // providerTag = SHA256("morpheus-sealed/provider/v1" ‖ provider address)
            // binds the proof to the provider this engine serves (a provider that
            // relays probes to someone else's engine is caught). No session is
            // created, so probes cannot evict consumer sessions, and the tag can
            // never equal a consumer's handshake key.
            guard let body = try? JSONSerialization.jsonObject(with: req.body) as? [String: Any],
                  let b64 = body["challenge"] as? String,
                  let challenge = Data(base64Encoded: b64), challenge.count == 32
            else { return respond(400, ["error": "challenge must be 32 bytes, base64"]) }
            let clientData = challenge + providerTag + Data(count: 32) + queue.sync { weightsHash }
            let keyId = queue.sync { self.keyId }
            service.generateAssertion(keyId, clientDataHash: Data(SHA256.hash(data: clientData))) { assertion, error in
                guard let assertion, error == nil else { return respond(503, ["error": "assertion refused"]) }
                respond(200, ["clientData": clientData.base64EncodedString(), "assertion": assertion.base64EncodedString()])
            }
        case ("POST", "/infer"):
            queue.sync { self.infer(req.body, respond: respond) }
        default:
            respond(404, ["error": "not found"])
        }
    }
}

enum EngineError: Error { case unsupported, keygen, attest }

extension Engine {
    /// P3a: decrypts one sealed prompt and returns sealed chunks. Until P4 binds a
    /// model, the reply proves decryption: the plaintext's length and SHA-256.
    /// Must be called on `queue`.
    fileprivate func infer(_ body: Data, respond: @escaping (Int, [String: Any]) -> Void) {
        guard let frame = try? JSONSerialization.jsonObject(with: body) as? [String: Any],
              let sid = frame["sessionId"] as? String,
              let seq = strictUInt64(frame["seq"]),
              let ct = (frame["ct"] as? String).flatMap({ Data(base64Encoded: $0) })
        else { return respond(400, ["error": "malformed frame"]) }
        guard let session = sessionsFor(sid) else { return respond(404, ["error": "unknown session"]) }
        // One generation per session (its reply frames must stay in order) and a
        // node-wide cap. Checked before decrypting, so a refusal consumes nothing.
        if model != nil && (session.generating || activeGenerations >= Self.maxGenerations) {
            return respond(409, ["error": "busy"])
        }
        do {
            guard seq == session.recvSeq else { throw EnvelopeError.sequence }
            let prompt = try session.keys.open(ct, seq: session.recvSeq)
            session.recvSeq += 1
            guard let model else {
                // Echo mode: prove decryption with the plaintext's length and SHA-256.
                let digest = SHA256.hash(data: prompt).map { String(format: "%02x", $0) }.joined()
                let frames = try sealFrames(session, sid: sid, chunks: ["sealed-engine: prompt received", "bytes=\(prompt.count)", "sha256=\(digest)"])
                return respond(200, ["frames": frames])
            }
            let request = try ChatRequest(json: prompt)
            session.generating = true
            activeGenerations += 1
            Task {
                defer {
                    self.queue.sync {
                        session.generating = false
                        self.activeGenerations -= 1
                    }
                }
                let chunks: [String]
                do {
                    chunks = try await model.complete(
                        messages: request.history, prompt: request.prompt,
                        maxTokens: request.maxTokens, temperature: request.temperature)
                } catch {
                    return respond(500, ["error": "inference failed"])
                }
                self.queue.sync {
                    do {
                        let frames = try self.sealFrames(session, sid: sid, chunks: chunks.isEmpty ? [""] : chunks)
                        respond(200, ["frames": frames])
                    } catch {
                        respond(500, ["error": "sealing failed"])
                    }
                }
            }
            return
        } catch {
            // A frame that fails authentication does not advance the counter, so a
            // forged or replayed frame cannot evict or desynchronise the consumer's
            // bound session. (An authentic frame whose plaintext is not a valid chat
            // request has advanced both sides' counters consistently.)
            respond(422, ["error": "rejected"])
        }
    }
}

extension Engine {
    /// Seals chunks as one response: flag 0x00 = more, 0x01 = final. Must be
    /// called on `queue`.
    fileprivate func sealFrames(_ session: Session, sid: String, chunks: [String]) throws -> [[String: Any]] {
        var frames: [[String: Any]] = []
        for (i, chunk) in chunks.enumerated() {
            let flag: UInt8 = i == chunks.count - 1 ? 1 : 0
            let seq = session.sendSeq
            let sealed = try session.keys.seal(Data([flag]) + Data(chunk.utf8), seq: seq)
            session.sendSeq += 1
            frames.append(["sessionId": sid, "seq": seq, "ct": sealed.base64EncodedString()])
        }
        return frames
    }
}

/// The decrypted OpenAI-style chat request: all messages but the last user
/// turn become history; the last user turn is the prompt.
struct ChatRequest {
    let history: [Chat.Message]
    let prompt: String
    let maxTokens: Int
    let temperature: Float

    init(json: Data) throws {
        guard let obj = try JSONSerialization.jsonObject(with: json) as? [String: Any],
              let raw = obj["messages"] as? [[String: Any]], !raw.isEmpty
        else { throw EnvelopeError.open }
        var msgs: [Chat.Message] = []
        for m in raw {
            guard let content = m["content"] as? String else { throw EnvelopeError.open } // no silent drops
            switch m["role"] as? String {
            case "system": msgs.append(.system(content))
            case "assistant": msgs.append(.assistant(content))
            case "user": msgs.append(.user(content))
            default: throw EnvelopeError.open
            }
        }
        guard let last = msgs.popLast(), last.role == .user else { throw EnvelopeError.open }
        history = msgs
        prompt = last.content
        maxTokens = min(max((obj["max_tokens"] as? Int) ?? 256, 1), 2048)
        let t = (obj["temperature"] as? Double) ?? 0.6
        temperature = Float(t.isFinite ? min(max(t, 0), 2) : 0.6)
    }
}

/// Accepts only a JSON integer in 0...UInt64.max (rejects 0.5, -1, strings, booleans).
private func strictUInt64(_ v: Any?) -> UInt64? {
    guard let n = v as? NSNumber, CFGetTypeID(n) != CFBooleanGetTypeID() else { return nil }
    let d = n.doubleValue
    guard d >= 0, d.rounded() == d, d < 1.8446744073709552e19 else { return nil }
    return n.uint64Value
}
