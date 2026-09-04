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
  scrollCoworkTranscriptToLatest,
  selectCoworkModelKey,
} from './Cowork';
import type {
  CoworkModelOption,
  CoworkProject,
  CoworkTask,
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

const persistedProject: CoworkProject = {
  id: 'project-1',
  name: 'Durable project',
  folderName: 'durable-project',
  instructions: 'Keep the implementation focused.',
  approvalMode: 'manual',
  createdAt: 10,
  updatedAt: 20,
};

const persistedApprovalPolicy = {
  schemaVersion: 1 as const,
  id: 'workspace' as const,
  mode: 'manual' as const,
  revision: 1,
  updatedAt: 20,
};

const persistedTask = (overrides: Partial<CoworkTask> = {}): CoworkTask => ({
  id: 'task-1',
  projectId: persistedProject.id,
  title: 'Continue the durable work',
  goal: 'Keep working after a session ends.',
  status: 'completed',
  model: {
    modelId: 'old-model',
    modelName: 'Original Model',
    isLocal: false,
    dataBoundary: 'independent-provider',
    sessionId: 'old-session',
    sessionEndsAt: 10,
  },
  plan: [],
  messages: [
    {
      id: 'latest-message',
      role: 'assistant',
      content: 'The latest saved result.',
      createdAt: 30,
      sequence: 51,
      author: {
        kind: 'model',
        modelId: 'old-model',
        modelName: 'Original Model',
        sessionId: 'old-session',
      },
    },
  ],
  activities: [],
  artifacts: [],
  createdAt: 10,
  updatedAt: 30,
  completedAt: 30,
  ...overrides,
});

const emptyExtensions = {
  skills: [],
  connectors: [],
  issues: [],
  executionAvailable: false as const,
  networkAccessPerformed: false as const,
};

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

  it('positions a newly opened transcript at its latest message immediately', () => {
    const element = { scrollHeight: 1_200, scrollTop: 0 };

    scrollCoworkTranscriptToLatest(element);

    expect(element.scrollTop).toBe(1_200);
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

  it('remounts a selected chat at its latest message after scrolling another chat upward', async () => {
    const previousCowork = window.cowork;
    const scrollHeight = vi
      .spyOn(HTMLElement.prototype, 'scrollHeight', 'get')
      .mockReturnValue(1_200);
    const clientHeight = vi
      .spyOn(HTMLElement.prototype, 'clientHeight', 'get')
      .mockReturnValue(240);
    const firstTask = persistedTask({ title: 'First chat' });
    const secondTask = persistedTask({
      id: 'task-2',
      title: 'Second chat',
      messages: [
        {
          id: 'second-latest',
          role: 'assistant',
          content: 'Newest message in the second chat.',
          createdAt: 40,
          sequence: 8,
        },
      ],
    });
    Object.defineProperty(window, 'cowork', {
      configurable: true,
      writable: true,
      value: {
        getApprovalPolicy: vi.fn(async () => persistedApprovalPolicy),
        listModelOptions: vi.fn(async () => []),
        listProjects: vi.fn(async () => [persistedProject]),
        listTasks: vi.fn(async () => [
          taskSummary({ title: firstTask.title, status: 'completed' }),
          taskSummary({
            id: secondTask.id,
            title: secondTask.title,
            status: 'completed',
            updatedAt: 15,
          }),
        ]),
        getTask: vi.fn(async (id: string) =>
          id === secondTask.id ? secondTask : firstTask,
        ),
        listSchedules: vi.fn(async () => []),
        listExtensions: vi.fn(async () => emptyExtensions),
        onTaskEvent: vi.fn(() => () => undefined),
      } as unknown as Window['cowork'],
    });

    try {
      render(
        createElement(
          MemoryRouter,
          { initialEntries: ['/workspace'] },
          createElement(Cowork),
        ),
      );

      await screen.findByText('The latest saved result.');
      const firstTranscript = document.querySelector(
        '.cowork-transcript',
      ) as HTMLDivElement;
      firstTranscript.scrollTop = 0;
      fireEvent.wheel(firstTranscript, { deltaY: -1 });

      fireEvent.click(screen.getByRole('button', { name: /Second chat/ }));
      await screen.findByText('Newest message in the second chat.');
      const secondTranscript = document.querySelector(
        '.cowork-transcript',
      ) as HTMLDivElement;

      expect(secondTranscript).not.toBe(firstTranscript);
      expect(secondTranscript.scrollTop).toBe(1_200);
    } finally {
      scrollHeight.mockRestore();
      clientHeight.mockRestore();
      Object.defineProperty(window, 'cowork', {
        configurable: true,
        writable: true,
        value: previousCowork,
      });
    }
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

  it('updates one approval policy for every existing and future Workspace task', async () => {
    const previousCowork = window.cowork;
    const updateApprovalPolicy = vi.fn(async () => ({
      ...persistedApprovalPolicy,
      mode: 'skip' as const,
      revision: 2,
      updatedAt: 30,
    }));
    Object.defineProperty(window, 'cowork', {
      configurable: true,
      writable: true,
      value: {
        getApprovalPolicy: vi.fn(async () => persistedApprovalPolicy),
        updateApprovalPolicy,
        listModelOptions: vi.fn(async () => []),
        listProjects: vi.fn(async () => [persistedProject]),
        listTasks: vi.fn(async () => []),
        listSchedules: vi.fn(async () => []),
        listExtensions: vi.fn(async () => emptyExtensions),
        onTaskEvent: vi.fn(() => () => undefined),
      } as unknown as Window['cowork'],
    });

    try {
      render(
        createElement(
          MemoryRouter,
          { initialEntries: ['/workspace'] },
          createElement(Cowork),
        ),
      );

      const policySelect = (await screen.findByLabelText(
        'Global Workspace approval policy',
      )) as HTMLSelectElement;
      expect(policySelect.value).toBe('manual');
      expect(
        screen.getByText(/Applies to every existing and future Workspace task/),
      ).toBeTruthy();

      fireEvent.change(policySelect, { target: { value: 'skip' } });

      await waitFor(() =>
        expect(updateApprovalPolicy).toHaveBeenCalledWith('skip', 1),
      );
      await waitFor(() => expect(policySelect.value).toBe('skip'));
    } finally {
      Object.defineProperty(window, 'cowork', {
        configurable: true,
        writable: true,
        value: previousCowork,
      });
    }
  });

  it('keeps project history available while an unavailable session is replaced', async () => {
    const previousCowork = window.cowork;
    const listProjects = vi.fn(async () => []);
    Object.defineProperty(window, 'cowork', {
      configurable: true,
      writable: true,
      value: {
        getApprovalPolicy: vi.fn(async () => persistedApprovalPolicy),
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

      await screen.findByText('That session is no longer active');
      expect(listProjects).toHaveBeenCalledTimes(1);
      expect(
        screen.getByRole('button', { name: 'Create project' }),
      ).not.toBeNull();

      const navigationToggle = screen.getByRole('button', {
        name: 'Open project navigation',
      });
      fireEvent.click(navigationToggle);
      expect(navigationToggle.getAttribute('aria-expanded')).toBe('true');
      expect(navigationToggle.getAttribute('aria-label')).toBe(
        'Close project navigation',
      );
      fireEvent.keyDown(window, { key: 'Escape' });
      expect(navigationToggle.getAttribute('aria-expanded')).toBe('false');

      const inspectorToggle = screen.getByRole('button', {
        name: 'Open task details',
      });
      fireEvent.click(inspectorToggle);
      expect(inspectorToggle.getAttribute('aria-expanded')).toBe('true');
      expect(inspectorToggle.getAttribute('aria-label')).toBe(
        'Close task details',
      );
      fireEvent.keyDown(window, { key: 'Escape' });
      expect(inspectorToggle.getAttribute('aria-expanded')).toBe('false');

      fireEvent.change(
        screen.getByLabelText('Choose another active Workspace session'),
        { target: { value: 'remote:model-1:session-1' } },
      );
      fireEvent.click(
        screen.getByRole('button', {
          name: 'Use session',
        }),
      );

      await waitFor(() =>
        expect(
          screen.queryByText('That session is no longer active'),
        ).toBeNull(),
      );
    } finally {
      Object.defineProperty(window, 'cowork', {
        configurable: true,
        writable: true,
        value: previousCowork,
      });
    }
  });

  it('renders saved model-attributed history without an active session and pages backward', async () => {
    const previousCowork = window.cowork;
    const task = persistedTask({ hasEarlierMessages: true });
    const listTaskMessages = vi.fn(async () => ({
      messages: [
        {
          id: 'earlier-user',
          role: 'user' as const,
          content: 'Earlier project request',
          createdAt: 15,
          sequence: 1,
        },
        {
          id: 'earlier-model',
          role: 'assistant' as const,
          content: 'Earlier saved answer',
          createdAt: 16,
          sequence: 2,
          author: {
            kind: 'model' as const,
            modelId: 'first-model',
            modelName: 'First Model',
            sessionId: 'first-session',
          },
        },
      ],
      hasMore: false,
    }));
    Object.defineProperty(window, 'cowork', {
      configurable: true,
      writable: true,
      value: {
        getApprovalPolicy: vi.fn(async () => persistedApprovalPolicy),
        listModelOptions: vi.fn(async () => []),
        listProjects: vi.fn(async () => [persistedProject]),
        listTasks: vi.fn(async () => [taskSummary({ status: 'completed' })]),
        getTask: vi.fn(async () => task),
        listTaskMessages,
        listSchedules: vi.fn(async () => []),
        listExtensions: vi.fn(async () => emptyExtensions),
        onTaskEvent: vi.fn(() => () => undefined),
      } as unknown as Window['cowork'],
    });

    try {
      render(
        createElement(
          MemoryRouter,
          { initialEntries: ['/workspace'] },
          createElement(Cowork),
        ),
      );

      await screen.findByText('The latest saved result.');
      expect(screen.getAllByText('Original Model')).toHaveLength(2);
      expect(
        screen.getByText('The original task session has ended'),
      ).toBeTruthy();
      expect(
        screen.getByText(/Project history remains available/),
      ).toBeTruthy();

      fireEvent.click(
        screen.getByRole('button', { name: 'Load earlier messages' }),
      );

      await screen.findByText('Earlier saved answer');
      expect(screen.getByText('Earlier project request')).toBeTruthy();
      expect(screen.getByText('First Model')).toBeTruthy();
      expect(listTaskMessages).toHaveBeenCalledWith('task-1', 51, 50);
      expect(
        screen.queryByRole('button', { name: 'Load earlier messages' }),
      ).toBeNull();
    } finally {
      Object.defineProperty(window, 'cowork', {
        configurable: true,
        writable: true,
        value: previousCowork,
      });
    }
  });

  it('allows a pending action to be denied after its task session expires', async () => {
    const previousCowork = window.cowork;
    const task = persistedTask({
      status: 'waiting_approval',
      pendingApproval: {
        id: 'approval-1',
        toolCall: {
          id: 'tool-call-1',
          type: 'function',
          function: {
            name: 'write_file',
            arguments: JSON.stringify({ path: 'report.txt' }),
          },
        },
        reason: 'Write the generated report',
        risk: 'write',
        createdAt: 25,
      },
    });
    const resolveApproval = vi.fn(async () => ({
      ...task,
      status: 'paused' as const,
      pendingApproval: undefined,
    }));
    Object.defineProperty(window, 'cowork', {
      configurable: true,
      writable: true,
      value: {
        getApprovalPolicy: vi.fn(async () => persistedApprovalPolicy),
        listModelOptions: vi.fn(async () => []),
        listProjects: vi.fn(async () => [persistedProject]),
        listTasks: vi.fn(async () => [
          taskSummary({ status: 'waiting_approval' }),
        ]),
        getTask: vi.fn(async () => task),
        resolveApproval,
        listSchedules: vi.fn(async () => []),
        listExtensions: vi.fn(async () => emptyExtensions),
        onTaskEvent: vi.fn(() => () => undefined),
      } as unknown as Window['cowork'],
    });

    try {
      render(
        createElement(
          MemoryRouter,
          { initialEntries: ['/workspace'] },
          createElement(Cowork),
        ),
      );

      const denyButton = await screen.findByRole('button', { name: 'Deny' });
      const approveButton = screen.getByRole('button', {
        name: 'Approve once',
      });
      expect((denyButton as HTMLButtonElement).disabled).toBe(false);
      expect((approveButton as HTMLButtonElement).disabled).toBe(true);

      fireEvent.click(denyButton);
      fireEvent.click(denyButton);

      await waitFor(() =>
        expect(resolveApproval).toHaveBeenCalledWith(
          'task-1',
          'approval-1',
          false,
        ),
      );
      expect(resolveApproval).toHaveBeenCalledTimes(1);
      await waitFor(() =>
        expect(screen.queryByRole('button', { name: 'Deny' })).toBeNull(),
      );
      expect(
        screen.queryByText(/approval request is no longer active/i),
      ).toBeNull();
    } finally {
      Object.defineProperty(window, 'cowork', {
        configurable: true,
        writable: true,
        value: previousCowork,
      });
    }
  });

  it('requires an explicit continuation before steering with a new session', async () => {
    const previousCowork = window.cowork;
    const replacement = activeModel({
      modelId: 'replacement-model',
      modelName: 'Replacement Model',
      sessionId: 'replacement-session',
    });
    const oldTask = persistedTask({ status: 'paused' });
    const reboundTask = persistedTask({
      status: 'paused',
      model: replacement,
    });
    const getTask = vi
      .fn()
      .mockResolvedValueOnce(oldTask)
      .mockResolvedValue(reboundTask);
    const rebindTask = vi.fn(async () => reboundTask);
    Object.defineProperty(window, 'cowork', {
      configurable: true,
      writable: true,
      value: {
        getApprovalPolicy: vi.fn(async () => persistedApprovalPolicy),
        listModelOptions: vi.fn(async () => [replacement]),
        listProjects: vi.fn(async () => [persistedProject]),
        listTasks: vi.fn(async () => [taskSummary({ status: 'paused' })]),
        getTask,
        rebindTask,
        listSchedules: vi.fn(async () => []),
        listExtensions: vi.fn(async () => emptyExtensions),
        onTaskEvent: vi.fn(() => () => undefined),
      } as unknown as Window['cowork'],
    });

    try {
      render(
        createElement(
          MemoryRouter,
          { initialEntries: ['/workspace'] },
          createElement(Cowork),
        ),
      );

      const continueButton = await screen.findByRole('button', {
        name: 'Continue with Replacement Model',
      });
      expect(
        (
          screen.getByPlaceholderText(
            /Choose an active session/,
          ) as HTMLTextAreaElement
        ).disabled,
      ).toBe(true);

      fireEvent.click(continueButton);
      await waitFor(() =>
        expect(rebindTask).toHaveBeenCalledWith('task-1', replacement),
      );
      await waitFor(() =>
        expect(
          screen.queryByText('The original task session has ended'),
        ).toBeNull(),
      );
      expect(
        (screen.getByPlaceholderText(/Give Workspace/) as HTMLTextAreaElement)
          .disabled,
      ).toBe(false);
    } finally {
      Object.defineProperty(window, 'cowork', {
        configurable: true,
        writable: true,
        value: previousCowork,
      });
    }
  });
});
