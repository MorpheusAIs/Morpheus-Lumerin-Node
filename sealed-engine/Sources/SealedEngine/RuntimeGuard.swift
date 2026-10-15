import CryptoKit
import Foundation
import MachO
import Metal

@_silgen_name("MTLSetShaderCachePath")
private func MTLSetShaderCachePath(_ path: NSString)
@_silgen_name("MTLGetShaderCachePath")
private func MTLGetShaderCachePath() -> Unmanaged<NSString>?
@_silgen_name("MTLGetModulesCachePath")
private func MTLGetModulesCachePath() -> Unmanaged<NSString>?

@_silgen_name("mlx_metal_set_source_compile_arguments")
private func mlx_metal_set_source_compile_arguments(_ args: UnsafePointer<CChar>) -> Int32

@_silgen_name("mlx_metal_set_source_compile_guard")
private func mlx_metal_set_source_compile_guard(_ guard: @convention(c) () -> Int32) -> Int32

/// Path checked by the source-compile guard; set once at startup, read from C.
private var bundleResourcesPath: UnsafeMutablePointer<CChar>?

/// 0 = nothing at the bundle's Resources path; anything else refuses. Also
/// refuses when the path cannot be checked for any reason but "absent".
private let resourcesAbsentGuard: @convention(c) () -> Int32 = {
    guard let p = bundleResourcesPath else { return 1 }
    var st = stat()
    if lstat(p, &st) == 0 { return 1 }
    return errno == ENOENT ? 0 : 1
}

@_silgen_name("mlx_metal_set_default_library_data")
private func mlx_metal_set_default_library_data(_ data: UnsafeRawPointer, _ size: Int) -> Int32

enum RuntimeGuard {
    private static let allowedEnv: Set<String> = [
        "HOME", "USER", "LOGNAME", "SHELL", "PATH", "TMPDIR", "PWD", "OLDPWD", "SHLVL",
        "TERM", "TERM_PROGRAM", "TERM_PROGRAM_VERSION", "TERM_SESSION_ID", "COLORTERM",
        "LANG", "MAIL", "_", "COMMAND_MODE", "XPC_FLAGS", "XPC_SERVICE_NAME",
        "SSH_CLIENT", "SSH_CONNECTION", "SSH_TTY", "SSH_AUTH_SOCK",
        "SEALED_ENGINE_PORT", "SEALED_MODEL_DIR", "SEALED_PROVIDER_ADDR",
    ]
    private static let allowedEnvPrefixes = ["LC_", "__CF"]

    static func enforceEnvironment() {
        let rejected = ProcessInfo.processInfo.environment.keys
            .filter { name in !allowedEnv.contains(name) && !allowedEnvPrefixes.contains { name.hasPrefix($0) } }
            .sorted()
        guard !rejected.isEmpty else { return }
        FileHandle.standardError.write(Data("refusing to start: environment variable(s) not allowed: \(rejected.joined(separator: ", "))\n".utf8))
        exit(4)
    }

    /// Metal keeps compiled GPU code in a per-app cache under the user's cache
    /// directory and, on a hit, runs it without compiling. That directory is
    /// writable by the operator's user, so a cached entry would be GPU code the
    /// attestation does not cover. With MTL_SHADER_CACHE_SIZE=0 Metal neither
    /// reads nor writes the cache (measured on macOS 27: a warm entry is
    /// ignored and compiled again). It must be set before Metal is first used;
    /// it is set here, in attested code, never taken from the launch
    /// environment (enforceEnvironment rejects the variable there).
    ///
    /// The compiler also keeps a module cache (the precompiled Metal standard
    /// library, `com.apple.metalfe`) that MTL_SHADER_CACHE_SIZE does not cover;
    /// Metal's compiler service read it on the engine's behalf (measured). So
    /// every cache path is also pointed at a new, empty, private directory for
    /// this launch: the compiler then loads Apple's prebuilt module from the
    /// sealed system volume and reads nothing from the directory but an empty
    /// timestamp (measured). MTLSetShaderCachePath is Metal SPI (exported,
    /// listed in the SDK's Metal.tbd, no public header).
    static func disableMetalShaderCache() {
        guard setenv("MTL_SHADER_CACHE_SIZE", "0", 1) == 0 else {
            refuse("could not disable the Metal shader cache")
        }
        var template = Array((NSTemporaryDirectory() + "sealed-mtl.XXXXXXXX").utf8CString)
        guard let dir = template.withUnsafeMutableBufferPointer({ mkdtemp($0.baseAddress!) }).map({ String(cString: $0) }) else {
            refuse("could not create a private Metal cache directory")
        }
        MTLSetShaderCachePath(dir as NSString)
        let modules = MTLGetModulesCachePath()?.takeUnretainedValue() as String?
        let shaders = MTLGetShaderCachePath()?.takeUnretainedValue() as String?
        guard modules == dir, shaders == dir else {
            refuse("Metal did not accept the private cache directory")
        }
        print("metal shader cache: disabled; module cache: fresh \(dir)")
    }

