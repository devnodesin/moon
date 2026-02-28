#!/usr/bin/env python3
"""
sync_api.py - Transform SPEC_API.md references to template includes

Reads SPEC_API.md, replaces markdown reference patterns like:
  See [Text](./SPEC_API/filename.md).
with template include statements:
  {{ include "filename.md" }}

Outputs to: cmd/moon/internal/handlers/templates/doc.md.tmpl
"""

import re
import sys
from pathlib import Path


def transform_spec_api(input_path: str, output_path: str) -> int:
    """
    Transform SPEC_API.md references to template includes.
    
    Args:
        input_path: Path to SPEC_API.md
        output_path: Path to output template file
        
    Returns:
        Number of replacements made
    """
    # Read input file
    try:
        with open(input_path, 'r', encoding='utf-8') as f:
            content = f.read()
    except FileNotFoundError:
        print(f"Error: Input file not found: {input_path}", file=sys.stderr)
        sys.exit(1)
    except IOError as e:
        print(f"Error reading {input_path}: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Pattern matches: See [<any text>](./SPEC_API/<filename>.md).
    # Captures the filename part
    pattern = r'See \[.*?\]\(\./SPEC_API/([^)]+\.md)\)\.'
    
    # Track replacements for logging
    replacements = []
    
    def replace_func(match):
        filename = match.group(1)
        replacements.append((match.group(0), f'{{{{ include "{filename}" }}}}'))
        return f'{{{{ include "{filename}" }}}}'
    
    # Perform replacement
    transformed_content = re.sub(pattern, replace_func, content)
    
    # Ensure output directory exists
    output_file = Path(output_path)
    output_file.parent.mkdir(parents=True, exist_ok=True)
    
    # Write output file
    try:
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(transformed_content)
    except IOError as e:
        print(f"Error writing {output_path}: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Log replacements
    if replacements:
        print(f"Made {len(replacements)} replacement(s):")
        for original, replacement in replacements:
            print(f"  '{original}' -> '{replacement}'")
    else:
        print("No replacements made (no matching patterns found)")
    
    print(f"\nOutput written to: {output_path}")
    return len(replacements)


def main():
    """Main entry point."""
    input_file = "SPEC_API.md"
    output_file = "cmd/moon/internal/handlers/templates/doc.md.tmpl"
    
    print(f"Transforming {input_file}...")
    transform_spec_api(input_file, output_file)


if __name__ == "__main__":
    main()
