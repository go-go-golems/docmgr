#!/usr/bin/env python3
"""
Upload markdown documents to a reMarkable tablet.

Workflow (per file):
1) Convert .md -> .pdf using pandoc with xelatex + DejaVu fonts (Unicode-safe)
2) Check if the destination PDF already exists on the device (rmapi ls / rmapi get)
3) Upload via rmapi put ai/YYYY/MM/DD/ (use --force to overwrite)
4) Cleanup temporary directory
"""

from __future__ import annotations

import argparse
import datetime as _dt
import os
import shutil
import subprocess
import sys
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import Optional


@dataclass(frozen=True)
class UploadTarget:
    md_path: Path
    pdf_name: str


def _run(
    argv: list[str],
    *,
    check: bool = True,
    capture: bool = False,
    cwd: Optional[Path] = None,
) -> subprocess.CompletedProcess[str]:
    try:
        return subprocess.run(
            argv,
            check=check,
            text=True,
            cwd=str(cwd) if cwd else None,
            stdout=subprocess.PIPE if capture else None,
            stderr=subprocess.STDOUT if capture else None,
        )
    except FileNotFoundError as e:
        raise RuntimeError(f"Command not found: {argv[0]}") from e


def _infer_ticket_date_from_path(ticket_dir: Path) -> Optional[str]:
    """
    Tries to infer YYYY/MM/DD from a ticket directory like:
      .../ttmp/2025/12/11/TICKET--slug
    Returns "YYYY/MM/DD" or None if not found.
    """
    parts = ticket_dir.parts
    # Find ".../ttmp/YYYY/MM/DD/<ticket>"
    for i in range(len(parts) - 4):
        if parts[i] == "ttmp":
            yyyy, mm, dd = parts[i + 1 : i + 4]
            if len(yyyy) == 4 and len(mm) == 2 and len(dd) == 2 and yyyy.isdigit() and mm.isdigit() and dd.isdigit():
                return f"{yyyy}/{mm}/{dd}"
    return None


def _default_date(ticket_dir: Path) -> str:
    inferred = _infer_ticket_date_from_path(ticket_dir)
    if inferred:
        return inferred
    today = _dt.date.today()
    return f"{today:%Y/%m/%d}"


def _normalize_rm_dir(date_ymd: str) -> str:
    date_ymd = date_ymd.strip().strip("/")
    # Accept YYYY/MM/DD or YYYY-MM-DD
    if "-" in date_ymd and "/" not in date_ymd:
        parts = date_ymd.split("-")
        if len(parts) == 3:
            date_ymd = "/".join(parts)
    parts = date_ymd.split("/")
    if len(parts) != 3 or any(not p.isdigit() for p in parts):
        raise ValueError(f"Invalid date format: {date_ymd!r} (expected YYYY/MM/DD or YYYY-MM-DD)")
    yyyy, mm, dd = parts
    if len(yyyy) != 4 or len(mm) != 2 or len(dd) != 2:
        raise ValueError(f"Invalid date format: {date_ymd!r} (expected YYYY/MM/DD)")
    return f"ai/{yyyy}/{mm}/{dd}/"


def _strip_yaml_frontmatter(md_text: str) -> str:
    """
    Removes a leading YAML frontmatter block delimited by:
      ---\n
      ...yaml...
      ---\n
    If no such block exists, returns input unchanged.

    This is intentionally simple and matches docmgr's strict delimiter logic.
    It also avoids pandoc failing on invalid YAML frontmatter (e.g. unquoted ':' in scalars).
    """
    if not md_text.startswith("---"):
        return md_text

    lines = md_text.splitlines()
    if not lines:
        return md_text
    if lines[0].strip() != "---":
        return md_text

    end_idx = -1
    for i in range(1, len(lines)):
        if lines[i].strip() == "---":
            end_idx = i
            break
    if end_idx < 0:
        return md_text

    body_lines = lines[end_idx + 1 :]
    return "\n".join(body_lines).lstrip("\n")


