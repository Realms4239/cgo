#!/usr/bin/env python3
"""Six-gate verifier for CGO. Standard library only; no network; no shell."""
from __future__ import annotations
import argparse, hashlib, json, os, re, shutil, subprocess, sys, tempfile, time
from dataclasses import dataclass
from pathlib import Path
ROOT = Path(__file__).resolve().parents[1]
PASS, FAIL, SKIP = "pass", "fail", "skip"
for _stream in (sys.stdout, sys.stderr):
    _stream.reconfigure(encoding="utf-8", errors="replace")
MAX_GATE_LINES = 650
CEILING_FILES = ("tools/verify.py", "tools/verify_test.py", ".githooks/pre-commit", ".githooks/pre-push")
SIZE_CEILING = 384 * 1024  # mesuré 2026-10-28 : plus gros fichier suivi 352 240 o
@dataclass
class Result:
    name: str
    status: str
    detail: str
@dataclass
class Gate:
    name: str
    proves: str
    paths: tuple
    run: object
    prove: object
    fast: bool = True
def _proc(args, timeout, cwd=ROOT, text=True, env=None):
    return subprocess.run([str(a) for a in args], cwd=str(cwd), capture_output=True, text=text,
                          encoding="utf-8" if text else None, errors="replace" if text else None,
                          timeout=timeout, check=False, env=env)
def _tail(text, limit=20):
    return "\n".join((text or "").strip().splitlines()[-limit:])
def _glob(root, name):
    return sorted(Path(root).glob(f"*/{name}")) if Path(root).is_dir() else []
def verdict(results):
    return 1 if any(r.status == FAIL for r in results) else 2 if any(r.status == SKIP for r in results) else 0
def line_ceiling_problem(root, files, maximum):
    counts = {n: len((Path(root) / n).read_text(encoding="utf-8", errors="replace").splitlines())
              for n in files if (Path(root) / n).exists()}
    total = sum(counts.values())
    return f"gate line ceiling exceeded: {total} > {maximum} ({counts})" if total > maximum else None
def missing_proofs(gates):
    return [g.name for g in gates if not callable(g.prove)]


# — frozen : l'audit du corpus est la seule source ------------------------
SEAL = "docs/frozen-manifests.sha256"
def seal_findings(root=ROOT):
    seal, runs = Path(root) / SEAL, Path(root) / "book/data/runs"
    if not seal.is_file():
        return [f"{SEAL}: anchor absent"]
    want = {}
    for line in seal.read_text(encoding="utf-8", errors="replace").splitlines():
        digest, sep, rel = line.partition("  ")
        if not sep or not rel.startswith("book/data/runs/") or not rel.endswith("/manifest.json"):
            return [f"{SEAL}: malformed line {line!r}"]
        want[rel] = digest
    got = {p.relative_to(root).as_posix(): hashlib.sha256(p.read_bytes()).hexdigest()
           for p in runs.glob("*/manifest.json")}
    return ([f"{r}: manifest changed" for r in sorted(want) if r in got and want[r] != got[r]]
            + [f"{r}: missing run" for r in sorted(want) if r not in got]
            + [f"{r}: run not in the anchor" for r in sorted(got) if r not in want])
def gate_frozen():
    audit, corpus = ROOT / "book/tools/audit-archive.py", ROOT / "book/data/runs"
    if not audit.is_file() or not corpus.is_dir():
        return Result("frozen", SKIP, "book/ absent: frozen corpus not checked")
    try:
        p = _proc([sys.executable, audit], 120)
    except subprocess.TimeoutExpired:
        return Result("frozen", FAIL, "audit timed out after 120 s")
    blob = (p.stdout or "") + (p.stderr or "")
    if p.returncode:
        return Result("frozen", FAIL, f"audit exit {p.returncode}:\n{_tail(blob)}")
    oks = [ln for ln in blob.splitlines() if ln.startswith("OK")]
    if len(oks) < 2:
        return Result("frozen", FAIL, f"audit exited 0 without two OK lines:\n{_tail(blob)}")
    findings = seal_findings()
    if findings:
        return Result("frozen", FAIL, f"{len(findings)} seal finding(s): " + "; ".join(findings[:5]))
    count = len(list(corpus.glob("*/manifest.json")))
    return Result("frozen", PASS, f"{' | '.join(oks)} | {SEAL}: {count} manifests")
