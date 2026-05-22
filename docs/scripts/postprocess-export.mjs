#!/usr/bin/env node
/**
 * Post-process a mint export directory: Pagefind index + UI hook + llms.txt.
 * Usage: SITE_URL=https://nodedocs.mor.org node scripts/postprocess-export.mjs <siteDir>
 */
import { execSync } from "child_process";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const siteDir = path.resolve(process.argv[2] ?? ".");
const docsDir = path.join(__dirname, "..");
const siteUrl = process.env.SITE_URL ?? "https://nodedocs.mor.org";

if (!fs.existsSync(siteDir)) {
  console.error(`Site directory not found: ${siteDir}`);
  process.exit(1);
}

console.log("Running Pagefind index…");
execSync(`npx pagefind --site "${siteDir}"`, { stdio: "inherit", cwd: docsDir });

console.log("Generating llms.txt…");
execSync(
  `node "${path.join(__dirname, "generate-llms-txt.mjs")}" "${docsDir}" "${siteDir}"`,
  { stdio: "inherit", env: { ...process.env, SITE_URL: siteUrl } }
);

const pagefindSnippet = `
<link href="/pagefind/pagefind-ui.css" rel="stylesheet">
<script src="/pagefind/pagefind-ui.js"></script>
<script>
  window.addEventListener("DOMContentLoaded", function () {
    if (typeof PagefindUI === "undefined") return;
    var mount = document.createElement("di" + "v");
    mount.id = "pagefind-ui";
    mount.style.cssText = "position:fixed;bottom:1rem;right:1rem;z-index:9999;max-width:420px;width:100%;";
    document.body.appendChild(mount);
    new PagefindUI({
      element: "#pagefind-ui",
      showSubResults: true,
      resetStyles: false
    });
  });
</script>
`;

function injectPagefind(htmlPath) {
  let html = fs.readFileSync(htmlPath, "utf8");
  if (html.includes("pagefind-ui.js")) return;
  if (html.includes("</body>")) {
    html = html.replace("</body>", `${pagefindSnippet}\n</body>`);
    fs.writeFileSync(htmlPath, html);
  }
}

function walkHtml(dir) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory() && entry.name !== "pagefind" && entry.name !== "_next") {
      walkHtml(full);
    } else if (entry.isFile() && entry.name.endsWith(".html")) {
      injectPagefind(full);
    }
  }
}

console.log("Injecting Pagefind UI…");
walkHtml(siteDir);

console.log("Post-process complete.");
