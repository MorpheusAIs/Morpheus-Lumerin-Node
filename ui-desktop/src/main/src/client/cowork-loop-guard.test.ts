import { describe, expect, it } from 'vitest'
import {
  createCoworkLoopGuardState,
  evaluateCoworkLoopGuard,
  revertCoworkLoopGuardMutation,
  type CoworkLoopGuardAction,
  type CoworkLoopGuardDecision,
  type CoworkLoopGuardState
} from './cowork-loop-guard'

function apply(
  state: CoworkLoopGuardState,
  toolName: string,
  input: Record<string, unknown> = {}
): CoworkLoopGuardDecision {
  return evaluateCoworkLoopGuard(state, { toolName, input })
}

function run(actions: CoworkLoopGuardAction[]): CoworkLoopGuardDecision[] {
  const decisions: CoworkLoopGuardDecision[] = []
  let state = createCoworkLoopGuardState()
  for (const action of actions) {
    const decision = evaluateCoworkLoopGuard(state, action)
    decisions.push(decision)
    state = decision.state
  }
  return decisions
}

describe('cowork loop guard', () => {
  it('blocks the third equivalent action even when plan updates separate attempts', () => {
    const write = {
      toolName: 'write_file',
      input: { path: './text_analyzer.py', content: 'same program' }
    }
    const decisions = run([
      write,
      { toolName: 'update_plan_step', input: { id: '1', status: 'in_progress' } },
      { ...write, input: { content: 'same program', path: 'text_analyzer.py' } },
      { toolName: 'set_plan', input: { steps: [] } },
      write
    ])

    expect(decisions.at(-1)).toMatchObject({
      blocked: true,
      reason: 'repeated-equivalent-action'
    })
  })

  it('blocks a sixth mutation to one destination when each payload changes', () => {
    const versions = ['one', 'two', 'three', 'four', 'five', 'six']
    const decisions = run(
      versions.map((version, index) => ({
        toolName: 'write_file',
        // Alternating the spelling proves normalisation, not a new destination.
        input: { path: index % 2 ? './calc.py' : 'calc.py', content: `version ${version}` }
      }))
    )

    expect(decisions.slice(0, -1).every((decision) => !decision.blocked)).toBe(true)
    expect(decisions.at(-1)).toMatchObject({ blocked: true, reason: 'repeated-destination' })
  })

  it('lets a run lay down many new files without a verification read between them', () => {
    // Building a directory skeleton is ordinary progress, not repetition. Counting
    // each new file against the blind-work budget paused real tasks partway.
    const decisions = run([
      { toolName: 'make_directory', input: { path: 'report' } },
      { toolName: 'write_file', input: { path: 'report/one.md', content: 'one' } },
      { toolName: 'update_plan_step', input: { id: '1', status: 'completed' } },
      { toolName: 'write_file', input: { path: 'report/two.md', content: 'two' } },
      { toolName: 'set_plan', input: { steps: [] } },
      { toolName: 'write_file', input: { path: 'report/three.md', content: 'three' } },
      { toolName: 'write_file', input: { path: 'report/four.md', content: 'four' } },
      { toolName: 'write_file', input: { path: 'report/five.md', content: 'five' } }
    ])

    expect(decisions.every((decision) => !decision.blocked)).toBe(true)
  })

  it('blocks a run that keeps rewriting the same destinations without ever reading', () => {
    const paths = ['a.md', 'b.md', 'c.md']
    const actions: CoworkLoopGuardAction[] = []
    for (let pass = 0; pass < 5; pass += 1) {
      for (const path of paths) {
        actions.push({ toolName: 'write_file', input: { path, content: `pass ${pass}` } })
      }
    }
    const decisions = run(actions)

    expect(decisions.some((decision) => decision.reason === 'unverified-mutation-limit')).toBe(true)
  })

  it('blocks identical generated content when it spreads to a fifth destination', () => {
    // Reusing one template across a few files is scaffolding; five is a stuck model.
    const decisions = run(
      ['a.py', 'b.py', 'c.py', 'd.py', 'e.py'].map((path) => ({
        toolName: 'write_file',
        input: { path, content: '[omitted after execution: 40 characters]' }
      }))
    )

    expect(decisions.at(-1)).toMatchObject({
      blocked: true,
      reason: 'repeated-content-across-destinations'
    })
  })

  it('blocks a repeating A-B action cycle before the generic unverified limit', () => {
    const first = { toolName: 'write_file', input: { path: 'a.py', content: 'alpha' } }
    const second = { toolName: 'write_file', input: { path: 'b.py', content: 'beta' } }
    const decisions = run([first, second, first, second])

    expect(decisions.at(-1)).toMatchObject({
      blocked: true,
      reason: 'alternating-action-cycle'
    })
  })

  it('resets unverified mutations after a successful verification read', () => {
    let state = createCoworkLoopGuardState()
    for (const path of ['one.py', 'two.py']) {
      state = apply(state, 'write_file', { path, content: path }).state
    }
    // A second visit to an already-written destination is what accumulates.
    state = apply(state, 'write_file', { path: 'one.py', content: 'revised' }).state
    expect(state.unverifiedMutations).toBe(1)

    const verified = apply(state, 'read_file', { path: 'one.py' })
    const nextWrite = apply(verified.state, 'write_file', {
      path: 'one.py',
      content: 'verified continuation'
    })

    expect(verified.state.unverifiedMutations).toBe(0)
    expect(verified.state.unverifiedDestinations).toEqual([])
    expect(nextWrite).toMatchObject({ blocked: false })
    expect(nextWrite.state.unverifiedMutations).toBe(0)
  })

  it('retains only hashes and counters, never raw paths, arguments, or generated content', () => {
    const secret = 'private generated payload that must not remain in guard state'
    const decision = apply(createCoworkLoopGuardState(), 'write_file', {
      path: 'private/final-report.txt',
      content: secret
    })
    const stored = JSON.stringify(decision.state)

    expect(stored).not.toContain(secret)
    expect(stored).not.toContain('private/final-report.txt')
    expect(decision.state.lastMutationHash).toMatch(/^[a-f0-9]{64}$/)
    expect(decision.state.lastDestinationHash).toMatch(/^[a-f0-9]{64}$/)
    expect(Object.keys(decision.state.contentDestinationsByHash)[0]).toMatch(/^[a-f0-9]{64}$/)
  })
})