def prove_frozen():
    if not (ROOT / "book").is_dir():
        return False, "book/ absent: the frozen proof cannot run"
    tmp = Path(tempfile.mkdtemp(prefix="cgo-frozen-"))
    try:
        dst = tmp / "book"
        shutil.copytree(ROOT / "book", dst, ignore=shutil.ignore_patterns("node_modules", ".git", "dist"))
        audit = dst / "tools/audit-archive.py"
        if _proc([sys.executable, audit], 120, cwd=tmp).returncode:
            return False, "control failed: the copied audit does not pass"
        seal, manifests = tmp / SEAL, sorted((dst / "data/runs").glob("*/manifest.json"))
        seal.parent.mkdir(parents=True, exist_ok=True)
        seal.write_bytes(("\n".join(f"{hashlib.sha256(p.read_bytes()).hexdigest()}  "
                                    f"{p.relative_to(tmp).as_posix()}" for p in manifests)
                          + "\n").encode("utf-8"))
        if seal_findings(tmp):
            return False, "control failed: the seal rejects the copied corpus"
        original = manifests[0].read_bytes()
        manifests[0].write_bytes(original + b"\n")
        if not seal_findings(tmp):
            return False, "SURVIVED: the seal accepts a rewritten manifest"
        if _proc([sys.executable, audit], 120, cwd=tmp).returncode:
            return False, "fault proof invalid: the audit rejects a whitespace-only manifest change"
        manifests[0].write_bytes(original)
        if seal_findings(tmp):
            return False, "control failed: the restored corpus no longer matches the seal"
        target = None
        for man in sorted((dst / "data/runs").glob("*/manifest.json")):
            try:
                files = json.loads(man.read_text(encoding="utf-8")).get("files", [])
            except (OSError, ValueError):
                continue
            target = next((man.parent / e["File"] for e in files
                           if e.get("File") and (man.parent / e["File"]).is_file()), None)
            if target:
                break
        if target is None:
            return False, "no manifest file found to mutate"
        with target.open("ab") as fh:
            fh.write(b"\n")
        if _proc([sys.executable, audit], 120, cwd=tmp).returncode == 0:
            return False, f"SURVIVED: audit exit 0 after mutating {target.relative_to(tmp)}"
        return True, f"control ok; seal caught a rewritten manifest; audit caught {target.relative_to(tmp)}"
    except Exception as exc:  # noqa: BLE001 — la preuve doit dire pourquoi
        return False, f"frozen proof error: {exc}"
    finally:
        shutil.rmtree(tmp, ignore_errors=True)


# — provenance : en-têtes CSV et enums de valuation -----------------------
BASES, SOURCES = {"measured", "estimated"}, {"ebpf", "tcpinfo", ""}
def csv_header(path):
    with Path(path).open("r", encoding="utf-8", errors="replace") as fh:
        return fh.readline().rstrip("\r\n")
def header_problems(csv_paths, reference):
    return [(str(p), "header not in the frozen eras") for p in csv_paths if csv_header(p) not in reference]
def valuation_problem(obj):
    if not isinstance(obj, dict):
        return "not a JSON object"
    basis, source = obj.get("basis"), obj.get("source", "")
    if basis not in BASES:
        return f"basis={basis!r} (allowed: measured, estimated)"
    if source not in SOURCES:
        return f"source={source!r} (allowed: ebpf, tcpinfo, empty)"
    return f"estimated with source={source!r}" if basis == "estimated" and source else None
