import { useEffect, useState } from 'react';
import styled from 'styled-components';
import {
  IconCheck,
  IconCopy,
  IconPencil,
  IconRefresh,
} from '@tabler/icons-react';
import {
  summarizeUsage,
  type MessageUsage as MessageUsageCounts,
} from '../../lib/messageUsage';

// Actions stay out of the way until the message is hovered or something inside
// it takes focus, so the transcript reads as prose rather than as a toolbar per
// paragraph. The row keeps its height whether or not they are showing, so
// revealing them never reflows the message above.
const Row = styled.div`
  display: flex;
  align-items: center;
  gap: 2px;
  margin-top: 4px;
`;

// Only the buttons fade. The usage figure beside them stays put, so revealing
// the actions never shifts anything.
const Buttons = styled.div`
  display: flex;
  align-items: center;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.12s ease;

  [data-message-row]:hover &,
  [data-message-row]:focus-within & {
    opacity: 1;
  }
`;

const ActionButton = styled.button.attrs({ type: 'button' })`
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 8px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: rgba(255, 255, 255, 0.42);
  font: inherit;
  font-size: 1.15rem;
  line-height: 1;
  cursor: pointer;
  transition:
    background 0.12s ease,
    color 0.12s ease;

  &:hover {
    background: rgba(255, 255, 255, 0.07);
    color: rgba(255, 255, 255, 0.82);
  }

  /* Keyboard users never hover, so the row has to be reachable by tab alone. */
  &:focus-visible {
    outline: 1px solid ${(p) => p.theme.colors.morMain};
    outline-offset: 1px;
    color: rgba(255, 255, 255, 0.82);
  }

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
`;

const Confirmed = styled.span`
  color: ${(p) => p.theme.colors.morMain};
`;

// What the turn cost, in the same row as the actions but not hidden with them.
// Spend is something a user checks without meaning to interact, so hiding it
// behind a hover would make the one number they might be watching the hardest
// one to see.
const Usage = styled.span`
  margin-left: 6px;
  color: rgba(255, 255, 255, 0.34);
  font-size: 1.1rem;
  line-height: 1;
  white-space: nowrap;
`;

export async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    // Clipboard access can be denied; the caller shows nothing rather than
    // claiming a copy that did not happen.
    return false;
  }
}

/**
 * The prompt that produced the message at `index`, or undefined when there is
 * none to find. Regenerating means re-asking the question, and the question is
 * the nearest user turn above the answer — not necessarily the one immediately
 * above it, since a run can interleave tool or status entries.
 *
 * Lives here rather than in Chat so it can be tested without mounting the
 * screen; it is the one piece of list reasoning the actions row needs.
 */
export function precedingUserText(
  messages: unknown,
  index: number,
): string | undefined {
  if (!Array.isArray(messages)) return undefined;
  for (let i = Math.min(index, messages.length) - 1; i >= 0; i -= 1) {
    const candidate = messages[i];
    if (candidate?.role !== 'user') continue;
    const text = candidate.text;
    // An image or audio turn has no text to resend, so it is not a prompt we
    // can replay; keep looking further up rather than sending an empty string.
    if (typeof text === 'string' && text.trim()) return text;
  }
  return undefined;
}

export type MessageActionsProps = {
  /** Raw text to place on the clipboard. Copy is hidden when this is empty. */
  text?: string;
  /** Present on user messages: load this message back into the composer. */
  onEdit?: () => void;
  /** Present on assistant messages: ask the model for another answer. */
  onRegenerate?: () => void;
  /** True while a request is in flight, when regenerating would race it. */
  busy?: boolean;
  /** Token counts for this answer, when the stream reported any. */
  usage?: MessageUsageCounts;
  /** MOR billed for the seconds this answer took, when the price is known. */
  costMor?: number;
};

export function MessageActions({
  text,
  onEdit,
  onRegenerate,
  busy = false,
  usage,
  costMor,
}: MessageActionsProps) {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) return undefined;
    const id = setTimeout(() => setCopied(false), 1600);
    return () => clearTimeout(id);
  }, [copied]);

  const canCopy = Boolean(text && text.trim());
  const usageSummary = summarizeUsage(usage, costMor);
  if (!canCopy && !onEdit && !onRegenerate && !usageSummary) {
    return null;
  }

  return (
    <Row data-testid="message-actions">
      <Buttons>
        {canCopy && (
          <ActionButton
            aria-label={copied ? 'Copied' : 'Copy message'}
            title="Copy message"
            onClick={async () => {
              if (await copyText(text as string)) setCopied(true);
            }}
          >
            {copied ? (
              <>
                <IconCheck size={14} stroke={2.2} />
                <Confirmed>Copied</Confirmed>
              </>
            ) : (
              <>
                <IconCopy size={14} stroke={1.8} />
                Copy
              </>
            )}
          </ActionButton>
        )}
        {onEdit && (
          <ActionButton
            aria-label="Edit and resend"
            title="Put this message back in the box to edit and send again"
            onClick={onEdit}
            disabled={busy}
          >
            <IconPencil size={14} stroke={1.8} />
            Edit
          </ActionButton>
        )}
        {onRegenerate && (
          <ActionButton
            aria-label="Regenerate response"
            title="Ask the model to answer again"
            onClick={onRegenerate}
            disabled={busy}
          >
            <IconRefresh size={14} stroke={1.8} />
            Retry
          </ActionButton>
        )}
      </Buttons>
      {usageSummary && (
        <Usage
          data-testid="message-usage"
          title="Tokens counted by your node, and the MOR billed for the seconds this answer took"
        >
          {usageSummary}
        </Usage>
      )}
    </Row>
  );
}

export default MessageActions;