    /// Runtime-generated kernel source is compiled against Metal's standard
    /// library, which the compiler always takes as a precompiled module. MLX
    /// had shader logging on (enabled whenever NDEBUG is undefined, as in this
    /// Release build), so no module on the system volume matched and the
    /// compiler built one in the cache directory and re-read it during the run
    /// (measured): a file the operator can replace. The MLX patch now leaves
    /// logging off. Apple's prebuilt modules on the sealed system volume are
    /// accepted when named explicitly and gave identical output (measured:
    /// greedy generations token-identical on a dense and a MoE model), and are
    /// never written. So every source compilation gets
    /// `-fno-implicit-modules -fmodule-file=<prebuilt module>`, chosen here by a
    /// trial compile with MLX's exact options. With implicit modules off the
    /// compiler cannot build a module of its own: it uses the pinned file or
    /// the compilation fails (measured), so it never falls back to a file the
    /// operator can write.
    static func pinMetalStandardLibraryModule() {
        guard #available(macOS 26, *) else { refuse("Metal language 4.0 is required") }
        let versions = "/System/Library/PrivateFrameworks/GPUCompiler.framework/Versions"
        let fm = FileManager.default
        var candidates: [String] = []
        for v in (try? fm.contentsOfDirectory(atPath: versions))?.sorted() ?? [] {
            let clang = "\(versions)/\(v)/Libraries/lib/clang"
            for c in (try? fm.contentsOfDirectory(atPath: clang))?.sorted() ?? [] {
                let prebuilt = "\(clang)/\(c)/include/metal/prebuilt_implicit_modules"
                for h in (try? fm.contentsOfDirectory(atPath: prebuilt))?.sorted() ?? [] {
                    candidates.append("\(prebuilt)/\(h)/monolithic_metal.pcm")
                }
            }
        }
        guard let device = MTLCreateSystemDefaultDevice() else { refuse("no Metal device") }
        let probe = """
        #include <metal_stdlib>
        using namespace metal;
        kernel void k(device float* a [[buffer(0)]], uint i [[thread_position_in_grid]]) { a[i] = exp(a[i]); }
        """
        func compiles(_ args: String) -> Bool {
            let opts = MTLCompileOptions()
            // Exactly MLX's options (device.cpp): language 4.0 on macOS 26+,
            // fastMathEnabled = false (not mathMode .safe, which keeps the fast
            // math functions and matches a different module), logging off.
            opts.languageVersion = .version4_0
            opts.fastMathEnabled = false
            opts.setValue(args, forKey: "additionalCompilerArguments")
            return (try? device.makeLibrary(source: probe, options: opts)) != nil
        }
        // The pin only means something if the compiler honours the arguments:
        // naming a module that does not exist must make the compile fail. If it
        // does not, the arguments are ignored and implicit modules would be used.
        guard !compiles("-fno-implicit-modules -fmodule-file=\(versions)/sealed-engine-nonexistent.pcm") else {
            refuse("Metal ignores the compiler arguments that pin the standard-library module")
        }
        for path in candidates {
            // The module must really live inside Apple's compiler framework on
            // the system volume, with no link leading elsewhere.
            guard let real = realpath(path, nil).map({ p -> String in defer { free(p) }; return String(cString: p) }),
                  real == path, real.hasPrefix(versions + "/"),
                  // It is embedded in one argument string: plain characters only.
                  real.unicodeScalars.allSatisfy({ CharacterSet(charactersIn: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/._-").contains($0) })
            else { continue }
            let args = "-fno-implicit-modules -fmodule-file=\(real)"
            guard compiles(args) else { continue }
            guard mlx_metal_set_source_compile_arguments(args) == 0 else {
                refuse("MLX rejected the Metal compile arguments")
            }
            print("metal stdlib module: \(real)")
            return
        }
        refuse("no prebuilt Metal standard-library module on the system volume accepts MLX's options")
    }

    /// Metal compiles runtime-generated kernel source against headers found in
    /// the app's Contents/Resources before the system headers (measured: a
    /// metal_stdlib placed there replaces Apple's for every such compile), and
    /// nothing checks bundle files at runtime. The engine therefore ships no
    /// Resources folder, refuses to start if one exists, and has MLX refuse
    /// each Metal source compilation if one has appeared since.
    static func guardBundleResources() {
        let path = Bundle.main.bundleURL.appendingPathComponent("Contents/Resources").path
        if let rp = Bundle.main.resourcePath, rp != path {
            refuse("unexpected bundle resource path \(rp)")
        }
        bundleResourcesPath = strdup(path)
        guard resourcesAbsentGuard() == 0 else {
            refuse("the app bundle has a Resources folder, which Metal would search for headers")
        }
        guard mlx_metal_set_source_compile_guard(resourcesAbsentGuard) == 0 else {
            refuse("MLX rejected the source-compile guard")
        }
        print("bundle resources: none; Metal source compiles guarded")
    }

    private static func refuse(_ why: String) -> Never {
        FileHandle.standardError.write(Data("refusing to start: \(why)\n".utf8))
        exit(6)
    }

    static func installEmbeddedMetalLibrary() {
        guard let header = _dyld_get_image_header(0) else {
            FileHandle.standardError.write(Data("refusing to start: embedded Metal library missing\n".utf8))
            exit(5)
        }
        let h = UnsafeRawPointer(header).assumingMemoryBound(to: mach_header_64.self)
        var size: UInt = 0
        let ptr = getsectiondata(h, "__TEXT", "__mlxlib", &size)
        guard let ptr, size > 0 else {
            FileHandle.standardError.write(Data("refusing to start: embedded Metal library missing\n".utf8))
            exit(5)
        }
        guard mlx_metal_set_default_library_data(ptr, Int(size)) == 0 else {
            FileHandle.standardError.write(Data("refusing to start: MLX rejected the embedded Metal library\n".utf8))
            exit(5)
        }
        let digest = SHA256.hash(data: UnsafeRawBufferPointer(start: ptr, count: Int(size)))
            .map { String(format: "%02x", $0) }.joined()
        print("metal library: embedded \(size) bytes sha256 \(digest)")
    }
}