def gate_provenance():
    frozen = ROOT / "book/data/runs"
    if not frozen.is_dir():
        return Result("provenance", SKIP, "book/data/runs absent: no reference header")
    reference = {csv_header(p) for p in _glob(frozen, "aqm_eval.csv")}
    if not reference:
        return Result("provenance", SKIP, "book/data/runs empty: no reference header")
    data_csvs = _glob(ROOT / "data/runs", "aqm_eval.csv")
    sidecars = _glob(ROOT / "data/runs", "valuation.json") + _glob(frozen, "valuation.json")
    bad = header_problems(data_csvs, reference)
    for path in sidecars:
        try:
            why = valuation_problem(json.loads(Path(path).read_text(encoding="utf-8")))
        except (OSError, ValueError) as exc:
            why = f"unreadable: {exc}"
        if why:
            bad.append((str(path), why))
    if bad:
        return Result("provenance", FAIL, f"{len(bad)} finding(s):\n" + "\n".join(f"{p}: {w}" for p, w in bad[:20]))
    side = f"{len(sidecars)} valuation sidecars checked" if sidecars else \
        "0 valuation sidecars checked; writer covered by tests"
    return Result("provenance", PASS,
                  f"{len(data_csvs)} headers checked against {len(reference)} frozen eras; {side}")
def prove_provenance():
    tmp = Path(tempfile.mkdtemp(prefix="cgo-prov-"))
    try:
        good = "run_id,event_id,profile"
        (tmp / "ref").mkdir()
        (tmp / "ref/aqm_eval.csv").write_text(good + "\n", encoding="utf-8")
        (tmp / "x").mkdir()
        (tmp / "x/aqm_eval.csv").write_text(good + ",extra\n", encoding="utf-8")
        checks = (
            (bool(header_problems([tmp / "x/aqm_eval.csv"], {good})), "extra column"),
            (valuation_problem({"basis": "guess"}) is not None, "basis=guess"),
            (valuation_problem({"basis": "estimated", "source": "ebpf"}) is not None, "estimated+ebpf"),
            (valuation_problem({"basis": "measured", "source": "ebpf"}) is None, "control measured+ebpf"),
            (valuation_problem({"basis": "estimated", "source": ""}) is None, "control estimated empty"),
            (not header_problems([tmp / "ref/aqm_eval.csv"], {good}), "control exact header"),
        )
        for ok, label in checks:
            if not ok:
                return False, f"SURVIVED or control failed: {label}"
        return True, "control ok; extra column and two bad sidecars caught"
    finally:
        shutil.rmtree(tmp, ignore_errors=True)


# — tests : la suite Go existante, jamais réécrite ------------------------
def gate_tests():
    if not (ROOT / "web/frontend/dist").is_dir():
        return Result("tests", SKIP, "web/frontend/dist absent: build the frontend first")
    go = shutil.which("go")
    if not go:
        return Result("tests", SKIP, "go not found on PATH")
    started = time.monotonic()
    for label, args, timeout in (("go vet", [go, "vet", "./..."], 300), ("go test", [go, "test", "./..."], 600)):
        try:
            p = _proc(args, timeout)
        except subprocess.TimeoutExpired:
            return Result("tests", FAIL, f"{label} timed out after {timeout} s")
        if p.returncode:
            return Result("tests", FAIL, f"{label} exit {p.returncode}:\n{_tail(p.stdout + p.stderr)}")
    return Result("tests", PASS, f"go vet + go test in {time.monotonic() - started:.1f} s")
def _module(path, body, name="ok_test.go"):
    path.mkdir()
    (path / "go.mod").write_text("module fixture\n\ngo 1.25\n", encoding="utf-8")
    (path / name).write_text(body, encoding="utf-8")
    return path
def prove_tests():
    go = shutil.which("go")
    if not go:
        return False, "go not found on PATH"
    tmp = Path(tempfile.mkdtemp(prefix="cgo-tests-"))
    try:
        ok = _module(tmp / "ok", 'package fixture\n\nimport "testing"\n\nfunc TestOK(t *testing.T) {}\n')
        bad = _module(tmp / "bad", 'package fixture\n\nimport "testing"\n\nfunc TestBad(t *testing.T) { t.Fatal("boom") }\n')
        vet = _module(tmp / "vet", 'package fixture\n\nimport "fmt"\n\nfunc F() { fmt.Printf("%d", "x") }\n', "vet.go")
        green = not _proc([go, "test", "./..."], 120, cwd=ok).returncode \
            and not _proc([go, "vet", "./..."], 120, cwd=ok).returncode
        if not green:
            return False, "control failed: passing module not green"
        if _proc([go, "test", "./..."], 120, cwd=bad).returncode == 0:
            return False, "SURVIVED: failing test module passed"
        if _proc([go, "vet", "./..."], 120, cwd=vet).returncode == 0:
            return False, "SURVIVED: vet error not caught"
        return True, "control ok; failing test and vet error caught"
    finally:
        shutil.rmtree(tmp, ignore_errors=True)


