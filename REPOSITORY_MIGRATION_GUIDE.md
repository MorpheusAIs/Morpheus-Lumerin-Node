# Repository Migration Guide: Lumerin-protocol → MorpheusAIs

## Overview

This guide outlines the complete migration process to consolidate all development work from the `Lumerin-protocol/Morpheus-Lumerin-Node` fork back to the primary `MorpheusAIs/Morpheus-Lumerin-Node` repository. This migration will preserve all history, branches, builds, containers, and CI/CD infrastructure.

## Current State Analysis

### Repository Structure
- **Fork**: `Lumerin-protocol/Morpheus-Lumerin-Node`
- **Origin**: `MorpheusAIs/Morpheus-Lumerin-Node`
- **Local branches**: 25+ active branches including `main`, `test`, `dev`, and numerous feature branches
- **Remote branches**: 150+ branches across various features, fixes, and CI/CD improvements
- **Tags**: Version scheme `v4.x.x` with test and release variants

### CI/CD Infrastructure
- **GitHub Actions**: Comprehensive build pipeline (`build.yml`)
- **Container Registry**: `ghcr.io/lumerin-protocol/morpheus-lumerin-node`
- **Build Artifacts**: Multi-platform executables (Linux, macOS, Windows)
- **Desktop UI**: Cross-platform Electron app builds
- **External Integrations**: GitLab deployment pipeline
- **Required Secrets**: `TEST_PRIVATE_KEY`, `GITLAB_TRIGGER_URL`, `GITLAB_TRIGGER_TOKEN`

### External Dependencies
- GitHub Container Registry (GHCR)
- GitLab deployment infrastructure
- Multi-platform build runners (Ubuntu, macOS, Windows)

## Migration Strategy Options

### Option 1: Repository Transfer (Recommended)

**Advantages:**
- Preserves complete history, all branches, tags, and metadata
- Maintains all GitHub-specific features (Issues, PRs, Actions, etc.)
- Cleanest migration with minimal disruption
- Preserves contributor statistics and blame history
- Automatic redirect from old URL to new location

**Process:**
1. **Prerequisites Check**
   - Ensure admin access to both organizations
   - Verify `MorpheusAIs/Morpheus-Lumerin-Node` can be deleted/renamed
   - Backup both repositories

2. **Pre-Migration Steps**
   - Archive or rename existing `MorpheusAIs/Morpheus-Lumerin-Node` if needed
   - Document current state of both repositories
   - Notify all contributors of upcoming migration

3. **Execute Transfer**
   - Use GitHub's repository transfer feature
   - Transfer `Lumerin-protocol/Morpheus-Lumerin-Node` → `MorpheusAIs/Morpheus-Lumerin-Node`

### Option 2: Manual History Consolidation

**Use when repository transfer is not feasible**

**Process:**
1. Add fork as remote and merge all branches
2. Manually recreate tags and releases
3. Migrate CI/CD infrastructure
4. Update all references

## Detailed Migration Steps

### Phase 1: Pre-Migration Preparation

#### 1.1 Repository Backup
```bash
# Backup current MorpheusAIs repository
git clone --mirror https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git morpheus-backup

# Backup Lumerin fork
git clone --mirror https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node.git lumerin-backup
```

#### 1.2 Audit Current State
```bash
# In the Lumerin fork, document all branches and tags
git branch -a > branches-list.txt
git tag -l > tags-list.txt
git log --oneline --all --graph > commit-history.txt
```

#### 1.3 Identify Dependencies
- [ ] GitHub Container Registry images
- [ ] GitLab deployment configurations  
- [ ] External service integrations
- [ ] Webhook configurations
- [ ] Third-party app integrations

### Phase 2: Repository Migration

#### 2.1 Option 1: GitHub Repository Transfer

1. **Navigate to Repository Settings**
   - Go to `Lumerin-protocol/Morpheus-Lumerin-Node`
   - Settings → General → Danger Zone → Transfer ownership

2. **Execute Transfer**
   - New owner: `MorpheusAIs`
   - Repository name: `Morpheus-Lumerin-Node`
   - Confirm transfer

3. **Post-Transfer Verification**
   ```bash
   # Verify all branches transferred
   git clone https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git
   cd Morpheus-Lumerin-Node
   git branch -a
   git tag -l
   ```

#### 2.2 Option 2: Manual History Merge

```bash
# Clone the original MorpheusAIs repo
git clone https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git
cd Morpheus-Lumerin-Node

# Add Lumerin fork as remote
git remote add lumerin-fork https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node.git

# Fetch all branches and tags
git fetch lumerin-fork
git fetch lumerin-fork --tags

# Push all branches from fork to origin
git push origin refs/remotes/lumerin-fork/*:refs/heads/*

# Push all tags
git push origin --tags
```

### Phase 3: CI/CD Infrastructure Migration

#### 3.1 GitHub Actions Setup

1. **Copy Workflow Files**
   ```bash
   # Ensure all workflow files are present
   cp -r .github/workflows/* /path/to/new/repo/.github/workflows/
   cp -r .github/actions/* /path/to/new/repo/.github/actions/
   ```

2. **Update Container Registry References**
   ```yaml
   # In .github/workflows/build.yml, update line 86:
   IMAGE_NAME="ghcr.io/morpheusais/morpheus-lumerin-node"
   ```

