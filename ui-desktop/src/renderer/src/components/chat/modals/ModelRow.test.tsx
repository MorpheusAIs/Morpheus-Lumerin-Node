import { render, screen } from '@testing-library/react';
import type { FC, PropsWithChildren } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import theme from '../../../ui/theme';
import ModelRow from './ModelRow';

const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

describe('ModelRow provider discovery state', () => {
  it('shows provider discovery as loading rather than falsely unavailable', () => {
    const { rerender } = render(
      <ThemeProvider theme={theme}>
        <ModelRow
          bidsLoading
          model={{ Id: 'model-1', Name: 'Model one', Tags: ['llm'] }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    const row = screen.getByRole('button', { name: /Model one/ });
    expect(row.getAttribute('aria-busy')).toBe('true');
    expect((row as HTMLButtonElement).disabled).toBe(true);
    expect(screen.getByText('Checking providers…')).toBeTruthy();
    expect(screen.queryByText('Unavailable')).toBeNull();

    rerender(
      <ThemeProvider theme={theme}>
        <ModelRow
          bidsLoading={false}
          model={{ Id: 'model-1', Name: 'Model one', Tags: ['llm'] }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    expect(screen.getByText('Unavailable')).toBeTruthy();
  });
});
