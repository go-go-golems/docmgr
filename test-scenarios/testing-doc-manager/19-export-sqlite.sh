#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"
OUT_SQLITE="${ROOT_DIR}/workspace-index.sqlite"

echo "==> Exporting workspace index sqlite to ${OUT_SQLITE}"
"${DOCMGR}" workspace export-sqlite --out "${OUT_SQLITE}" --force --root ttmp

if [[ ! -s "${OUT_SQLITE}" ]]; then
  echo "Expected sqlite file to exist and be non-empty: ${OUT_SQLITE}" >&2
  exit 1
fi

python3 - "${OUT_SQLITE}" <<'PY'
import sqlite3
import sys

path = sys.argv[1]
conn = sqlite3.connect(path)
cur = conn.cursor()

def table_exists(name: str) -> bool:
    cur.execute("SELECT 1 FROM sqlite_master WHERE type='table' AND name=? LIMIT 1", (name,))
    return cur.fetchone() is not None

assert table_exists("README"), "README table missing"

cur.execute("SELECT COUNT(*) FROM README")
count = cur.fetchone()[0]
assert count >= 1, f"expected README rows >= 1, got {count}"

cur.execute("SELECT content FROM README WHERE name='__about__.md'")
row = cur.fetchone()
assert row and row[0].strip(), "__about__.md missing or empty"

# One known embedded doc should be present
cur.execute("SELECT COUNT(*) FROM README WHERE name='docmgr-how-to-use.md'")
how_to_use = cur.fetchone()[0]
assert how_to_use == 1, f"expected docmgr-how-to-use.md to exist in README, got {how_to_use}"

print("[ok] README table exists and contains embedded docs")
PY


