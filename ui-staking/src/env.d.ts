import type { Env } from "../env.ts";

declare global {
  namespace NodeJS {
    interface ProcessEnv extends Env {}
  }
}
