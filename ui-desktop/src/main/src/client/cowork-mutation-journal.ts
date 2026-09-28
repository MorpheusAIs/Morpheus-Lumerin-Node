import { createHash } from 'node:crypto'
import type { CoworkToolCall, CoworkToolExecution } from './cowork.types'

function canonicalJson(value: unknown): string {
  if (value === null || typeof value !== 'object') return JSON.stringify(value) ?? 'null'
  if (Array.isArray(value)) return `[${value.map(canonicalJson).join(',')}]`
  const object = value as Record<string, unknown>
  return `{${Object.keys(object)
    .sort()
    .map((key) => `${JSON.stringify(key)}:${canonicalJson(object[key])}`)
    .join(',')}}`
}

export function mutationArgumentsHash(toolName: string, input: Record<string, unknown>): string {
  return createHash('sha256')
    .update(toolName)
    .update('\0')
    .update(canonicalJson(input))
    .digest('hex')
}

export function executionMatchesToolCall(
  execution: CoworkToolExecution,
  toolCall: CoworkToolCall
): boolean {
  if (execution.toolName !== toolCall.function.name) return false
  try {
    const input = JSON.parse(toolCall.function.arguments || '{}')
    if (!input || typeof input !== 'object' || Array.isArray(input)) return false
    return execution.argumentsHash === mutationArgumentsHash(toolCall.function.name, input)
  } catch {
    return false
  }
}
