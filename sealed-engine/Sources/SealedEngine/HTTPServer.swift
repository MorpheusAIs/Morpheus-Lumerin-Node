import Foundation
import Network

struct HTTPRequest {
    let method: String
    let path: String
    let body: Data
}

/// Minimal HTTP/1.1 server bound to loopback only; the provider proxy-router
/// is the only intended client. One request per connection.
final class HTTPServer {
    typealias Handler = (HTTPRequest, @escaping (Int, [String: Any]) -> Void) -> Void

    private let listener: NWListener
    private let handler: Handler
    private let queue = DispatchQueue(label: "sealed-engine.http")
    private static let maxRequest = 1 << 20

    init(port: UInt16, handler: @escaping Handler) throws {
        let params = NWParameters.tcp
        params.requiredLocalEndpoint = .hostPort(host: "127.0.0.1", port: NWEndpoint.Port(rawValue: port)!)
        listener = try NWListener(using: params)
        self.handler = handler
    }

    func start() {
        listener.newConnectionHandler = { [weak self] conn in
            guard let self else { return }
            conn.start(queue: self.queue)
            self.read(conn, buffer: Data())
        }
        listener.start(queue: queue)
    }

    private func read(_ conn: NWConnection, buffer: Data) {
        conn.receive(minimumIncompleteLength: 1, maximumLength: 65536) { [weak self] data, _, complete, error in
            guard let self else { return }
            var buf = buffer
            if let data { buf.append(data) }
            if error != nil || buf.count > Self.maxRequest { return conn.cancel() }
            if let req = Self.parse(buf) {
                self.handler(req) { status, json in self.send(conn, status: status, json: json) }
            } else if complete {
                conn.cancel()
            } else {
                self.read(conn, buffer: buf)
            }
        }
    }

    /// Returns a request once headers and the full Content-Length body are present.
    private static func parse(_ buf: Data) -> HTTPRequest? {
        guard let sep = buf.range(of: Data("\r\n\r\n".utf8)),
              let head = String(data: buf[..<sep.lowerBound], encoding: .utf8) else { return nil }
        let lines = head.components(separatedBy: "\r\n")
        let parts = lines.first?.split(separator: " ") ?? []
        guard parts.count >= 2 else { return nil }
        var length = 0
        for line in lines.dropFirst() {
            let kv = line.split(separator: ":", maxSplits: 1).map { $0.trimmingCharacters(in: .whitespaces) }
            if kv.count == 2, kv[0].lowercased() == "content-length" { length = Int(kv[1]) ?? 0 }
        }
        guard length >= 0, length <= maxRequest else { return HTTPRequest(method: "", path: "", body: Data()) }
        let body = buf[sep.upperBound...]
        guard body.count >= length else { return nil }
        return HTTPRequest(method: String(parts[0]), path: String(parts[1]), body: Data(body.prefix(length)))
    }

    private func send(_ conn: NWConnection, status: Int, json: [String: Any]) {
        let body = (try? JSONSerialization.data(withJSONObject: json, options: [.sortedKeys])) ?? Data("{}".utf8)
        let reason = [200: "OK", 400: "Bad Request", 404: "Not Found", 409: "Conflict", 422: "Unprocessable Content", 503: "Service Unavailable"][status] ?? "Error"
        var head = "HTTP/1.1 \(status) \(reason)\r\nContent-Type: application/json\r\n"
        head += "Content-Length: \(body.count)\r\nConnection: close\r\n\r\n"
        conn.send(content: Data(head.utf8) + body, completion: .contentProcessed { _ in conn.cancel() })
    }
}
