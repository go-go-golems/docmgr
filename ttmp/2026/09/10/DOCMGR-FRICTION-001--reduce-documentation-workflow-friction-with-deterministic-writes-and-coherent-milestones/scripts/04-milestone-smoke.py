#!/usr/bin/env python3
"""Pinned-binary CLI smoke; all writes are confined to a temporary workspace."""
import json
import pathlib
import subprocess
import sys
import tempfile

binary = str(pathlib.Path(sys.argv[1]).resolve())
with tempfile.TemporaryDirectory(prefix='docmgr-milestone-') as tmp:
    root = pathlib.Path(tmp)
    def run(*args, ok=True):
        p = subprocess.run([binary, *args], cwd=root, capture_output=True, text=True)
        assert (p.returncode == 0) == ok, (args, p.stdout, p.stderr)
        return p.stdout
    run('init', '--seed-vocabulary')
    run('ticket', 'create-ticket', '--ticket', 'SMOKE', '--title', 'Milestone smoke', '--topics', 'docmgr')
    ticket = next((root/'ttmp').glob('*/*/*/SMOKE--*'))
    (ticket/'tasks.md').write_text('# Tasks\n\n- [ ] Verify <!-- t:ab12 -->\n')
    args = ('milestone', 'record', '--ticket', 'SMOKE', '--operation-id', 'verified', '--summary', 'Verified fixture', '--phase', 'validate', '--task-id', 'ab12', '--next', 'Review')
    run(*args, '--dry-run')
    assert not (ticket/'.docmgr-operations').exists()
    run(*args)
    original = json.loads((ticket/'.docmgr-operations/verified.json').read_text())['receipt']
    rows = json.loads(run(*args, '--with-glaze-output', '--output', 'json'))
    assert rows[0]['receipt'] == original
    view = json.loads(run('ticket', 'resume', '--ticket', 'SMOKE', '--with-glaze-output', '--output', 'json'))[0]['resume']
    assert view['remaining'] == [] and view['next'] == 'Review' and view['conflicts'] == []
    run('ticket', 'close', '--ticket', 'SMOKE', '--operation-id', 'close-smoke')
    before = {p.name: p.read_bytes() for p in (ticket/'index.md', ticket/'changelog.md')}
    run('ticket', 'close', '--ticket', 'SMOKE')
    assert before == {p.name: p.read_bytes() for p in (ticket/'index.md', ticket/'changelog.md')}
    run('help', 'milestone-workflows')
print('PASS: dry-run, bare/JSON replay, resume, close no-op, embedded help')
