#!/usr/bin/env node
/**
 * Post-process a mint export directory: Pagefind index + navbar search +
 * llms.txt / llms-full.txt / per-page *.md.
 *
 * Mintlify's built-in search (docs.json "search") targets Mintlify Cloud and prompts
 * for CLI login on self-hosted S3/CloudFront exports. We use Pagefind (static index)
 * in the top navbar instead.
 *
 * Per-page *.md and the Fargate MCP at /mcp replace Mintlify cloud agent endpoints
 * that do not ship with `mint export`.
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

console.log("Generating llms.txt / llms-full.txt / per-page *.md…");
execSync(
  `node "${path.join(__dirname, "generate-llms-txt.mjs")}" "${docsDir}" "${siteDir}"`,
  { stdio: "inherit", env: { ...process.env, SITE_URL: siteUrl } }
);

// Agent discovery aids (Mintlify cloud hosts these; mint export does not).
const llmsTxtPath = path.join(siteDir, "llms.txt");
if (fs.existsSync(llmsTxtPath)) {
  const wellKnownDir = path.join(siteDir, ".well-known");
  fs.mkdirSync(wellKnownDir, { recursive: true });
  fs.copyFileSync(llmsTxtPath, path.join(wellKnownDir, "llms.txt"));
  const robots = [
    "User-agent: *",
    "Allow: /",
    "",
    `# LLM / agent corpus`,
    `Sitemap: ${siteUrl}/llms.txt`,
    `# Full docs as plain text: ${siteUrl}/llms-full.txt`,
    `# Per-page markdown: append .md to any docs path (or fetch without Accept: text/html)`,
    "",
  ].join("\n");
  fs.writeFileSync(path.join(siteDir, "robots.txt"), robots);
  console.log("Wrote robots.txt and .well-known/llms.txt");
}

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
  /* Mobile: icon-only trigger so the navbar does not grow a desktop ⌘K pill. */
  #nodedocs-pagefind-mobile-trigger.pf-nodedocs-mobile {
    display: inline-flex;
    align-items: center;
  }
  #nodedocs-pagefind-mobile-trigger.pf-nodedocs-mobile .pf-trigger-btn {
    width: 2rem;
    height: 2rem;
    min-width: 2rem;
    padding: 0;
    justify-content: center;
    border-radius: 0.5rem;
  }
</style>
<script type="module">
  import "/pagefind/pagefind-component-ui.js";

  /**
   * Mintlify hydration / client-nav re-renders the navbar (and sometimes body).
   * Pagefind's component registry does not deregister disconnected utilities, and
   * pagefind-modal-trigger.openModal() always uses getUtilities("modal")[0].
   * After a remount that orphan is detached, so mobile taps set aria-expanded but
   * never call showModal() on the live dialog (reproduced on iOS WebKit / Brave).
   */

  function findDesktopSlot() {
    const entry = document.getElementById("search-bar-entry");
    if (entry?.parentElement) return entry.parentElement;
    return (
      document.querySelector("#navbar .justify-center") ??
      document.querySelector("header .justify-center")
    );
  }

  /** Keep the modal outside <body> so Mintlify body remounts do not detach it. */
  function getPagefindRoot() {
    let root = document.getElementById("nodedocs-pagefind-root");
    if (!root) {
      root = document.createElement("div");
      root.id = "nodedocs-pagefind-root";
      document.documentElement.appendChild(root);
    }
    return root;
  }

  function ensurePagefindModal() {
    const root = getPagefindRoot();
    let modal = root.querySelector("pagefind-modal") ?? document.querySelector("pagefind-modal");
    if (!modal) {
      modal = document.createElement("pagefind-modal");
      modal.setAttribute("instance", "nodedocs");
      root.appendChild(modal);
      return modal;
    }
    if (modal.parentElement !== root) {
      root.appendChild(modal);
    }
    return modal;
  }

  /** Open the in-document modal; repair stuck Pagefind _isOpen / detached dialogEl. */
  function openLivePagefindModal() {
    const modal = ensurePagefindModal();
    if (!modal || typeof modal.open !== "function") return;

    const dialog = modal.dialogEl;
    if (modal._isOpen && !dialog?.open) {
      modal._isOpen = false;
    }
    if (dialog && !document.contains(dialog) && typeof modal.render === "function") {
      modal._isOpen = false;
      modal.render();
    }
    try {
      modal.open();
    } catch {
      modal._isOpen = false;
      if (typeof modal.render === "function") modal.render();
      modal.open();
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
      searchbox.setAttribute("instance", "nodedocs");
      searchbox.setAttribute("placeholder", "Search documentation…");
      host.replaceChildren(searchbox);
    }
  }

  function patchModalTrigger(trigger) {
    if (trigger.dataset.nodedocsOpenPatched === "1") return;
    trigger.dataset.nodedocsOpenPatched = "1";
    // Bypass Pagefind's stale getUtilities("modal")[0] lookup after remounts.
    trigger.openModal = function nodedocsOpenModal() {
      openLivePagefindModal();
      this.buttonEl?.setAttribute("aria-expanded", "true");
    };
  }

  function ensureMobileTrigger() {
    const mobileBtn = document.getElementById("search-bar-entry-mobile");
    if (!mobileBtn?.parentElement) return;

    const slot = mobileBtn.parentElement;
    let trigger = slot.querySelector("#nodedocs-pagefind-mobile-trigger");
    if (!trigger) {
      trigger = document.createElement("pagefind-modal-trigger");
      trigger.id = "nodedocs-pagefind-mobile-trigger";
      trigger.className = "pf-nodedocs-mobile";
      trigger.setAttribute("instance", "nodedocs");
      trigger.setAttribute("placeholder", "Search documentation");
      trigger.setAttribute("compact", "");
      trigger.setAttribute("hide-shortcut", "");
      mobileBtn.insertAdjacentElement("afterend", trigger);
    }
    patchModalTrigger(trigger);
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
