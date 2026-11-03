import { getModelModality, isClosed } from '../chat/utils';
import { normalizeModelList } from '../../store/utils/modelMetadata';

// Shapes of the proxy-router video documents (proxy-router/internal/videojob).

export type VideoGuideField = {
  name: string;
  required?: boolean;
  type?: string;
  enum?: string[];
  minimum?: number;
  maximum?: number;
  description?: string;
};

export type VideoGuide = {
  object: string;
  model_name?: string;
  defaults: {
    duration_seconds: number;
    aspect_ratio: string;
    resolution: string;
    generate_audio: boolean;
  };
  fields: VideoGuideField[];
};

export type VideoGuideOffer = {
  provider: string;
  bidId: string;
  status?: string;
  guide?: VideoGuide;
  error?: string;
};

export type VideoResolved = {
  prompt: string;
  duration_seconds: number;
  aspect_ratio: string;
  resolution: string;
  generate_audio: boolean;
  has_image: boolean;
};

export type VideoQuote = {
  provider: string;
  bid_id: string;
  price_per_second_wei: string;
  // The session to open for this job alone: quote_seconds raised to the
  // contract minimum.
  session_seconds?: number;
  // What the job uses of a session; the provider adds these up against the
  // session length.
  quote_seconds?: number;
  session_cost_wei?: string;
  resolved?: VideoResolved;
  error?: string;
};

export type VideoJob = {
  id: string;
  model: string;
  status: string;
  session_id: string;
  quote_seconds: number;
  vendor_task_id: string;
  resolved: VideoResolved;
  video_url?: string;
  expires_at?: string;
  error?: { code?: string; message: string } | null;
  created_at: string;
  updated_at: string;
};

export type VideoForm = {
  minDuration: number;
  maxDuration: number;
  aspectRatios: string[];
  resolutions: string[];
  audio: boolean;
  image: boolean;
  defaults: VideoGuide['defaults'];
};

const TERMINAL_STATUSES = new Set([
  'succeeded',
  'failed',
  'expired',
  'cancelled',
]);

export const isTerminalStatus = (status?: string) =>
  TERMINAL_STATUSES.has(String(status));

export const selectVideoModels = (models: unknown) =>
  normalizeModelList(models).filter(
    (model) => !model.IsDeleted && getModelModality(model) === 'video',
  );

export const sameId = (a?: string, b?: string) =>
  !!a && !!b && a.toLowerCase() === b.toLowerCase();

/** The first rated provider, other than this wallet, that published a guide. */
export const guideOffer = (offers: VideoGuideOffer[], ownAddress?: string) =>
  offers.find((offer) => offer.guide && !sameId(offer.provider, ownAddress));

/** What the job form may offer, read back from the guide's fields. */
export const videoForm = (guide?: VideoGuide | null): VideoForm | null => {
  if (!guide?.defaults || !Array.isArray(guide.fields)) return null;
  const field = (name: string) => guide.fields.find((f) => f?.name === name);
  const duration = field('duration_seconds');
  const defaults = guide.defaults;
  return {
    minDuration: duration?.minimum ?? defaults.duration_seconds,
    maxDuration: duration?.maximum ?? defaults.duration_seconds,
    aspectRatios: field('aspect_ratio')?.enum ?? [defaults.aspect_ratio],
    resolutions: field('resolution')?.enum ?? [defaults.resolution],
    audio: !!field('generate_audio'),
    // The app sends a first frame from a local file, so only inline frames.
    image: !!field('image_b64'),
    defaults,
  };
};

/** This wallet's open sessions for the model, newest first. */
export const openModelSessions = (sessions: any[], modelId?: string) =>
  (sessions ?? [])
    .filter((s) => !isClosed(s) && sameId(s?.ModelAgentId, modelId))
    .sort((a, b) => Number(b.OpenedAt) - Number(a.OpenedAt));

// The provider accepts a job while the quotes of the session's jobs add up to
// no more than the session's length, with this much slack.
const SESSION_SLACK_SECONDS = 10;

/**
 * A job holds session time once the provider may have taken it: confirmed by
 * the vendor, or still on its way. A job the provider refused holds none.
 */
const holdsSessionTime = (job: VideoJob) =>
  !!job.vendor_task_id || !isTerminalStatus(job.status);

/**
 * How much of the session its clips use. The node learns a job's quote only
 * from the provider's receipt, so while a job on the session is unconfirmed
 * the time left is not known (`pending`).
 */
export const sessionBudget = (session: any, jobs: VideoJob[]) => {
  const length = Number(session?.EndsAt) - Number(session?.OpenedAt);
  const onSession = (jobs ?? []).filter(
    (job) => sameId(job.session_id, session?.Id) && holdsSessionTime(job),
  );
  const used = onSession.reduce(
    (sum, job) => sum + (Number(job.quote_seconds) || 0),
    0,
  );
  return {
    length,
    used,
    left: Math.max(0, length + SESSION_SLACK_SECONDS - used),
    pending: onSession.some((job) => !job.vendor_task_id),
  };
};

/** The session time a quoted job uses once it is accepted. */
export const jobSeconds = (quote?: VideoQuote) =>
  quote && !quote.error
    ? (quote.quote_seconds ?? quote.session_seconds)
    : undefined;

export const usableQuotes = (quotes: VideoQuote[], ownAddress?: string) =>
  (quotes ?? []).filter(
    (q) =>
      !q.error &&
      !!q.session_seconds &&
      !!q.session_cost_wei &&
      !sameId(q.provider, ownAddress),
  );

export const cheapestQuote = (quotes: VideoQuote[], ownAddress?: string) =>
  usableQuotes(quotes, ownAddress).sort((a, b) => {
    const diff =
      BigInt(a.session_cost_wei ?? 0) - BigInt(b.session_cost_wei ?? 0);
    return diff < 0n ? -1 : diff > 0n ? 1 : 0;
  })[0];

export const formatSeconds = (total: number) => {
  const seconds = Math.max(0, Math.round(total));
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  if (h) return m ? `${h} h ${m} min` : `${h} h`;
  if (m) return s ? `${m} min ${s} s` : `${m} min`;
  return `${s} s`;
};

/** A live timer: 3:44, or 1:02:05 past an hour. */
export const formatClock = (total: number) => {
  const seconds = Math.max(0, Math.floor(total));
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const pad = (n: number) => String(n).padStart(2, '0');
  return h
    ? `${h}:${pad(m)}:${pad(seconds % 60)}`
    : `${m}:${pad(seconds % 60)}`;
};

export const STATUS_LABELS: Record<string, string> = {
  queued: 'Queued',
  running: 'Rendering',
  succeeded: 'Ready',
  failed: 'Failed',
  expired: 'Expired',
  cancelled: 'Cancelled',
};

export const jobErrorText = (job: VideoJob) => {
  switch (job.error?.code) {
    case 'session_closed_early':
      return 'The session was closed before this clip was delivered.';
    case 'invalid_video_url':
      return 'The provider sent a link that is not a public https link.';
  }
  return job.error?.message ?? '';
};

export const linkExpired = (job: VideoJob, now = Date.now()) =>
  !!job.expires_at && new Date(job.expires_at).getTime() <= now;
