import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { IconDownload, IconPhoto, IconPlayerPlay } from '@tabler/icons-react';
import withChatState from '../../store/hocs/withChatState';
import { queryKeys } from '../../store/queries';
import { abbreviateAddress, smartRound } from '../../utils';
import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';
import { Btn } from '../common/Btn';
import QueryError from '../common/QueryError';
import * as S from './Video.styles';
import {
  STATUS_LABELS,
  VideoGuideOffer,
  VideoJob,
  VideoQuote,
  cheapestQuote,
  formatClock,
  formatSeconds,
  guideOffer,
  isTerminalStatus,
  jobErrorText,
  jobSeconds,
  linkExpired,
  openModelSessions,
  sameId,
  selectVideoModels,
  sessionBudget,
  usableQuotes,
  videoForm,
} from './videoJobs';

const MAX_PROMPT_BYTES = 4000;
const MAX_FRAME_BYTES = 10 * 1024 * 1024;
const FRAME_TYPES = ['image/jpeg', 'image/png', 'image/webp'];
// The prompt does not change the price, so prices show before it is written.
const QUOTE_PROMPT = 'quote';
const ACTIVE_JOB_POLL_MS = 4000;
// After a submit that timed out, look again for the job the node may still be
// sending.
const LATE_JOB_RECHECK_MS = [15_000, 60_000];

const errorText = (error: any, fallback: string) =>
  (typeof error?.message === 'string' && error.message) || fallback;

// What sendToMainProcess rejects with when its own timer fires. A node error
// that merely says "timed out" means the clip was refused, so only this counts.
const IPC_TIMEOUT_MESSAGE = 'Operation timed out. Please try again later.';
const timedOut = (error: any) => errorText(error, '') === IPC_TIMEOUT_MESSAGE;

const useNow = (intervalMs: number) => {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), intervalMs);
    return () => clearInterval(timer);
  }, [intervalMs]);
  return now;
};

/** Seconds left until `until` (unix seconds), ticking on screen. */
const Countdown = ({ until }: { until: number }) => {
  const now = useNow(1000);
  return <>{formatClock(until - now / 1000)}</>;
};

/** Time since `since`, ticking on screen. */
const Elapsed = ({ since }: { since: string }) => {
  const now = useNow(1000);
  return <>{formatClock((now - Date.parse(since)) / 1000)}</>;
};

const readFrame = (file: File) =>
  new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result));
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });

const durationOptions = (min: number, max: number) => {
  const options: number[] = [];
  for (let s = min; s <= max && options.length < 120; s += 1) options.push(s);
  return options;
};

// One clip crosses IPC at a time: every call in flight on a channel receives
// every answer, so parallel loads would copy each clip into each listener.
let clipQueue: Promise<unknown> = Promise.resolve();
const loadClipQueued = (
  client: any,
  id: string,
  cancelled: () => boolean,
): Promise<{ data: ArrayBuffer; mimeType: string }> => {
  // A load whose card is gone by its turn is skipped, not downloaded.
  const run = clipQueue.then(() =>
    cancelled()
      ? Promise.reject(new Error('The clip load was cancelled.'))
      : client.loadVideoClip({ id }),
  );
  clipQueue = run.catch(() => undefined);
  return run;
};

type JobCardProps = {
  job: VideoJob;
  modelName: string;
  client: any;
  toasts: any;
  now: number;
};