# — exit : le contrat 0/1/2 du CLI et du runner ---------------------------
SAFE_INVOCATIONS = (
    ((), "usage, no argument"),
    (("--nope",), "unknown command: usage then exit 2"),
    (("version",), "prints the version only"),
    (("shape",), "exits 2 before any tc call"),
    (("service",), "prints install usage, exits 2"),
)
REFUSAL = re.compile(r"(?i)no such file|not found|unknown|refus|erreur|error")
def classify_exit(code, out, err):
    if code not in (0, 1, 2):
        return f"unlisted exit code {code}"
    blob = (out or "") + "\n" + (err or "")
    if "panic:" in blob or "Traceback (most recent call last)" in blob:
        return "panic or traceback in output"
    if code == 0 and REFUSAL.search(blob):
        return "exit 0 while printing a refusal"
    return None
def verdict_contract():
    checks = (([Result("x", PASS, "")], 0), ([Result("x", PASS, ""), Result("y", SKIP, "")], 2),
              ([Result("x", FAIL, ""), Result("y", SKIP, "")], 1))
    for results, want in checks:
        if verdict(results) != want:
            return f"verdict contract broken: wanted {want}, got {verdict(results)}"
    return None
def gate_exit():
    go = shutil.which("go")
    if not go:
        return Result("exit", SKIP, "go not found on PATH")
    tmp = Path(tempfile.mkdtemp(prefix="cgo-exit-"))
    try:
        binary = tmp / ("cgo.exe" if os.name == "nt" else "cgo")
        build = _proc([go, "build", "-o", binary, "./cmd/cgo"], 300)
        if build.returncode:
            return Result("exit", FAIL, f"go build failed:\n{_tail(build.stdout + build.stderr)}")
        findings = []
        # Aucun identifiant dans l'environnement des commandes sondées.
        drop = re.compile(r"(?i)key|token|secret|password|credential")
        env = {k: v for k, v in os.environ.items() if not drop.search(k)}
        for args, reason in SAFE_INVOCATIONS:
            try:
                p = _proc([binary, *args], 30, env=env)
            except subprocess.TimeoutExpired:
                findings.append(f"{args}: timed out after 30 s")
                continue
            why = classify_exit(p.returncode, p.stdout, p.stderr)
            if why:
                findings.append(f"{args}: {why} ({reason})")
        if verdict_contract():
            findings.append(verdict_contract())
        if findings:
            return Result("exit", FAIL, "\n".join(findings[:20]))
        return Result("exit", PASS, f"{len(SAFE_INVOCATIONS)} safe invocations; verdict contract ok")
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
def prove_exit():
    fixtures = ((3, "", "", True), (0, "", "Traceback (most recent call last):", True),
                (0, "no such file", "", True), (2, "", "", False), (1, "usage", "", False))
    for code, out, err, must_fail in fixtures:
        why = classify_exit(code, out, err)
        if must_fail and not why:
            return False, f"SURVIVED: exit {code} / {err or out!r} accepted"
        if not must_fail and why:
            return False, f"control failed: exit {code} rejected ({why})"
    if verdict_contract():
        return False, verdict_contract()
    return True, "control ok; exit 3, panic, refusal and verdict contract caught"


