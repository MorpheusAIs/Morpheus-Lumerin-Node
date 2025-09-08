# Repository Migration Guide: Lumerin-protocol → MorpheusAIs

## ⚠️ CRITICAL MIGRATION NOTICE

This guide contains **IRREVERSIBLE ACTIONS** that will permanently transfer your repository. Read all warnings and complete all backups before proceeding.

## Overview

This guide provides a step-by-step checklist to transfer the complete `Lumerin-protocol/Morpheus-Lumerin-Node` repository to `MorpheusAIs/Morpheus-Lumerin-Node` using GitHub's repository transfer feature. This preserves all history, branches, tags, issues, PRs, and metadata.

## Current State Analysis

### Repository Structure
- **Source**: `Lumerin-protocol/Morpheus-Lumerin-Node` (150+ branches, complete development history)
- **Target**: `MorpheusAIs/Morpheus-Lumerin-Node` (will be replaced)
- **Branches**: 25+ local branches, 150+ remote branches across features/fixes
- **Tags**: Version scheme `v4.x.x` with test and release variants
- **Build Status**: ✅ Production code builds successfully, tests consolidated

### CI/CD Infrastructure
- **GitHub Actions**: Comprehensive build pipeline (`build.yml`)
- **Container Registry**: `ghcr.io/lumerin-protocol/morpheus-lumerin-node`
- **Build Artifacts**: Multi-platform executables (Linux, macOS, Windows)
- **Desktop UI**: Cross-platform Electron app builds
- **External Integrations**: GitLab deployment pipeline
- **Required Secrets**: `TEST_PRIVATE_KEY`, `GITLAB_TRIGGER_URL`, `GITLAB_TRIGGER_TOKEN`

## 🚨 ONE-WAY DOORS & IRREVERSIBLE ACTIONS

### ⛔ POINT OF NO RETURN
Once you execute the repository transfer:
- **Cannot be undone** without GitHub support intervention
- **Original MorpheusAIs repository will be permanently deleted**
- **All URLs immediately redirect** to the transferred repository
- **All existing clones/forks** will need remote URL updates

### 🔄 REVERSIBLE ACTIONS (Safe to test)
- Repository backups
- CI/CD configuration updates
- Container registry changes
- Documentation updates

## 📋 STEP-BY-STEP MIGRATION CHECKLIST

### 🔒 PHASE 1: BACKUP & SAFETY (REQUIRED - DO NOT SKIP)

#### ✅ Step 1.1: Create Complete Repository Backups
**⚠️ CRITICAL: Complete this BEFORE any other steps**

```bash
# Create backup directory
mkdir ~/morpheus-migration-backups
cd ~/morpheus-migration-backups

# Backup 1: Complete MorpheusAIs repository (will be deleted)
git clone --mirror https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git morpheus-original-backup
echo "✅ MorpheusAIs backup created: $(date)" >> backup-log.txt

# Backup 2: Complete Lumerin fork (source of truth)
git clone --mirror https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node.git lumerin-source-backup
echo "✅ Lumerin fork backup created: $(date)" >> backup-log.txt

# Backup 3: Export all issues, PRs, and metadata (if needed)
# Use GitHub CLI or API to export issues/PRs if you need them from MorpheusAIs repo
```

**Verification Checklist:**
- [ ] `morpheus-original-backup` directory exists and contains `.git` folder
- [ ] `lumerin-source-backup` directory exists and contains `.git` folder  
- [ ] Both backups are at least 100MB+ (contains full history)
- [ ] `backup-log.txt` shows successful completion timestamps

#### ✅ Step 1.2: Document Current State
```bash
cd lumerin-source-backup
# Document all branches
git branch -a > ../pre-migration-branches.txt
# Document all tags  
git tag -l > ../pre-migration-tags.txt
# Document recent commits
git log --oneline --all --graph -50 > ../pre-migration-commits.txt
# Document repository size
du -sh . > ../pre-migration-size.txt
```

