const fs = require("fs").promises;
const path = require("path");
const core = require("@actions/core");
const github = require("@actions/github");

async function run() {
  try {
    const releaseId = core.getInput("release-id");
    const artifactPath = core.getInput("path");

    const files = await fs.readdir(artifactPath);

    for (const file of files) {
      const filePath = path.join(artifactPath, file);
      const stats = await fs.stat(filePath);

      if (stats.isDirectory()) {
        console.log(`Skipping directory: ${file}`);
        continue;
      }

      console.log(`Uploading file: ${file}`);
      const octokit = github.getOctokit(process.env.GITHUB_TOKEN);

      await octokit.rest.repos.uploadReleaseAsset({
        owner: github.context.repo.owner,
        repo: github.context.repo.repo,
        release_id: releaseId,
        name: file,
        data: await fs.readFile(filePath),
      });
    }
  } catch (error) {
    core.setFailed(`Error uploading assets: ${error.message}`);
  }
}

run();
