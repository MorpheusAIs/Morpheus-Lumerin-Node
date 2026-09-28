import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import TermsAndConditions from './TermsAndConditions';

describe('bundled terms and conditions', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('renders the real legal Markdown immediately without fetch being available', () => {
    vi.stubGlobal('fetch', undefined);

    render(<TermsAndConditions />);

    expect(
      screen.getByRole('heading', { level: 1, name: 'Morpheus Terms of Use' }),
    ).toBeVisible();
    expect(
      screen.getByText(/Last Reviewed Date: February 26, 2025/),
    ).toBeVisible();
    expect(
      screen.getByText(/THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY/),
    ).toBeVisible();
  });

  it('does not request an asset URL when network access is blocked', () => {
    const blockedFetch = vi.fn(() => {
      throw new Error('Network and data-URL fetches are blocked');
    });
    vi.stubGlobal('fetch', blockedFetch);

    render(<TermsAndConditions />);

    expect(
      screen.getByText(/Permission is hereby granted, free of charge/),
    ).toBeVisible();
    expect(
      screen.getByText(/THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY/),
    ).toBeVisible();
    expect(blockedFetch).not.toHaveBeenCalled();
  });
});