**Verification Checklist:**
- [ ] `pre-migration-branches.txt` shows 150+ branches
- [ ] `pre-migration-tags.txt` shows version tags (v4.x.x format)
- [ ] `pre-migration-commits.txt` shows recent development activity
- [ ] All documentation files created successfully

#### ✅ Step 1.3: Verify Admin Access & Permissions
**⚠️ CRITICAL: Confirm you have the necessary permissions**

**Pre-flight Checklist:**
- [ ] Admin access to `Lumerin-protocol` organization
- [ ] Admin access to `MorpheusAIs` organization  
- [ ] Ability to delete/transfer repositories in both orgs
- [ ] Access to all required secrets (`TEST_PRIVATE_KEY`, `GITLAB_TRIGGER_URL`, `GITLAB_TRIGGER_TOKEN`)
- [ ] GitLab deployment access (if needed for updates)

#### ✅ Step 1.4: Prepare Rollback Plan
**Create recovery instructions in case something goes wrong:**

```bash
# Create rollback script
cat > ~/morpheus-migration-backups/ROLLBACK-INSTRUCTIONS.md << 'EOF'
# EMERGENCY ROLLBACK PROCEDURE

If migration fails and you need to restore:

1. Restore MorpheusAIs repository from backup:
   cd ~/morpheus-migration-backups/morpheus-original-backup
   git push --mirror https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git

2. Contact GitHub support if repository transfer cannot be undone

3. Restore CI/CD secrets and configurations

4. Update all team members about rollback

BACKUP CREATED: $(date)
EOF
```

**Verification Checklist:**
- [ ] Rollback instructions created and reviewed
- [ ] Team contact list ready for emergency notifications
- [ ] GitHub support contact information available

### ⚠️ PHASE 2: FINAL PREPARATION (LAST CHANCE TO ABORT)

#### ✅ Step 2.1: Final Team Notification
**Send notification to all contributors 24 hours before migration:**

```
Subject: FINAL NOTICE - Repository Migration Tomorrow

The Morpheus-Lumerin-Node repository will be migrated from 
Lumerin-protocol to MorpheusAIs organization tomorrow.

WHAT WILL CHANGE:
- Repository URL remains the same (automatic redirect)
- All history, branches, and tags preserved
- CI/CD may be temporarily disrupted

WHAT YOU NEED TO DO:
- Nothing immediately (automatic redirects work)
- After migration: git remote set-url origin https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git

Migration window: [INSERT TIME]
Estimated downtime: 30 minutes
```

**Verification Checklist:**
- [ ] All active contributors notified
- [ ] Migration time window scheduled and communicated
- [ ] Emergency contact information shared

#### ✅ Step 2.2: Pre-Migration System Check
```bash
# Verify current state hasn't changed
cd /path/to/your/local/repo
git fetch --all
git status
git log --oneline -5

# Confirm latest changes are pushed
git push origin dev
git push origin test  
git push origin main
```

**Verification Checklist:**
- [ ] All local changes pushed to remote
- [ ] No uncommitted changes in working directory
- [ ] Latest CI/CD build completed successfully
- [ ] All critical branches are up-to-date

### 🔥 PHASE 3: REPOSITORY TRANSFER EXECUTION (POINT OF NO RETURN)

#### ⚠️ FINAL WARNING BEFORE EXECUTION
**Once you click "Transfer repository", this action CANNOT be undone without GitHub support intervention.**

**Pre-Execution Checklist:**
- [ ] All backups completed and verified
- [ ] Team notified and ready
- [ ] Admin access confirmed to both organizations
- [ ] Rollback plan prepared and reviewed
- [ ] **You are 100% ready to proceed**

#### ✅ Step 3.1: Execute Repository Transfer
**🚨 IRREVERSIBLE ACTION - READ CAREFULLY**

1. **Navigate to Source Repository**
   - Go to https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node
   - Click "Settings" tab
   - Scroll to "Danger Zone" section

