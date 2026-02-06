---
name: markdown-documentation-standards
description: Markdown documentation standards, file organization, and formatting guidelines for platform-go project documentation
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
---


# Markdown Documentation Standards

This skill ensures all documentation follows consistent formatting, structure, and best practices. All documentation must be written in English without emoji or special characters.

## Table of Contents

1. [When to Use](#when-to-use)
2. [Quick Rules](#quick-rules)
   - [File Organization](#file-organization)
   - [Header Structure](#header-structure)
   - [Formatting Rules](#formatting-rules)
   - [Link Format](#link-format)
   - [Content Structure](#content-structure)
3. [Best Practices](#best-practices)
   - [No Unnecessary Elements](#no-unnecessary-elements)
   - [Code Examples](#code-examples)
   - [Table Format](#table-format)
   - [Line Length](#line-length)
   - [Paragraph Structure](#paragraph-structure)
   - [File Naming](#file-naming)
4. [Common Sections](#common-sections)
   - [Prerequisites](#prerequisites)
   - [Installation](#installation)
   - [Configuration](#configuration)
   - [Examples](#examples)
   - [Troubleshooting](#troubleshooting)
5. [Documentation Checklist](#documentation-checklist)
6. [Keep It Simple](#keep-it-simple)


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


## Common Sections

### Prerequisites

```markdown
## Prerequisites

- Go 1.21 or later
- Docker 20.10 or later
- PostgreSQL 13+
```

### Installation

```markdown
## Installation

1. Clone repository
2. Install dependencies
3. Configure environment
4. Run setup
```

### Configuration

```markdown
## Configuration

Key configuration options with defaults.
```

### Examples

```markdown
## Examples

### Basic Usage
...
### Advanced Usage
...
```

### Troubleshooting

```markdown
## Troubleshooting

### Problem: Connection timeout
**Solution**: Check network settings...

### Problem: Authentication fails
**Solution**: Verify credentials...
```


## Keep It Simple

Goal: Clear, scannable, maintainable documentation.

- Use short sentences
- Use lists instead of paragraphs
- Use bold for key terms
- Use code formatting for technical terms
- Remove unnecessary words
- One topic per file
