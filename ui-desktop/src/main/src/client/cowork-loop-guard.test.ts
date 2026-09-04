import { describe, expect, it } from 'vitest'
import {
  createCoworkLoopGuardState,
  evaluateCoworkLoopGuard,
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

  it('blocks a third mutation to one destination when each payload changes', () => {
    const decisions = run([
      { toolName: 'write_file', input: { path: 'calc.py', content: 'version one' } },
      { toolName: 'write_file', input: { path: './calc.py', content: 'version two' } },
      { toolName: 'write_file', input: { path: 'calc.py', content: 'version three' } }
    ])

    expect(decisions.at(-1)).toMatchObject({ blocked: true, reason: 'repeated-destination' })
  })

  it('blocks a fourth unverified mutation and does not let plan churn reset the count', () => {
    const decisions = run([
      { toolName: 'write_file', input: { path: 'one.py', content: 'one' } },
      { toolName: 'update_plan_step', input: { id: '1', status: 'completed' } },
      { toolName: 'write_file', input: { path: 'two.py', content: 'two' } },
      { toolName: 'set_plan', input: { steps: [] } },
      { toolName: 'write_file', input: { path: 'three.py', content: 'three' } },
      { toolName: 'update_plan_step', input: { id: '2', status: 'in_progress' } },
      { toolName: 'write_file', input: { path: 'four.py', content: 'four' } }
    ])

    expect(decisions.at(-1)).toMatchObject({
      blocked: true,
      reason: 'unverified-mutation-limit'
    })
  })

  it('blocks identical generated content when it spreads to a third destination', () => {
    const decisions = run(
      ['simple_calculator.py', 'calculator_input.py', 'calc.py'].map((path) => ({
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
    for (const path of ['one.py', 'two.py', 'three.py']) {
      state = apply(state, 'write_file', { path, content: path }).state
    }
    const verified = apply(state, 'read_file', { path: 'three.py' })
    const nextWrite = apply(verified.state, 'write_file', {
      path: 'four.py',
      content: 'verified continuation'
    })

    expect(verified.state.unverifiedMutations).toBe(0)
    expect(nextWrite).toMatchObject({ blocked: false })
    expect(nextWrite.state.unverifiedMutations).toBe(1)
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
