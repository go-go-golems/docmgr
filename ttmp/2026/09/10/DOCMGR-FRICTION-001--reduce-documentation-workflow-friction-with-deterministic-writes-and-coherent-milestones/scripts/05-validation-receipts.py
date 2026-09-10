#!/usr/bin/env python3
"""Inventory committed implementation and retained test evidence, not a test runner.
Run after the documented tests. This script does not infer semantic completion.
"""
import hashlib
import json
import subprocess
import sys
from pathlib import Path

ticket = Path(__file__).resolve().parents[1]
repo = Path(subprocess.check_output(['git', 'rev-parse', '--show-toplevel'], cwd=ticket, text=True).strip())
skills = Path(sys.argv[1]).resolve()
skills_ticket = next((skills/'ttmp/2026/09/10').glob('SKILLS-FRICTION-001--*'))

def git(root, *args):
    return subprocess.check_output(['git', *args], cwd=root, text=True).strip()

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def inventory(root, base):
    names = git(root, 'diff', '--name-only', base, 'HEAD').splitlines()
    return {name: sha(root/name) for name in names if not name.startswith('ttmp/') and (root/name).is_file()}

def retain(root, data):
    (root/'sources/implementation-validation.json').write_text(json.dumps(data, indent=2)+'\n')

assert 'PASS:' in (ticket/'sources/implementation-cli-smoke.log').read_text()
assert 'FAIL' not in (ticket/'sources/implementation-tagged-tests.log').read_text()
assert 'FAIL' not in (ticket/'sources/implementation-race-tests.log').read_text()
assert '[ok] Scenario completed' in (ticket/'sources/implementation-scenarios.log').read_text()
assert 'affected by 0 vulnerabilities' in (ticket/'sources/implementation-vuln-after.log').read_text()
retain(ticket, {
    'schema_version': 1,
    'ticket': 'DOCMGR-FRICTION-001',
    'code_revision': git(repo, 'rev-parse', 'HEAD'),
    'binary_sha256': sha(Path('/tmp/docmgr-friction-local')),
    'checks': {
        'full_go_tests_and_aggregate_lint': 'passed in final code commit hook; includes pinned Glazed analyzer',
        'tagged_tests': 'go test -tags sqlite_fts5 ./...: passed',
        'race_tests': 'go test -race ./internal/operations ./pkg/commands ./internal/httpapi: passed',
        'static_and_build': 'go vet ./... and go build ./...: passed',
        'logcopter': 'make logcopter-check: passed',
        'cli': 'pinned binary smoke: passed',
        'scenarios': 'pinned binary repository scenario suite: passed at integration checkpoint 1516abc; later resume follow-up covered by focused tests and fresh CLI smoke',
        'frontend': 'pnpm exec tsc -b and targeted ESLint: passed',
        'security': 'govulncheck: zero reachable findings; one imported-package and five module advisories remain without reachable calls',
    },
    'limits': ['cooperative locks, not atomic multi-file visibility', 'process-crash recovery, not power-loss durability', 'uncooperative editor races remain outside contract', 'global installed binary not replaced; no push or new upload requested'],
    'committed_code_sha256': inventory(repo, '311498b'),
    'evidence_sha256': {p.name: sha(p) for p in sorted((ticket/'sources').glob('implementation-*.log'))},
})
assert 'Ran 10 tests' in (skills_ticket/'sources/implementation-tests.log').read_text()
assert json.loads((skills_ticket/'sources/implementation-skill-check.json').read_text()) == {'errors': []}
retain(skills_ticket, {
    'schema_version': 1,
    'ticket': 'SKILLS-FRICTION-001',
    'code_revision': git(skills, 'rev-parse', 'HEAD'),
    'checks': {'tests': '10 passed', 'owned_skill_conventions': 'passed', 'preexisting_modified_files': 'SHA256 unchanged; shared diary inode preserved'},
    'remaining': {'vewh': 'requires equivalent real post-adoption sessions; no synthetic fixture proves agent time/token savings'},
    'committed_code_sha256': inventory(skills, 'a879bfe'),
    'evidence_sha256': {p.name: sha(p) for p in sorted((skills_ticket/'sources').glob('implementation-*')) if p.name != 'implementation-validation.json'},
})
print('Wrote scoped implementation evidence and inventories for both tickets.')
