#!/usr/bin/env python3
"""Executa a suite Go e registra cobertura, ambiente e revisao da medicao."""
from datetime import date
import json
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[2]


def main():
    engine = ROOT / "Codigo/poc-engine"
    result = subprocess.run(["go", "test", "-count=1", "-v", "-cover", "./..."],
                            cwd=engine, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(result.stdout, end="")
    result.check_returncode()
    coverage = float(re.findall(r"coverage: ([\d.]+)%", result.stdout)[-1])
    if coverage < 80:
        raise SystemExit("Cobertura abaixo da meta de 80%")
    baseline = ROOT / "Docs/baseline"
    baseline.mkdir(exist_ok=True)
    (baseline / "coverage_current.txt").write_text(result.stdout)
    record = {"coverage_pct": coverage,
              "command": "go test -count=1 -v -cover ./...",
              "go_version": subprocess.check_output(["go", "version"], text=True).strip(),
              "base_revision": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip(),
              "working_tree_changes": bool(subprocess.check_output(["git", "diff", "--name-only"], cwd=ROOT)),
              "date": date.today().isoformat()}
    (baseline / "coverage_current.json").write_text(json.dumps(record, indent=2) + "\n")
    tex = r"\providecommand{\pocCoverage}{" + str(coverage).replace(".", "{,}") + "}\n"
    (ROOT / "Docs/monografia/resultados_verificacao.tex").write_text(tex)
    readme = ROOT / "README.md"
    content = re.sub(r"PoC%20Tests-PASS%20\([\d.]+%25%20cov", f"PoC%20Tests-PASS%20({coverage:.1f}%25%20cov", readme.read_text())
    content = re.sub(r"Suíte completa de validação com `[\d,]+%` de cobertura",
                     f"Suíte completa de validação com `{coverage:.1f}%` de cobertura".replace(".", ","), content)
    readme.write_text(content)


if __name__ == "__main__":
    main()
