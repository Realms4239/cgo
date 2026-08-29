#!/usr/bin/env node
"use strict";

// postinstall: download the prebuilt cgo binary for this platform from the
// GitHub release and write it to bin/meteolink-<os>-<arch>(.exe).
// Node >= 16 compatible (https module, no fetch, no external deps).

const fs = require("fs");
const path = require("path");
const https = require("https");
const zlib = require("zlib");
const os = require("os");

const VERSION = require("./package.json").version;
const REPO = "Realms4239/cgo";
const BASE_URL = `https://github.com/${REPO}/releases/download/v${VERSION}`;

// {platform, arch} -> {os, arch} segments used in goreleaser asset names.
const PLATFORMS = {
  "darwin-arm64": { os: "darwin", arch: "arm64" },
  "darwin-x64": { os: "darwin", arch: "amd64" },
  "linux-x64": { os: "linux", arch: "amd64" },
  "linux-arm64": { os: "linux", arch: "arm64" },
  "win32-x64": { os: "windows", arch: "amd64" },
};

function platformKey() {
  return `${process.platform}-${os.arch()}`;
}

function binaryName() {
  const p = PLATFORMS[platformKey()];
  if (!p) return null;
  return `meteolink-${p.os}-${p.arch}${p.os === "windows" ? ".exe" : ""}`;
}

function assetName() {
  const p = PLATFORMS[platformKey()];
  if (!p) return null;
  // goreleaser archives: cgo-<os>-<arch>.tar.gz (unix), .zip (windows)
  return `cgo-${p.os}-${p.arch}${p.os === "windows" ? ".zip" : ".tar.gz"}`;
}

// ---------------------------------------------------------------------------
// Minimal archive extraction (builtins only).
// ---------------------------------------------------------------------------

// Extract a single named file from an uncompressed-in-memory tar stream.
// Handles ustar/GNU tar with short names (goreleaser asset entries qualify);
// GNU long-name / pax extended header blocks (typeflag 'L'/'x'/'g') are
// consumed and long names honored for the following entry.
function extractFromTarGz(buf, wanted) {
  const raw = zlib.gunzipSync(buf);
  let off = 0;
  let pendingLongName = null;
  while (off + 512 <= raw.length) {
    const header = raw.subarray(off, off + 512);
    if (header.every((b) => b === 0)) break;
    const nameField = header.subarray(0, 100).toString("utf8").replace(/\0.*$/, "");
    const sizeField = header.subarray(124, 136).toString("ascii").replace(/[\0 ]*$/, "");
    const size = parseInt(sizeField, 8) || 0;
    const typeflag = String.fromCharCode(header[156] || 48);
    off += 512;
    if (typeflag === "L" || typeflag === "x" || typeflag === "g") {
      const payload = raw.subarray(off, off + size).toString("utf8").replace(/\0.*$/, "");
      if (typeflag === "L") pendingLongName = payload;
      off += Math.ceil(size / 512) * 512;
      continue;
    }
    const name = pendingLongName !== null ? pendingLongName : nameField;
    pendingLongName = null;
    const base = name.split("/").pop();
    if (typeflag === 48 || typeflag === "0") {
      // regular file
      if (base === wanted) return raw.subarray(off, off + size);
      off += Math.ceil(size / 512) * 512;
    }
  }
  return null;
}

