#!/usr/bin/env node
//@ts-check

import esbuild from "esbuild";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { livereloadPlugin } from "@jgoz/esbuild-plugin-livereload";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function copyPublicFolder() {
  const publicDir = path.resolve(__dirname, "public");
  const distDir = path.resolve(__dirname, "dist");

  fs.mkdirSync(distDir, { recursive: true });

  const files = fs.readdirSync(publicDir);
  for (const file of files) {
    const srcPath = path.join(publicDir, file);
    const destPath = path.join(distDir, file);

    fs.copyFileSync(srcPath, destPath);
  }
}

/**  @type {import('esbuild').BuildOptions} */
const esbuildOptions = {
  logLevel: "info",
  entryPoints: ["src/index.tsx"],
  bundle: true,
  outfile: "dist/index.js",
  define: {
    "process.env.NODE_ENV": '"production"',
  },
  plugins: [livereloadPlugin()],
};

async function main() {
  const isWatch = process.argv.includes("-w");

  const ctx = await esbuild.context({
    entryPoints: ["src/index.tsx"],
    bundle: true,
    outdir: "dist",
    define: {
      "process.env.NODE_ENV": '"development"',
      "process.env.REACT_APP_STAKING_ADDR": '"0x959922be3caee4b8cd9a407cc3ac1c251c2007b1"',
      "process.env.REACT_APP_ETH_NODE_URL": '"http://0.0.0.0:8545"',
      "process.env.REACT_APP_MOR_ADDR": '"0x5fbdb2315678afecb367f032d93f642f64180aa3"',
      "process.env.REACT_APP_LMR_ADDR": '"0x0B306BF915C4d645ff596e518fAf3F9669b97016"',
    },
  });
  copyPublicFolder();

  await ctx.watch();

  const { host, port } = await ctx.serve({
    servedir: "dist",
  });

  console.log(`Serving on http://${host}:${port}`);
}

main();