2. **Initiate Transfer**
   - Click "Transfer ownership"
   - **Target owner**: `MorpheusAIs` 
   - **Repository name**: `Morpheus-Lumerin-Node` (EXACT match)
   - **Confirmation**: Type the repository name exactly
   - ⚠️ **PAUSE**: This will permanently delete the existing MorpheusAIs repository
   - Click "I understand, transfer this repository"

3. **Monitor Transfer Progress**
   ```bash
   # Watch for completion (usually 1-5 minutes)
   # Check new location becomes available:
   curl -I https://github.com/MorpheusAIs/Morpheus-Lumerin-Node
   ```

**Transfer Completion Checklist:**
- [ ] Transfer completed without errors
- [ ] New repository URL accessible: https://github.com/MorpheusAIs/Morpheus-Lumerin-Node
- [ ] Old URL redirects automatically: https://github.com/Lumerin-protocol/Morpheus-Lumerin-Node
- [ ] All branches visible in new location
- [ ] All tags preserved
- [ ] Issues and PRs transferred (if any)

#### ✅ Step 3.2: Immediate Post-Transfer Verification
**Verify transfer was successful:**

```bash
# Clone from new location to verify
git clone https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git morpheus-verification
cd morpheus-verification

# Verify all branches transferred
git branch -r | wc -l  # Should show 150+ branches
git tag | wc -l        # Should show version tags

# Verify recent commits
git log --oneline -10

# Verify specific important branches
git checkout dev && git log --oneline -3
git checkout test && git log --oneline -3  
git checkout main && git log --oneline -3

# Document verification results
echo "Transfer verified at: $(date)" >> ../transfer-verification.log
echo "Branches: $(git branch -r | wc -l)" >> ../transfer-verification.log
echo "Tags: $(git tag | wc -l)" >> ../transfer-verification.log
```

**Post-Transfer Verification Checklist:**
- [ ] Repository accessible at new URL
- [ ] All 150+ branches transferred
- [ ] All version tags present (v4.x.x format)
- [ ] Recent development history intact
- [ ] `dev`, `test`, and `main` branches have expected commits
- [ ] File structure and content unchanged

### 🔧 PHASE 4: CI/CD INFRASTRUCTURE MIGRATION

#### ✅ Step 4.1: Update GitHub Actions Configuration
**Update container registry and repository references:**

```bash
cd morpheus-verification  # Use the verification clone

# Update container registry reference
sed -i 's/ghcr.io\/lumerin-protocol\/morpheus-lumerin-node/ghcr.io\/morpheusais\/morpheus-lumerin-node/g' .github/workflows/build.yml

# Update repository condition checks  
sed -i "s/github.repository != 'MorpheusAIs\/Morpheus-Lumerin-Node'/github.repository == 'MorpheusAIs\/Morpheus-Lumerin-Node'/g" .github/workflows/build.yml

# Commit the changes
git add .github/workflows/build.yml
git commit -m "Update CI/CD configuration for MorpheusAIs organization

- Update container registry to ghcr.io/morpheusais/morpheus-lumerin-node  
- Update repository checks for new organization
- Prepare for first CI/CD run in new location"

git push origin dev
```

#### ✅ Step 4.2: Configure Repository Secrets
**Add required secrets to new repository:**

1. **Navigate to Repository Settings**
   - Go to https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/settings/secrets/actions

2. **Add Required Secrets** (one by one):
   - `TEST_PRIVATE_KEY`: [Your test private key]
   - `GITLAB_TRIGGER_URL`: [Your GitLab trigger URL]  
   - `GITLAB_TRIGGER_TOKEN`: [Your GitLab trigger token]

3. **Verify Secrets Added:**
   - [ ] `TEST_PRIVATE_KEY` added
   - [ ] `GITLAB_TRIGGER_URL` added
   - [ ] `GITLAB_TRIGGER_TOKEN` added
   - [ ] `GITHUB_TOKEN` available (automatic)

#### ✅ Step 4.3: Test Initial CI/CD Run
**Trigger first build to verify everything works:**