// Extract a single named file from a zip buffer using the central directory.
// Supports stored (method 0) and deflate (method 8) entries.
function extractFromZip(buf, wanted) {
  // Locate End Of Central Directory record (scan backwards for signature).
  const EOCD_SIG = 0x06054b50;
  let eocd = -1;
  for (let i = buf.length - 22; i >= Math.max(0, buf.length - 22 - 0xffff); i--) {
    if (buf.readUInt32LE(i) === EOCD_SIG) {
      eocd = i;
      break;
    }
  }
  if (eocd < 0) throw new Error("invalid zip archive (no EOCD record)");
  const entries = buf.readUInt16LE(eocd + 10);
  let ptr = buf.readUInt32LE(eocd + 16);
  for (let i = 0; i < entries; i++) {
    if (buf.readUInt32LE(ptr) !== 0x02014b50) throw new Error("invalid zip central directory");
    const method = buf.readUInt16LE(ptr + 10);
    const compressedSize = buf.readUInt32LE(ptr + 20);
    const nameLen = buf.readUInt16LE(ptr + 28);
    const extraLen = buf.readUInt16LE(ptr + 30);
    const commentLen = buf.readUInt16LE(ptr + 32);
    const localOffset = buf.readUInt32LE(ptr + 42);
    const name = buf.subarray(ptr + 46, ptr + 46 + nameLen).toString("utf8");
    if (name.split("/").pop() === wanted) {
      // Local file header: name/extra lengths may differ from central record.
      const lNameLen = buf.readUInt16LE(localOffset + 26);
      const lExtraLen = buf.readUInt16LE(localOffset + 28);
      const dataStart = localOffset + 30 + lNameLen + lExtraLen;
      const data = buf.subarray(dataStart, dataStart + compressedSize);
      return method === 0 ? Buffer.from(data) : zlib.inflateRawSync(data);
    }
    ptr += 46 + nameLen + extraLen + commentLen;
  }
  return null;
}

// ---------------------------------------------------------------------------
// Download + install.
// ---------------------------------------------------------------------------

function httpsGet(url, redirects) {
  redirects = redirects || 0;
  if (redirects > 5) return Promise.reject(new Error("too many redirects"));
  return new Promise((resolve, reject) => {
    https
      .get(url, { headers: { "User-Agent": `meteolink-npm/${VERSION}` } }, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          res.resume();
          resolve(httpsGet(new URL(res.headers.location, url).href, redirects + 1));
          return;
        }
        if (res.statusCode !== 200) {
          res.resume();
          const err = new Error(`HTTP ${res.statusCode} for ${url}`);
          err.statusCode = res.statusCode;
          reject(err);
          return;
        }
        const chunks = [];
        res.on("data", (c) => chunks.push(c));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

async function main() {
  const p = PLATFORMS[platformKey()];
  if (!p) {
    console.warn(`meteolink: unsupported platform ${platformKey()} — skipping binary install.`);
    return;
  }
  const outPath = path.join(__dirname, "bin", binaryName());
  if (fs.existsSync(outPath) && fs.statSync(outPath).size > 0) {
    console.log(`meteolink: binary already present (${binaryName()})`);
    return;
  }
  const asset = assetName();
  const url = `${BASE_URL}/${asset}`;
  console.log(`meteolink: downloading ${url}`);
  let buf;
  try {
    buf = await httpsGet(url);
  } catch (err) {
    if (err.statusCode === 404) {
      console.error(`meteolink: release asset not found: ${url}`);
      console.error(`meteolink: is v${VERSION} published with the ${asset} asset?`);
      console.error("meteolink: see https://github.com/Realms4239/cgo/releases");
    } else {
      console.error(`meteolink: download failed: ${err.message}`);
    }
    process.exit(1);
  }
  const binName = p.os === "windows" ? "cgo.exe" : "cgo";
  const binary = p.os === "windows" ? extractFromZip(buf, binName) : extractFromTarGz(buf, binName);
  if (!binary || binary.length === 0) {
    console.error(`meteolink: could not extract '${binName}' from ${asset}`);
    process.exit(1);
  }
  fs.mkdirSync(path.dirname(outPath), { recursive: true });
  fs.writeFileSync(outPath, binary);
  if (p.os !== "windows") fs.chmodSync(outPath, 0o755);
  console.log(`meteolink: installed ${binaryName()} (${binary.length} bytes)`);
}

main().catch((err) => {
  console.error(`meteolink: install failed: ${err.message}`);
  process.exit(1);
});

