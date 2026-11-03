import { describe, expect, it } from 'vitest';
import { VideoJob, sessionBudget } from './videoJobs';

const session = { Id: '0xAA', OpenedAt: 1_000, EndsAt: 1_844 };

const job = (fields: Partial<VideoJob>): VideoJob =>
  ({
    session_id: '0xaa',
    status: 'running',
    vendor_task_id: 'task',
    quote_seconds: 300,
    ...fields,
  }) as VideoJob;

// The provider refuses a job once the quotes of the session's jobs would add up
// to more than its length (plus 10 s), so the app must count the same jobs.
describe('sessionBudget', () => {
  it.each([
    ['no jobs', [], 0, 854, false],
    ['a confirmed job', [job({})], 300, 554, false],
    [
      'a failed job the vendor took',
      [job({ status: 'failed' })],
      300,
      554,
      false,
    ],
    // The node knows a job's quote only from the receipt.
    [
      'a job still on its way',
      [job({ vendor_task_id: '', status: 'queued', quote_seconds: 0 })],
      0,
      854,
      true,
    ],
    [
      'a job the provider refused',
      [job({ vendor_task_id: '', status: 'failed', quote_seconds: 0 })],
      0,
      854,
      false,
    ],
    ['a job on another session', [job({ session_id: '0xbb' })], 0, 854, false],
    [
      'more than the session holds',
      [job({ quote_seconds: 600 }), job({ quote_seconds: 600 })],
      1200,
      0,
      false,
    ],
  ])('%s', (_name, jobs, used, left, pending) => {
    expect(sessionBudget(session, jobs as VideoJob[])).toEqual({
      length: 844,
      used,
      left,
      pending,
    });
  });
});
