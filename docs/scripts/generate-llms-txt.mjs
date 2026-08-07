#!/usr/bin/env node
/**
 * Generate llms.txt, llms-full.txt, and per-page *.md from docs.json + MDX.
 * Usage: SITE_URL=https://nodedocs.mor.org node scripts/generate-llms-txt.mjs [docsDir] [outDir]
 *
 * Self-hosted S3 export cannot use Mintlify cloud's dynamic markdown / Accept
 * negotiation. This script emits clean per-page *.md (MDX components stripped
 * to plain markdown) plus llms indexes so CloudFront can serve non-browser
 * clients readable text at the same page URLs.
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const docsDir = path.resolve(process.argv[2] ?? path.join(__dirname, ".."));
const outDir = path.resolve(process.argv[3] ?? docsDir);
const siteUrl = (process.env.SITE_URL ?? "https://nodedocs.mor.org").replace(/\/$/, "");

const docsJson = JSON.parse(fs.readFileSync(path.join(docsDir, "docs.json"), "utf8"));

const AGENT_INSTRUCTIONS = [
  "## Agent Instructions",
  "",
  `- Non-browser fetches of page URLs on this site return clean Markdown (not the JS UI). Prefer \`${siteUrl}/llms-full.txt\` for the full corpus, or \`${siteUrl}/llms.txt\` for the index.`,
  `- Per-page Markdown is also at \`<page-url>.md\` (homepage: \`${siteUrl}/index.md\`).`,
  `- Docs search MCP: \`${siteUrl}/mcp\` (discovery: \`${siteUrl}/.well-known/mcp\`).`,
  "- Never invent contract addresses, chain IDs, token addresses, or live bid/model counts. Cite Networks and tokens; link active.mor.org for live data.",
  "- Never claim Morpheus runs inference — independent providers do. Opening a session escrows MOR; it does not spend it.",
  "",
].join("\n");

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
  const slugs = ["index"];
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

function slugToMdUrl(slug) {
  return slug === "index" ? `${siteUrl}/index.md` : `${siteUrl}/${slug}.md`;
}

function slugToMdxPath(slug) {
  if (slug === "index") return path.join(docsDir, "index.mdx");
  return path.join(docsDir, `${slug}.mdx`);
}

/**
 * Convert Mintlify MDX components to plain markdown for agent consumption.
 * Leaves fenced code blocks (including mermaid) untouched.
 */
function mdxToCleanMarkdown(body) {
  const fences = [];
  let text = body.replace(/```[\s\S]*?```/g, (block) => {
    const token = `@@FENCE_${fences.length}@@`;
    fences.push(block);
    return token;
  });

  // Accordion / AccordionGroup
  text = text.replace(/<\/?AccordionGroup>/g, "");
  text = text.replace(
    /<Accordion\s+title="([^"]*)"\s*>\s*([\s\S]*?)\s*<\/Accordion>/g,
    (_, title, inner) => `\n### ${title.trim()}\n\n${inner.trim()}\n`
  );

  // Steps
  text = text.replace(/<\/?Steps>/g, "");
  text = text.replace(/<Step\s+title="([^"]*)"\s*>\s*([\s\S]*?)\s*<\/Step>/g, (_, title, inner) => {
    return `\n### ${title.trim()}\n\n${inner.trim()}\n`;
  });
  text = text.replace(/<\/?Step>/g, "");

  // Cards
  text = text.replace(/<CardGroup[^>]*>/g, "");
  text = text.replace(/<\/CardGroup>/g, "");
  text = text.replace(
    /<Card\s+([^>]*)>\s*([\s\S]*?)\s*<\/Card>/g,
    (_, attrs, inner) => {
      const title = (attrs.match(/title="([^"]*)"/) || [])[1] || "";
      const href = (attrs.match(/href="([^"]*)"/) || [])[1] || "";
      const bodyInner = inner.trim();
      if (title && href) return `\n- **[${title}](${href})** — ${bodyInner}\n`;
      if (title) return `\n- **${title}** — ${bodyInner}\n`;
      return `\n${bodyInner}\n`;
    }
  );

  // Callouts → blockquotes
  for (const tag of ["Note", "Warning", "Tip", "Info", "Check", "Danger"]) {
    const re = new RegExp(`<${tag}[^>]*>\\s*([\\s\\S]*?)\\s*<\\/${tag}>`, "g");
    text = text.replace(re, (_, inner) => {
      const quoted = inner
        .trim()
        .split("\n")
        .map((line) => `> ${line}`)
        .join("\n");
      return `\n${quoted}\n`;
    });
  }

  // Tabs (keep content, drop chrome)
  text = text.replace(/<\/?Tabs>/g, "");
  text = text.replace(/<Tab\s+title="([^"]*)"\s*>\s*([\s\S]*?)\s*<\/Tab>/g, (_, title, inner) => {
    return `\n### ${title.trim()}\n\n${inner.trim()}\n`;
  });

  // Frame / ResponseField / ParamField wrappers — keep inner text
  text = text.replace(/<\/?(?:Frame|ResponseField|ParamField)[^>]*>/g, "");

  // Self-closing or leftover JSX-ish tags that shouldn't appear in agent md
  text = text.replace(/<\/?[A-Z][A-Za-z0-9]*\b[^>]*\/?>/g, "");

  // Collapse excess blank lines
  text = text.replace(/\n{3,}/g, "\n\n");

  for (let i = 0; i < fences.length; i++) {
    text = text.replace(`@@FENCE_${i}@@`, fences[i]);
  }
  return text.trim();
}

function writePageMarkdown(slug, title, url, body) {
  const mdPath =
    slug === "index"
      ? path.join(outDir, "index.md")
      : path.join(outDir, `${slug}.md`);
  fs.mkdirSync(path.dirname(mdPath), { recursive: true });
  const md = [
    `# ${title}`,
    "",
    `Source: ${url}`,
    "",
    AGENT_INSTRUCTIONS.trimEnd(),
    "",
    body.trim(),
    "",
  ].join("\n");
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
  const mdUrl = slugToMdUrl(slug);
  const clean = mdxToCleanMarkdown(body);

  entries.push({ title, description, url: mdUrl, htmlUrl: url });
  fullSections.push(`# ${title}\n\nSource: ${url}\n\n${clean}\n`);
  writePageMarkdown(slug, title, url, clean);
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
  AGENT_INSTRUCTIONS.trimEnd(),
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
  AGENT_INSTRUCTIONS.trimEnd(),
  "",
  ...fullSections,
].join("\n\n");

fs.mkdirSync(outDir, { recursive: true });
fs.writeFileSync(path.join(outDir, "llms.txt"), llmsTxt);
fs.writeFileSync(path.join(outDir, "llms-full.txt"), llmsFullTxt);

console.log(
  `Wrote llms.txt (${entries.length} pages), llms-full.txt, and ${mdCount} *.md files to ${outDir}`
);
