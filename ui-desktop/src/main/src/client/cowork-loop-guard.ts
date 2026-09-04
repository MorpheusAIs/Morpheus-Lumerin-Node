import { createHash } from 'node:crypto'

const MUTATION_TOOLS = new Set([
  'write_file',
  'make_directory',
  'copy_file',
  'move_file',
  'delete_file',
  'create_docx',
  'create_xlsx',
  'create_pptx',
  'create_pdf'
])

const GENERATED_PAYLOAD_TOOLS = new Set([
  'write_file',
  'create_docx',
  'create_xlsx',
  'create_pptx',
  'create_pdf'
])

const VERIFICATION_TOOLS = new Set([
  'inspect_file',
  'read_file',
  'read_document',
  'search_files',
  'analyze_csv'
])

const NEUTRAL_PLAN_TOOLS = new Set(['set_plan', 'update_plan_step'])

export type CoworkLoopGuardReason =
  | 'repeated-equivalent-action'
  | 'repeated-destination'
  | 'unverified-mutation-limit'
  | 'repeated-content-across-destinations'
  | 'alternating-action-cycle'

export interface CoworkLoopGuardState {
  version: 1
  unverifiedMutations: number
  lastMutationHash?: string
  equivalentMutationStreak: number
  lastDestinationHash?: string
  destinationMutationStreak: number
  recentMutationHashes: string[]
  contentDestinationsByHash: Record<string, string[]>
}

export interface CoworkLoopGuardAction {
  toolName: string
  input: Record<string, unknown>
}

export interface CoworkLoopGuardDecision {
  state: CoworkLoopGuardState
  blocked: boolean
  reason?: CoworkLoopGuardReason
  message?: string
}

export function createCoworkLoopGuardState(): CoworkLoopGuardState {
  return {
    version: 1,
    unverifiedMutations: 0,
    equivalentMutationStreak: 0,
    destinationMutationStreak: 0,
    recentMutationHashes: [],
    contentDestinationsByHash: {}
  }
}

function canonicalJson(value: unknown): string {
  if (value === null) return 'null'
  if (typeof value === 'string' || typeof value === 'boolean') return JSON.stringify(value)
  if (typeof value === 'number') return Number.isFinite(value) ? JSON.stringify(value) : 'null'
  if (Array.isArray(value)) return `[${value.map(canonicalJson).join(',')}]`
  if (value && typeof value === 'object') {
    const object = value as Record<string, unknown>
    return `{${Object.keys(object)
      .sort()
      .map((key) => `${JSON.stringify(key)}:${canonicalJson(object[key])}`)
      .join(',')}}`
  }
  return 'null'
}

function hash(label: string, value: unknown): string {
  return createHash('sha256').update(label).update('\0').update(canonicalJson(value)).digest('hex')
}

function normalizedRelativePath(value: unknown): string | undefined {
  if (typeof value !== 'string' || !value.trim()) return undefined
  const parts: string[] = []
  for (const part of value.trim().replaceAll('\\', '/').split('/')) {
    if (!part || part === '.') continue
    if (part === '..') parts.pop()
    else parts.push(part)
  }
  return parts.join('/')
}

function canonicalInput(input: Record<string, unknown>): Record<string, unknown> {
  const canonical = { ...input }
  for (const key of ['path', 'source', 'destination']) {
    if (key in canonical) canonical[key] = normalizedRelativePath(canonical[key]) ?? ''
  }
  return canonical
}

function destination(input: Record<string, unknown>): string | undefined {
  return normalizedRelativePath(input.path ?? input.destination)
}

function generatedPayload(input: Record<string, unknown>): Record<string, unknown> {
  const payload = { ...input }
  delete payload.path
  delete payload.destination
  return payload
}

function clonedState(state: CoworkLoopGuardState): CoworkLoopGuardState {
  return {
    ...state,
    recentMutationHashes: [...state.recentMutationHashes],
    contentDestinationsByHash: Object.fromEntries(
      Object.entries(state.contentDestinationsByHash).map(([key, values]) => [key, [...values]])
    )
  }
}

