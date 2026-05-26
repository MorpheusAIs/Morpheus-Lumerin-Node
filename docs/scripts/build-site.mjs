#!/usr/bin/env node
/**
 * Build static docs site: mint export → postprocess (Pagefind navbar search + llms.txt).
 * Usage: SITE_URL=https://nodedocs.mor.org node scripts/build-site.mjs [outDir]
 */
import { execSync } from "child_process";
import fs from "fs";
import os from "os";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const docsDir = path.join(__dirname, "..");
const outDir = path.resolve(process.argv[2] ?? path.join(docsDir, "..", ".site"));
const siteUrl = process.env.SITE_URL ?? "https://nodedocs.mor.org";
const zipPath = path.join(os.tmpdir(), `nodedocs-${Date.now()}.zip`);

console.log(`Building docs site → ${outDir}`);
console.log(`Canonical URL: ${siteUrl}`);

execSync(`npx mint export --output "${zipPath}"`, {
  stdio: "inherit",
  cwd: docsDir,
});

fs.rmSync(outDir, { recursive: true, force: true });
fs.mkdirSync(outDir, { recursive: true });
execSync(`unzip -q -o "${zipPath}" -d "${outDir}"`, { stdio: "inherit" });
fs.rmSync(zipPath, { force: true });

execSync(`node "${path.join(__dirname, "postprocess-export.mjs")}" "${outDir}"`, {
  stdio: "inherit",
  env: { ...process.env, SITE_URL: siteUrl },
});

console.log(`Site ready at ${outDir}`);
