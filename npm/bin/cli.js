#!/usr/bin/env node
"use strict";

// meteolink — lance le binaire cgo téléchargé au postinstall (archive unique
// cgo-<os>-<arch>). Tous les sous-commandes passent tels quels : tui, setup,
// kit, run, --serve, top, audit…
const { spawn } = require("child_process");
const path = require("path");
const os = require("os");

const PLATFORMS = {
  "darwin-arm64": "meteolink-darwin-arm64",
  "darwin-x64": "meteolink-darwin-amd64",
  "linux-x64": "meteolink-linux-amd64",
  "linux-arm64": "meteolink-linux-arm64",
  "win32-x64": "meteolink-windows-amd64.exe",
};

const key = `${process.platform}-${os.arch()}`;
const bin = PLATFORMS[key];
if (!bin) {
  console.error(`meteolink: plateforme non supportée ${key}`);
  process.exit(1);
}

const child = spawn(path.join(__dirname, bin), process.argv.slice(2), {
  stdio: "inherit",
});
child.on("exit", (code, signal) => {
  if (signal) process.kill(process.pid, signal);
  else process.exit(code ?? 0);
});
