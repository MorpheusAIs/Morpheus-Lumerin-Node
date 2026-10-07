import CryptoKit
import Foundation
import Hub
import MLX
import MLXLLM
import MLXLMCommon
import MLXNN
import Tokenizers

/// In-process model host. The model directory is read ONCE into memory (an
/// explicit allowlist of files; anything else refuses to start), the weights
/// manifest hash is computed over exactly those bytes, and the model and
/// tokenizer are built from the same buffers. Nothing is re-read from disk, so
/// the operator cannot swap, add or FIFO-serve weights between hashing and
/// loading: the attested hash is the weights in memory.
final class ModelHost: @unchecked Sendable {
    let container: ModelContainer
    let weightsHash: Data

    private init(container: ModelContainer, weightsHash: Data) {
        self.container = container
        self.weightsHash = weightsHash
    }

    static func load(directory: URL) async throws -> ModelHost {
        let snapshot = try ModelSnapshot(directory: directory)
        let container = try await buildContainer(snapshot)
        return ModelHost(container: container, weightsHash: snapshot.manifestHash)
    }

    /// Mirrors mlx-swift-lm's LLMModelFactory._load + loadWeights, fed from the
    /// in-memory snapshot instead of the filesystem.
    private static func buildContainer(_ s: ModelSnapshot) async throws -> ModelContainer {
        let configData = try s.require("config.json")
        let baseConfig = try JSONDecoder.json5().decode(BaseConfiguration.self, from: configData)
        let model = try await LLMTypeRegistry.shared.createModel(
            configuration: configData, modelType: baseConfig.modelType)

        // Weights: each hashed .safetensors buffer, in manifest order; a tensor
        // name appearing in two files is refused rather than silently overwritten.
        var weights = [String: MLXArray]()
        var metadata = [String: String]()
        for name in s.names where name.hasSuffix(".safetensors") {
            let (w, m) = try loadArraysAndMetadata(data: s.files[name]!)
            for (key, value) in w {
                guard weights[key] == nil else { throw ModelError.duplicateTensor(key) }
                weights[key] = value
            }
            if metadata.isEmpty { metadata = m }
        }
        guard !weights.isEmpty else { throw ModelError.missing("*.safetensors") }
        weights = model.sanitize(weights: weights, metadata: metadata)
        if let perLayer = baseConfig.perLayerQuantization {
            quantize(model: model) { path, _ in
                weights["\(path).scales"] != nil ? perLayer.quantization(layer: path)?.asTuple : nil
            }
        }
        try model.update(parameters: ModuleParameters.unflattened(weights), verify: [.all])
        eval(model)

        // Tokenizer from the in-memory tokenizer_config.json + tokenizer.json.
        let tokenizer = TokenizerBridge(try AutoTokenizer.from(
            tokenizerConfig: Config(try jsonObject(s.require("tokenizer_config.json"))),
            tokenizerData: Config(try jsonObject(s.require("tokenizer.json")))))

        var eos = Set(baseConfig.eosTokenIds?.values ?? [])
        if let gen = s.files["generation_config.json"],
           let genConfig = try? JSONDecoder.json5().decode(GenerationConfigFile.self, from: gen),
           let ids = genConfig.eosTokenIds?.values {
            eos = Set(ids)
        }
        let configuration = ModelConfiguration(
            directory: s.directory, eosTokenIds: eos,
            toolCallFormat: ToolCallFormat.infer(from: baseConfig.modelType, configData: configData))
        let messageGenerator: any MessageGenerator =
            (model as? LLMModel)?.messageGenerator(tokenizer: tokenizer) ?? DefaultMessageGenerator()
        let processor = ChatInputProcessor(tokenizer: tokenizer, messageGenerator: messageGenerator)
        return ModelContainer(context: ModelContext(
            configuration: configuration, model: model, processor: processor, tokenizer: tokenizer))
    }

    private static func jsonObject(_ data: Data) throws -> [NSString: Any] {
        guard let obj = try JSONSerialization.jsonObject(with: data) as? [NSString: Any] else {
            throw ModelError.malformed
        }
        return obj
    }

    /// Runs a chat completion in-process and returns the streamed text chunks.
    func complete(messages: [Chat.Message], prompt: String, maxTokens: Int, temperature: Float) async throws -> [String] {
        let session = ChatSession(
            container, history: messages,
            generateParameters: GenerateParameters(maxTokens: maxTokens, temperature: temperature))
        var chunks: [String] = []
        for try await chunk in session.streamResponse(to: prompt) {
            chunks.append(chunk)
        }
        return chunks
    }
}

