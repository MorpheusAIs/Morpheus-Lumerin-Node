const USERNAME_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._@-]{0,127}$/u
const TOKEN_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$/u

export function validateAgentUsername(value: unknown): string {
  if (
    typeof value !== 'string' ||
    !USERNAME_PATTERN.test(value) ||
    value.toLowerCase() === 'admin'
  ) {
    throw new Error('Agent username is invalid.')
  }
  return value
}

export function validateAgentToken(value: unknown): string {
  if (typeof value !== 'string' || !TOKEN_PATTERN.test(value)) {
    throw new Error('Agent token is invalid.')
  }
  return value
}

export function validateAgentDecision(value: unknown): boolean {
  if (typeof value !== 'boolean') throw new Error('Agent confirmation decision is invalid.')
  return value
}
