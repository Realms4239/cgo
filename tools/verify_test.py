#!/usr/bin/env python3
"""Unit tests for tools/verify.py. No network, no bench, no real tree mutation."""
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import verify


class VerifyTests(unittest.TestCase):
    def test_verdict(self):
        r = verify.Result
        self.assertEqual(verify.verdict([r("a", verify.PASS, "")]), 0)
        self.assertEqual(verify.verdict([r("a", verify.PASS, ""), r("b", verify.SKIP, "")]), 2)
        self.assertEqual(verify.verdict([r("a", verify.FAIL, ""), r("b", verify.SKIP, "")]), 1)

    def test_staged_selection(self):
        cases = (
            (["book/data/runs/x"], ["frozen", "hygiene"]),
            (["pkg/api/x.go"], ["provenance", "exit", "hygiene"]),
            (["docs/x.md"], ["hygiene", "links"]),
            (["tools/verify.py"], list(verify.ALL_NAMES)),
            (["README.md"], ["hygiene"]),
        )
        for paths, want in cases:
            self.assertEqual(verify.select_staged(paths), want, paths)

    def test_provenance(self):
        with tempfile.TemporaryDirectory() as tmp:
            csv = Path(tmp) / "aqm_eval.csv"
            csv.write_text("a,b\n1,2\n", encoding="utf-8")
            self.assertEqual(verify.header_problems([csv], {"a,b"}), [])
            self.assertTrue(verify.header_problems([csv], {"a,c"}))
        self.assertIsNotNone(verify.valuation_problem({"basis": "guess"}))
        self.assertIsNotNone(verify.valuation_problem({"basis": "estimated", "source": "ebpf"}))
        self.assertIsNone(verify.valuation_problem({"basis": "measured", "source": "ebpf"}))

    def test_hygiene(self):
        key = "-----BEGIN " + "RSA PRIVATE KEY-----"
        token = "ghp_" + "A" * 36
        address = ".".join(("100", "99", "1", "2"))
        self.assertTrue(verify.scan_text("k", key.encode()))
        self.assertTrue(verify.scan_text("t", token.encode()))
        self.assertTrue(verify.scan_text("n", address.encode()))
        self.assertTrue(verify.path_problem("note.docx"))
        self.assertFalse(verify.scan_text("c", b'KEY=<your-key>\n10.30.0.2\npassword = ""\n'))
        self.assertIsNone(verify.path_problem("pkg/api/x.go"))

    def test_links(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / "docs").mkdir()
            (root / "AGENTS.md").write_text("see `docs/gone.md`\n", encoding="utf-8")
            self.assertTrue(verify.missing_links(root, ("AGENTS.md",)))
            (root / "docs/gone.md").write_text("x\n", encoding="utf-8")
            self.assertEqual(verify.missing_links(root, ("AGENTS.md",)), [])

    def test_exit_classifier(self):
        self.assertTrue(verify.classify_exit(3, "", ""))
        self.assertTrue(verify.classify_exit(0, "", "Traceback (most recent call last):"))
        self.assertTrue(verify.classify_exit(0, "no such file", ""))
        self.assertIsNone(verify.classify_exit(2, "", ""))
        self.assertIsNone(verify.verdict_contract())

    def test_missing_proof(self):
        class Bare:
            name = "bare"
            prove = None
        self.assertEqual(verify.missing_proofs([Bare()]), ["bare"])

    def test_line_ceiling(self):
        with tempfile.TemporaryDirectory() as tmp:
            big = Path(tmp) / "big.py"
            big.write_text("x\n" * 999, encoding="utf-8")
            self.assertIn("999", verify.line_ceiling_problem(tmp, ("big.py",), 100))
            big.write_text("x\n", encoding="utf-8")
            self.assertIsNone(verify.line_ceiling_problem(tmp, ("big.py",), 100))


if __name__ == "__main__":
    unittest.main()