/// Mirrors mlx-swift-lm's (private) LLMUserInputProcessor, except that a
/// model without a chat template is refused rather than silently fed plain
/// text (which would change model behaviour without anyone noticing).
struct ChatInputProcessor: UserInputProcessor {
    let tokenizer: any MLXLMCommon.Tokenizer
    let messageGenerator: any MessageGenerator

    func prepare(input: UserInput) throws -> LMInput {
        let messages = messageGenerator.generate(from: input)
        let tokens = try tokenizer.applyChatTemplate(
            messages: messages, tools: input.tools, additionalContext: input.additionalContext)
        return LMInput(tokens: MLXArray(tokens))
    }
}

enum ModelError: Error { case unexpectedEntry(String), notRegular(String), missing(String), duplicateTensor(String), malformed }

/// The model directory read once into memory. Allowed entries: the files below
/// and `*.safetensors`; any other entry (hidden files, subdirectories,
/// anything that is not a regular file after resolving symlinks — FIFOs,
/// sockets, devices) refuses to start.
///
/// Manifest hash (must match proxy-router sealed.WeightsManifestHash):
///   SHA256 over, for each file sorted by name (byte order):
///   name ‖ 0x00 ‖ SHA256(contents) ‖ 0x0A
struct ModelSnapshot {
    static let allowedNames: Set<String> = [
        "config.json", "generation_config.json", "tokenizer.json", "tokenizer_config.json",
        "special_tokens_map.json", "added_tokens.json", "vocab.json", "merges.txt",
        "chat_template.jinja", "model.safetensors.index.json",
    ]

    let directory: URL
    let names: [String]
    let files: [String: Data]
    let manifestHash: Data

    init(directory: URL) throws {
        self.directory = directory
        let fm = FileManager.default
        var files: [String: Data] = [:]
        for name in try fm.contentsOfDirectory(atPath: directory.path) {
            guard Self.allowedNames.contains(name) || (name.hasSuffix(".safetensors") && !name.hasPrefix(".") && !name.contains("/")) else {
                throw ModelError.unexpectedEntry(name)
            }
            let resolved = directory.appendingPathComponent(name).resolvingSymlinksInPath()
            guard (try fm.attributesOfItem(atPath: resolved.path)[.type] as? FileAttributeType) == .typeRegular else {
                throw ModelError.notRegular(name)
            }
            files[name] = try Data(contentsOf: resolved, options: .uncached)
        }
        let names = files.keys.sorted { Array($0.utf8).lexicographicallyPrecedes(Array($1.utf8)) }
        var manifest = SHA256()
        for name in names {
            manifest.update(data: Data(name.utf8) + Data([0]) + Data(SHA256.hash(data: files[name]!)) + Data([0x0A]))
        }
        self.names = names
        self.files = files
        self.manifestHash = Data(manifest.finalize())
    }

    func require(_ name: String) throws -> Data {
        guard let d = files[name] else { throw ModelError.missing(name) }
        return d
    }
}

struct TokenizerBridge: MLXLMCommon.Tokenizer {
    private let upstream: any Tokenizers.Tokenizer
    init(_ upstream: any Tokenizers.Tokenizer) { self.upstream = upstream }

    func encode(text: String, addSpecialTokens: Bool) -> [Int] {
        upstream.encode(text: text, addSpecialTokens: addSpecialTokens)
    }
    func decode(tokenIds: [Int], skipSpecialTokens: Bool) -> String {
        upstream.decode(tokens: tokenIds, skipSpecialTokens: skipSpecialTokens)
    }
    func convertTokenToId(_ token: String) -> Int? { upstream.convertTokenToId(token) }
    func convertIdToToken(_ id: Int) -> String? { upstream.convertIdToToken(id) }
    var bosToken: String? { upstream.bosToken }
    var eosToken: String? { upstream.eosToken }
    var unknownToken: String? { upstream.unknownToken }

    func applyChatTemplate(
        messages: [[String: any Sendable]], tools: [[String: any Sendable]]?, additionalContext: [String: any Sendable]?
    ) throws -> [Int] {
        do {
            return try upstream.applyChatTemplate(messages: messages, tools: tools, additionalContext: additionalContext)
        } catch Tokenizers.TokenizerError.missingChatTemplate {
            throw MLXLMCommon.TokenizerError.missingChatTemplate
        }
    }
}
