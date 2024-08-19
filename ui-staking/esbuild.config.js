#!/usr/bin/env node
//@ts-check

import esbuild from "esbuild";
import { livereloadPlugin } from "@jgoz/esbuild-plugin-livereload";
import { copy } from "esbuild-plugin-copy";

async function main() {
  const ctx = await esbuild.context({
    logLevel: "info",
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
    inject: ["./src/react-shim.ts"],
    plugins: [
      livereloadPlugin(),
      copy({
        resolveFrom: "out",
        assets: {
          from: "./public/**/*",
          to: ".",
        },
        watch: true,
      }),
    ],
    loader: {
      ".png": "file",
    },
  });

  await ctx.watch();

  const { host, port } = await ctx.serve({
    servedir: "dist",
    fallback: "dist/index.html",
  });

  console.log(`Serving on http://${host}:${port}`);
}

main();
