#!/usr/bin/env node
/**
 * Generate llms.txt, llms-full.txt, and per-page *.md from docs.json + MDX.
 * Usage: SITE_URL=https://nodedocs.mor.org node scripts/generate-llms-txt.mjs [docsDir] [outDir]
 *
 * Per-page markdown makes "View as Markdown" / Accept: text/markdown work on the
 * self-hosted S3 site (Mintlify cloud serves these dynamically; mint export does not).
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const docsDir = path.resolve(process.argv[2] ?? path.join(__dirname, ".."));
const outDir = path.resolve(process.argv[3] ?? docsDir);
const siteUrl = (process.env.SITE_URL ?? "https://nodedocs.mor.org").replace(/\/$/, "");

const docsJson = JSON.parse(fs.readFileSync(path.join(docsDir, "docs.json"), "utf8"));

function walkNavItems(items, slugs) {
  for (const item of items ?? []) {
    if (typeof item === "string") {
      slugs.push(item);
    } else if (item && Array.isArray(item.pages)) {
      walkNavItems(item.pages, slugs);
    }
  }
}

function collectSlugs() {
  const slugs = [];
  for (const tab of docsJson.navigation?.tabs ?? []) {
    for (const group of tab.groups ?? []) {
      walkNavItems(group.pages, slugs);
    }
  }
  return [...new Set(slugs)];
}

function parseFrontmatter(filePath) {
  const raw = fs.readFileSync(filePath, "utf8");
  const match = raw.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n([\s\S]*)$/);
  if (!match) return { meta: {}, body: raw };

  const meta = {};
  for (const line of match[1].split("\n")) {
    const kv = line.match(/^(\w+):\s*"?(.+?)"?\s*$/);
    if (kv) meta[kv[1]] = kv[2];
  }
  return { meta, body: match[2] };
}

function slugToUrl(slug) {
  return slug === "index" ? siteUrl : `${siteUrl}/${slug}`;
}

function slugToMdxPath(slug) {
  if (slug === "index") return path.join(docsDir, "index.mdx");
  return path.join(docsDir, `${slug}.mdx`);
}

function writePageMarkdown(slug, title, url, body) {
  const mdPath =
    slug === "index"
      ? path.join(outDir, "index.md")
      : path.join(outDir, `${slug}.md`);
  fs.mkdirSync(path.dirname(mdPath), { recursive: true });
  const md = `# ${title}\n\nSource: ${url}\n\n${body.trim()}\n`;
  fs.writeFileSync(mdPath, md);
}

const entries = [];
const fullSections = [];
let mdCount = 0;

for (const slug of collectSlugs()) {
  const mdxPath = slugToMdxPath(slug);
  if (!fs.existsSync(mdxPath)) continue;

  const { meta, body } = parseFrontmatter(mdxPath);
  const title = meta.title ?? meta.sidebarTitle ?? slug;
  const description = meta.description ?? "";
  const url = slugToUrl(slug);

  entries.push({ title, description, url });
  fullSections.push(`# ${title}\n\nSource: ${url}\n\n${body.trim()}\n`);
  writePageMarkdown(slug, title, url, body);
  mdCount += 1;
}

const siteName = docsJson.name ?? "Morpheus Lumerin Node Docs";
const siteDescription =
  docsJson.description ??
  "Canonical documentation for the Morpheus Lumerin Node.";

const llmsTxt = [
  `# ${siteName}`,
  "",
  `> ${siteDescription}`,
  "",
  "## Pages",
  "",
  ...entries.map((e) =>
    e.description
      ? `- [${e.title}](${e.url}): ${e.description}`
      : `- [${e.title}](${e.url})`
  ),
  "",
].join("\n");

const llmsFullTxt = [
  `# ${siteName} — full text export`,
  "",
  `> ${siteDescription}`,
  "",
  ...fullSections,
].join("\n\n");

fs.mkdirSync(outDir, { recursive: true });
fs.writeFileSync(path.join(outDir, "llms.txt"), llmsTxt);
fs.writeFileSync(path.join(outDir, "llms-full.txt"), llmsFullTxt);

console.log(
  `Wrote llms.txt (${entries.length} pages), llms-full.txt, and ${mdCount} *.md files to ${outDir}`
);