```bash
# Make a small change to trigger CI/CD
echo "# Repository successfully migrated to MorpheusAIs organization" >> README.md
git add README.md
git commit -m "Test CI/CD pipeline after migration"
git push origin dev
```

**Monitor the build:**
- Go to https://github.com/MorpheusAIs/Morpheus-Lumerin-Node/actions
- Verify build starts and completes successfully
- Check container registry: https://github.com/orgs/MorpheusAIs/packages

**CI/CD Verification Checklist:**
- [ ] GitHub Actions workflow triggered successfully
- [ ] Build completes without errors
- [ ] Container images pushed to `ghcr.io/morpheusais/morpheus-lumerin-node`
- [ ] All build artifacts generated
- [ ] GitLab deployment triggered (if applicable)

### 👥 PHASE 5: TEAM TRANSITION

#### ✅ Step 5.1: Update Team Repositories
**Send updated instructions to all contributors:**

```
Subject: ✅ MIGRATION COMPLETE - Update Your Local Repository

The repository migration is complete! 

NEW REPOSITORY LOCATION: https://github.com/MorpheusAIs/Morpheus-Lumerin-Node

ACTION REQUIRED - Update your local repository:

1. Update remote URL:
   git remote set-url origin https://github.com/MorpheusAIs/Morpheus-Lumerin-Node.git

2. Verify and fetch:
   git remote -v
   git fetch origin
   git pull origin main

3. Update your feature branches:
   git checkout your-feature-branch
   git rebase origin/main

OLD URLs STILL WORK (automatic redirect), but please update for best performance.

Questions? Reply to this email.
```

#### ✅ Step 5.2: Update Documentation & References
**Update all documentation to reflect new location:**

```bash
# Update README and documentation files
find . -name "*.md" -type f -exec grep -l "Lumerin-protocol/Morpheus-Lumerin-Node" {} \;

# Update container registry references in documentation  
find . -name "*.md" -type f -exec sed -i 's/ghcr.io\/lumerin-protocol\/morpheus-lumerin-node/ghcr.io\/morpheusais\/morpheus-lumerin-node/g' {} \;

# Commit documentation updates
git add .
git commit -m "Update documentation for new repository location"
git push origin dev
```

### 🏁 PHASE 6: FINALIZATION & CLEANUP

#### ✅ Step 6.1: Archive Source Repository (Optional)
**If you want to archive the original Lumerin-protocol organization reference:**

1. **The original repository is now gone** (transferred to MorpheusAIs)
2. **Automatic redirects are in place** from old URLs to new location
3. **No further action needed** - GitHub handles the redirection

#### ✅ Step 6.2: Final Verification & Documentation
**Complete final checks and document the migration:**

```bash
# Create migration completion report
cat > ~/morpheus-migration-backups/MIGRATION-COMPLETE-REPORT.md << EOF
# MIGRATION COMPLETION REPORT

Migration Date: $(date)
Source: Lumerin-protocol/Morpheus-Lumerin-Node  
Target: MorpheusAIs/Morpheus-Lumerin-Node

## Transfer Results
- Repository transferred successfully: ✅
- All branches preserved: ✅ ($(git branch -r | wc -l) branches)
- All tags preserved: ✅ ($(git tag | wc -l) tags)
- CI/CD pipeline working: ✅
- Container registry updated: ✅
- Team notified: ✅

## New Repository Details
URL: https://github.com/MorpheusAIs/Morpheus-Lumerin-Node
Container Registry: ghcr.io/morpheusais/morpheus-lumerin-node
Status: Active and operational

Migration completed successfully!
EOF
```

**Final Verification Checklist:**
- [ ] Repository transfer completed successfully
- [ ] All branches and tags preserved  
- [ ] CI/CD pipeline operational
- [ ] Container registry updated and working
- [ ] Team members notified and updated
- [ ] Documentation updated
- [ ] Migration report created
- [ ] Backups safely stored

## 📊 SUCCESS METRICS & MONITORING

### ✅ Migration Success Criteria
**All items must be checked for successful migration:**

