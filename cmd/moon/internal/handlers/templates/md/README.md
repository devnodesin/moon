# Markdown Includes Directory

This directory contains Markdown files that can be included in the main documentation template (`doc.md.tmpl`).

## Usage

To include a Markdown file in the template, use the `include` template function:

```go
{{ include "filename.md" }}
```

For example, to include `example.md`:

```go
{{ include "example.md" }}
```

## Features

- **Dynamic Inclusion**: Files are read from the embedded filesystem at runtime
- **Error Handling**: Missing files are handled gracefully with warning logs and HTML comments
- **Extensibility**: Add new Markdown files to this directory and reference them in the template
- **Formatting Preservation**: All Markdown formatting is preserved in the output

## File Naming

- Files must have a `.md` extension
- Use descriptive names (e.g., `quickstart.md`, `footer.md`, `security-notes.md`)
- Follow snake_case naming convention for consistency

## Error Handling

If an included file is missing or unreadable:
- A warning is logged to the console
- An HTML comment is inserted in the output: `<!-- Error: Failed to include filename.md -->`
- Template rendering continues normally

## Examples

### Creating a New Include

1. Create a new `.md` file in this directory:
   ```bash
   echo "# My Section" > my-section.md
   ```

2. Reference it in `doc.md.tmpl`:
   ```go
   {{ include "my-section.md" }}
   ```

3. The content will be injected when the template is rendered

### Common Use Cases

- **Footer**: `footer.md` - Common footer content
- **Sections**: `section1.md`, `section2.md` - Modular content sections
- **Quickstart**: `quickstart.md` - Getting started guide
- **FAQ**: `faq.md` - Frequently asked questions
- **Security**: `security-notes.md` - Security guidelines
