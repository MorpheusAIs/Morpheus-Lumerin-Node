import { useEffect, useRef, useState } from 'react';
import styled from 'styled-components';

// A model that reasons can sit silent for minutes before its first token. With
// nothing on screen but a spinner, that is indistinguishable from a hung app,
// and the honest fix is to say what is happening and for how long — the elapsed
// count is the part that tells the user the app is still alive.

export type ActivityPhase =
  | 'connecting'
  | 'thinking'
  | 'tool'
  | 'writing'
  | 'transcribing'
  | 'speaking';

const PHASE_LABELS: Record<ActivityPhase, string> = {
  connecting: 'Working on it',
  thinking: 'Thinking',
  tool: 'Running a command',
  writing: 'Writing the answer',
  transcribing: 'Transcribing the audio',
  speaking: 'Generating the audio',
};

export function formatElapsed(ms: number): string {
  const totalSeconds = Math.max(0, Math.floor(ms / 1000));
  if (totalSeconds < 60) return `${totalSeconds}s`;
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes < 60) return `${minutes}m ${seconds}s`;
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

/**
 * Milliseconds since the work started, ticking while `active` stays true and
 * holding its final value once it goes false. Returns 0 before the first run,
 * so a caller can treat 0 as "nothing to show".
 *
 * `startedAtMs` is when the work actually began, which is not the same as when
 * this hook was mounted. A Workspace run outlives the screen that shows it, so
 * a timer anchored to mount restarted from zero every time the user left the
 * project and came back, and a run that had been going for ten minutes claimed
 * to be four seconds old. Passing the run's own stored start makes the count
 * describe the run rather than the visit. Chat, where the request and the view
 * do begin together, passes nothing and anchors to now.
 */
export function useElapsed(
  active: boolean,
  tickMs = 250,
  startedAtMs?: number,
): number {
  const startedAt = useRef<number | undefined>(undefined);
  const [elapsed, setElapsed] = useState(0);

  useEffect(() => {
    if (!active) {
      startedAt.current = undefined;
      return undefined;
    }
    const now = Date.now();
    // A stored start in the future would count backwards, so it is treated as
    // now. Clock changes between launches make that reachable.
    const anchor =
      startedAtMs !== undefined && Number.isFinite(startedAtMs)
        ? Math.min(startedAtMs, now)
        : now;
    startedAt.current = anchor;
    setElapsed(now - anchor);
    const id = setInterval(() => {
      if (startedAt.current !== undefined) {
        setElapsed(Date.now() - startedAt.current);
      }
    }, tickMs);
    return () => clearInterval(id);
  }, [active, tickMs, startedAtMs]);

  return elapsed;
}

const OPEN_REASONING_RE =
  /<(think|thinking|thought|reasoning|reflection)(?:\s[^>]*)?>/gi;
const CLOSE_REASONING_RE =
  /<\/(think|thinking|thought|reasoning|reflection)\s*>/gi;

/** True when the stream is currently inside an unterminated reasoning tag. */
export function isInsideReasoning(text: string): boolean {
  const opens = (text.match(OPEN_REASONING_RE) || []).length;
  const closes = (text.match(CLOSE_REASONING_RE) || []).length;
  return opens > closes;
}

/**
 * What to tell the user the model is doing, from the text received so far.
 * Deliberately derived rather than tracked: the stream is the only thing that
 * actually knows, and a separate flag would drift from it.
 */
export function derivePhase(streamedText: string | undefined): ActivityPhase {
  const text = streamedText ?? '';
  if (text.trim().length === 0) return 'connecting';
  if (isInsideReasoning(text)) return 'thinking';
  // Everything before the last closing tag is reasoning; if nothing followed it
  // the model has finished thinking but has not started the answer yet.
  const afterReasoning = text.split(/<\/(?:think|thinking|thought|reasoning|reflection)\s*>/i).pop() ?? '';
  if (afterReasoning.trim().length === 0) return 'connecting';
  return 'writing';
}

/**
 * Text of the in-flight assistant reply, or undefined while the user's own
 * message is still the last one — i.e. before the model has said anything.
 */
export function lastAssistantText(messages: unknown): string | undefined {
  if (!Array.isArray(messages) || messages.length === 0) return undefined;
  const last = messages[messages.length - 1];
  if (!last || last.role !== 'assistant') return undefined;
  return typeof last.text === 'string' ? last.text : undefined;
}

const StatusRow = styled.div`
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin: 4px 0 12px;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.65);
  font-size: 1.3rem;
`;

const Pulse = styled.span`
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex: none;
  background: ${(p) => p.theme?.colors?.morMain ?? 'rgba(32, 220, 142, 1)'};
  animation: activityPulse 1.3s ease-in-out infinite;

  @keyframes activityPulse {
    0%,
    100% {
      opacity: 0.25;
      transform: scale(0.8);
    }
    50% {
      opacity: 1;
      transform: scale(1.15);
    }
  }
`;

const Elapsed = styled.span`
  color: rgba(255, 255, 255, 0.4);
  font-variant-numeric: tabular-nums;
`;

const Detail = styled.span`
  color: rgba(255, 255, 255, 0.45);
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`;

export function ActivityStatus({
  phase,
  elapsedMs,
  detail,
}: {
  phase: ActivityPhase;
  elapsedMs: number;
  detail?: string;
}) {
  return (
    <StatusRow role="status" aria-live="polite">
      <Pulse />
      <span>{PHASE_LABELS[phase]}</span>
      {detail && <Detail>{detail}</Detail>}
      {elapsedMs > 0 && <Elapsed>{formatElapsed(elapsedMs)}</Elapsed>}
    </StatusRow>
  );
}