- [ ] **Repository Transfer**: 100% of branches and tags migrated
- [ ] **CI/CD Operational**: All pipelines running without errors
- [ ] **Container Registry**: New registry receiving builds successfully  
- [ ] **Team Transition**: All contributors updated and operational
- [ ] **External Integrations**: GitLab and other services working
- [ ] **Documentation**: All references updated to new location
- [ ] **Build Functionality**: No loss of deployment capabilities
- [ ] **History Preservation**: Complete git history and blame intact

### 📈 Post-Migration Monitoring (First 30 Days)

#### Week 1: Critical Monitoring
- [ ] Daily CI/CD pipeline health checks
- [ ] Monitor container registry usage and access
- [ ] Track team member transition completion
- [ ] Address any immediate issues or questions
- [ ] Verify external service integrations

#### Week 2-4: Optimization & Cleanup  
- [ ] Optimize CI/CD performance in new environment
- [ ] Update any remaining external documentation
- [ ] Collect feedback from contributors
- [ ] Document lessons learned
- [ ] Archive migration artifacts and backups

## 🆘 EMERGENCY PROCEDURES

### If Migration Fails During Transfer
1. **DO NOT PANIC** - GitHub transfers are robust
2. **Contact GitHub Support** immediately if transfer hangs or fails
3. **Reference your backup locations** in support ticket
4. **Do not attempt to recreate repositories** until GitHub responds
5. **Notify team** of delay and provide status updates

### If CI/CD Fails After Migration
1. **Check secrets configuration** first (most common issue)
2. **Verify repository conditions** in workflow files
3. **Test with manual workflow dispatch** to isolate issues
4. **Roll back to previous container registry** temporarily if needed
5. **Use backup repositories** to compare configurations

### Emergency Contacts
- **GitHub Support**: https://support.github.com
- **Team Lead**: [Insert contact information]
- **DevOps Lead**: [Insert contact information]
- **Backup Location**: `~/morpheus-migration-backups/`

## 🎯 ESTIMATED TIMELINE

### Repository Transfer Method (Recommended)
- **Preparation & Backups**: 2-4 hours
- **Team Notification**: 24 hours advance notice
- **Transfer Execution**: 5-30 minutes
- **CI/CD Configuration**: 1-2 hours  
- **Verification & Testing**: 2-4 hours
- **Team Communication**: 1 hour

**Total Active Time**: 1 day
**Total Calendar Time**: 2-3 days (including notifications)

## 🏆 FINAL SUCCESS CONFIRMATION

### Migration Complete When ALL Items Checked:

#### Core Migration ✅
- [ ] Repository successfully transferred to MorpheusAIs organization
- [ ] All 150+ branches preserved and accessible
- [ ] All version tags (v4.x.x) present and correct
- [ ] Complete git history and contributor attribution intact
- [ ] Automatic redirects working from old URLs

#### Infrastructure ✅  
- [ ] GitHub Actions workflows running successfully
- [ ] Container images building and pushing to new registry
- [ ] All required secrets configured and working
- [ ] GitLab integration operational (if applicable)
- [ ] External services updated and functional

#### Team & Documentation ✅
- [ ] All contributors notified and repositories updated
- [ ] Documentation updated with new URLs and references
- [ ] README and installation guides reflect new location
- [ ] Migration completion report created and archived
- [ ] Lessons learned documented for future reference

---

## 🎉 CONGRATULATIONS!

**When all checklists are complete, your repository migration is successful!**

Your development team can now continue building amazing software in the MorpheusAIs organization with:
- ✅ Complete development history preserved
- ✅ All contributors and permissions maintained  
- ✅ Fully operational CI/CD pipeline
- ✅ Seamless developer experience
- ✅ Professional, clean migration executed

**The Morpheus-Lumerin-Node project is now home in the MorpheusAIs organization! 🚀**

---

*Migration Guide Version: 2.0*  
*Last Updated: Migration Preparation Phase*  
*Status: Ready for Execution*
