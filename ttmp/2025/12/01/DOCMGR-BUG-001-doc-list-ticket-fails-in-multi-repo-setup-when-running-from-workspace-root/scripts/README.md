# Bug Reproduction Scripts

## 01-reproduce-multi-repo-bug.sh

Reproduces DOCMGR-BUG-001: `docmgr doc list --ticket` fails in multi-repo setup when running from workspace root.

### Usage

```bash
# From the docmgr directory
cd /path/to/docmgr

# Run the reproduction script
./ttmp/2025/12/01/DOCMGR-BUG-001-.../scripts/01-reproduce-multi-repo-bug.sh [TEMP_DIR]

# Or specify a custom temp directory
./ttmp/2025/12/01/DOCMGR-BUG-001-.../scripts/01-reproduce-multi-repo-bug.sh /tmp/my-test
```

### What it does

1. Creates a multi-repo workspace structure:
   ```
   workspace/
   ├── .ttmp.yaml          (points to project1/ttmp)
   ├── project1/
   │   └── ttmp/
   └── project2/
       └── ttmp/
   ```

2. Initializes `project1/ttmp` with `docmgr init`

3. Creates a ticket `SOME-TICKET` and adds a test document

4. Tests `doc list --ticket SOME-TICKET`:
   - From `project1/ttmp`: ✅ Should work
   - From `workspace/`: ❌ Currently fails (BUG)

### Expected Output

The script will show:
- ✅ Success when running from `project1/ttmp`
- ❌ Failure when running from `workspace/` (confirms the bug)

### Environment Variables

- `DOCMGR_PATH`: Path to docmgr binary (default: `./docmgr`)
  ```bash
  DOCMGR_PATH=/usr/local/bin/docmgr ./01-reproduce-multi-repo-bug.sh
  ```

### Cleanup

The script creates a temporary directory (default: `/tmp/docmgr-bug-repro`) that can be safely deleted after testing.