def _normalize_list_spacing(md_text: str) -> str:
    """
    Ensures proper spacing before bullet/numbered lists so pandoc recognizes them.
    
    Inserts a blank line before list items that are not preceded by a blank line.
    Handles both unordered (-, *, +) and ordered (1., 2., etc.) lists.
    """
    lines = md_text.splitlines()
    if not lines:
        return md_text
    
    result = []
    for i, line in enumerate(lines):
        stripped = line.lstrip()
        
        # Check if this line is a list item (bullet or numbered)
        is_list_item = False
        if stripped:
            # Unordered list: starts with -, *, or + followed by space
            if stripped[0] in "-*+" and len(stripped) > 1 and stripped[1] == " ":
                is_list_item = True
            # Ordered list: starts with digit(s) followed by . and space
            elif stripped[0].isdigit():
                # Check for pattern like "1. " or "12. "
                j = 1
                while j < len(stripped) and stripped[j].isdigit():
                    j += 1
                if j < len(stripped) and stripped[j] == ".":
                    if j + 1 < len(stripped) and stripped[j + 1] == " ":
                        is_list_item = True
        
        # If this is a list item and previous line is not blank and not a list item
        if is_list_item and i > 0:
            prev_line = lines[i - 1].strip()
            prev_is_list = False
            if prev_line:
                prev_stripped = prev_line.lstrip()
                if prev_stripped:
                    # Check if previous line is also a list item
                    if prev_stripped[0] in "-*+" and len(prev_stripped) > 1 and prev_stripped[1] == " ":
                        prev_is_list = True
                    elif prev_stripped[0].isdigit():
                        j = 1
                        while j < len(prev_stripped) and prev_stripped[j].isdigit():
                            j += 1
                        if j < len(prev_stripped) and prev_stripped[j] == ".":
                            if j + 1 < len(prev_stripped) and prev_stripped[j + 1] == " ":
                                prev_is_list = True
            
            # Insert blank line if previous line is not blank and not a list item
            if prev_line and not prev_is_list:
                result.append("")
        
        result.append(line)
    
    return "\n".join(result)


def _pandoc_to_pdf(md_path: Path, out_pdf: Path) -> None:
    # Strip YAML frontmatter before pandoc to avoid YAML parse errors and
    # to prevent frontmatter metadata from appearing in the rendered PDF.
    md_text = md_path.read_text(encoding="utf-8", errors="replace")
    body_text = _strip_yaml_frontmatter(md_text)
    # Normalize list spacing to ensure pandoc recognizes lists properly
    body_text = _normalize_list_spacing(body_text)
    
    # Always use a temp file since we may have modified the content
    tmp_md = out_pdf.with_suffix(".input.md")
    tmp_md.write_text(body_text, encoding="utf-8")
    input_path = tmp_md

    # Create a LaTeX header file with better list formatting
    header_content = r"""
\usepackage{enumitem}
\setlist[itemize]{leftmargin=*,topsep=0.5em,itemsep=0.3em,parsep=0.2em}
\setlist[enumerate]{leftmargin=*,topsep=0.5em,itemsep=0.3em,parsep=0.2em}
\usepackage{geometry}
\geometry{margin=1in}
"""
    header_file = out_pdf.with_suffix(".header.tex")
    header_file.write_text(header_content, encoding="utf-8")

    argv = [
        "pandoc",
        str(input_path),
        "-o",
        str(out_pdf),
        "--pdf-engine=xelatex",
        "--standalone",
        "-H",
        str(header_file),  # Include custom header for better list formatting
        "-V",
        "mainfont=DejaVu Sans",
        "-V",
        "monofont=DejaVu Sans Mono",
        "-V",
        "geometry:margin=1in",  # Set margins
    ]
    _run(argv, check=True, capture=False)
    
    # Clean up temp files
    header_file.unlink(missing_ok=True)
    tmp_md.unlink(missing_ok=True)


