module.exports.main = async function main({ github, context, core }) {
  const owner = context.repo.owner
  const repo = context.repo.repo
  const pr = context.payload.pull_request
  const base = pr.base.ref
  const body = pr.body || ''
  const promote = /^(promote:|release\b)/i.test(pr.title || '') && (base === 'test' || base === 'main')

  if (context.payload.action !== 'closed') {
    await syncNeedsIssueLabel({ github, context, core, owner, repo, pr, base, body, promote })
  }

  if (context.payload.action === 'opened' && !promote) {
    for (const n of refsIn(body)) {
      await note(github, core, owner, repo, n, `Pull request #${pr.number} opened against \`${base}\`.`)
    }
  }

  if (context.payload.action === 'closed' && pr.merged) {
    const refs = new Set(refsIn(body))
    const commits = await github.paginate(github.rest.pulls.listCommits, {
      owner,
      repo,
      pull_number: pr.number,
      per_page: 100,
    })
    for (const commit of commits) {
      for (const n of refsIn(commit.commit.message || '')) refs.add(n)
    }
    const status = base === 'main' ? 'Shipped' : base === 'test' ? 'On Test' : base === 'dev' ? 'On Dev' : null
    for (const n of refs) {
      if (base === 'main') {
        await note(github, core, owner, repo, n, `Shipped on \`main\` in #${pr.number}.`)
        await quietly(core, n, () => github.rest.issues.update({
          owner,
          repo,
          issue_number: Number(n),
          state: 'closed',
          state_reason: 'completed',
        }))
      } else {
        await note(github, core, owner, repo, n, `Landed on \`${base}\` in #${pr.number}. The issue stays open until \`main\` deploys.`)
        await quietly(core, n, async () => {
          const issue = await github.rest.issues.get({ owner, repo, issue_number: Number(n) })
          if (issue.data.state === 'closed') {
            await github.rest.issues.update({
              owner,
              repo,
              issue_number: Number(n),
              state: 'open',
              state_reason: 'reopened',
            })
          }
        })
      }
      if (status) await setProjectStatus({ github, core, owner, repo, issueNumber: n, status })
    }
  }
}

function refsIn(text) {
  const found = []
  const re = /Refs\s+#(\d+)/gi
  let match
  while ((match = re.exec(text || ''))) found.push(match[1])
  return found
}

async function syncNeedsIssueLabel({ github, context, owner, repo, pr, base, body, promote }) {
  if (promote || base !== 'dev') return
  const linked = refsIn(body).length > 0 || /-\s*\[[xX]\]\s*No tracked issue/.test(body)
  if (linked) {
    await quietly(null, pr.number, () => github.rest.issues.removeLabel({
      owner,
      repo,
      issue_number: pr.number,
      name: 'needs-issue',
    }))
    return
  }
  await github.rest.issues.addLabels({
    owner,
    repo,
    issue_number: pr.number,
    labels: ['needs-issue'],
  })
  if (context.payload.action === 'opened') {
    await github.rest.issues.createComment({
      owner,
      repo,
      issue_number: pr.number,
      body: 'Add `Refs #123` in the pull request body, or check **No tracked issue** and give a one-line reason.',
    })
  }
}

async function note(github, core, owner, repo, issueNumber, body) {
  await quietly(core, issueNumber, () => github.rest.issues.createComment({
    owner,
    repo,
    issue_number: Number(issueNumber),
    body,
  }))
}

async function quietly(core, issueNumber, fn) {
  try {
    await fn()
  } catch (err) {
    const message = `Issue #${issueNumber}: ${err.message}`
    if (core) core.warning(message)
  }
}

async function setProjectStatus({ github, core, owner, repo, issueNumber, status }) {
  const projectNumber = process.env.PROJECT_NUMBER
  if (!projectNumber) return
  try {
    const projectData = await github.graphql(
      `query($org: String!, $number: Int!) {
        organization(login: $org) {
          projectV2(number: $number) {
            id
            field(name: "Status") {
              ... on ProjectV2SingleSelectField {
                id
                options { id name }
              }
            }
          }
        }
      }`,
      { org: owner, number: Number(projectNumber) },
    )
    const project = projectData.organization && projectData.organization.projectV2
    if (!project || !project.field) return
    const option = project.field.options.find((item) => item.name === status)
    if (!option) {
      core.warning(`Project has no status "${status}"`)
      return
    }
    const issue = await github.rest.issues.get({ owner, repo, issue_number: Number(issueNumber) })
    const contentId = issue.data.node_id
    const items = await github.graphql(
      `query($id: ID!) {
        node(id: $id) {
          ... on Issue {
            projectItems(first: 20) {
              nodes { id project { id } }
            }
          }
        }
      }`,
      { id: contentId },
    )
    let itemId = items.node.projectItems.nodes.find((node) => node.project.id === project.id)
    itemId = itemId && itemId.id
    if (!itemId) {
      const added = await github.graphql(
        `mutation($project: ID!, $content: ID!) {
          addProjectV2ItemById(input: { projectId: $project, contentId: $content }) {
            item { id }
          }
        }`,
        { project: project.id, content: contentId },
      )
      itemId = added.addProjectV2ItemById.item.id
    }
    await github.graphql(
      `mutation($project: ID!, $item: ID!, $field: ID!, $option: String!) {
        updateProjectV2ItemFieldValue(input: {
          projectId: $project
          itemId: $item
          fieldId: $field
          value: { singleSelectOptionId: $option }
        }) { projectV2Item { id } }
      }`,
      { project: project.id, item: itemId, field: project.field.id, option: option.id },
    )
  } catch (err) {
    core.warning(`Project update skipped: ${err.message}`)
  }
}
