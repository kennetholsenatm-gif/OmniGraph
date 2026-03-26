# Wiki Source (Companion to GitHub Wiki)

This directory contains the canonical source for OmniGraph's documentation wiki. The markdown files here are meant to be copied into the repository's **GitHub Wiki**, or kept as the authoritative wiki text in the main repository.

## Wiki Structure

```
wiki/
├── Home.md                    # Main entry point
├── Getting-Started.md         # Installation and quick start
├── CLI-Reference.md           # Complete command reference
├── Configuration.md           # Schema and configuration files
├── Web-UI.md                  # Browser-based interface guide
├── Lifecycle.md               # End-to-end deployment workflow
├── Integrations.md            # External tool integrations
├── Architecture.md            # System design and components
├── Infrastructure-IR.md       # Intent Reference model
├── Execution-Matrix.md        # Runners and execution environments
├── Declarative-Reconciliation.md  # Kubernetes-style resource management
├── Architecture-Decisions.md  # ADR summaries and rationale
├── Pipeline-Runs.md           # Run artifact schema and details
├── Inventory-Sources.md       # Triangulated inventory management
├── Troubleshooting.md         # Common issues and solutions
├── _Sidebar.md                # Navigation sidebar for GitHub Wiki
└── README.md                  # This file
```

## Publishing to GitHub Wiki

### Option 1: Manual Copy

1. Enable Wiki on the GitHub repository (Settings → Features)
2. Create a new page in the GitHub Wiki for each file in this directory
3. Copy the content from each `.md` file into the corresponding wiki page
4. Use the filename (without `.md`) as the page title

Example:
- `Home.md` → Wiki page titled "Home"
- `Getting-Started.md` → Wiki page titled "Getting Started"
- `_Sidebar.md` → Wiki page titled "_Sidebar"

### Option 2: Git-based Wiki

1. Clone the wiki git remote:
   ```bash
   git clone https://github.com/kennetholsenatm-gif/OmniGraph.wiki.git
   ```

2. Copy files from this directory into the cloned wiki repository:
   ```bash
   cp wiki/*.md OmniGraph.wiki/
   ```

3. Commit and push:
   ```bash
   cd OmniGraph.wiki
   git add .
   git commit -m "Update wiki content"
   git push
   ```

### Option 3: Automated Sync

You can set up a GitHub Action to automatically sync changes from this directory to the wiki:

```yaml
# .github/workflows/sync-wiki.yml
name: Sync Wiki

on:
  push:
    paths:
      - 'wiki/**'

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Sync wiki
        uses: SwiftDocOrg/github-wiki-publish-action@v1
        with:
          path: wiki
        env:
          GH_PERSONAL_ACCESS_TOKEN: ${{ secrets.GH_PERSONAL_ACCESS_TOKEN }}
```

## Editing Guidelines

### Page Structure

Each wiki page should follow this structure:

```markdown
# Page Title

Brief introduction paragraph.

## Section 1

Content...

## Section 2

Content...

## Related Documentation

- [Link 1](Page-Name)
- [Link 2](https://github.com/...)

## Next Steps

- [Next Page](Next-Page)
```

### Formatting

- Use Markdown headers (`#`, `##`, `###`) for structure
- Use code blocks with language hints for syntax highlighting
- Use tables for structured data
- Use bullet points and numbered lists for clarity
- Use bold and italic for emphasis

### Cross-References

- Link to other wiki pages using `[Page Name](Page-Name)` format
- Link to GitHub docs using `[Description](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/docs/file.md)`
- Link to external resources using full URLs

### Code Examples

Always include language hints for code blocks:

```bash
# Command example
omnigraph validate .omnigraph.schema
```

```yaml
# YAML example
apiVersion: omnigraph/v1
kind: Schema
```

```go
// Go example
func main() {
    // ...
}
```

## Maintenance

### Adding New Pages

1. Create a new `.md` file in this directory
2. Use a descriptive filename (e.g., `New-Feature.md`)
3. Add the page to `_Sidebar.md` in the appropriate section
4. Update `Home.md` if the page should be prominently linked
5. Commit and push changes

### Updating Existing Pages

1. Edit the `.md` file in this directory
2. Ensure cross-references are still valid
3. Update any related pages if necessary
4. Commit and push changes

### Removing Pages

1. Delete the `.md` file from this directory
2. Remove the page from `_Sidebar.md`
3. Remove any links to the page from other files
4. Commit and push changes

## Documentation Sources

This wiki references documentation from multiple locations:

- **wiki/** (this directory) - User-facing documentation
- **docs/** - Technical specifications and ADRs
- **README.md** - Project overview and quick start
- **CONTRIBUTING.md** - Contribution guidelines

## Best Practices

1. **Keep it concise** - Wiki pages should be scannable and focused
2. **Use examples** - Show, don't just tell
3. **Cross-reference** - Link to related pages and external docs
4. **Stay current** - Update pages when features change
5. **Test instructions** - Verify that code examples work
6. **Use consistent formatting** - Follow the style guide

## Style Guide

### Headers

- Use sentence case for headers
- Keep headers descriptive but concise
- Use `##` for main sections, `###` for subsections

### Code

- Use inline code for commands, file names, and short snippets
- Use code blocks for multi-line examples
- Always specify the language for syntax highlighting

### Links

- Use descriptive link text (not "click here")
- Check that links are valid before committing
- Prefer relative links for internal documentation

### Lists

- Use bullet points for unordered items
- Use numbered lists for sequential steps
- Keep list items parallel in structure

## Versioning

This wiki is versioned alongside the main repository. When releasing a new version:

1. Tag the repository with the version number
2. Update version references in wiki pages if necessary
3. Consider creating version-specific documentation branches for major releases

## Questions?

If you have questions about the wiki structure or content, please open an issue or start a discussion on GitHub.