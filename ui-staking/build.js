#!/usr/bin/env node
//@ts-check

import esbuild from "esbuild";
import { copy } from "esbuild-plugin-copy";
import config from "./esbuild.config.js";

async function main() {
  await esbuild.build({
    ...config,
    plugins: [
      ...(config.plugins ? config.plugins : []),
      copy({
        resolveFrom: "out",
        assets: {
          from: "./public/**/*",
          to: ".",
        },
      }),
    ],
    minify: true,
    treeShaking: true,
  });

  console.log("Build complete");
}

main();
