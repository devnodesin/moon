import os
import sys
import glob


def main():
    if len(sys.argv) != 2:
        print("Usage: python mk-spec.py <directory>")
        sys.exit(1)

    src_dir = sys.argv[1]
    if not os.path.isdir(src_dir):
        print(f"Error: {src_dir} is not a directory.")
        sys.exit(1)

    # Find all Markdown files with a numeric prefix (e.g., 000-*.md, 080-*.md)
    pattern = os.path.join(src_dir, "[0-9][0-9][0-9]-*.md")
    files = sorted(glob.glob(pattern))
    if not files:
        print(f"No files matching '[0-9][0-9][0-9]-*.md' found in {src_dir}")
        sys.exit(1)

    # Prepare output directory and filename (out dir in current working directory)
    out_dir = os.path.join(os.getcwd(), "out")
    os.makedirs(out_dir, exist_ok=True)
    src_dirname = os.path.basename(os.path.normpath(src_dir))
    out_file = os.path.join(out_dir, f"{src_dirname}.md")

    with open(out_file, "w", encoding="utf-8") as outfile:
        for fname in files:
            with open(fname, "r", encoding="utf-8") as infile:
                outfile.write(infile.read())
                outfile.write("\n\n")  # Separate chapters
    print(f"Combined {len(files)} files into {out_file}")

if __name__ == "__main__":
    main()