def _rm_ls(remote_dir: str) -> tuple[int, str]:
    """
    Returns (exit_code, output). rmapi prints to stdout; capture both.
    """
    cp = _run(["rmapi", "ls", remote_dir], check=False, capture=True)
    return cp.returncode, (cp.stdout or "")


def _rm_get(remote_path: str) -> int:
    """
    Returns exit code for rmapi get. We use this as a more direct existence probe.
    """
    cp = _run(["rmapi", "get", remote_path], check=False, capture=True)
    return cp.returncode


def _remote_file_exists(remote_dir: str, pdf_name: str) -> bool:
    # First try ls (cheap, doesn't download)
    rc, out = _rm_ls(remote_dir)
    if rc == 0:
        # rmapi ls output formats vary; do a conservative contains check.
        # We match both "name.pdf" and "name" styles.
        if pdf_name in out:
            return True
        # Sometimes rmapi prints without extension in some modes; keep strict default.
        return False

    # If ls fails (e.g. directory missing), fall back to a direct get probe.
    # This may download if it exists; we accept that for now.
    remote_path = remote_dir.rstrip("/") + "/" + pdf_name
    return _rm_get(remote_path) == 0


def _upload_pdf(local_pdf: Path, remote_dir: str, *, force: bool) -> None:
    argv = ["rmapi", "put", str(local_pdf), remote_dir]
    if force:
        argv.append("--force")
    _run(argv, check=True, capture=False)


def _targets_from_args(
    md_files: list[str],
    *,
    ticket_dir: Path,
) -> list[UploadTarget]:
    if md_files:
        targets: list[UploadTarget] = []
        for f in md_files:
            p = Path(f).expanduser()
            if not p.is_absolute():
                p = (Path.cwd() / p).resolve()
            targets.append(UploadTarget(md_path=p, pdf_name=p.with_suffix(".pdf").name))
        return targets

    # Defaults: the two documents mentioned in the ticket
    bug = ticket_dir / "reference" / "01-bug-report-doc-relate-fails-on-non-docmgr-markdown-files.md"
    analysis = ticket_dir / "analysis" / "02-code-flow-analysis-frontmatter-validation-failure.md"
    return [
        UploadTarget(md_path=bug, pdf_name=bug.with_suffix(".pdf").name),
        UploadTarget(md_path=analysis, pdf_name=analysis.with_suffix(".pdf").name),
    ]


def _ensure_file_exists(p: Path) -> None:
    if not p.exists():
        raise FileNotFoundError(str(p))
    if not p.is_file():
        raise FileNotFoundError(f"Not a file: {p}")


