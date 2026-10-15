import CryptoKit
import Foundation

/// Engine side of envelope v1 — must match proxy-router/internal/sealed/sealed.go.
///   shared = X25519(engine_session_priv, client_eph_pub)   (all-zero rejected)
///   okm    = HKDF-SHA256(shared, salt: challenge, info: "morpheus-sealed/v1" ‖ client_pub ‖ engine_pub, 64)
///   c2e = okm[0..<32], e2c = okm[32..<64]
///   frame n = ChaCha20-Poly1305(key_dir, nonce: 0x00000000 ‖ UInt64.bigEndian(n), aad: "morpheus-sealed/v1" ‖ dir)
struct EnvelopeKeys {
    static let version = "morpheus-sealed/v1"
    let c2e: SymmetricKey
    let e2c: SymmetricKey
    let clientPub: Data

    init(enginePriv: Curve25519.KeyAgreement.PrivateKey, clientPub: Data, challenge: Data) throws {
        guard clientPub.count == 32, challenge.count == 32 else { throw EnvelopeError.key }
        let peer = try Curve25519.KeyAgreement.PublicKey(rawRepresentation: clientPub)
        let shared = try enginePriv.sharedSecretFromKeyAgreement(with: peer)
        let allZero = shared.withUnsafeBytes { $0.allSatisfy { $0 == 0 } }
        guard !allZero else { throw EnvelopeError.key }
        let info = Data(Self.version.utf8) + clientPub + enginePriv.publicKey.rawRepresentation
        let okm = shared.hkdfDerivedSymmetricKey(using: SHA256.self, salt: challenge, sharedInfo: info, outputByteCount: 64)
        let bytes = okm.withUnsafeBytes { Data($0) }
        c2e = SymmetricKey(data: bytes.prefix(32))
        e2c = SymmetricKey(data: bytes.suffix(32))
        self.clientPub = clientPub
    }

    static func nonce(_ seq: UInt64) throws -> ChaChaPoly.Nonce {
        var be = seq.bigEndian
        return try ChaChaPoly.Nonce(data: Data(count: 4) + Data(bytes: &be, count: 8))
    }

    func open(_ ct: Data, seq: UInt64) throws -> Data {
        guard ct.count >= 16 else { throw EnvelopeError.open }
        let box = try ChaChaPoly.SealedBox(nonce: Self.nonce(seq), ciphertext: ct.dropLast(16), tag: ct.suffix(16))
        do {
            return try ChaChaPoly.open(box, using: c2e, authenticating: Data((Self.version + "c2e").utf8))
        } catch {
            throw EnvelopeError.open
        }
    }

    func seal(_ pt: Data, seq: UInt64) throws -> Data {
        let box = try ChaChaPoly.seal(pt, using: e2c, nonce: Self.nonce(seq), authenticating: Data((Self.version + "e2c").utf8))
        return box.ciphertext + box.tag
    }
}

enum EnvelopeError: Error { case key, open, sequence, session }