describe('mutations that never reached disk', () => {
  /** Mirrors the runner: evaluate, then roll back whatever the tool failed to do. */
  function runWithOutcomes(
    actions: (CoworkLoopGuardAction & { failed?: boolean })[]
  ): CoworkLoopGuardDecision[] {
    const decisions: CoworkLoopGuardDecision[] = []
    let state = createCoworkLoopGuardState()
    for (const action of actions) {
      const before = state
      const decision = evaluateCoworkLoopGuard(state, action)
      decisions.push(decision)
      state = decision.blocked
        ? decision.state
        : action.failed
          ? revertCoworkLoopGuardMutation(decision.state, before)
          : decision.state
    }
    return decisions
  }

  it('lets a model keep correcting a rejected write to the same destination', () => {
    // The schema error it was fixing is the whole reason it retried. Counting a
    // rejected call as a mutation paused the run on the turn after the model was
    // finally told what to fix.
    // Comfortably past the streak a successful write would trip, so this fails
    // if a rejected call is ever counted as having touched the destination.
    const decisions = runWithOutcomes([
      ...[80, 72, 64, 56, 48, 40, 32].map((width) => ({
        toolName: 'create_xlsx',
        input: { path: 'docs/M.xlsx', columnWidths: [width] },
        failed: true
      })),
      { toolName: 'create_xlsx', input: { path: 'docs/M.xlsx', columnWidths: [24] } }
    ])
    expect(decisions.every((decision) => !decision.blocked)).toBe(true)
  })

  it('still catches a model resending the identical failing call', () => {
    const call = { toolName: 'create_xlsx', input: { path: 'docs/M.xlsx' }, failed: true }
    const decisions = runWithOutcomes([call, call, call])
    expect(decisions.at(-1)?.reason).toBe('repeated-equivalent-action')
  })

  it('still catches genuine rewrites of a destination that did reach disk', () => {
    const decisions = runWithOutcomes(
      ['one', 'two', 'three', 'four', 'five', 'six'].map((content) => ({
        toolName: 'write_file',
        input: { path: 'a.md', content }
      }))
    )
    expect(decisions.at(-1)?.reason).toBe('repeated-destination')
  })

  it('does not let a failure erase progress already recorded for other files', () => {
    const decisions = runWithOutcomes([
      { toolName: 'write_file', input: { path: 'b.md', content: 'kept' } },
      { toolName: 'write_file', input: { path: 'b.md', content: 'boom' }, failed: true },
      { toolName: 'write_file', input: { path: 'b.md', content: 'again' } }
    ])
    // Two writes landed on b.md, so the destination streak is 2, not reset to 1.
    expect(decisions.at(-1)?.state.destinationMutationStreak).toBe(2)
    expect(decisions.every((decision) => !decision.blocked)).toBe(true)
  })
})
