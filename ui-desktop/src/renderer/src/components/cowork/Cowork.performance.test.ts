import { createElement } from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it, vi } from 'vitest';
import Cowork, {
  createCoworkTaskEventBatch,
  isActiveMarketplaceCoworkModel,
  isNearCoworkBottom,
  mergeCoworkTaskEvent,
  observeCoworkAutoScroll,
  selectCoworkModelKey,
} from './Cowork';
import type {
  CoworkModelOption,
  CoworkTaskEvent,
  CoworkTaskSummary,
} from './types';

const taskEvent = (
  overrides: Partial<CoworkTaskEvent> = {},
): CoworkTaskEvent => ({
  taskId: 'task-1',
  projectId: 'project-1',
  title: 'Task one',
  status: 'running',
  createdAt: 10,
  updatedAt: 20,
  ...overrides,
});

const taskSummary = (
  overrides: Partial<CoworkTaskSummary> = {},
): CoworkTaskSummary => ({
  id: 'task-1',
  projectId: 'project-1',
  title: 'Task one',
  status: 'running',
  createdAt: 10,
  updatedAt: 20,
  ...overrides,
});

describe('Cowork performance helpers', () => {
  it('uses the configured near-bottom threshold for transcript following', () => {
    expect(
      isNearCoworkBottom({
        clientHeight: 100,
        scrollHeight: 1_000,
        scrollTop: 804,
      }),
    ).toBe(true);
    expect(
      isNearCoworkBottom({
        clientHeight: 100,
        scrollHeight: 1_000,
        scrollTop: 803,
      }),
    ).toBe(false);
  });

  it('disables following on upward input and removes stable listeners', () => {
    const addEventListener = vi.fn();
    const removeEventListener = vi.fn();
    const element = {
      addEventListener,
      clientHeight: 100,
      removeEventListener,
      scrollHeight: 1_000,
      scrollTop: 850,
    } as unknown as HTMLElement;
    const onAutoScrollChange = vi.fn();

    const cleanup = observeCoworkAutoScroll(element, onAutoScrollChange);
    const scrollListener = addEventListener.mock.calls.find(
      ([eventName]) => eventName === 'scroll',
    )?.[1] as EventListener;
    const wheelListener = addEventListener.mock.calls.find(
      ([eventName]) => eventName === 'wheel',
    )?.[1] as EventListener;

    scrollListener(new Event('scroll'));
    wheelListener(new WheelEvent('wheel', { deltaY: -1 }));
    cleanup();

    expect(onAutoScrollChange).toHaveBeenNthCalledWith(1, true);
    expect(onAutoScrollChange).toHaveBeenNthCalledWith(2, true);
    expect(onAutoScrollChange).toHaveBeenNthCalledWith(3, false);
    expect(removeEventListener).toHaveBeenCalledWith('scroll', scrollListener);
    expect(removeEventListener).toHaveBeenCalledWith('wheel', wheelListener);
  });

  it('moves an updated task into place without mutating the existing rail', () => {
    const current: CoworkTaskSummary[] = [
      {
        id: 'task-2',
        projectId: 'project-1',
        title: 'Task two',
        status: 'completed',
        createdAt: 5,
        updatedAt: 30,
      },
      {
        id: 'task-1',
        projectId: 'project-1',
        title: 'Old title',
        status: 'queued',
        createdAt: 10,
        updatedAt: 20,
        startedAt: 15,
      },
    ];

    const next = mergeCoworkTaskEvent(
      current,
      taskEvent({ title: 'New title', updatedAt: 40 }),
    );

    expect(next.map((task) => task.id)).toEqual(['task-1', 'task-2']);
    expect(next[0]).toMatchObject({
      title: 'New title',
      status: 'running',
      startedAt: 15,
      updatedAt: 40,
    });
    expect(current[1]).toMatchObject({ title: 'Old title', updatedAt: 20 });
  });

  it('ignores stale or unchanged task events without scheduling a render', () => {
    const current: CoworkTaskSummary[] = [taskSummary()];

    expect(
      mergeCoworkTaskEvent(
        current,
        taskEvent({ status: 'queued', updatedAt: 19 }),
      ),
    ).toBe(current);
    expect(mergeCoworkTaskEvent(current, taskEvent())).toBe(current);
  });

  it('coalesces a burst to one frame and keeps each task newest event', () => {
    let nextHandle = 0;
    const frames = new Map<number, FrameRequestCallback>();
    const requestFrame = vi.fn((callback: FrameRequestCallback) => {
      const handle = ++nextHandle;
      frames.set(handle, callback);
      return handle;
    });
    const cancelFrame = vi.fn((handle: number) => frames.delete(handle));
    const commits: CoworkTaskEvent[][] = [];
    const batch = createCoworkTaskEventBatch(
      (events) => commits.push(events),
      requestFrame,
      cancelFrame,
    );

    batch.push(taskEvent({ updatedAt: 20 }));
    batch.push(taskEvent({ status: 'completed', updatedAt: 30 }));
    batch.push(taskEvent({ taskId: 'task-2', updatedAt: 25 }));

    expect(requestFrame).toHaveBeenCalledTimes(1);
    expect(commits).toEqual([]);
    frames.get(1)?.(16);
    expect(commits).toHaveLength(1);
    expect(commits[0]).toHaveLength(2);
    expect(commits[0].find((event) => event.taskId === 'task-1')).toMatchObject(
      {
        status: 'completed',
        updatedAt: 30,
      },
    );

    batch.push(taskEvent({ updatedAt: 40 }));
    batch.cancel();
    frames.get(2)?.(32);
    expect(cancelFrame).toHaveBeenCalledWith(2);
    expect(commits).toHaveLength(1);
  });
});