function resetMutationSequence(state: CoworkLoopGuardState, resetUnverified: boolean): void {
  if (resetUnverified) state.unverifiedMutations = 0
  delete state.lastMutationHash
  state.equivalentMutationStreak = 0
  delete state.lastDestinationHash
  state.destinationMutationStreak = 0
  state.recentMutationHashes = []
}

function blocked(
  state: CoworkLoopGuardState,
  reason: CoworkLoopGuardReason,
  message: string
): CoworkLoopGuardDecision {
  return { state, blocked: true, reason, message }
}

/**
 * Records one accepted tool action and decides whether a file mutation must be paused.
 * Callers should record verification tools only after a successful read. State contains
 * hashes and counters only; generated content and raw arguments are never retained.
 */
export function evaluateCoworkLoopGuard(
  current: Readonly<CoworkLoopGuardState>,
  action: CoworkLoopGuardAction
): CoworkLoopGuardDecision {
  const state = clonedState(current as CoworkLoopGuardState)

  if (VERIFICATION_TOOLS.has(action.toolName)) {
    resetMutationSequence(state, true)
    return { state, blocked: false }
  }
  if (NEUTRAL_PLAN_TOOLS.has(action.toolName)) return { state, blocked: false }
  if (!MUTATION_TOOLS.has(action.toolName)) {
    resetMutationSequence(state, false)
    return { state, blocked: false }
  }

  const input = canonicalInput(action.input)
  const mutationHash = hash(`mutation:${action.toolName}`, input)
  const destinationValue = destination(input)
  const destinationHash = destinationValue
    ? hash('destination', destinationValue)
    : hash('destination', '')

  state.equivalentMutationStreak =
    state.lastMutationHash === mutationHash ? state.equivalentMutationStreak + 1 : 1
  state.lastMutationHash = mutationHash
  state.destinationMutationStreak =
    state.lastDestinationHash === destinationHash ? state.destinationMutationStreak + 1 : 1
  state.lastDestinationHash = destinationHash
  state.unverifiedMutations += 1
  state.recentMutationHashes = [...state.recentMutationHashes, mutationHash].slice(-4)

  let repeatedContentAcrossDestinations = false
  if (GENERATED_PAYLOAD_TOOLS.has(action.toolName)) {
    const contentHash = hash(`generated-payload:${action.toolName}`, generatedPayload(input))
    const destinations = state.contentDestinationsByHash[contentHash] ?? []
    if (!destinations.includes(destinationHash)) {
      state.contentDestinationsByHash[contentHash] = [...destinations, destinationHash]
      repeatedContentAcrossDestinations = destinations.length >= 2
    }
  }

  const recent = state.recentMutationHashes
  const alternatingCycle =
    recent.length === 4 &&
    recent[0] === recent[2] &&
    recent[1] === recent[3] &&
    recent[0] !== recent[1]

  if (state.equivalentMutationStreak >= 3) {
    return blocked(
      state,
      'repeated-equivalent-action',
      'Paused after the same file action was requested three times consecutively. Inspect the current files before continuing.'
    )
  }
  if (state.destinationMutationStreak >= 3) {
    return blocked(
      state,
      'repeated-destination',
      'Paused after three consecutive mutations targeted the same destination. Inspect the current file before continuing.'
    )
  }
  if (repeatedContentAcrossDestinations) {
    return blocked(
      state,
      'repeated-content-across-destinations',
      'Paused after equivalent generated content was sent to three different destinations. Inspect the current files before continuing.'
    )
  }
  if (alternatingCycle) {
    return blocked(
      state,
      'alternating-action-cycle',
      'Paused after detecting a repeating A-B file-action cycle. Inspect the current files before continuing.'
    )
  }
  if (state.unverifiedMutations >= 4) {
    return blocked(
      state,
      'unverified-mutation-limit',
      'Paused before a fourth file mutation without a successful verification read. Inspect the current files before continuing.'
    )
  }

  return { state, blocked: false }
}