def main(argv: Optional[list[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="Convert markdown to PDF (pandoc/xelatex + DejaVu) and upload to reMarkable via rmapi.",
    )
    parser.add_argument(
        "md",
        nargs="*",
        help="Markdown files to upload. If omitted, uploads the ticket's bug report + analysis doc.",
    )
    parser.add_argument(
        "--ticket-dir",
        default="",
        help="Ticket directory to use for default documents (when no md args are provided).",
    )
    parser.add_argument(
        "--ticket",
        default="",
        help="Ticket id to locate under --root (best-effort name match; looks for a matching directory with index.md).",
    )
    parser.add_argument(
        "--root",
        default="ttmp",
        help="Docs root directory to search for tickets when using --ticket (default: ttmp; resolved from CWD).",
    )
    parser.add_argument(
        "--date",
        default=None,
        help="Destination date folder under ai/ (YYYY/MM/DD or YYYY-MM-DD). Defaults to the ticket date if inferable, else today's date.",
    )
    parser.add_argument(
        "--force",
        action="store_true",
        help="Overwrite existing PDFs on the device (passed to `rmapi put --force`).",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print what would be done, but don't run pandoc/rmapi.",
    )
    parser.add_argument(
        "--pdf-only",
        action="store_true",
        help="Only generate PDF, don't upload to reMarkable. PDF is saved to current directory or --output-dir if specified.",
    )
    parser.add_argument(
        "--output-dir",
        default=None,
        help="Output directory for PDF when using --pdf-only (default: current directory).",
    )

    args = parser.parse_args(argv)

    script_path = Path(__file__).resolve()
    ticket_dir = script_path.parent.parent  # works when script lives under <ticket>/scripts/

    # When installed in PATH, script_path won't live under a ticket directory.
    # Allow overriding via --ticket-dir or --ticket/--root.
    if args.ticket_dir:
        td = Path(args.ticket_dir).expanduser()
        if not td.is_absolute():
            td = (Path.cwd() / td).resolve()
        ticket_dir = td
    elif args.ticket:
        root = Path(args.root).expanduser()
        if not root.is_absolute():
            root = (Path.cwd() / root).resolve()
        if not root.exists() or not root.is_dir():
            print(f"ERROR: docs root not found: {root}", file=sys.stderr)
            return 2
        needle = args.ticket.lower().strip()
        found: Optional[Path] = None
        for p in root.rglob("*"):
            if not p.is_dir():
                continue
            if needle and needle in p.name.lower() and (p / "index.md").exists():
                found = p
                break
        if found is None:
            print(f"ERROR: ticket not found under {root} matching {args.ticket!r}", file=sys.stderr)
            return 2
        ticket_dir = found

    date_ymd = args.date or _default_date(ticket_dir)
    remote_dir = _normalize_rm_dir(date_ymd)

    targets = _targets_from_args(args.md, ticket_dir=ticket_dir)

    print(f"Ticket dir: {ticket_dir}")
    if not args.pdf_only:
        print(f"Remote dir: {remote_dir}")
    print(f"Force: {bool(args.force)}")
    print(f"Dry-run: {bool(args.dry_run)}")
    print(f"PDF-only: {bool(args.pdf_only)}")

    # Determine output directory for PDF-only mode
    output_dir = Path.cwd()
    if args.output_dir:
        output_dir = Path(args.output_dir).expanduser()
        if not output_dir.is_absolute():
            output_dir = (Path.cwd() / output_dir).resolve()
        output_dir.mkdir(parents=True, exist_ok=True)

    for t in targets:
        try:
            _ensure_file_exists(t.md_path)
        except FileNotFoundError:
            print(f"\nERROR: markdown file not found: {t.md_path}", file=sys.stderr)
            return 2

        print(f"\n=== {t.md_path} ===")
        print(f"PDF name: {t.pdf_name}")

        # Skip reMarkable existence check if PDF-only mode
        if not args.pdf_only and not args.force and not args.dry_run:
            exists = _remote_file_exists(remote_dir, t.pdf_name)
            if exists:
                print(
                    f"SKIP: `{remote_dir}{t.pdf_name}` already exists on reMarkable.\n"
                    f"Re-run with `--force` to overwrite.",
                    file=sys.stderr,
                )
                continue

        # Determine output PDF location
        if args.pdf_only:
            out_pdf = output_dir / t.pdf_name
            tmpdir = None
        else:
            tmpdir = Path(tempfile.mkdtemp(prefix="docmgr-remarkable-"))
            out_pdf = tmpdir / t.pdf_name

        try:
            if args.dry_run:
                print(f"DRY: pandoc {t.md_path} -> {out_pdf} (xelatex, DejaVu fonts)")
                if not args.pdf_only:
                    print(f"DRY: rmapi put {out_pdf} {remote_dir}" + (" --force" if args.force else ""))
                continue

            _pandoc_to_pdf(t.md_path, out_pdf)
            
            if args.pdf_only:
                print(f"OK: generated {out_pdf}")
            else:
                _upload_pdf(out_pdf, remote_dir, force=bool(args.force))
                print(f"OK: uploaded {t.pdf_name} -> {remote_dir}")
        finally:
            # Clean up temp directory only if not PDF-only mode
            if tmpdir is not None:
                shutil.rmtree(tmpdir, ignore_errors=True)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())


