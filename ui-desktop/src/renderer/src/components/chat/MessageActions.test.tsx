import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { FC, PropsWithChildren } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import theme from '../../ui/theme';
import { MessageActions, precedingUserText } from './MessageActions';

// styled-components v4's own typings predate React 18 and do not satisfy the
// JSX element constraint; the rest of the suite casts the same way.
const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

const renderActions = (props: Parameters<typeof MessageActions>[0]) =>
  render(
    <ThemeProvider theme={theme}>
      <MessageActions {...props} />
    </ThemeProvider>,
  );

describe('precedingUserText', () => {
  const messages = [
    { role: 'user', text: 'first question' },
    { role: 'assistant', text: 'first answer' },
    { role: 'user', text: 'second question' },
    { role: 'assistant', text: 'second answer' },
  ];

  it('finds the question directly above an answer', () => {
    expect(precedingUserText(messages, 3)).toBe('second question');
    expect(precedingUserText(messages, 1)).toBe('first question');
  });

  it('skips back past non-user turns', () => {
    const withTool = [
      { role: 'user', text: 'do the thing' },
      { role: 'assistant', text: 'calling a tool' },
      { role: 'assistant', text: 'done' },
    ];
    expect(precedingUserText(withTool, 2)).toBe('do the thing');
  });

  it('skips user turns that carry no resendable text', () => {
    const withImage = [
      { role: 'user', text: 'describe this' },
      { role: 'user', text: '   ' },
      { role: 'assistant', text: 'a cat' },
    ];
    expect(precedingUserText(withImage, 2)).toBe('describe this');
  });

  it('is undefined when nothing precedes the message', () => {
    expect(precedingUserText(messages, 0)).toBeUndefined();
    expect(precedingUserText([{ role: 'assistant', text: 'hi' }], 1)).toBe(
      undefined,
    );
  });

  it('tolerates malformed input', () => {
    expect(precedingUserText(undefined, 2)).toBeUndefined();
    expect(precedingUserText(null, 0)).toBeUndefined();
    expect(precedingUserText([null, undefined], 2)).toBeUndefined();
  });

  it('does not run off the end of a shorter list', () => {
    expect(precedingUserText(messages, 99)).toBe('second question');
  });
});

describe('MessageActions', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders nothing when there is no text and no action', () => {
    const { container } = renderActions({ text: '   ' });
    expect(container).toBeEmptyDOMElement();
  });

  it('copies the message text and confirms', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText },
      configurable: true,
    });

    renderActions({ text: 'the answer' });
    await userEvent.click(screen.getByLabelText('Copy message'));

    expect(writeText).toHaveBeenCalledWith('the answer');
    expect(await screen.findByText('Copied')).toBeInTheDocument();
  });

  // A rejection inside the click handler would otherwise surface as an
  // unhandled rejection, and the label must not claim a copy that never landed.
  it('does not claim a copy the clipboard refused', async () => {
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: vi.fn().mockRejectedValue(new Error('denied')) },
      configurable: true,
    });

    renderActions({ text: 'the answer' });
    await userEvent.click(screen.getByLabelText('Copy message'));

    expect(screen.queryByText('Copied')).not.toBeInTheDocument();
  });

  it('shows edit only when an edit handler is given', async () => {
    const onEdit = vi.fn();

    const { rerender } = renderActions({ text: 'hi' });
    expect(screen.queryByLabelText('Edit and resend')).not.toBeInTheDocument();

    rerender(
      <ThemeProvider theme={theme}>
        <MessageActions text="hi" onEdit={onEdit} />
      </ThemeProvider>,
    );
    await userEvent.click(screen.getByLabelText('Edit and resend'));
    expect(onEdit).toHaveBeenCalledTimes(1);
  });

  it('shows regenerate only when a regenerate handler is given', async () => {
    const onRegenerate = vi.fn();

    renderActions({ text: 'hi', onRegenerate });
    await userEvent.click(screen.getByLabelText('Regenerate response'));
    expect(onRegenerate).toHaveBeenCalledTimes(1);
  });

  it('disables the resending actions while a request is in flight', () => {
    renderActions({
      text: 'hi',
      onEdit: vi.fn(),
      onRegenerate: vi.fn(),
      busy: true,
    });
    expect(screen.getByLabelText('Edit and resend')).toBeDisabled();
    expect(screen.getByLabelText('Regenerate response')).toBeDisabled();
    // Copy never touches the network, so it stays available.
    expect(screen.getByLabelText('Copy message')).toBeEnabled();
  });

  it('offers actions on a media message that carries no text', () => {
    renderActions({ text: '', onRegenerate: vi.fn() });
    expect(screen.queryByLabelText('Copy message')).not.toBeInTheDocument();
    expect(screen.getByLabelText('Regenerate response')).toBeInTheDocument();
  });
});

describe('usage on a message', () => {
  const usage = {
    completionTokens: 318,
    promptTokens: 1_204,
    totalTokens: 1_522,
  };

  it('shows tokens and cost under the answer', () => {
    renderActions({ text: 'the answer', usage, costMor: 0.002 });
    expect(screen.getByTestId('message-usage')).toHaveTextContent(
      '1,204 in · 318 out · 0.002 MOR',
    );
  });

  it('shows tokens alone when the price is unknown', () => {
    // A local model has no bid, so quoting a MOR figure would be an invention.
    renderActions({ text: 'the answer', usage });
    expect(screen.getByTestId('message-usage')).toHaveTextContent(
      '1,204 in · 318 out',
    );
    expect(screen.getByTestId('message-usage')).not.toHaveTextContent('MOR');
  });

  it('shows nothing when the stream reported no usage', () => {
    renderActions({ text: 'the answer' });
    expect(screen.queryByTestId('message-usage')).not.toBeInTheDocument();
  });

  it('renders usage even on a turn with no text and no actions', () => {
    // Usage is the whole reason the row exists in that case, so the early
    // return that hides an empty row must not hide this too.
    renderActions({ text: '', costMor: 0.002 });
    expect(screen.getByTestId('message-usage')).toHaveTextContent('0.002 MOR');
  });

  it('keeps usage visible rather than hiding it with the hover actions', () => {
    renderActions({ text: 'the answer', usage, onRegenerate: vi.fn() });
    // The buttons live in their own faded wrapper; usage is a sibling of it,
    // not a child, so it is never affected by the hover rule.
    const row = screen.getByTestId('message-actions');
    const usageNode = screen.getByTestId('message-usage');
    expect(usageNode.parentElement).toBe(row);
    expect(row).toContainElement(screen.getByLabelText('Regenerate response'));
    expect(usageNode).not.toContainElement(
      screen.getByLabelText('Regenerate response'),
    );
  });
});
