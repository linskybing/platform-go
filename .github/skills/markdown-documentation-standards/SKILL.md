---
name: markdown-documentation-standards
description: Markdown documentation standards, file organization, and formatting guidelines for platform-go project documentation
---

# Markdown Documentation Standards

This skill ensures all documentation follows consistent formatting, structure, and best practices. All documentation must be written in English without emoji or special characters.

## When to Use

Apply this skill when:
- Writing README files
- Creating user documentation
- Writing API documentation
- Creating implementation guides
- Documenting configuration
- Writing troubleshooting guides
- Creating architectural documentation

## Quick Rules

### File Organization

```
docs/
├─ README.md (project overview)
├─ QUICK_START.md (5-minute setup)
├─ INSTALL.md (detailed installation)
├─ API.md (API documentation)
├─ DEPLOYMENT.md (deployment guide)
├─ TROUBLESHOOTING.md (common issues)
└─ architecture/
   ├─ overview.md
   └─ components.md
```

### Header Structure

Use consistent hierarchy:
```markdown
# Main Title (H1 - one per file)

## Section (H2)

### Subsection (H3)

#### Detail (H4 - rarely needed)
```

Do NOT skip header levels. Always go from H1 → H2 → H3 in sequence.

### Formatting Rules

**Text Style**:
- Use `code` for file names, commands, package names
- Use **bold** for emphasis on important concepts
- Use _italic_ for technical terms when first introduced
- Use `inline code` for variables, function names, paths

**Code Blocks**:
```markdown
# Specify language
javascript
function example() {
  return true;
}
```

Always include the language identifier for syntax highlighting.

**Lists**:
- Use `-` for unordered lists (consistent)
- Indent with 2 spaces for nested items
- Use numbered lists `1.` `2.` for sequential steps

### Link Format

```markdown
[Display Text](path/to/file.md)
[API Documentation](./API.md)
[External Link](https://example.com)
```

Keep link text descriptive and short.

### Content Structure

**Every documentation file should have**:

1. Title (H1)
2. Brief description (1-2 sentences)
3. Table of contents (if file is long)
4. Main content sections
5. Examples section (when applicable)
6. Related links section (at end)

**Example structure**:
```markdown
# Feature Name

Brief description of what this feature does.

## Overview

Key concepts and terminology.

## Installation

Steps to set up feature.

## Configuration

How to configure the feature.

## Examples

Common use cases with code.

## Troubleshooting

Common issues and solutions.

## Related Documentation

- [Link 1](./file1.md)
- [Link 2](./file2.md)
```

### No Unnecessary Elements

- **No emoji** - Professional documentation only
- **No non-English content** - All documentation in English
- **No ASCII art diagrams** - Use external tools or description
- **No flowcharts** - Text description or external tools only
- **No colored text** - Markdown does not support it properly
- **No HTML markup** - Use markdown syntax only
- **No special characters** - Stick to standard ASCII

Keep documentation clean, professional, and universally readable.

### Code Examples

```bash
# Shell examples show commands and output
$ command here
expected output here
```

```go
// Go code examples must be valid
func Example() error {
    return nil
}
```

Each code block should be:
- Valid and runnable
- Under 20 lines (reference only)
- Properly formatted with indentation

### Table Format

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Row 1    | Data     | Data     |
| Row 2    | Data     | Data     |

Use pipes for clear separation. Keep tables simple (max 4 columns).

### Line Length

Keep lines under 100 characters where possible. Exception: URLs.

### Paragraph Structure

- One idea per paragraph
- Maximum 3 sentences per paragraph
- Use blank lines between sections
- Lead with key information

### File Naming

Use lowercase with hyphens:
- Good: `deployment-guide.md`, `api-reference.md`
- Bad: `DeploymentGuide.md`, `API_Reference.md`, `api reference.md`

### Common Sections

#### Prerequisites
```markdown
## Prerequisites

- Go 1.21 or later
- Docker 20.10 or later
- PostgreSQL 13+
```

#### Installation
```markdown
## Installation

1. Clone repository
2. Install dependencies
3. Configure environment
4. Run setup
```

#### Configuration
```markdown
## Configuration

Key configuration options with defaults.
```

#### Examples
```markdown
## Examples

### Basic Usage
...
### Advanced Usage
...
```

#### Troubleshooting
```markdown
## Troubleshooting

### Problem: Connection timeout
**Solution**: Check network settings...

### Problem: Authentication fails
**Solution**: Verify credentials...
```

### Documentation Checklist

Before publishing documentation:

- [ ] Title clearly states topic
- [ ] Headings are properly hierarchical
- [ ] Code blocks have language identifiers
- [ ] All links work and are relative paths
- [ ] No markdown syntax errors
- [ ] Line length under 100 characters
- [ ] Consistent formatting throughout
- [ ] No unnecessary blank lines
- [ ] Table content is accurate
- [ ] Examples are runnable
- [ ] Spelling and grammar correct
- [ ] No emoji or special characters
- [ ] Related links section complete

### Keep It Simple

Goal: Clear, scannable, maintainable documentation.

- Use short sentences
- Use lists instead of paragraphs
- Use bold for key terms
- Use code formatting for technical terms
- Remove unnecessary words
- One topic per file

---

Last Updated: 2026-02-02
