import { useMemo } from 'react';
import styled from 'styled-components';
import { IconAlertTriangle, IconRefresh } from '@tabler/icons-react';
import {
  analyzeToolCallMarkup,
  TOOL_MARKUP_REPLY_NOTICE,
} from '../../../../main/src/client/tool-call-markup';
import { ThinkingMessageBody } from './ThinkingMessageBody';

const Notice = styled.div`
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid rgba(255, 196, 0, 0.35);
  background: rgba(255, 196, 0, 0.08);
  color: rgba(255, 255, 255, 0.85);
  font-size: 1.3rem;
  line-height: 1.45;

  svg {
    flex-shrink: 0;
    margin-top: 2px;
  }
`;

const RetryButton = styled.button.attrs({ type: 'button' })`
  display: inline-flex;
  align-items: center;
  gap: 5px;
  margin-top: 8px;
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
`;

/**
 * Renders an assistant reply, never showing raw tool-call markup as if it were
 * the answer. A model whose serving stack has no tool parser prints its chat
 * template's call syntax (`<tool_calls>…list_files…</tool_calls>`) as text;
 * Chat cannot run tools, so that is a failed turn, shown as one with a retry.
 * Prose around the markup is kept and the markup itself is dropped.
 */
export function AssistantReplyBody({
  text,
  onRetry,
  busy,
}: {
  text: string;
  onRetry?: () => void;
  busy?: boolean;
}) {
  const verdict = useMemo(() => analyzeToolCallMarkup(text), [text]);

  if (verdict.markupOnly) {
    return (
      <Notice role="alert" data-testid="tool-markup-notice">
        <IconAlertTriangle size={18} />
        <div>
          <div>{TOOL_MARKUP_REPLY_NOTICE}</div>
          {onRetry && (
            <RetryButton onClick={onRetry} disabled={busy}>
              <IconRefresh size={14} />
              Retry
            </RetryButton>
          )}
        </div>
      </Notice>
    );
  }

  return <ThinkingMessageBody text={verdict.found ? verdict.cleaned : text} />;
}