# — hygiene : secrets, artefacts, adresses, taille ------------------------
KEY_RX = re.compile(r"-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----")
TOKEN_RX = re.compile(r"ghp_[A-Za-z0-9]{30,}|github[_-]?pat|sk-[A-Za-z0-9]{20,}|AKIA[0-9A-Z]{16}|xox[baprs]-")
NET_RX = re.compile(r"\b100\.(6[4-9]|[7-9][0-9]|1[01][0-9]|12[0-7])\.\d{1,3}\.\d{1,3}\b")
# Deux adresses nommées volontairement dans AGENTS.md et les registres
# append-only : on ne réécrit pas l'historique. Toute autre échoue.
ALLOWED_ADDRESSES = {"100.114.101.13", "100.72.251.32"}
ALLOWED_ARTIFACTS = {"pkg/results/runs-frozen.tar.gz", "modules/results/runs-frozen.tar.gz"}  # //go:embed : le build en dépend
# 18 archives suivies avant la règle « backup/ jamais commité » : dette nommée,
# tout nouveau fichier sous backup/ échoue.
LEGACY_ARCHIVES = frozenset("""
backup/runs-20260902-001002.tar.gz backup/runs-20260903-232145.tar.gz backup/runs-20260904-004749.tar.gz
backup/runs-20260904-024318.tar.gz backup/runs-20260904-034927.tar.gz backup/runs-20260904-053037.tar.gz
backup/runs-20260904-081048.tar.gz backup/runs-20260904-085058.tar.gz backup/runs-20260904-093525.tar.gz
backup/runs-20260904-101207.tar.gz backup/runs-20260904-104558.tar.gz backup/runs-20260904-110337.tar.gz
backup/runs-20260904-111003.tar.gz backup/runs-20260904-113340.tar.gz backup/runs-20260904-115508.tar.gz
backup/runs-20260904-121855.tar.gz backup/runs-20260904-124132.tar.gz backup/runs-20260904-125750.tar.gz
""".split())
USER_DOC_RX = re.compile(r"^(session-.*|handoff-.*|critical-review-.*)\.md$")
FORBIDDEN_PREFIXES = ("backup/", "data/", "book/", "testbed-bootstrap/", "node_modules/", ".worktrees/")
def path_problem(path):
    if path in ALLOWED_ARTIFACTS or path in LEGACY_ARCHIVES:
        return None
    if re.match(r"unp[^/]*/", path) or path.startswith(FORBIDDEN_PREFIXES):
        return "forbidden tracked path"
    if path == "epbfBBR.md" or USER_DOC_RX.match(path):
        return "forbidden user document"
    return "forbidden artifact" if path.endswith((".docx", ".pdf", ".tar.gz", ".zip", ".exe")) else None
def scan_text(name, data):
    text = data.decode("utf-8", "ignore")
    findings = []
    if KEY_RX.search(text):
        findings.append((name, "private key block"))
    if TOKEN_RX.search(text):
        findings.append((name, "token shape"))
    if any(m.group(0) not in ALLOWED_ADDRESSES for m in NET_RX.finditer(text)):
        findings.append((name, "tailnet address"))
    return findings
def hygiene_findings(paths, root):
    findings = []
    for path in paths:
        why = path_problem(path)
        if why:
            findings.append(f"{path}: {why}")
            continue
        full = Path(root) / path
        try:
            size = full.stat().st_size
            if size > SIZE_CEILING and not path.startswith("docs/img/"):
                findings.append(f"{path}: oversized ({size} bytes > {SIZE_CEILING})")
                continue
            data = full.read_bytes() if size <= 2_000_000 else b""
        except OSError:
            continue
        findings += [f"{name}: {rule}" for name, rule in scan_text(path, data)]
    return findings
def gate_hygiene():
    try:
        raw = _proc(["git", "ls-files", "-z"], 60, text=False).stdout.decode("utf-8", "replace")
    except (OSError, subprocess.SubprocessError) as exc:
        return Result("hygiene", SKIP, f"git unavailable: {exc}")
    paths = [p for p in raw.split("\0") if p]
    findings = hygiene_findings(paths, ROOT)
    if findings:
        return Result("hygiene", FAIL, f"{len(findings)} finding(s):\n" + "\n".join(findings[:20]))
    return Result("hygiene", PASS, f"{len(paths)} tracked files clean; legacy allowlist: "
                                  f"{len(LEGACY_ARCHIVES)} backup archives and pkg/results/runs-frozen.tar.gz; "
                                  f"{len(ALLOWED_ADDRESSES)} named tailnet addresses")