describe('Cowork marketplace session selection', () => {
  const activeModel = (
    overrides: Partial<CoworkModelOption> = {},
  ): CoworkModelOption => ({
    modelId: 'model-1',
    modelName: 'Marketplace model',
    isLocal: false,
    sessionId: 'session-1',
    sessionEndsAt: Date.now() + 60_000,
    source: 'marketplace',
    dataBoundary: 'independent-provider',
    visionCapability: 'declared',
    ...overrides,
  });

  it('rejects local, configured, missing, and expired sessions', () => {
    const now = 1_000;
    expect(isActiveMarketplaceCoworkModel(activeModel(), now)).toBe(true);
    expect(
      isActiveMarketplaceCoworkModel(
        activeModel({
          isLocal: true,
          source: 'local',
          dataBoundary: 'on-device',
          sessionId: undefined,
        }),
        now,
      ),
    ).toBe(false);
    expect(
      isActiveMarketplaceCoworkModel(activeModel({ sessionEndsAt: now }), now),
    ).toBe(false);
    expect(
      isActiveMarketplaceCoworkModel(
        activeModel({ sessionId: undefined }),
        now,
      ),
    ).toBe(false);
  });

  it('preselects the exact active session passed from Chat', () => {
    const first = activeModel();
    const requested = activeModel({
      modelId: 'model-2',
      modelName: 'Requested model',
      sessionId: 'session-2',
    });

    expect(selectCoworkModelKey([first, requested], '', 'session-2')).toBe(
      'remote:model-2:session-2',
    );
  });

  it('never falls back when the exact Chat handoff is unavailable', () => {
    const first = activeModel();
    const second = activeModel({
      modelId: 'model-2',
      sessionId: 'session-2',
    });

    expect(
      selectCoworkModelKey(
        [first, second],
        'remote:model-2:session-2',
        'missing',
      ),
    ).toBe('');
    expect(selectCoworkModelKey([first], '', 'missing')).toBe('');
  });

  it('keeps an active explicit selection and only defaults on first entry', () => {
    const first = activeModel();
    const second = activeModel({
      modelId: 'model-2',
      sessionId: 'session-2',
    });

    expect(
      selectCoworkModelKey([first, second], 'remote:model-2:session-2'),
    ).toBe('remote:model-2:session-2');
    expect(selectCoworkModelKey([first], '')).toBe('remote:model-1:session-1');
    expect(selectCoworkModelKey([first], 'remote:model-2:session-2')).toBe('');
  });

  it('keeps the whole workspace locked until another session is chosen explicitly', async () => {
    const previousCowork = window.cowork;
    const listProjects = vi.fn(async () => []);
    Object.defineProperty(window, 'cowork', {
      configurable: true,
      writable: true,
      value: {
        listModelOptions: vi.fn(async () => [activeModel()]),
        listProjects,
        onTaskEvent: vi.fn(() => () => undefined),
      } as unknown as Window['cowork'],
    });

    try {
      render(
        createElement(
          MemoryRouter,
          { initialEntries: ['/cowork?sessionId=missing'] },
          createElement(Cowork),
        ),
      );

      await screen.findByRole('heading', {
        name: 'That Cowork session is no longer active',
      });
      expect(listProjects).not.toHaveBeenCalled();
      expect(
        screen.queryByRole('button', { name: 'Create project' }),
      ).toBeNull();

      fireEvent.change(
        screen.getByLabelText('Choose another active Cowork session'),
        { target: { value: 'remote:model-1:session-1' } },
      );
      fireEvent.click(
        screen.getByRole('button', {
          name: 'Use selected active session',
        }),
      );

      await waitFor(() => expect(listProjects).toHaveBeenCalledTimes(1));
      expect(
        screen.getByRole('button', { name: 'Create project' }),
      ).not.toBeNull();
    } finally {
      Object.defineProperty(window, 'cowork', {
        configurable: true,
        writable: true,
        value: previousCowork,
      });
    }
  });
});
