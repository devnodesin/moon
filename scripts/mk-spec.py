"""Build a merged Markdown spec by expanding include directives."""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


INCLUDE_PATTERN = re.compile(r"\{\{\s*include\s+([^}\r\n]+?)\s*\}\}")


def parse_args() -> argparse.Namespace:
    """Parse command-line arguments."""
    parser = argparse.ArgumentParser(
        description="Expand {{ include ... }} directives into a single Markdown file."
    )
    parser.add_argument(
        "-i",
        "--input",
        default="./moon-llms.src.md",
        help="Input Markdown file (default: ./moon-llms.src.md)",
    )
    parser.add_argument(
        "-o",
        "--output",
        default="./moon-llms.md",
        help="Output Markdown file (default: ./moon-llms.md)",
    )
    return parser.parse_args()


def resolve_include_path(raw_path: str, base_dir: Path) -> Path:
    """Resolve include path relative to the input file directory."""
    include_path = Path(raw_path)
    if include_path.is_absolute():
        return include_path
    return base_dir / include_path


def expand_includes(markdown: str, base_dir: Path) -> str:
    """Replace include directives with file contents when files exist."""

    def replacer(match: re.Match[str]) -> str:
        directive = match.group(0)
        raw_include_path = match.group(1).strip()
        include_path = resolve_include_path(raw_include_path, base_dir)

        if include_path.is_file():
            print(f"✅ include: {raw_include_path}")
            return include_path.read_text(encoding="utf-8")

        print(f"❌ missing: {raw_include_path}")
        return directive

    return INCLUDE_PATTERN.sub(replacer, markdown)


def run(input_filename: str, output_filename: str) -> int:
    """Build output Markdown file by expanding include directives."""
    input_path = Path(input_filename)
    output_path = Path(output_filename)

    source_markdown = input_path.read_text(encoding="utf-8")
    rendered_markdown = expand_includes(source_markdown, input_path.parent)

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(rendered_markdown, encoding="utf-8")
    print(f"✅ wrote: {output_path}")
    return 0


def configure_utf8_output() -> None:
    """Ensure console streams can emit emoji on Windows terminals."""
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8")
    if hasattr(sys.stderr, "reconfigure"):
        sys.stderr.reconfigure(encoding="utf-8")


def main() -> int:
    """Program entrypoint."""
    configure_utf8_output()
    args = parse_args()
    try:
        return run(args.input, args.output)
    except OSError as err:
        print(f"❌ error: {err}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