def prove_hygiene():
    key, token = "-----BEGIN " + "RSA PRIVATE KEY-----", "ghp_" + "A" * 36
    address = ".".join(("100", "99", "1", "2"))
    checks = (
        (bool(scan_text("k", key.encode())), "private key block"),
        (bool(scan_text("t", token.encode())), "token shape"),
        (bool(scan_text("n", address.encode())), "tailnet address"),
        (bool(path_problem("note.docx")), "tracked .docx"),
        (path_problem("pkg/api/x.go") is None, "control ordinary Go file"),
        (not scan_text("c", b'KEY=<your-key>\npassword = ""\n10.30.0.2\n'), "control placeholders"),
    )
    for ok, label in checks:
        if not ok:
            return False, f"SURVIVED or control failed: {label}"
    return True, "control ok; key, token, tailnet address and .docx caught"


# — links : AGENTS.md, registres et index pointent des fichiers réels -----
LINK_FILES = ("AGENTS.md", "docs/VERIFY.md", "docs/CHECKPOINTS.md", "docs/ERRORS.md",
              "docs/superpowers/plans/README.md")
def link_candidates(text, root):
    for token in re.findall(r"`([^`\n]+)`", text):
        tok = token.strip()
        if not tok or " " in tok or any(c in tok for c in "*<>{}:"):
            continue
        if tok.startswith(("http", "git ", "go ", "make ", "python ", "-", ".")) or "/" not in tok:
            continue
        first = tok.split("/")[0]
        if first[:1].isalpha() and (Path(root) / first).exists():
            yield tok
def missing_links(root, docs):
    out = []
    for name in docs:
        path = Path(root) / name
        if not path.exists():
            out.append((name, "scanned file is absent"))
            continue
        for tok in link_candidates(path.read_text(encoding="utf-8", errors="replace"), root):
            if not (Path(root) / tok.rstrip("/")).exists():
                out.append((name, tok))
    return out
def missing_plan_links(text, root=ROOT):
    dirs = (Path(root) / "docs/superpowers/specs", Path(root) / "docs/superpowers/plans")
    return [tok for tok in re.findall(r"`([^`\n]+\.md)`", text)
            if "/" not in tok and re.match(r"2026-\d{2}-\d{2}-.*\.md$", tok)
            and not any((d / tok).exists() for d in dirs)]
def gate_links():
    absent = [name for name in LINK_FILES if not (ROOT / name).exists()]
    if absent:
        return Result("links", SKIP, f"{absent[0]} absent: links not checked")
    index = ROOT / "docs/superpowers/plans/README.md"
    bad = missing_links(ROOT, LINK_FILES) + [
        (index.name, tok) for tok in missing_plan_links(index.read_text(encoding="utf-8", errors="replace"))]
    if bad:
        return Result("links", FAIL, f"{len(bad)} missing path(s):\n" + "\n".join(f"{s}: {t}" for s, t in bad[:20]))
    return Result("links", PASS, f"{len(LINK_FILES)} documents scanned; plan and spec links resolve")
def prove_links():
    tmp = Path(tempfile.mkdtemp(prefix="cgo-links-"))
    try:
        (tmp / "docs").mkdir()
        (tmp / "AGENTS.md").write_text("see `docs/gone.md`\n", encoding="utf-8")
        if not missing_links(tmp, ("AGENTS.md",)):
            return False, "SURVIVED: a missing path was not caught"
        (tmp / "docs/gone.md").write_text("x\n", encoding="utf-8")
        if missing_links(tmp, ("AGENTS.md",)):
            return False, "control failed: an existing path was rejected"
        (tmp / "docs/superpowers/plans").mkdir(parents=True)
        (tmp / "docs/superpowers/specs").mkdir(parents=True)
        if not missing_plan_links("`2026-01-01-gone-design.md`", tmp):
            return False, "SURVIVED: a missing plan link was not caught"
        (tmp / "docs/superpowers/specs/2026-01-01-live-design.md").write_text("x\n", encoding="utf-8")
        if missing_plan_links("`2026-01-01-live-design.md`", tmp):
            return False, "control failed: an existing spec link was rejected"
        if missing_plan_links("`make verify` `https://example.org/x`", tmp):
            return False, "control failed: a command or URL was treated as a path"
        return True, "control ok; missing path and missing plan link caught"
    finally:
        shutil.rmtree(tmp, ignore_errors=True)


