import { beforeEach, describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { FC, PropsWithChildren } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import theme from '../../ui/theme';
import { CustomInstructions } from './CustomInstructions';
import {
  CUSTOM_INSTRUCTIONS_KEY,
  CUSTOM_INSTRUCTIONS_MAX_LENGTH,
} from '../../lib/customInstructions';

// styled-components v4's own typings predate React 18 and do not satisfy the
// JSX element constraint; the rest of the suite casts the same way.
const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

const renderPage = () =>
  render(
    <ThemeProvider theme={theme}>
      <CustomInstructions />
    </ThemeProvider>,
  );

const editor = () =>
  screen.getByLabelText('Your instructions') as HTMLTextAreaElement;

describe('CustomInstructions', () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it('starts empty when nothing has been saved', () => {
    renderPage();
    expect(editor().value).toBe('');
    expect(screen.getByText('Save instructions')).toBeDisabled();
  });

  it('loads the stored instruction on mount', () => {
    window.localStorage.setItem(CUSTOM_INSTRUCTIONS_KEY, 'Be concise.');
    renderPage();
    expect(editor().value).toBe('Be concise.');
  });

  it('enables save only once the text differs from what is stored', async () => {
    const user = userEvent.setup();
    window.localStorage.setItem(CUSTOM_INSTRUCTIONS_KEY, 'Be concise.');
    renderPage();
    const save = screen.getByText('Save instructions');
    expect(save).toBeDisabled();
    await user.type(editor(), ' Always.');
    expect(save).toBeEnabled();
  });

  it('persists the trimmed value and confirms the save', async () => {
    const user = userEvent.setup();
    renderPage();
    await user.type(editor(), '  Use metric units.  ');
    await user.click(screen.getByText('Save instructions'));
    expect(window.localStorage.getItem(CUSTOM_INSTRUCTIONS_KEY)).toBe(
      'Use metric units.',
    );
    expect(screen.getByRole('status')).toHaveTextContent('Saved');
    expect(screen.getByText('Save instructions')).toBeDisabled();
  });

  it('clears both the box and storage', async () => {
    const user = userEvent.setup();
    window.localStorage.setItem(CUSTOM_INSTRUCTIONS_KEY, 'Be concise.');
    renderPage();
    await user.click(screen.getByText('Clear'));
    expect(editor().value).toBe('');
    expect(window.localStorage.getItem(CUSTOM_INSTRUCTIONS_KEY)).toBeNull();
  });

  it('disables clear when there is nothing to clear', () => {
    renderPage();
    expect(screen.getByText('Clear')).toBeDisabled();
  });

  it('blocks saving past the length limit', async () => {
    const user = userEvent.setup();
    renderPage();
    const box = editor();
    // Typing 4000+ characters one keystroke at a time is far too slow, so the
    // value is set directly and the change event fired as the browser would.
    await user.click(box);
    const long = 'x'.repeat(CUSTOM_INSTRUCTIONS_MAX_LENGTH + 1);
    await user.paste(long);
    expect(
      screen.getByText(
        `${long.length} / ${CUSTOM_INSTRUCTIONS_MAX_LENGTH}`,
      ),
    ).toBeInTheDocument();
    expect(screen.getByText('Save instructions')).toBeDisabled();
  });
});
