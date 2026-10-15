// A-13 probe: does a spawner-supplied LaunchCodeRequirement naming a CD hash
// make the kernel refuse an ad-hoc child whose CD hash differs?
// usage: a13 <child> <cdhashHex20 | bogus-team | none> <marker> [dumpfile]
// With a dumpfile, the encoded requirement bytes are written there (for a13spi).
import Foundation
import LightweightCodeRequirements

func hex(_ s: String) -> Data {
  var d = Data(); var i = s.startIndex
  while i < s.endIndex { let j = s.index(i, offsetBy: 2); d.append(UInt8(s[i..<j], radix: 16)!); i = j }
  return d
}

let args = CommandLine.arguments
let p = Process()
p.executableURL = URL(fileURLWithPath: args[1])
p.arguments = [args[3]]
do {
  if args[2] == "bogus-team" {
    p.launchRequirement = try LaunchCodeRequirement.allOf { TeamIdentifier("ZZZZZZZZZZ") }
  } else if args[2] != "none" {
    p.launchRequirement = try LaunchCodeRequirement.allOf { CodeDirectoryHash(hex(args[2])) }
  }
} catch { print("REQ BUILD ERROR: \(error)"); exit(2) }
if let d = (p as NSObject).value(forKey: "launchRequirementData") as? Data {
  print("launchRequirementData \(d.count) bytes")
  if args.count > 4 { try? d.write(to: URL(fileURLWithPath: args[4])); print("dumped to \(args[4])"); exit(0) }
}
do {
  try p.run(); p.waitUntilExit()
  print("RESULT reason=\(p.terminationReason == .exit ? "exit" : "signal") status=\(p.terminationStatus)")
} catch { print("RESULT spawn-error: \(error)") }
