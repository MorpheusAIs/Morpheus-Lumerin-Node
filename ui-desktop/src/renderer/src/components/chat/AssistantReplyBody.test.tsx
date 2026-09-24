import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { FC, PropsWithChildren } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import theme from '../../ui/theme';
import { AssistantReplyBody } from './AssistantReplyBody';
import {
  LEAKED_PLACEHOLDER_REPLY,
  LEAKED_PLAN_UPDATE,
  LEAKED_TOOL_CALLS_BLOCK,
} from '../../../../main/src/client/tool-call-markup.fixtures';

const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

const renderReply = (text: string, onRetry?: () => void, busy?: boolean) =>
  render(
    <ThemeProvider theme={theme}>
      <AssistantReplyBody text={text} onRetry={onRetry} busy={busy} />
    </ThemeProvider>,
  );

describe('AssistantReplyBody', () => {
  it.each([
    ['<tool_calls> list_files block', LEAKED_TOOL_CALLS_BLOCK],
    ['plan-update markup', LEAKED_PLAN_UPDATE],
  ])('shows a retry state instead of the %s', async (_label, text) => {
    const onRetry = vi.fn();
    const { container } = renderReply(text, onRetry);

    expect(screen.getByTestId('tool-markup-notice')).toBeTruthy();
    expect(container.textContent).not.toContain('list_files');
    expect(container.textContent).not.toContain('update_plan_step');
    expect(container.textContent).not.toContain('"arguments"');

    await userEvent.click(screen.getByRole('button', { name: /retry/i }));
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it('keeps the prose and drops the markup from a mixed reply', () => {
    const { container } = renderReply(LEAKED_PLACEHOLDER_REPLY, vi.fn());

    expect(screen.queryByTestId('tool-markup-notice')).toBeNull();
    expect(container.textContent).toContain(
      "I'll start by looking at the project.",
    );
    expect(container.textContent).not.toContain('list_files');
    expect(container.textContent).not.toContain('in_progress');
  });

  it('disables retry while a reply is still streaming', () => {
    renderReply(LEAKED_TOOL_CALLS_BLOCK, vi.fn(), true);
    expect(
      (screen.getByRole('button', { name: /retry/i }) as HTMLButtonElement)
        .disabled,
    ).toBe(true);
  });

  it('renders an ordinary answer unchanged', () => {
    const { container } = renderReply(
      'Use `<tool_calls>` only when the parser is missing.',
    );
    expect(screen.queryByTestId('tool-markup-notice')).toBeNull();
    expect(container.textContent).toContain('<tool_calls>');
  });
});
