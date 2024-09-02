import type { BuildOptions } from "esbuild";
import { getAndValidateEnv } from "./env.ts";

export default {
  logLevel: "info",
  entryPoints: ["src/index.tsx"],
  bundle: true,
  outdir: "dist",
  define: {
    ...getAndValidateEnv().full,
  },
  inject: ["./src/config/react-shim.ts"],
  plugins: [],
  loader: {
    ".png": "file",
  },
} as BuildOptions;
