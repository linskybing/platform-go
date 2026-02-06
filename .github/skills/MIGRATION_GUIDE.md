---
title: Skills Migration Guide
date: 2024-01-15
---

# Skills Migration Guide: 17 → 5 Consolidated Skills

This document outlines the migration from the original 17 scattered skills to the new 5 consolidated skill sets.

## Timeline

- **Before**: 17 separate skill directories (`.github/skills/`)
- **After**: 5 consolidated skill sets (`.github/skills-consolidated/`)
- **Transition Period**: Skills maintained in both locations for backward compatibility
- **Final**: Archive original skills after verification

## Original Skills → New Consolidated Skills

### 1. Code Quality (4 original skills)
| Original | Maps To | Status |
|----------|---------|--------|
| golang-production-standards | code-quality | ✓ Merged |
| testing-best-practices | code-quality | ✓ Merged |
| error-handling-guide | code-quality | ✓ Merged |
| code-validation-standards | code-quality | ✓ Merged |

### 2. Architecture (4 original skills)
| Original | Maps To | Status |
|----------|---------|--------|
| api-design-patterns | architecture | ✓ Merged |
| file-structure-guidelines | code-quality + architecture | ✓ Merged |
| database-best-practices | architecture | ✓ Merged |
| production-readiness-checklist | architecture | ✓ Merged |

### 3. Operations (4 original skills)
| Original | Maps To | Status |
|----------|---------|--------|
| cicd-pipeline-optimization | operations | ✓ Merged |
| kubernetes-integration | operations | ✓ Merged |
| redis-caching | operations | ✓ Merged |
| monitoring-observability | operations | ✓ Merged |

### 4. Security & Compliance (4 original skills)
| Original | Maps To | Status |
|----------|---------|--------|
| security-best-practices | security-compliance | ✓ Merged |
| access-control-best-practices | security-compliance | ✓ Merged |
| unified-authentication-strategy | security-compliance | ✓ Merged |
| automigration-apikey-initialization | security-compliance | ✓ Merged |

### 5. Documentation (1 original skill)
| Original | Maps To | Status |
|----------|---------|--------|
| markdown-documentation-standards | documentation | ✓ Merged |

## Migration Checklist

### For Developers
- [ ] Update local documentation references (old skills → new skills)
- [ ] Update PR templates to reference new consolidated skills
- [ ] Update code comments with new skill paths
- [ ] Run validation scripts from new consolidated location

### For CI/CD
- [ ] Update GitHub Actions workflows to use new skill paths
- [ ] Update pre-commit hooks to reference new scripts
- [ ] Test all validation scripts in new location

### For Documentation
- [ ] Update README.md to reference new skills structure
- [ ] Update CONTRIBUTING.md with new skill guidelines
- [ ] Create link redirects from old to new skill docs

### Before Cleanup
- [ ] Verify all new skill content is complete
- [ ] Test all scripts in consolidated location
- [ ] Update all internal references
- [ ] Create git commit with migration
- [ ] Run full test suite

### Cleanup
- [ ] Archive old `.github/skills/` directory
- [ ] Create final git commit
- [ ] Push to main branch
- [ ] Update deployment documentation

## Using the New Skills

### Example: Adding Auth to an Endpoint

**Old Approach:**
```
1. Check unified-authentication-strategy/SKILL.md
2. Check access-control-best-practices/SKILL.md
3. Check security-best-practices/SKILL.md
(3 different files, scattered knowledge)
```

**New Approach:**
```
1. Check security-compliance/SKILL.md (all 3 consolidated)
(Single source of truth)
```

### Example: Setting Up Testing

**Old Approach:**
```
1. Check testing-best-practices/SKILL.md
2. Check code-validation-standards/SKILL.md
3. Check golang-production-standards/SKILL.md
(Multiple file reviews needed)
```

**New Approach:**
```
1. Check code-quality/SKILL.md (all consolidated)
(Single comprehensive reference)
```

### Example: K8s Deployment

**Old Approach:**
```
1. Check kubernetes-integration/SKILL.md
2. Check cicd-pipeline-optimization/SKILL.md
3. Check monitoring-observability/SKILL.md
4. Check redis-caching/SKILL.md
(Need to jump between files)
```

**New Approach:**
```
1. Check operations/SKILL.md (all consolidated)
(One-stop reference)
```

## File Structure Comparison

### Before (17 scattered skills)
```
.github/skills/
├── access-control-best-practices/
├── api-design-patterns/
├── automigration-apikey-initialization/
├── cicd-pipeline-optimization/
├── code-validation-standards/
├── database-best-practices/
├── error-handling-guide/
├── file-structure-guidelines/
├── golang-production-standards/
├── kubernetes-integration/
├── markdown-documentation-standards/
├── monitoring-observability/
├── production-readiness-checklist/
├── redis-caching/
├── security-best-practices/
├── testing-best-practices/
└── unified-authentication-strategy/
(17 directories with scattered knowledge)
```

### After (5 consolidated skills)
```
.github/skills-consolidated/
├── code-quality/         (5 skills merged)
├── architecture/         (4 skills merged)
├── operations/           (4 skills merged)
├── security-compliance/  (4 skills merged)
└── documentation/        (1 skill)
(5 directories with unified knowledge)
```

## References

### Documentation Links
- Old Skills: `.github/skills/` (deprecated)
- New Skills: `.github/skills-consolidated/` (current)
- Master Index: `.github/skills-consolidated/README.md`

### Related Files
- Project README: [README.md](../../README.md)
- Contribution Guide: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- API Standards: [docs/API_STANDARDS.md](../../docs/API_STANDARDS.md)

## FAQ

### Q: When will the old skills be removed?
**A:** Old skills will remain for 1 month for backward compatibility, then archived in `.github/skills-archive/`.

### Q: Can I still reference old skills?
**A:** Yes, but we recommend updating to new consolidated skills. Update internal links to point to new location.

### Q: Will the old scripts still work?
**A:** Yes, all scripts have been copied to the new consolidated location with the same functionality.

### Q: Where do I find a specific topic now?
**A:** Use the skill mapping table above or the master index in `.github/skills-consolidated/README.md`.

### Q: How do I update my documentation?
**A:** Replace old skill paths (e.g., `golang-production-standards/`) with new paths (e.g., `code-quality/`).

## Support

For questions about the migration:
1. Check the master index: `.github/skills-consolidated/README.md`
2. Review the relevant consolidated skill
3. Check the mapping table above
4. Open an issue with `[skills-migration]` tag

---

**Migration Date**: 2024-01-15
**Status**: ✓ Consolidation Complete
**Next Steps**: Archive old skills directory