const VideoJobCard = ({
  job,
  modelName,
  client,
  toasts,
  now,
}: JobCardProps) => {
  const [clipUrl, setClipUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const lastStatus = useRef(job.status);
  const mounted = useRef(true);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  useEffect(
    () => () => {
      if (clipUrl) URL.revokeObjectURL(clipUrl);
    },
    [clipUrl],
  );

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const clip = await loadClipQueued(client, job.id, () => !mounted.current);
      // A blob made after the card is gone would never be revoked.
      if (!mounted.current) return;
      setClipUrl(
        URL.createObjectURL(new Blob([clip.data], { type: clip.mimeType })),
      );
    } catch (error) {
      if (mounted.current) {
        toasts.toast(
          'error',
          errorText(error, 'The clip could not be loaded.'),
        );
      }
    } finally {
      if (mounted.current) setLoading(false);
    }
  }, [client, job.id, toasts]);

  // A clip that finishes while the screen is open loads without a click.
  useEffect(() => {
    if (lastStatus.current !== job.status && job.status === 'succeeded') {
      void load();
    }
    lastStatus.current = job.status;
  }, [job.status, load]);

  const save = async () => {
    setSaving(true);
    try {
      const result = await client.saveVideoClip({ id: job.id });
      if (result?.saved) toasts.toast('success', 'Video saved.');
    } catch (error) {
      toasts.toast('error', errorText(error, 'The video was not saved.'));
    } finally {
      if (mounted.current) setSaving(false);
    }
  };

  const resolved = job.resolved;
  // The node fills in the job's settings from the provider's receipt.
  const confirmed = !!resolved?.duration_seconds;
  const expired = linkExpired(job, now);
  const problem = jobErrorText(job);

  return (
    <S.JobCard>
      <S.JobHeader>
        <p>
          {confirmed
            ? resolved.prompt
            : isTerminalStatus(job.status)
              ? 'The provider did not accept this clip.'
              : 'Waiting for the provider to confirm this clip.'}
        </p>
        <S.StatusBadge $status={job.status}>
          {!isTerminalStatus(job.status) && <S.Spinner aria-hidden="true" />}
          {STATUS_LABELS[job.status] ?? job.status}
          {!isTerminalStatus(job.status) && (
            <span>
              <Elapsed since={job.created_at} />
            </span>
          )}
        </S.StatusBadge>
      </S.JobHeader>
      <S.JobMeta>
        <span>{modelName}</span>
        {confirmed && (
          <span>
            {resolved.duration_seconds} s · {resolved.resolution} ·{' '}
            {resolved.aspect_ratio}
            {resolved.generate_audio ? ' · sound' : ''}
            {resolved.has_image ? ' · first frame' : ''}
          </span>
        )}
        <span>{new Date(job.created_at).toLocaleString()}</span>
        <span>
          Session <S.Mono>{abbreviateAddress(job.session_id, 4)}</S.Mono>
        </span>
      </S.JobMeta>
      {problem && <S.Hint $tone="error">{problem}</S.Hint>}
      {clipUrl && <S.Player src={clipUrl} controls />}
      {job.status === 'succeeded' && (
        <S.Actions>
          {!clipUrl && (
            <S.SecondaryBtn onClick={load} disabled={loading || expired}>
              <IconPlayerPlay size={16} aria-hidden="true" />
              {loading ? 'Loading…' : 'Play'}
            </S.SecondaryBtn>
          )}
          <S.SecondaryBtn onClick={save} disabled={saving || expired}>
            <IconDownload size={16} aria-hidden="true" />
            {saving ? 'Saving…' : 'Save…'}
          </S.SecondaryBtn>
          {job.expires_at && (
            <S.Hint>
              {expired
                ? 'The vendor link has expired.'
                : `The vendor link works until ${new Date(job.expires_at).toLocaleString()}.`}
            </S.Hint>
          )}
        </S.Actions>
      )}
    </S.JobCard>
  );
};

