#!/usr/bin/env node
/**
 * Post-process a mint export directory: Pagefind index + navbar search + llms.txt.
 *
 * Mintlify's built-in search (docs.json "search") targets Mintlify Cloud and prompts
 * for CLI login on self-hosted S3/CloudFront exports. We use Pagefind (static index)
 * in the top navbar instead.
 *
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
<link href="/pagefind/pagefind-component-ui.css" rel="stylesheet">
<style>
  /* Mintlify export still renders cloud-search buttons; we replace them with Pagefind. */
  #search-bar-entry,
  #search-bar-entry-mobile {
    display: none !important;
  }
  #nodedocs-pagefind-host {
    width: 100%;
    max-width: 36rem;
    min-width: 12rem;
    flex: 1 1 auto;
  }
  #nodedocs-pagefind-host pagefind-searchbox {
    display: block;
    width: 100%;
  }
</style>
<script type="module">
  import "/pagefind/pagefind-component-ui.js";

  /** Mintlify client-nav re-renders the navbar; re-mount when the slot is replaced. */
  function findDesktopSlot() {
    const entry = document.getElementById("search-bar-entry");
    if (entry?.parentElement) return entry.parentElement;
    return (
      document.querySelector("#navbar .justify-center") ??
      document.querySelector("header .justify-center")
    );
  }

  function ensurePagefindModal() {
    if (!document.querySelector("pagefind-modal")) {
      document.body.appendChild(document.createElement("pagefind-modal"));
    }
  }

  function ensureDesktopSearchbox() {
    const desktopSlot = findDesktopSlot();
    if (!desktopSlot) return;

    let host = desktopSlot.querySelector("#nodedocs-pagefind-host");
    if (!host) {
      host = document.createElement("div");
      host.id = "nodedocs-pagefind-host";
      desktopSlot.appendChild(host);
    }

    if (!host.querySelector("pagefind-searchbox")) {
      const searchbox = document.createElement("pagefind-searchbox");
      searchbox.setAttribute("placeholder", "Search documentation…");
      host.replaceChildren(searchbox);
    }
  }

  function ensureMobileTrigger() {
    const mobileBtn = document.getElementById("search-bar-entry-mobile");
    if (!mobileBtn?.parentElement) return;

    const slot = mobileBtn.parentElement;
    if (!slot.querySelector("pagefind-modal-trigger")) {
      const trigger = document.createElement("pagefind-modal-trigger");
      trigger.setAttribute("aria-label", "Search documentation");
      mobileBtn.insertAdjacentElement("afterend", trigger);
    }
  }

  function mountNavbarSearch() {
    ensurePagefindModal();
    ensureDesktopSearchbox();
    ensureMobileTrigger();
  }

  let mountTimer;
  function scheduleMountNavbarSearch() {
    clearTimeout(mountTimer);
    mountTimer = setTimeout(mountNavbarSearch, 50);
  }

  function watchNavbar() {
    const navbar =
      document.getElementById("navbar") ??
      document.querySelector("header nav") ??
      document.querySelector("header");
    if (!navbar) return;

    new MutationObserver(() => scheduleMountNavbarSearch()).observe(navbar, {
      childList: true,
      subtree: true,
    });
  }

  function patchHistory() {
    for (const method of ["pushState", "replaceState"]) {
      const original = history[method].bind(history);
      history[method] = function (...args) {
        const result = original(...args);
        scheduleMountNavbarSearch();
        return result;
      };
    }
    window.addEventListener("popstate", scheduleMountNavbarSearch);
    window.addEventListener("pageshow", scheduleMountNavbarSearch);
  }

  function initPagefindNavbar() {
    mountNavbarSearch();
    watchNavbar();
    patchHistory();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initPagefindNavbar);
  } else {
    initPagefindNavbar();
  }
</script>
`;

function injectPagefind(htmlPath) {
  let html = fs.readFileSync(htmlPath, "utf8");
  if (html.includes("pagefind-component-ui.js")) return;
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

console.log("Injecting Pagefind navbar search…");
walkHtml(siteDir);

console.log("Post-process complete.");
