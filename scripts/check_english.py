#!/usr/bin/env python3
"""Fail when source code contains Spanish.

The codebase is English-only (identifiers, comments, logs, error messages).
User-facing Spanish copy lives in frontend/src/locales/es and is excluded.
Intentional Spanish data (sample names, fixtures) is listed in
scripts/english-allowlist.txt as `path-glob:regex` lines.

Usage: python3 scripts/check_english.py [--list]
"""

import fnmatch
import pathlib
import re
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
EXTENSIONS = {".go", ".ts", ".tsx", ".sql", ".css", ".yml", ".yaml", ".sh", ".py", ".html"}
EXCLUDED = [
    "frontend/src/locales/*",
    "frontend/src/types/api.ts",
    "docs/archive/*",
    "backend-go/internal/platform/database/sqlc/*",
    "scripts/check_english.py",
]

# Accented letters plus common Spanish words that are not English words.
SPANISH = re.compile(
    r"[áéíóúñÁÉÍÓÚÑ¿¡]"
    r"|\b(?:turnos?|clientes?|servicios?|nombre|apellido|fecha|precio|contrasenia"
    r"|administrador(?:es)?|estadisticas?|reservas?|ajustes|negocio|ingresos?"
    r"|cantidad|telefono|deberia|invalid[ao]s?|valid[ao]s?|tambien|segun"
    r"|que|para|los|las|una|del|por|cuando|desde|hasta|ahora|pero|nunca"
    r"|siempre|aqui|vacio|esta|este|estos|cada|todos|solo|debe|puede|hace"
    r"|de|se|al|un|con|abrir|crear|leer|guardar|borrar|buscar|mostrar|cargar|falla|error interno)\b",
    re.IGNORECASE,
)


def tracked_files():
    out = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard"],
        cwd=ROOT, capture_output=True, text=True, check=True,
    ).stdout.splitlines()
    for rel in out:
        p = ROOT / rel
        if p.suffix in EXTENSIONS and p.is_file() and not any(fnmatch.fnmatch(rel, g) for g in EXCLUDED):
            yield rel, p


def load_allowlist():
    rules = []
    path = ROOT / "scripts" / "english-allowlist.txt"
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if line and not line.startswith("#"):
            glob, _, rx = line.partition(":")
            rules.append((glob, re.compile(rx)))
    return rules


def main():
    allow = load_allowlist()
    hits = []
    for rel, p in tracked_files():
        for n, line in enumerate(p.read_text(encoding="utf-8", errors="ignore").splitlines(), 1):
            if SPANISH.search(line) and not any(fnmatch.fnmatch(rel, g) and rx.search(line) for g, rx in allow):
                hits.append(f"{rel}:{n}: {line.strip()[:140]}")
    if hits:
        print("\n".join(hits))
        print(f"\ncheck_english: {len(hits)} line(s) with Spanish. Translate them, or add an entry to scripts/english-allowlist.txt for intentional data.")
        return 1
    print("check_english: OK")
    return 0


if __name__ == "__main__":
    sys.exit(main())