export const Video = (props: any) => {
  const { client, address, toasts } = props;
  const queryClient = useQueryClient();
  const now = useNow(15_000);
  const actionInFlight = useRef(false);
  const fileInput = useRef<HTMLInputElement | null>(null);
  const formSeeded = useRef(false);

  const modelsQuery = useQuery({
    queryKey: queryKeys.allModels,
    queryFn: () => props.getAllModels(),
    select: selectVideoModels,
  });
  const models = modelsQuery.data ?? [];
  const [modelId, setModelId] = useState('');
  const model = models.find((m) => m.Id === modelId) ?? models[0];
  const activeModelId: string = model?.Id ?? '';
  const modelNames = useMemo(
    () => new Map(models.map((m) => [String(m.Id).toLowerCase(), m.Name])),
    [models],
  );

  const sessionsQuery = useQuery({
    queryKey: queryKeys.sessions(address),
    queryFn: () => props.getSessionsByUser(address),
    enabled: !!address,
  });

  const jobsQuery = useQuery<VideoJob[]>({
    queryKey: queryKeys.videoJobs,
    queryFn: () => client.getVideoJobs(),
    refetchInterval: (query) =>
      (query.state.data ?? []).some((job) => !isTerminalStatus(job.status))
        ? ACTIVE_JOB_POLL_MS
        : false,
  });
  const jobs = jobsQuery.data ?? [];
  // Until both lists are in, the page cannot tell an open session from none,
  // or how much of a session its clips already use.
  const sessionsKnown = sessionsQuery.data !== undefined;
  const jobsKnown = jobsQuery.data !== undefined;

  // Clips go to the newest open session; older ones can only be closed here,
  // since Chat does not list video models.
  const [session, ...otherSessions] = openModelSessions(
    sessionsQuery.data ?? [],
    activeModelId,
  );

  const guideQuery = useQuery<{ offers: VideoGuideOffer[] }>({
    queryKey: queryKeys.videoGuide(address, activeModelId),
    queryFn: () => client.getVideoGuide({ modelId: activeModelId }),
    enabled: !!activeModelId,
    staleTime: 60_000,
  });
  const offers = guideQuery.data?.offers ?? [];
  // Clips on an open session go to its provider, so its guide sets the form.
  const sessionOffer = session
    ? offers.find((o) => o.guide && sameId(o.provider, session.Provider))
    : undefined;
  const guide = (sessionOffer ?? guideOffer(offers, address))?.guide;
  const form = useMemo(() => videoForm(guide), [guide]);

  const [prompt, setPrompt] = useState('');
  const [duration, setDuration] = useState(0);
  const [aspectRatio, setAspectRatio] = useState('');
  const [resolution, setResolution] = useState('');
  const [audio, setAudio] = useState(false);
  const [frame, setFrame] = useState<{ name: string; dataUri: string } | null>(
    null,
  );
  const [pickedBid, setPickedBid] = useState('');
  const [busy, setBusy] = useState<string | null>(null);

  // The first guide sets the defaults. A later one, from another model or a
  // refetch that reached another provider, keeps every choice it still allows.
  useEffect(() => {
    if (!form) return;
    const seeded = formSeeded.current;
    formSeeded.current = true;
    const { defaults } = form;
    setDuration((d) =>
      seeded && d >= form.minDuration && d <= form.maxDuration
        ? d
        : defaults.duration_seconds,
    );
    setAspectRatio((r) =>
      seeded && form.aspectRatios.includes(r) ? r : defaults.aspect_ratio,
    );
    setResolution((r) =>
      seeded && form.resolutions.includes(r) ? r : defaults.resolution,
    );
    setAudio((a) => form.audio && (seeded ? a : defaults.generate_audio));
    if (!form.image) setFrame(null);
  }, [form]);

  const settings =
    form && duration && aspectRatio && resolution
      ? {
          duration_seconds: duration,
          aspect_ratio: aspectRatio,
          resolution,
          ...(form.audio ? { generate_audio: audio } : {}),
        }
      : null;

  const quoteQuery = useQuery<VideoQuote[]>({
    queryKey: queryKeys.videoQuote(
      activeModelId,
      JSON.stringify({ ...settings, image: !!frame }),
    ),
    queryFn: () =>
      client.quoteVideoJob({
        modelId: activeModelId,
        job: { prompt: QUOTE_PROMPT, ...settings },
        hasImage: !!frame,
      }),
    enabled: !!activeModelId && !!settings,
    staleTime: 30_000,
  });
  const quotes = (quoteQuery.data ?? []).filter(
    (q) => !sameId(q.provider, address),
  );

  const budget = session ? sessionBudget(session, jobs) : null;
  const sessionQuote = session
    ? quotes.find((q) => sameId(q.provider, session.Provider))
    : undefined;
  const neededSeconds = jobSeconds(sessionQuote);
  const sessionFits =
    jobsKnown &&
    !!budget &&
    !budget.pending &&
    !!neededSeconds &&
    neededSeconds <= budget.left;
  const sessionRendering =
    !!session &&
    jobs.some(
      (job) =>
        sameId(job.session_id, session.Id) && !isTerminalStatus(job.status),
    );
  // The session's clips took all its time and none is still rendering: closing
  // it is the next step, not generating.
  const closeIsNext =
    jobsKnown &&
    !!budget &&
    !budget.pending &&
    !!neededSeconds &&
    !sessionFits &&
    !sessionRendering;

  const picked =
    usableQuotes(quotes, address).find((q) => q.bid_id === pickedBid) ??
    cheapestQuote(quotes, address);

  const trimmedPrompt = prompt.trim();
  const promptBytes = new TextEncoder().encode(trimmedPrompt).length;
  const promptOk = promptBytes > 0 && promptBytes <= MAX_PROMPT_BYTES;
  const canGenerate =
    !!address &&
    sessionsKnown &&
    !!settings &&
    promptOk &&
    !busy &&
    (session ? sessionFits : !!picked);

  const refreshSessions = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions(address) }),
      queryClient.invalidateQueries({ queryKey: queryKeys.balances(address) }),
    ]);

  const refreshJobs = () =>
    queryClient.invalidateQueries({ queryKey: queryKeys.videoJobs });

  const submit = async (sessionId: string) => {
    setBusy('Submitting the clip…');
    try {
      await client.createVideoJob({
        sessionId,
        job: {
          prompt: trimmedPrompt,
          ...settings,
          ...(frame ? { image_b64: frame.dataUri } : {}),
        },
      });
      toasts.toast(
        'success',
        'Clip submitted. Rendering usually takes a few minutes.',
      );
    } catch (error) {
      if (timedOut(error)) {
        // The node may still be sending it: it keeps a job the provider may
        // have received and sends it again with the same id.
        toasts.toast(
          'info',
          'The node is still sending the clip. It appears under Clips if the provider took it.',
          { autoClose: 12000 },
        );
        for (const delay of LATE_JOB_RECHECK_MS) {
          setTimeout(() => void refreshJobs(), delay);
        }
      } else {
        toasts.toast('error', errorText(error, 'The clip was not submitted.'), {
          autoClose: 12000,
        });
      }
    } finally {
      // The session's budget includes the new job before the next click.
      await refreshJobs();
    }
  };

  const generate = async () => {
    if (!canGenerate || actionInFlight.current) return;
    actionInFlight.current = true;
    try {
      if (session) {
        await submit(session.Id);
        return;
      }
      setBusy('Opening the session…');
      // The session is opened for exactly the quoted length, on the provider
      // whose quote was picked: the provider refuses a job its session cannot
      // pay for.
      const opened = await props.onOpenSession({
        modelId: activeModelId,
        duration: picked!.session_seconds,
        bidId: picked!.bid_id,
        provider: picked!.provider,
      });
      if (typeof opened === 'string') {
        await submit(opened);
      } else if (opened?.existingSessionID) {
        // The session list was stale: one was already open, and nothing was
        // sent. It shows above once the list reloads.
        toasts.toast(
          'info',
          'The clip was not sent. Your open session for this model is shown above: generate again to use it.',
          { autoClose: 12000 },
        );
      }
      await refreshSessions();
    } finally {
      actionInFlight.current = false;
      setBusy(null);
    }
  };

  const close = async (sessionId: string) => {
    if (actionInFlight.current) return;
    actionInFlight.current = true;
    setBusy('Closing the session…');
    try {
      await props.closeSession(sessionId);
    } finally {
      await refreshSessions();
      actionInFlight.current = false;
      setBusy(null);
    }
  };

  const pickFrame = async (file?: File | null) => {
    if (!file) return;
    if (!FRAME_TYPES.includes(file.type)) {
      toasts.toast(
        'error',
        'The first frame must be a JPEG, PNG or WebP image.',
      );
      return;
    }
    if (file.size > MAX_FRAME_BYTES) {
      toasts.toast('error', 'The first frame must be 10 MB or smaller.');
      return;
    }
    try {
      setFrame({ name: file.name, dataUri: await readFrame(file) });
    } catch {
      toasts.toast('error', 'The image could not be read.');
    }
  };

  const renderSession = (s: any, title: string, closeButton = true) => {
    const rowBudget = sessionBudget(s, jobs);
    const rendering = jobs.some(
      (job) => sameId(job.session_id, s.Id) && !isTerminalStatus(job.status),
    );
    return (
      <div key={s.Id}>
        <S.SessionBox>
          <div>
            <strong>{title}</strong>
            <small>
              <S.Mono>{abbreviateAddress(s.Id, 6)}</S.Mono> · provider{' '}
              <S.Mono>{abbreviateAddress(s.Provider, 4)}</S.Mono>
            </small>
            <small>
              Ends in <Countdown until={Number(s.EndsAt)} />
            </small>
            <small>
              Clip time reserved: {formatSeconds(rowBudget.used)} of{' '}
              {formatSeconds(rowBudget.length)}
            </small>
          </div>
          {closeButton && (
            <S.SecondaryBtn onClick={() => close(s.Id)} disabled={!!busy}>
              Close session
            </S.SecondaryBtn>
          )}
        </S.SessionBox>
        {rendering && (
          <S.Hint $tone="warning">
            A clip is still rendering on this session. Closing it now loses that
            clip.
          </S.Hint>
        )}
      </div>
    );
  };

  const renderSessionPrice = () => {
    if (!budget) return null;
    if (quoteQuery.error && !sessionQuote) {
      return (
        <QueryError
          error={quoteQuery.error}
          what="the price"
          onRetry={() => quoteQuery.refetch()}
        />
      );
    }
    if (quoteQuery.isFetching && !sessionQuote) {
      return <S.Hint>Getting the price…</S.Hint>;
    }
    if (!sessionQuote) {
      return (
        <S.Hint $tone="error">
          The session's provider did not price this clip.
        </S.Hint>
      );
    }
    if (sessionQuote.error) {
      return <S.Hint $tone="error">{sessionQuote.error}</S.Hint>;
    }
    if (!jobsKnown) {
      return jobsQuery.error ? (
        <S.Hint $tone="error">
          The clips could not be loaded, so the session's time left is unknown.
        </S.Hint>
      ) : (
        <S.Hint>Checking this session's clips…</S.Hint>
      );
    }
    if (budget.pending) {
      return (
        <S.Hint $tone="warning">
          Waiting for the provider to confirm the previous clip before another
          one goes on this session.
        </S.Hint>
      );
    }
    if (sessionFits) {
      return (
        <S.Hint>
          This clip reserves {formatSeconds(neededSeconds!)} of the session.
        </S.Hint>
      );
    }
    // Not a fault. A clip reserves its whole quote however fast it renders,
    // so a session can still be running with no room for another clip.
    return sessionRendering ? (
      <S.Hint>
        The clip on this session reserves the rest of its clip time, so the next
        clip needs a new session. When the clip is ready, close this one: there
        is no need to wait for it to end.
      </S.Hint>
    ) : (
      <S.Hint>
        Every clip reserves its full quote, even when it renders faster, so this
        session has no room for another one
        {budget.left > 0
          ? ` (${formatSeconds(budget.left)} free, this clip needs ${formatSeconds(neededSeconds!)})`
          : ''}
        . Close it now to open a new session for the next clip: there is no need
        to wait for it to end.
      </S.Hint>
    );
  };

  const renderPrices = () => {
    if (!sessionsKnown) {
      return sessionsQuery.error ? (
        <QueryError
          error={sessionsQuery.error}
          what="your sessions"
          onRetry={() => sessionsQuery.refetch()}
        />
      ) : (
        <S.Hint>Checking your open sessions…</S.Hint>
      );
    }
    if (session) {
      return (
        <>
          {renderSession(session, 'Open session', !closeIsNext)}
          {renderSessionPrice()}
          {otherSessions.map((s) => renderSession(s, 'Other open session'))}
        </>
      );
    }

    return (
      <>
        <S.Hint>Session cost at each provider's bid price</S.Hint>
        {quoteQuery.error ? (
          <QueryError
            error={quoteQuery.error}
            what="prices"
            onRetry={() => quoteQuery.refetch()}
          />
        ) : quoteQuery.isPending ? (
          <S.Hint>Getting prices…</S.Hint>
        ) : !quotes.length ? (
          <S.Hint $tone="warning">No provider priced this clip.</S.Hint>
        ) : (
          <S.QuoteList role="radiogroup" aria-label="Provider">
            {quotes.map((q) => {
              const ok =
                !q.error && !!q.session_seconds && !!q.session_cost_wei;
              return (
                <S.QuoteRow key={q.bid_id} $disabled={!ok}>
                  <input
                    type="radio"
                    name="video-quote"
                    disabled={!ok}
                    checked={ok && picked?.bid_id === q.bid_id}
                    onChange={() => setPickedBid(q.bid_id)}
                  />
                  <span>
                    <S.Mono>{abbreviateAddress(q.provider, 4)}</S.Mono>
                    <small>
                      {ok
                        ? `Session ${formatSeconds(q.session_seconds!)}`
                        : q.error || 'No price'}
                    </small>
                  </span>
                  {ok && <strong>{smartRound(q.session_cost_wei)} MOR</strong>}
                </S.QuoteRow>
              );
            })}
          </S.QuoteList>
        )}
      </>
    );
  };

  const renderForm = () => {
    if (guideQuery.isPending) {
      return <S.Hint>Asking this model's providers for its limits…</S.Hint>;
    }
    if (guideQuery.error) {
      return (
        <QueryError
          error={guideQuery.error}
          what="the model's limits"
          onRetry={() => guideQuery.refetch()}
        />
      );
    }
    if (!form) {
      return (
        <>
          <S.Hint $tone="warning">
            No provider of this model published its video limits. They may be
            offline or run an older version.
          </S.Hint>
          <S.Actions>
            <S.SecondaryBtn onClick={() => guideQuery.refetch()}>
              Ask again
            </S.SecondaryBtn>
          </S.Actions>
        </>
      );
    }

    return (
      <>
        {/* The job sent is the one shown: nothing changes while it is sent. */}
        <S.Fieldset disabled={!!busy}>
          <S.Field>
            Prompt
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="A red fox running through fresh snow at dawn, slow motion"
            />
          </S.Field>
          {promptBytes > MAX_PROMPT_BYTES && (
            <S.Hint $tone="error">
              The prompt is {promptBytes} bytes; the limit is {MAX_PROMPT_BYTES}
              .
            </S.Hint>
          )}
          <S.FieldRow>
            <S.Field>
              Length
              <select
                value={duration}
                onChange={(e) => setDuration(Number(e.target.value))}
              >
                {durationOptions(form.minDuration, form.maxDuration).map(
                  (s) => (
                    <option key={s} value={s}>
                      {s} s
                    </option>
                  ),
                )}
              </select>
            </S.Field>
            <S.Field>
              Resolution
              <select
                value={resolution}
                onChange={(e) => setResolution(e.target.value)}
              >
                {form.resolutions.map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </select>
            </S.Field>
            <S.Field>
              Aspect ratio
              <select
                value={aspectRatio}
                onChange={(e) => setAspectRatio(e.target.value)}
              >
                {form.aspectRatios.map((r) => (
                  <option key={r} value={r}>
                    {r === 'adaptive' ? 'Adaptive' : r}
                  </option>
                ))}
              </select>
            </S.Field>
          </S.FieldRow>
          {form.audio && (
            <S.Checkbox>
              <input
                type="checkbox"
                checked={audio}
                onChange={(e) => setAudio(e.target.checked)}
              />
              Generate sound
            </S.Checkbox>
          )}
          {form.image &&
            (frame ? (
              <S.FramePreview>
                <img src={frame.dataUri} alt="First frame" />
                <span>{frame.name}</span>
                <S.SecondaryBtn onClick={() => setFrame(null)}>
                  Remove
                </S.SecondaryBtn>
              </S.FramePreview>
            ) : (
              <S.Actions>
                <S.SecondaryBtn onClick={() => fileInput.current?.click()}>
                  <IconPhoto size={16} aria-hidden="true" />
                  Add a first frame
                </S.SecondaryBtn>
              </S.Actions>
            ))}
          <input
            ref={fileInput}
            type="file"
            accept={FRAME_TYPES.join(',')}
            hidden
            onChange={(e) => {
              void pickFrame(e.target.files?.[0]);
              e.target.value = '';
            }}
          />
        </S.Fieldset>
        {renderPrices()}
        {closeIsNext ? (
          <Btn block onClick={() => close(session.Id)} disabled={!!busy}>
            {busy ?? 'Close session'}
          </Btn>
        ) : (
          <Btn block onClick={generate} disabled={!canGenerate}>
            {busy ??
              (session
                ? 'Generate in this session'
                : 'Open session & generate')}
          </Btn>
        )}
        {!session && (
          <S.Hint>
            Opening the session holds MOR in escrow for its length. The next
            window shows the amount before anything is sent.
          </S.Hint>
        )}
      </>
    );
  };

  return (
    <View data-testid="video-container">
      <LayoutHeader title="Video" />
      <QueryError
        error={modelsQuery.error}
        what="models"
        onRetry={() => modelsQuery.refetch()}
      />
      <S.Layout>
        <S.Card aria-labelledby="video-new-clip">
          <S.CardTitle id="video-new-clip">New clip</S.CardTitle>
          {modelsQuery.isPending ? (
            <S.EmptyState>Loading video models…</S.EmptyState>
          ) : !models.length ? (
            <S.EmptyState>
              No video models are registered on the marketplace yet.
            </S.EmptyState>
          ) : (
            <>
              <S.Field>
                Model
                <select
                  value={activeModelId}
                  disabled={!!busy}
                  onChange={(e) => {
                    setModelId(e.target.value);
                    setPickedBid('');
                  }}
                >
                  {models.map((m) => (
                    <option key={m.Id} value={m.Id}>
                      {m.Name}
                    </option>
                  ))}
                </select>
              </S.Field>
              {renderForm()}
            </>
          )}
        </S.Card>
        <S.Card aria-labelledby="video-clips">
          <S.CardTitle id="video-clips">Clips</S.CardTitle>
          <QueryError
            error={jobsQuery.error}
            what="clips"
            onRetry={() => jobsQuery.refetch()}
          />
          {jobsQuery.isPending ? (
            <S.EmptyState>Loading clips…</S.EmptyState>
          ) : !jobs.length ? (
            <S.EmptyState>Clips you generate appear here.</S.EmptyState>
          ) : (
            <S.JobList>
              {jobs.map((job) => (
                <VideoJobCard
                  key={job.id}
                  job={job}
                  modelName={
                    modelNames.get(String(job.model).toLowerCase()) ??
                    abbreviateAddress(job.model, 4)
                  }
                  client={client}
                  toasts={toasts}
                  now={now}
                />
              ))}
            </S.JobList>
          )}
        </S.Card>
      </S.Layout>
    </View>
  );
};

export default withChatState(Video);
