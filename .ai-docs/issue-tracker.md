# Issue tracker

A change is done when it is on `main` and production has deployed. The issue stays open through `dev` and `test`.

Maintainers follow the full practice in the private repo `MorpheusAIs/MorpheusMaintainerContext`, file `.ai-docs/how-we-work.md`. This page is the local rules for this repository.

## Filing

Use the issue forms:

- **Bug report** for something broken
- **Feature request** for a change or a proposal
- **Docs** for a wrong or missing nodedocs page

Questions go to Discussions. Blank issues are closed.

A new issue is labeled `needs-triage`. During triage, set one area label and the Priority field, set the status on the **Morpheus Work** project, then clear `needs-triage`.

Area labels: `area: proxy-router`, `area: ui-desktop`, `area: cli`, `area: docs`, `area: contracts`, `area: agents`.

Priority is the GitHub field in the issue sidebar: **Urgent** now, **High** next, **Medium** backlog, **Low** someday. The issue form sets Type (Bug, Feature, Task). Effort is optional.

## Pull requests

Feature, docs, and fix pull requests target `dev`. A promote pull request targets `test` or `main` and its title starts with `promote:`. A release pull request targets `main` and its title starts with `Release`.

The body contains a line `Refs #123`, or the checked no-issue line for a chore, revert, or CI-only change.

Promote pull requests are merge commits.

When this work is part of an initiative, create the issue as a sub-issue of that initiative:

```shell
gh issue create --repo MorpheusAIs/Morpheus-Lumerin-Node \
  --title "TITLE" --body "BODY" \
  --parent https://github.com/MorpheusAIs/MorpheusMaintainerContext/issues/N
```

## Status

The public project **Morpheus work** holds issues from this repo. Status for this repo:

Triage → Needs Info → Backlog → Ready → In Progress → On Dev → On Test → Shipped

Shipped means the commit is on `main`. The workflow comments on the linked issue at each hop and closes it on `main`.

## Setup checklist

In the repository:

- [x] Issue forms for bug, feature, and docs
- [x] Pull request template asks for `Refs #` or the no-issue line
- [x] `CONTRIBUTING.md` and `AGENTS.md` state that feature pull requests target `dev`
- [x] Workflow fails a pull request that does not target `dev`, unless it is a promote
- [x] Workflow comments on the linked issue and closes it only on `main`
- [x] Labels created: `needs-triage`, `needs-info`, `needs-issue`, and the `area:` labels
- [x] Discussions enabled
- [x] Public project **Morpheus Work** is https://github.com/orgs/MorpheusAIs/projects/3
- [x] Repository variable `MORPHEUS_WORK_PROJECT` is `3`
- [ ] Open issues added to the project in Triage
- [ ] After this workflow is on `dev`, the **PR target** check is required on `dev`
- [ ] Later, the missing-issue reminder becomes a required check

Project status moves need a token. The built-in Actions token cannot edit an org project.

- [ ] Repository secret `MORPHEUS_WORK_TOKEN`: a fine-grained token with organization Projects read and write, and repository Issues and Pull requests read and write. Without it the workflow still comments and closes, and skips the status move.
