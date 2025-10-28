# Template Embedding

This directory contains the embedded template filesystem for Technocrat.

## Structure

- `fs.go` - Embeds templates using Go's `embed` package
- `templates/` - **Auto-copied** from `../../templates/` (gitignored)

## Why the Copy?

Go's `embed` directive requires files to be in the same directory tree as the package.
Since `embed` doesn't support `../` paths, we copy templates here during build.

The `templates/` subdirectory is in `.gitignore` - the source of truth is `/templates/` at repo root.

## Build Process

Before building, run:

```bash
cp -r ../../templates ./templates
```

Or use the build script which handles this automatically.
