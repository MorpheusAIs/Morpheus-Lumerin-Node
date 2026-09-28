import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { FC, PropsWithChildren } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import theme from '../../ui/theme';
import { ThinkingMessageBody } from './ThinkingMessageBody';

// styled-components v4's own typings predate React 18 and do not satisfy the
// JSX element constraint; the rest of the suite casts the same way.
const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

const renderBody = (text: string) =>
  render(
    <ThemeProvider theme={theme}>
      <ThinkingMessageBody text={text} />
    </ThemeProvider>,
  );

describe('ThinkingMessageBody markdown', () => {
  beforeEach(() => vi.restoreAllMocks());

  // Pipe tables are the single most common way a model formats a comparison,
  // and without remark-gfm react-markdown renders them as a wall of pipes.
  it('renders a GFM pipe table as a real table', () => {
    renderBody(
      ['| Model | Price |', '| --- | --- |', '| llama2 | 0.4 |'].join('\n'),
    );

    const table = document.querySelector('table');
    expect(table).not.toBeNull();
    expect(screen.getByText('Model').tagName).toBe('TH');
    expect(screen.getByText('llama2').tagName).toBe('TD');
  });

  it('renders the other GFM constructs the old parser dropped', () => {
    renderBody('~~gone~~ and https://example.com');

    expect(document.querySelector('del')?.textContent).toBe('gone');
    // Autolink literals are GFM-only; plain CommonMark leaves them as text.
    expect(document.querySelector('a')?.getAttribute('href')).toBe(
      'https://example.com',
    );
  });

  it('gives a fenced code block a language label and a copy button', () => {
    renderBody('```bash\nyarn install\n```');

    expect(screen.getByText('bash')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Copy code' })).toBeTruthy();
  });

  it('copies the code and confirms, without the trailing newline', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText },
      configurable: true,
    });

    renderBody('```bash\nyarn install\n```');
    await userEvent.click(screen.getByRole('button', { name: 'Copy code' }));

    expect(writeText).toHaveBeenCalledWith('yarn install');
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'Copied' })).toBeTruthy(),
    );
  });

  // A clipboard rejection inside a click handler would otherwise surface as an
  // unhandled rejection and, in dev, an error overlay over the whole message.
  it('survives a clipboard that refuses', async () => {
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: vi.fn().mockRejectedValue(new Error('denied')) },
      configurable: true,
    });

    renderBody('```bash\nyarn install\n```');
    await userEvent.click(screen.getByRole('button', { name: 'Copy code' }));

    expect(screen.getByRole('button', { name: 'Copy code' })).toBeTruthy();
  });

  it('leaves inline code alone', () => {
    renderBody('run `yarn install` first');

    expect(screen.queryByRole('button', { name: 'Copy code' })).toBeNull();
    expect(document.querySelector('code')?.textContent).toBe('yarn install');
  });

  it('still renders tables inside a reasoning block', () => {
    renderBody('<think>\n| a |\n| --- |\n| b |\n</think>done');

    expect(document.querySelector('table')).not.toBeNull();
    expect(screen.getByText('done')).toBeTruthy();
  });
});
