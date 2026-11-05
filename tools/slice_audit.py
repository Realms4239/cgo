#!/usr/bin/env python3
"""Audit de la matrice de slices : exécute chaque vérification, contrôle ses marqueurs, résume.

Lecture seule : le script n'écrit aucun fichier et n'appelle jamais git.
Sur Windows, shell=True passerait par cmd.exe, qui ne comprend ni /c/cgo ni
test -s ; les commandes de la matrice sont donc confiées à Git Bash (bash -c).
"""
from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MATRIX = ROOT / "tools" / "slice_matrix.json"
CWD = "C:/cgo"
TAIL_LINES = 25

for _stream in (sys.stdout, sys.stderr):
    _stream.reconfigure(encoding="utf-8", errors="replace")


def find_bash() -> str | None:
    """Chemin de Git Bash, ou None hors Windows (shell=True convient alors)."""
    if os.name != "nt":
        return None
    candidates = (
        "bash",
        r"C:\Program Files\Git\bin\bash.exe",
        r"C:\Program Files\Git\usr\bin\bash.exe",
    )
    for candidate in candidates:
        found = shutil.which(candidate) or (candidate if Path(candidate).is_file() else None)
        if found:
            return found
    return None


def execute(check: dict, bash: str | None) -> subprocess.CompletedProcess:
    """Lance une vérification avec son délai propre ; aucun shell intermédiaire."""
    common = dict(
        cwd=CWD,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
        timeout=check["timeout_s"],
    )
    if bash:
        return subprocess.run([bash, "-c", check["cmd"]], shell=False, **common)
    return subprocess.run(check["cmd"], shell=True, **common)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Audit read-only de tools/slice_matrix.json.")
    parser.add_argument("--list", action="store_true", help="liste les slices sans rien exécuter")
    parser.add_argument("--only", metavar="IDS", help="sous-ensemble d'identifiants, ex. 3,8,15")
    parser.add_argument("--no-global", action="store_true", help="saute global_checks")
    args = parser.parse_args(argv)

    matrix = json.loads(MATRIX.read_text(encoding="utf-8"))
    slices = matrix["slices"]
    global_checks = matrix.get("global_checks", [])

    if args.list:
        for entry in slices:
            print(f"{entry['id']}\t{entry['name']}\trunnable={entry['runnable']}\t"
                  f"checks={len(entry['checks'])}")
        return 0

    if args.only:
        wanted = [part.strip() for part in args.only.split(",") if part.strip()]
        known = {entry["id"] for entry in slices}
        unknown = [part for part in wanted if part not in known]
        if unknown:
            parser.error(f"unknown slice id(s): {', '.join(unknown)}")
        keep = set(wanted)
        slices = [entry for entry in slices if entry["id"] in keep]

    bash = find_bash()
    if os.name == "nt" and bash is None:
        print("ERROR: Git Bash introuvable; les commandes de la matrice sont en syntaxe POSIX.",
              file=sys.stderr)
        return 2

    started = time.monotonic()
    passed = failed = skipped = 0
    failing: list[str] = []

    def record_failure(sid: str) -> None:
        if sid not in failing:
            failing.append(sid)

    def run_check(sid: str, check: dict) -> None:
        nonlocal passed, failed
        begin = time.monotonic()
        try:
            proc = execute(check, bash)
        except subprocess.TimeoutExpired:
            elapsed = time.monotonic() - begin
            failed += 1
            record_failure(sid)
            print(f"FAIL {sid} {check['label']} ({elapsed:.1f}s)")
            print(f"    timeout after {check['timeout_s']} s")
            return
        elapsed = time.monotonic() - begin
        combined = (proc.stdout or "") + (proc.stderr or "")
        missing = [marker for marker in check.get("expect", []) if marker not in combined]
        if proc.returncode == 0 and not missing:
            passed += 1
            print(f"PASS {sid} {check['label']} ({elapsed:.1f}s)")
            return
        failed += 1
        record_failure(sid)
        print(f"FAIL {sid} {check['label']} ({elapsed:.1f}s)")
        print(f"    returncode: {proc.returncode}")
        print(f"    missing: {missing if missing else '(none)'}")
        tail = combined.rstrip("\n").splitlines()[-TAIL_LINES:]
        print(f"    output (last {TAIL_LINES} lines):")
        for line in tail:
            print(f"    {line}")

    for entry in slices:
        if entry["runnable"] == "no":
            skipped += 1
            print(f"SKIP {entry['id']} {entry['name']} — {entry['blocked_by']}")
            continue
        for check in entry["checks"]:
            run_check(entry["id"], check)

    if not args.no_global:
        for check in global_checks:
            run_check("global", check)

    total = time.monotonic() - started
    print(f"SUMMARY: {passed} passed, {failed} failed, {skipped} skipped, total {total:.1f}s")
    print("FAILING SLICES: " + (", ".join(failing) if failing else "none"))
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