# — registre --------------------------------------------------------------
GATES = (
    Gate("frozen", "the frozen corpus audit still attests", ("book/", "data/"), gate_frozen, prove_frozen),
    Gate("provenance", "CSV headers and valuation enums cannot lie", ("pkg/", "cmd/"), gate_provenance, prove_provenance),
    Gate("tests", "go vet and the Go suite pass", (), gate_tests, prove_tests, fast=False),
    Gate("exit", "the CLI keeps the 0/1/2 contract", ("pkg/", "cmd/"), gate_exit, prove_exit),
    Gate("hygiene", "no tracked secret, artifact or oversized file", ("tools/",), gate_hygiene, prove_hygiene),
    Gate("links", "AGENTS.md and the ledgers name existing paths", ("docs/", "AGENTS.md"), gate_links, prove_links),
)
ALL_NAMES = tuple(g.name for g in GATES)
def select_staged(changed):
    if any(p in ("tools/verify.py", "tools/verify_test.py") for p in changed):
        return list(ALL_NAMES)
    names = {"hygiene"}
    for gate in GATES:
        if gate.paths and any(p.startswith(gate.paths) for p in changed):
            names.add(gate.name)
    return [n for n in ALL_NAMES if n in names]
def staged_paths():
    try:
        raw = _proc(["git", "diff", "--cached", "--name-only", "-z"], 60, text=False).stdout
        return [p for p in raw.decode("utf-8", "replace").split("\0") if p]
    except (OSError, subprocess.SubprocessError):
        return []
def main(argv=None):
    parser = argparse.ArgumentParser(description="CGO six-gate verifier")
    parser.add_argument("--fast", action="store_true", help="all gates except tests")
    parser.add_argument("--staged", action="store_true", help="select gates from staged paths")
    parser.add_argument("--only", action="append", metavar="NAME", help="named gate (repeatable)")
    parser.add_argument("--list", action="store_true", help="list gates and their proofs")
    parser.add_argument("--selftest", action="store_true", help="run every gate fault proof")
    args = parser.parse_args(argv)
    problem = line_ceiling_problem(ROOT, CEILING_FILES, MAX_GATE_LINES)
    if problem:
        print(problem, file=sys.stderr)
        return 2
    if missing_proofs(GATES):
        print(f"gate without a fault proof: {', '.join(missing_proofs(GATES))}", file=sys.stderr)
        return 2
    if args.list:
        for gate in GATES:
            print(f"{gate.name:11} {'fast' if gate.fast else 'full':4} proof  {gate.proves}")
        return 0
    if args.only:
        unknown = [n for n in args.only if n not in ALL_NAMES]
        if unknown:
            print(f"unknown gate: {', '.join(unknown)} (known: {', '.join(ALL_NAMES)})", file=sys.stderr)
            return 2
        selected = [g for g in GATES if g.name in args.only]
    elif args.staged:
        chosen = set(select_staged(staged_paths()))
        selected = [g for g in GATES if g.name in chosen]
    elif args.fast:
        selected = [g for g in GATES if g.fast]
    else:
        selected = list(GATES)
    if args.selftest:
        survived = 0
        for gate in selected:
            caught, detail = gate.prove()
            print(f"{'caught  ' if caught else 'SURVIVED'} {gate.name:11} {detail}")
            survived += not caught
        return 1 if survived else 0
    results = [gate.run() for gate in selected]
    for result in results:
        print(f"{result.status.upper():4} {result.name:11} {result.detail}")
    failed = sum(r.status == FAIL for r in results)
    skipped = sum(r.status == SKIP for r in results)
    print(f"\n{len(results) - failed - skipped} passed, {failed} failed, {skipped} not run")
    return verdict(results)


if __name__ == "__main__":
    sys.exit(main())
