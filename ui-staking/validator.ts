import * as Ajv from "ajv";
import { default as addFormats } from "ajv-formats";

export function newAjv(): Ajv.Ajv {
  //@ts-ignore
  return addFormats(new Ajv.Ajv({ coerceTypes: true }), [
    "date-time",
    "time",
    "date",
    "email",
    "hostname",
    "ipv4",
    "ipv6",
    "uri",
    "uri-reference",
    "uuid",
    "uri-template",
    "json-pointer",
    "relative-json-pointer",
    "regex",
  ]) as Ajv.Ajv;
}
