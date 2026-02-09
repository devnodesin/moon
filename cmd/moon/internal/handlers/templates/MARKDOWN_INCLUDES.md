# Markdown File Inclusion in Templates

This feature allows you to include external Markdown files into the main documentation template (`doc.md.tmpl`). This makes it easier to maintain modular, reusable content sections.

## Overview

The system supports including Markdown files from the `md/` subdirectory using a custom template function. Files are embedded at build time and loaded from the embedded filesystem at runtime.

## Usage

### Basic Inclusion

In your template file (`doc.md.tmpl`), use the `include` function:

```go
{{ include "filename.md" }}
```

### Example

To include a footer file:

```go
{{ include "footer.md" }}
```

The contents of `templates/md/footer.md` will be injected at that position in the rendered output.

## Features

1. **Embedded Files**: All Markdown files in the `md/` directory are embedded into the binary at build time using Go's `embed` package
2. **Runtime Loading**: Files are read from the embedded filesystem when the template is executed
3. **Error Handling**: Missing or unreadable files are handled gracefully:
   - A warning is logged to the console
   - An HTML comment is inserted: `<!-- Error: Failed to include filename.md -->`
   - Template rendering continues normally
4. **Format Preservation**: All Markdown formatting from included files is preserved in the output

## File Organization

```
cmd/moon/internal/handlers/templates/
├── doc.md.tmpl              # Main template file
├── md/                      # Markdown includes directory
│   ├── README.md           # Documentation for includes
│   ├── example.md          # Example include file
│   ├── footer.md           # Footer content
│   └── troubleshooting.md  # Troubleshooting guide
└── MARKDOWN_INCLUDES.md     # This file
```

## Creating New Includes

1. Create a new `.md` file in the `templates/md/` directory:
   ```bash
   echo "# My Custom Section" > templates/md/my-section.md
   ```

2. Add content to your file using standard Markdown syntax

3. Reference it in `doc.md.tmpl`:
   ```go
   {{ include "my-section.md" }}
   ```

4. Rebuild the application to embed the new file

## Naming Conventions

- Use descriptive filenames (e.g., `quickstart.md`, `api-reference.md`)
- Follow snake_case naming convention
- Always use `.md` extension
- Avoid special characters in filenames

## Common Use Cases

### Footer Content
```go
{{ include "footer.md" }}
```

### Modular Sections
```go
{{ include "introduction.md" }}
{{ include "getting-started.md" }}
{{ include "advanced-usage.md" }}
```

### Conditional Inclusion
```go
{{if .JWTEnabled}}
{{ include "jwt-authentication.md" }}
{{end}}
```

### Multiple Files
```go
{{ include "section1.md" }}

## Additional Content

{{ include "section2.md" }}
{{ include "section3.md" }}
```

## Technical Implementation

The inclusion mechanism is implemented in `doc.go`:

1. **Embedding**: The `md/` directory is embedded using:
   ```go
   //go:embed templates/md/*.md
   var mdFiles embed.FS
   ```

2. **Template Function**: A custom template function reads files:
   ```go
   funcMap := template.FuncMap{
       "include": func(filename string) (string, error) {
           content, err := mdFiles.ReadFile("templates/md/" + filename)
           if err != nil {
               log.Printf("WARNING: Failed to read markdown file %s: %v", filename, err)
               return fmt.Sprintf("<!-- Error: Failed to include %s -->", filename), nil
           }
           return string(content), nil
       },
   }
   ```

3. **Template Parsing**: The function is registered when parsing the template:
   ```go
   tmpl, err := template.New("doc").Funcs(funcMap).Parse(docTemplateContent)
   ```

## Error Handling

### Missing File
When a file doesn't exist:
```
WARNING: Failed to read markdown file nonexistent.md: file does not exist
```
Output contains: `<!-- Error: Failed to include nonexistent.md -->`

### Read Error
When a file can't be read:
```
WARNING: Failed to read markdown file corrupted.md: read error
```
Output contains: `<!-- Error: Failed to include corrupted.md -->`

### Template Continues
The template rendering process continues regardless of include errors, ensuring documentation is always available.

## Performance Considerations

- **Zero Runtime Overhead**: Files are embedded at compile time
- **No File I/O**: No disk access at runtime (files read from embedded FS)
- **Caching**: Rendered documentation is cached after first request
- **Efficient**: No performance impact compared to inline content

## Testing

The feature includes comprehensive test coverage:

- `TestDocHandler_MarkdownIncludeFunction`: Verifies include function setup
- `TestDocHandler_IncludeExistingFile`: Tests successful file inclusion
- `TestDocHandler_IncludeFileHandlesErrors`: Validates error handling
- `TestDocHandler_MarkdownIncludesInHTML`: Ensures HTML conversion works

Run tests with:
```bash
go test -v ./cmd/moon/internal/handlers -run TestDocHandler
```

## Best Practices

1. **Keep Includes Focused**: Each file should cover a specific topic
2. **Use Descriptive Names**: Make filenames self-documenting
3. **Document Dependencies**: Note if one include references another
4. **Version Control**: Commit all `.md` files in the `md/` directory
5. **Review Changes**: Test documentation after modifying includes
6. **Avoid Circular References**: Don't create include loops
7. **Maintain Consistency**: Follow the same style across all includes

## Migration Guide

### Before (Inline Content)
```go
## Footer

For support, visit [GitHub](https://github.com/devnodesin/moon).
```

### After (Using Includes)
```go
{{ include "footer.md" }}
```

Create `templates/md/footer.md`:
```markdown
## Footer

For support, visit [GitHub](https://github.com/devnodesin/moon).
```

## Future Enhancements

Potential improvements for future versions:

1. Support for nested includes
2. Parameter passing to includes
3. Conditional inclusion helpers
4. Include caching optimization
5. Hot-reload in development mode
6. Include dependency tracking
7. Markdown preprocessing options

## Support

For questions or issues:
- Open an issue on [GitHub](https://github.com/devnodesin/moon/issues)
- Reference this documentation: `templates/MARKDOWN_INCLUDES.md`
- Check example files in `templates/md/`
