Upload markdown (or a ticket document) to a reMarkable device as a PDF.

This guide assumes you have the working uploader script installed at:

- `python3 /home/manuel/.local/bin/remarkable_upload.py`

What it does:
- Converts `.md` → `.pdf` via `pandoc` + `xelatex` (DejaVu fonts)
- Strips YAML frontmatter before conversion (prevents pandoc YAML parse issues and keeps metadata out of the PDF)
- Uploads via `rmapi` into `ai/YYYY/MM/DD/`
- Avoids overwriting unless you pass `--force`

## The mental model (for people new to this)

You always do the same 3 steps:

1) **Choose the markdown file(s)** you want to upload.
2) **Dry-run** to confirm:
   - destination folder (`ai/YYYY/MM/DD/`)
   - PDF filename(s)
   - the exact pandoc/rmapi commands that would run
3) **Run the real upload** (and only use `--force` if you explicitly want to overwrite).

## Quick Start (upload one markdown file)

1) Dry-run:

```bash
python3 /home/manuel/.local/bin/remarkable_upload.py --dry-run /abs/path/to/doc.md
```

2) Upload (no overwrite):

```bash
python3 /home/manuel/.local/bin/remarkable_upload.py /abs/path/to/doc.md
```

3) If it reports the PDF already exists, ask before overwriting, then:

```bash
python3 /home/manuel/.local/bin/remarkable_upload.py --force /abs/path/to/doc.md
```

## Upload a document that lives inside a ticket directory

The script can infer the destination date folder from ticket paths like:

- `.../ttmp/YYYY/MM/DD/TICKET--slug/...`

Recommended pattern (explicit doc path + ticket-dir for context):

```bash
python3 /home/manuel/.local/bin/remarkable_upload.py \
  --ticket-dir /abs/path/to/moments/ttmp/YYYY/MM/DD/TICKET--slug \
  --dry-run \
  /abs/path/to/moments/ttmp/YYYY/MM/DD/TICKET--slug/design-doc/01-something.md
```

Then run the same command without `--dry-run`.

## Choose / override the destination folder

Destination is always `ai/YYYY/MM/DD/`.

How the date is chosen:
- If the script can infer a ticket date from the ticket directory path: it uses that.
- Otherwise: it uses today’s date.

Override explicitly:

```bash
python3 /home/manuel/.local/bin/remarkable_upload.py --date 2025/12/11 /abs/path/to/doc.md
```

## Upload multiple markdown files

Just pass multiple paths:

```bash
python3 /home/manuel/.local/bin/remarkable_upload.py --dry-run /abs/a.md /abs/b.md
python3 /home/manuel/.local/bin/remarkable_upload.py /abs/a.md /abs/b.md
```

## Troubleshooting

- **`rmapi` not logged in / auth fails**:
  - Run `rmapi` manually once and complete authentication.
  - Then re-run the upload.

- **`pandoc` missing**:
  - Install pandoc.

- **`xelatex` missing**:
  - Install a TeX distribution that provides `xelatex`.

- **File not found**:
  - Use absolute paths to the markdown file.

- **A PDF already exists on the device**:
  - The script will skip without `--force`.
  - Only use `--force` if you explicitly want to overwrite.

## Getting Annotated PDFs from reMarkable

If you need to download annotated PDFs from the device, use the `remarks` tool (installed at `~/code/others/remarks`).

```bash
# 1. Download the .rmdoc archive from reMarkable
rmapi get ai/YYYY/MM/DD/DOCUMENT_NAME

# 2. Extract the archive and process with remarks
testdir=$(mktemp -d)
unzip -q DOCUMENT_NAME.rmdoc -d "$testdir"
outdir=$(mktemp -d)
cd ~/code/others/remarks
poetry run remarks "$testdir" "$outdir"

# 3. Copy the generated PDF (and markdown if needed)
cp "$outdir"/*.pdf ./DOCUMENT_NAME-annotated.pdf
cp "$outdir"/*.md ./DOCUMENT_NAME-annotated.md  # Optional: highlights in markdown

# 4. Cleanup
rm -rf "$testdir" "$outdir"
