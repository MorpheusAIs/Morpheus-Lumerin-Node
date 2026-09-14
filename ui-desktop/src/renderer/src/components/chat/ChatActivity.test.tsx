import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, render, renderHook, screen } from '@testing-library/react';
import {
  ActivityStatus,
  derivePhase,
  formatElapsed,
  isInsideReasoning,
  lastAssistantText,
  useElapsed,
} from './ChatActivity';

describe('formatElapsed', () => {
  it('counts in seconds below a minute', () => {
    expect(formatElapsed(0)).toBe('0s');
    expect(formatElapsed(999)).toBe('0s');
    expect(formatElapsed(1000)).toBe('1s');
    expect(formatElapsed(59_999)).toBe('59s');
  });

  it('switches to minutes and then hours', () => {
    expect(formatElapsed(60_000)).toBe('1m 0s');
    expect(formatElapsed(95_000)).toBe('1m 35s');
    expect(formatElapsed(3_600_000)).toBe('1h 0m');
    expect(formatElapsed(7_500_000)).toBe('2h 5m');
  });

  // A clock skew between two Date.now() calls can hand us a negative span, and
  // "-1s" in the header would look like a bug in the app rather than the clock.
  it('never reports negative time', () => {
    expect(formatElapsed(-5000)).toBe('0s');
  });
});

describe('useElapsed', () => {
  afterEach(() => vi.useRealTimers());

  const useFakeClock = (now: number) => {
    vi.useFakeTimers();
    vi.setSystemTime(now);
  };

  it('reports nothing while the work is not running', () => {
    useFakeClock(1_000_000);
    const { result } = renderHook(() => useElapsed(false));
    expect(result.current).toBe(0);
  });

  it('counts from now when no start is given', () => {
    useFakeClock(1_000_000);
    const { result } = renderHook(() => useElapsed(true));

    expect(result.current).toBe(0);
    act(() => void vi.advanceTimersByTime(3_000));
    expect(result.current).toBe(3_000);
  });

  // The Workspace bug: a run outlives the screen that shows it, so a timer
  // anchored to mount restarted at zero every time the user came back to the
  // project and a ten minute run claimed to be seconds old.
  it('counts from the run start when one is given', () => {
    useFakeClock(1_000_000);
    const { result } = renderHook(() => useElapsed(true, 250, 1_000_000 - 600_000));

    expect(result.current).toBe(600_000);
    act(() => void vi.advanceTimersByTime(1_000));
    expect(result.current).toBe(601_000);
  });

  it('keeps counting from the same start across a remount', () => {
    useFakeClock(1_000_000);
    const startedAt = 1_000_000 - 120_000;

    const first = renderHook(() => useElapsed(true, 250, startedAt));
    expect(first.result.current).toBe(120_000);
    first.unmount();

    act(() => void vi.advanceTimersByTime(5_000));
    const second = renderHook(() => useElapsed(true, 250, startedAt));
    expect(second.result.current).toBe(125_000);
  });

  it('treats a start in the future as now rather than counting backwards', () => {
    // Reachable through a clock change between launches, and a negative count
    // would read as a bug in the app rather than in the clock.
    useFakeClock(1_000_000);
    const { result } = renderHook(() => useElapsed(true, 250, 1_500_000));
    expect(result.current).toBe(0);
  });

  it('ignores a start that is not a real number', () => {
    useFakeClock(1_000_000);
    const { result } = renderHook(() => useElapsed(true, 250, Number.NaN));
    expect(result.current).toBe(0);
  });

  it('holds its last value once the work stops', () => {
    useFakeClock(1_000_000);
    const { result, rerender } = renderHook(
      ({ active }) => useElapsed(active, 250, 1_000_000 - 30_000),
      { initialProps: { active: true } },
    );

    expect(result.current).toBe(30_000);
    rerender({ active: false });
    act(() => void vi.advanceTimersByTime(10_000));
    expect(result.current).toBe(30_000);
  });
});

describe('isInsideReasoning', () => {
  it('is true only while a reasoning tag is unterminated', () => {
    expect(isInsideReasoning('<think>weighing it')).toBe(true);
    expect(isInsideReasoning('<think>weighing it</think>')).toBe(false);
    expect(isInsideReasoning('plain text')).toBe(false);
  });

  it('handles the other tag spellings and attributes', () => {
    expect(isInsideReasoning('<reasoning depth="2">hm')).toBe(true);
    expect(isInsideReasoning('<REFLECTION>hm</REFLECTION>')).toBe(false);
  });
});

describe('derivePhase', () => {
  it('reports connecting before any token arrives', () => {
    expect(derivePhase(undefined)).toBe('connecting');
    expect(derivePhase('')).toBe('connecting');
    expect(derivePhase('   \n ')).toBe('connecting');
  });

  it('reports thinking inside an open reasoning block', () => {
    expect(derivePhase('<think>let me check')).toBe('thinking');
  });

  // The gap between "done reasoning" and "first word of the answer" is exactly
  // the moment a spinner looks most like a freeze, so it gets a label too.
  it('falls back to connecting once reasoning closes with nothing after it', () => {
    expect(derivePhase('<think>done</think>')).toBe('connecting');
    expect(derivePhase('<think>done</think>\n\n')).toBe('connecting');
  });

  it('reports writing once the answer starts', () => {
    expect(derivePhase('<think>done</think>Here is the answer')).toBe(
      'writing',
    );
    expect(derivePhase('Here is the answer')).toBe('writing');
  });
});

describe('lastAssistantText', () => {
  it('returns the text of a trailing assistant message', () => {
    expect(
      lastAssistantText([
        { role: 'user', text: 'hi' },
        { role: 'assistant', text: 'hello' },
      ]),
    ).toBe('hello');
  });

  it('returns undefined while the user message is still last', () => {
    expect(lastAssistantText([{ role: 'user', text: 'hi' }])).toBeUndefined();
  });

  it('tolerates a missing or malformed list', () => {
    expect(lastAssistantText(undefined)).toBeUndefined();
    expect(lastAssistantText([])).toBeUndefined();
    expect(lastAssistantText([null])).toBeUndefined();
    expect(
      lastAssistantText([{ role: 'assistant', text: 42 }]),
    ).toBeUndefined();
  });
});

describe('ActivityStatus', () => {
  it('announces the phase politely for screen readers', () => {
    render(<ActivityStatus phase="tool" elapsedMs={5000} />);

    const row = screen.getByRole('status');
    expect(row.getAttribute('aria-live')).toBe('polite');
    expect(screen.getByText('Running a command')).toBeTruthy();
    expect(screen.getByText('5s')).toBeTruthy();
  });

  it('hides the counter before the first tick', () => {
    render(<ActivityStatus phase="connecting" elapsedMs={0} />);

    expect(screen.getByText('Working on it')).toBeTruthy();
    expect(screen.queryByText('0s')).toBeNull();
  });

  it('shows the detail when one is given', () => {
    render(
      <ActivityStatus phase="tool" elapsedMs={1000} detail="Reading the CSV" />,
    );

    expect(screen.getByText('Reading the CSV')).toBeTruthy();
  });
});
