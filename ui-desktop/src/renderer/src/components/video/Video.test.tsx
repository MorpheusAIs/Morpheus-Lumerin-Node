import { describe, expect, it, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { FC, PropsWithChildren } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import theme from '../../ui/theme';
import { Video } from './Video';

const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

const MODEL = '0x' + '1'.repeat(64);
const PROVIDER = '0x' + 'a'.repeat(40);
const BID = '0x' + 'b'.repeat(64);
const SESSION = '0x' + 'c'.repeat(64);

const guide = {
  object: 'video.guide',
  defaults: {
    duration_seconds: 5,
    aspect_ratio: '16:9',
    resolution: '720p',
    generate_audio: false,
  },
  fields: [
    { name: 'prompt', required: true },
    { name: 'duration_seconds', minimum: 4, maximum: 12 },
    { name: 'aspect_ratio', enum: ['16:9', '9:16'] },
    { name: 'resolution', enum: ['480p', '720p'] },
  ],
};

const renderVideo = (sessions: any[] | Promise<any[]>) => {
  const client = {
    getVideoJobs: vi.fn().mockResolvedValue([]),
    getVideoGuide: vi.fn().mockResolvedValue({
      offers: [{ provider: PROVIDER, bidId: BID, guide }],
    }),
    quoteVideoJob: vi.fn().mockResolvedValue([
      {
        provider: PROVIDER,
        bid_id: BID,
        price_per_second_wei: '1000000000000',
        session_seconds: 844,
        session_cost_wei: '844000000000000',
      },
    ]),
    createVideoJob: vi.fn().mockResolvedValue({ id: 'job-1' }),
  };
  const props = {
    client,
    address: '0x' + 'e'.repeat(40),
    toasts: { toast: vi.fn() },
    getAllModels: vi
      .fn()
      .mockResolvedValue([
        { Id: MODEL, Name: 'Seedance', ModelType: 'VIDEO', Tags: ['video'] },
      ]),
    getSessionsByUser: vi.fn().mockReturnValue(Promise.resolve(sessions)),
    onOpenSession: vi.fn().mockResolvedValue(SESSION),
    closeSession: vi.fn(),
  };
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={queryClient}>
      <ThemeProvider theme={theme}>
        <Video {...props} />
      </ThemeProvider>
    </QueryClientProvider>,
  );
  return props;
};

// The provider refuses a job its session cannot pay for, so the session is
// opened for exactly the quoted length, on the provider whose quote was taken.
describe('Video', () => {
  it('opens a session for the quoted length on the quoted bid, then submits the job on it', async () => {
    const user = userEvent.setup();
    const props = renderVideo([]);

    await user.type(await screen.findByLabelText('Prompt'), 'A fox in snow');
    await screen.findByText('Session 14 min 4 s');
    await user.click(screen.getByText('Open session & generate'));

    await waitFor(() =>
      expect(props.client.createVideoJob).toHaveBeenCalledWith({
        sessionId: SESSION,
        job: {
          prompt: 'A fox in snow',
          duration_seconds: 5,
          aspect_ratio: '16:9',
          resolution: '720p',
        },
      }),
    );
    expect(props.onOpenSession).toHaveBeenCalledWith({
      modelId: MODEL,
      duration: 844,
      bidId: BID,
      provider: PROVIDER,
    });
  });

  it('does not submit on an open session that has no time left for the clip', async () => {
    const user = userEvent.setup();
    const now = Math.floor(Date.now() / 1000);
    const props = renderVideo([
      {
        Id: SESSION,
        ModelAgentId: MODEL,
        Provider: PROVIDER,
        OpenedAt: now - 100,
        EndsAt: now + 200,
        ClosedAt: 0,
      },
    ]);

    await user.type(await screen.findByLabelText('Prompt'), 'A fox in snow');
    await screen.findByText(/no room for another one/);
    // Closing is the next step: there is no button to submit on it.
    expect(screen.queryByText('Generate in this session')).toBeNull();
    expect(props.client.createVideoJob).not.toHaveBeenCalled();
    expect(props.onOpenSession).not.toHaveBeenCalled();
  });

  it('does not offer to open a session before it knows whether one is open', async () => {
    const user = userEvent.setup();
    const props = renderVideo(new Promise<any[]>(() => {}));

    await user.type(await screen.findByLabelText('Prompt'), 'A fox in snow');
    await screen.findByText('Checking your open sessions…');
    expect(screen.getByText('Open session & generate')).toBeDisabled();
    expect(props.onOpenSession).not.toHaveBeenCalled();
  });
});
