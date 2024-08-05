import readline from "readline";
import yargs from "yargs-parser";
import { Address, isAddress as viemIsAddress } from "viem";

/** Retuns variabe from env, args or prompts for it */
export async function getVar<T>(params: Params<T>): Promise<T> {
  const { envName, argName, prompt, validator = noop } = params;

  // wraps the validator to provide a better context
  async function validatePrint<T>(value: string): Promise<T> {
    try {
      const v = (await validator(value)) as T;
      console.log(`${params.argName}: ${params.secret ? formatSecret(v) : v}`);
      return v;
    } catch (error) {
      throw new Error(
        `Invalid value provided for ${argName} (${envName}): "${value}"\n ${error}`,
      );
    }
  }

  if (envName) {
    const envValue = process.env[envName];
    if (envValue) {
      return await validatePrint(envValue);
    }
  }

  if (argName) {
    const argValue = yargs(process.argv.slice(2))[argName] as
      | string
      | undefined;
    if (argValue) {
      return await validatePrint(argValue);
    }
  }

  if (prompt) {
    const promptValue = await promptConsole(prompt);
    return await validatePrint(promptValue);
  }

  throw new Error("No value provided for " + argName);
}

type Params<T = undefined> = {
  envName?: string;
  argName?: string;
  prompt?: string;
  validator?: (input: string) => Promise<T>;
  secret?: boolean;
};

function promptConsole(prompt: string): Promise<string> {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });
  return new Promise<string>((resolve) => {
    rl.question(`${prompt}: `, (answer) => {
      rl.close();
      resolve(answer);
    });
  });
}

// function to prepare secret before printing by replacing all chars with "*"
function formatSecret(value: any): string {
  if (value.toString) {
    return "*".repeat(value.toString().length);
  }
  return "***";
}

export async function isDefined(value: string): Promise<string> {
  if (value === "") {
    throw new Error("Value is empty");
  }
  return value;
}

export async function isAddress(value: string): Promise<Address> {
  if (!viemIsAddress(value)) {
    throw new Error("Value is not a valid address");
  }
  return value;
}

export async function isBigInt(value: string): Promise<bigint> {
  return BigInt(value);
}

export async function noop(value: string): Promise<string> {
  return value;
}
