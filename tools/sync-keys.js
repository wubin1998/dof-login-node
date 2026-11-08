// Copies the RSA private key from privatekey.pem into the embedded key
// constants used by both clients. Also verifies that privatekey.pem and
// publickey.pem are a matching pair.
//
//   node tools/sync-keys.js          # write the embedded keys
//   node tools/sync-keys.js --check  # only verify, do not modify
//
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

const root = path.join(__dirname, "..");
const privPath = path.join(root, "privatekey.pem");
const pubPath = path.join(root, "publickey.pem");

const priv = fs.readFileSync(privPath, "utf8").replace(/\r\n/g, "\n").trim() + "\n";
const pub = fs.readFileSync(pubPath, "utf8");

const norm = (s) => s.replace(/\r\n/g, "\n").trim();

if (!priv.includes("-----BEGIN RSA PRIVATE KEY-----")) {
  throw new Error("privatekey.pem 必须是 PKCS#1 格式（BEGIN RSA PRIVATE KEY），不能是 PKCS#8。");
}
if (!pub.includes("-----BEGIN PUBLIC KEY-----")) {
  throw new Error("publickey.pem 必须是 SPKI 格式（BEGIN PUBLIC KEY）。");
}

const pubFromPriv = crypto.createPublicKey(priv);
const pubFile = crypto.createPublicKey(pub);
const a = pubFromPriv.export({ format: "jwk" });
const b = pubFile.export({ format: "jwk" });
if (a.n !== b.n || a.e !== b.e) {
  throw new Error("privatekey.pem 和 publickey.pem 不是同一对密钥。");
}

const checkOnly = process.argv.includes("--check");
const targets = [
  { file: "client/token.go", re: /const privateKeyPEM = `[\s\S]*?`/ },
  { file: "client-native/token.go", re: /const privateKeyPEM = `[\s\S]*?`/ },
  { file: "legacy-node/main.js", re: /this\.PRIVATE_KEY = `[\s\S]*?`;/ },
];

let ok = true;
for (const target of targets) {
  const file = path.join(root, target.file);
  if (!fs.existsSync(file)) {
    console.log("skip (missing): " + target.file);
    continue;
  }

  const src = fs.readFileSync(file, "utf8");
  const match = src.match(target.re);
  if (!match) {
    console.log("ERROR: private key block not found in " + target.file);
    ok = false;
    continue;
  }

  const want = target.file.endsWith(".go")
    ? "const privateKeyPEM = `\n" + priv + "`"
    : "this.PRIVATE_KEY = `\n" + priv + "`;";

  if (checkOnly) {
    const pemRe = /-----BEGIN RSA PRIVATE KEY-----[\s\S]*?-----END RSA PRIVATE KEY-----/;
    const currentKey = (match[0].match(pemRe) || [""])[0];
    if (!currentKey || norm(currentKey) !== norm(priv)) {
      console.log("MISMATCH: " + target.file);
      ok = false;
    } else {
      console.log("OK: " + target.file);
    }
    continue;
  }

  if (match[0] !== want) {
    fs.writeFileSync(file, src.replace(target.re, want));
    console.log("updated: " + target.file);
  } else {
    console.log("unchanged: " + target.file);
  }
}

if (checkOnly) {
  if (!ok) {
    process.exitCode = 1;
  } else {
    console.log("all embedded private keys match privatekey.pem");
  }
} else {
  console.log("done. next: run go test ./... in client and client-native, then rebuild.");
}
