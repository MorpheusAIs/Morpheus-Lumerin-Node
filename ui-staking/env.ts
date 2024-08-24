import { env } from "@dotenv-run/core";
import { Type, type Static } from "@sinclair/typebox";
import { newAjv } from "./validator.ts";

// Eth address field validator
const TypeEthAddress = Type.String({ pattern: "^0x[a-fA-F0-9]{40}$" });

// Environment variables schema
const EnvSchema = Type.Object({
  NODE_ENV: Type.Union([Type.Literal("development"), Type.Literal("production")]),
  REACT_APP_CHAIN_ID: Type.Number(),
  REACT_APP_ETH_NODE_URL: Type.String({ format: "uri" }),
  REACT_APP_LMR_ADDR: TypeEthAddress,
  REACT_APP_MOR_ADDR: TypeEthAddress,
  REACT_APP_STAKING_ADDR: TypeEthAddress,
  REACT_APP_WALLET_CONNECT_PROJECT_ID: Type.String({ minLength: 1 }),
});

// Inferred type of environment variables
export type Env = Static<typeof EnvSchema>;

// Get and validate environment variables
export function getAndValidateEnv() {
  const ajv = newAjv();
  const validate = ajv.compile(EnvSchema);

  const e = env({
    prefix: "REACT_APP_",
    verbose: false,
    files: [".env"],
  });

  if (!validate(e.raw)) {
    throw new Error(
      `Invalid environment variables: ${ajv.errorsText(validate.errors, {
        dataVar: "ENV",
        separator: ".",
      })}`
    );
  }

  return e;
}
