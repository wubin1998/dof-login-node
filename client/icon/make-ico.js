// Builds a multi-resolution .ico from the PNG files rendered from icon.html.
// Usage: node make-ico.js
const fs = require("fs");
const path = require("path");

const sizes = [16, 24, 32, 48, 64, 128, 256];
const images = sizes.map((size) => {
  const file = path.join(__dirname, `icon-${size}.png`);
  return { size, data: fs.readFileSync(file) };
});

const header = Buffer.alloc(6);
header.writeUInt16LE(0, 0); // reserved
header.writeUInt16LE(1, 2); // icon
header.writeUInt16LE(images.length, 4);

const entries = [];
let offset = 6 + images.length * 16;

for (const image of images) {
  const entry = Buffer.alloc(16);
  entry.writeUInt8(image.size >= 256 ? 0 : image.size, 0);
  entry.writeUInt8(image.size >= 256 ? 0 : image.size, 1);
  entry.writeUInt8(0, 2); // palette
  entry.writeUInt8(0, 3); // reserved
  entry.writeUInt16LE(1, 4); // color planes
  entry.writeUInt16LE(32, 6); // bits per pixel
  entry.writeUInt32LE(image.data.length, 8);
  entry.writeUInt32LE(offset, 12);
  entries.push(entry);
  offset += image.data.length;
}

const out = Buffer.concat([header, ...entries, ...images.map((image) => image.data)]);
fs.writeFileSync(path.join(__dirname, "..", "icon.ico"), out);
console.log(`wrote icon.ico (${out.length} bytes, ${images.length} sizes)`);