3. **Configure Repository Secrets**
   - Navigate to Settings → Secrets and variables → Actions
   - Add required secrets:
     - `TEST_PRIVATE_KEY`
     - `GITLAB_TRIGGER_URL` 
     - `GITLAB_TRIGGER_TOKEN`
     - `GITHUB_TOKEN` (automatically available)

4. **Update Workflow Conditions**
   ```yaml
   # Update repository checks in build.yml (multiple locations):
   github.repository == 'MorpheusAIs/Morpheus-Lumerin-Node'
   ```

#### 3.2 Container Registry Migration

1. **Retag Existing Images**
   ```bash
   # Pull existing images
   docker pull ghcr.io/lumerin-protocol/morpheus-lumerin-node:latest
   
   # Retag for new registry
   docker tag ghcr.io/lumerin-protocol/morpheus-lumerin-node:latest ghcr.io/morpheusais/morpheus-lumerin-node:latest
   
   # Push to new registry
   docker push ghcr.io/morpheusais/morpheus-lumerin-node:latest
   ```

2. **Update Documentation**
   - Update all references in docs to new container registry
   - Update deployment scripts and configurations

### Phase 4: External Integration Updates

#### 4.1 GitLab Integration
- Update GitLab pipeline configurations to reference new repository
- Verify webhook configurations
- Test deployment triggers

#### 4.2 Documentation Updates
- Update all README files with new repository URLs
- Update installation instructions
- Update container registry references

### Phase 5: Developer Transition

#### 5.1 Developer Communication

**Email Template:**
```
Subject: Repository Migration - Action Required

The Morpheus-Lumerin-Node repository has been migrated from Lumerin-protocol to MorpheusAIs organization.

New Repository URL: https://github.com/MorpheusAIs/Morpheus-Lumerin-Node

Action Required:
1. Update your local repository remote URL
2. Fetch latest changes
3. Rebase/merge your feature branches

See migration guide for detailed instructions.
```

#### 5.2 Local Repository Updates

**For Contributors:**
```bash
# Update remote URL
git remote set-url origin https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git

# Verify the change
git remote -v

# Fetch latest changes
git fetch origin

# Update local main branch
git checkout main
git pull origin main

# Rebase feature branches (for each feature branch)
git checkout feature/your-branch
git rebase origin/main
```

#### 5.3 Branch Cleanup
```bash
# Clean up old remote references
git remote prune origin

# Remove old remote if using manual merge approach
git remote remove lumerin-fork
```

### Phase 6: Fork Deprecation

#### 6.1 Archive Original Fork
1. Navigate to `Lumerin-protocol/Morpheus-Lumerin-Node`
2. Settings → General → Archive this repository
3. Add archive notice to README

#### 6.2 Redirect Notice
Update the archived repository README:
```markdown
# ⚠️ REPOSITORY MOVED

This repository has been migrated to: **https://github.com/MorpheusAIs/Morpheus-Lumerin-Node**

All development, issues, and releases now happen in the MorpheusAIs organization.

Please update your bookmarks and local git remotes:
```bash
git remote set-url origin https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git
```
```

## Post-Migration Checklist

### Immediate Verification
- [ ] All branches present in new repository
- [ ] All tags and releases migrated
- [ ] CI/CD pipelines functioning
- [ ] Container builds working
- [ ] External integrations operational
- [ ] Developer access permissions configured

### First Week
- [ ] Monitor CI/CD pipeline runs
- [ ] Verify container registry access
- [ ] Test deployment processes
- [ ] Address any contributor issues
- [ ] Update external documentation/links

### First Month
- [ ] Confirm all contributors have migrated
- [ ] Archive old repository
- [ ] Update any external service integrations
- [ ] Review and optimize new CI/CD setup
- [ ] Document lessons learned

## Rollback Plan

In case of issues during migration:

1. **Repository Transfer Rollback**
   - Contact GitHub support for transfer reversal
   - Restore from backup repositories

2. **CI/CD Rollback**
   - Revert workflow file changes
   - Restore original container registry settings
   - Reconfigure secrets if needed

3. **Communication**
   - Notify all contributors of rollback
   - Provide updated instructions
   - Document issues for future attempts

## Risk Mitigation

### High-Risk Items
1. **Data Loss**: Complete repository backups before starting
2. **CI/CD Disruption**: Test workflows in fork before migration
3. **Container Registry Issues**: Maintain both registries during transition
4. **Developer Confusion**: Clear communication and documentation

### Contingency Plans
- Maintain parallel CI/CD during transition period
- Keep original fork accessible until migration verified
- Have rollback procedures documented and tested

## Timeline Estimate

- **Planning & Preparation**: 2-3 days
- **Migration Execution**: 1 day
- **CI/CD Configuration**: 1-2 days
- **Developer Transition**: 1 week
- **Verification & Cleanup**: 1 week

**Total Estimated Duration**: 2-3 weeks

## Success Metrics

- [ ] 100% of branches and tags migrated
- [ ] All CI/CD pipelines operational
- [ ] All contributors successfully transitioned
- [ ] No loss of build/deployment functionality
- [ ] External integrations working
- [ ] Documentation updated and accurate

---

*This migration guide ensures a comprehensive transition while minimizing disruption to development workflows and maintaining the integrity of the project's history.*
