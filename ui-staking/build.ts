#!/usr/bin/env ts-node

import esbuild from "esbuild";
import { copy } from "esbuild-plugin-copy";
import config from "./esbuild.config.ts";
import fs from "node:fs";

async function main() {
  const metafile = await esbuild.build({
    ...config,
    define: {
      ...config.define,
      "process.env.NODE_ENV": "'production'",
    },
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
    metafile: true,
    // mainFields: ["module", "main"],
  });

  fs.writeFileSync("dist/metafile", JSON.stringify(metafile.metafile));

  console.log("Build complete");
}

main();
