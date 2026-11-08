// Generates a new 2048-bit RSA key pair for the login token.
//
//   node tools/generate-keys.js
//
// The private key is written in PKCS#1 format (BEGIN RSA PRIVATE KEY), which
// is what client/token.go and client-native/token.go parse.
const fs = require("fs");
const path = require("path");
const { generateKeyPairSync } = require("crypto");

const root = path.join(__dirname, "..");
const privPath = path.join(root, "privatekey.pem");
const pubPath = path.join(root, "publickey.pem");

if (fs.existsSync(privPath) || fs.existsSync(pubPath)) {
  const backup = path.join(root, "keys-backup-" + Date.now());
  fs.mkdirSync(backup, { recursive: true });
  if (fs.existsSync(privPath)) fs.copyFileSync(privPath, path.join(backup, "privatekey.pem"));
  if (fs.existsSync(pubPath)) fs.copyFileSync(pubPath, path.join(backup, "publickey.pem"));
  console.log("existing keys backed up to " + backup);
}

const { privateKey, publicKey } = generateKeyPairSync("rsa", {
  modulusLength: 2048,
  privateKeyEncoding: { type: "pkcs1", format: "pem" },
  publicKeyEncoding: { type: "spki", format: "pem" },
});

fs.writeFileSync(privPath, privateKey);
fs.writeFileSync(pubPath, publicKey);

console.log("generated privatekey.pem (PKCS#1) and publickey.pem (SPKI)");
console.log("next: node tools/sync-keys.js");
