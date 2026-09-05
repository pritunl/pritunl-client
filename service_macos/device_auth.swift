import CryptoKit
import LocalAuthentication
import Security
import Foundation
import Darwin.C

let userPresence = false

// Secure enclave operations can fail transiently for a short period
// after the system wakes from sleep while the enclave and keychain
// services come back up. Retry a bounded number of times before failing.
// The total retry budget must stay below the open timeout used by the
// service and the helper lifetime enforced by the client.
let retryAttempts = 6
let retryDelayUsec: UInt32 = 400_000

struct Input: Codable {
  var key_data: String
}

struct Input2: Codable {
  var sign_data: String
}

struct Output: Codable {
  var key_data: String
  var public_key: String
}

struct Output2: Codable {
  var signature: String
}

enum DeviceAuthError: Error, CustomStringConvertible {
  case missingInput
  case invalidKeyData
  case invalidSignData
  case enclaveUnavailable
  case authFailed(String)

  var description: String {
    switch self {
    case .missingInput:
      return "missing input"
    case .invalidKeyData:
      return "invalid key data"
    case .invalidSignData:
      return "invalid sign data"
    case .enclaveUnavailable:
      return "secure enclave not available"
    case .authFailed(let msg):
      return "authentication failed: " + msg
    }
  }
}

let encoder = JSONEncoder()
let decoder = JSONDecoder()

func logErr(_ msg: String) {
  FileHandle.standardError.write((msg + "\n").data(using: .utf8)!)
}

func fail(_ msg: String) -> Never {
  logErr(msg)
  exit(1)
}

func retry<T>(_ label: String, _ op: () throws -> T) throws -> T {
  var lastErr: Error = DeviceAuthError.enclaveUnavailable
  for attempt in 1...retryAttempts {
    do {
      return try op()
    } catch {
      lastErr = error
      logErr("\(label) attempt \(attempt) failed: \(error)")
      if attempt < retryAttempts {
        usleep(retryDelayUsec)
      }
    }
  }
  throw lastErr
}

func loadKey(
  _ keyData: String,
  _ authContext: LAContext,
  _ accessControl: SecAccessControl
) throws -> SecureEnclave.P256.Signing.PrivateKey {
  if keyData == "" {
    return try CryptoKit.SecureEnclave.P256.Signing.PrivateKey(
      accessControl: accessControl,
      authenticationContext: authContext
    )
  }

  guard let keyDataRep = Data(base64Encoded: keyData) else {
    throw DeviceAuthError.invalidKeyData
  }

  return try CryptoKit.SecureEnclave.P256.Signing.PrivateKey(
    dataRepresentation: keyDataRep,
    authenticationContext: authContext
  )
}

func run(
  _ input: Input,
  _ authContext: LAContext,
  _ accessControl: SecAccessControl
) throws {
  let enclaveKey = try retry("load key") {
    try loadKey(input.key_data, authContext, accessControl)
  }

  var output = Output(
    key_data: enclaveKey.dataRepresentation.base64EncodedString(),
    public_key: enclaveKey.publicKey.derRepresentation.base64EncodedString()
  )

  let outputData = try encoder.encode(output)
  let outputStr = String(decoding: outputData, as: UTF8.self)
  print(outputStr)
  fflush(stdout)
  output.key_data = ""

  // An open only request closes stdin after reading the key output,
  // exit cleanly without signing.
  guard let input2Str = readLine() else {
    return
  }
  let input2Data = input2Str.trimmingCharacters(
    in: .whitespacesAndNewlines).data(using: .utf8)!
  let input2 = try decoder.decode(Input2.self, from: input2Data)

  guard let signDataBytes = Data(base64Encoded: input2.sign_data) else {
    throw DeviceAuthError.invalidSignData
  }

  let signature = try retry("sign") {
    try enclaveKey.signature(for: signDataBytes)
  }

  let output2 = Output2(
    signature: signature.derRepresentation.base64EncodedString()
  )
  let output2Data = try encoder.encode(output2)
  let output2Str = String(decoding: output2Data, as: UTF8.self)
  print(output2Str)
  fflush(stdout)
}

guard let inputStr = readLine() else {
  fail("missing input")
}
let inputData = inputStr.trimmingCharacters(
  in: .whitespacesAndNewlines).data(using: .utf8)!

let input: Input
do {
  input = try decoder.decode(Input.self, from: inputData)
} catch {
  fail("invalid input: \(error)")
}

do {
  try retry("enclave available") {
    if !SecureEnclave.isAvailable {
      throw DeviceAuthError.enclaveUnavailable
    }
  }
} catch {
  fail("secure enclave not available")
}

let authContext = LAContext()
let accessControl = SecAccessControlCreateWithFlags(
  kCFAllocatorDefault,
  kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
  [.privateKeyUsage],
  nil
)!

if (userPresence) {
  var authError: Error?
  let waiters = DispatchGroup()
  waiters.enter()

  authContext.evaluatePolicy(
    LAPolicy.deviceOwnerAuthentication,
    localizedReason: "authenticate device"
  ) { (success: Bool, err: Error?) -> Void in
    defer {
      waiters.leave()
    }

    if let err = err {
      authError = DeviceAuthError.authFailed("\(err)")
      return
    }

    do {
      try run(input, authContext, accessControl)
    } catch {
      authError = error
    }
  }

  waiters.wait()

  if let authError = authError {
    fail("device auth error: \(authError)")
  }
} else {
  do {
    try run(input, authContext, accessControl)
  } catch {
    fail("device auth error: \(error)")
  }
}

exit(0)
