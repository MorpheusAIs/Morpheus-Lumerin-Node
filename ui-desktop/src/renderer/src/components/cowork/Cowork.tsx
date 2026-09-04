import {
  FormEvent,
  memo,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import {
  IconActivity,
  IconAlertTriangle,
  IconCalendarTime,
  IconCheck,
  IconChevronRight,
  IconCircleCheck,
  IconCircleDashed,
  IconClock,
  IconCloud,
  IconDeviceLaptop,
  IconExternalLink,
  IconFile,
  IconFolder,
  IconFolderPlus,
  IconListCheck,
  IconLayoutSidebarRight,
  IconLoader2,
  IconMenu2,
  IconPaperclip,
  IconPencil,
  IconPlayerPause,
  IconPlayerPlay,
  IconPlayerStop,
  IconPlus,
  IconPuzzle,
  IconRefresh,
  IconRobot,
  IconSend,
  IconShieldCheck,
  IconSparkles,
  IconTrash,
  IconX,
} from '@tabler/icons-react';
import ReactMarkdown from 'react-markdown';
import { useLocation, useNavigate } from 'react-router';
import type {
  CoworkActivity,
  CoworkApprovalMode,
  CoworkApprovalPolicy,
  CoworkArtifact,
  CoworkArtifactPreview,
  CoworkDisplayMessage,
  CoworkExtensionCatalog,
  CoworkModelOption,
  CoworkPlanStep,
  CoworkProject,
  CoworkSchedule,
  CoworkScheduleCadence,
  CoworkTask,
  CoworkTaskEvent,
  CoworkTaskSummary,
  CoworkTaskStatus,
} from './types';
import './Cowork.css';

const COWORK_BOTTOM_THRESHOLD_PX = 96;

type CoworkScrollMetrics = Pick<
  HTMLElement,
  'clientHeight' | 'scrollHeight' | 'scrollTop'
>;

export const isNearCoworkBottom = (
  element: CoworkScrollMetrics,
  threshold = COWORK_BOTTOM_THRESHOLD_PX,
): boolean =>
  element.scrollHeight - element.scrollTop - element.clientHeight <= threshold;

export const scrollCoworkTranscriptToLatest = (
  element: Pick<HTMLElement, 'scrollHeight' | 'scrollTop'>,
): void => {
  element.scrollTop = element.scrollHeight;
};

export const observeCoworkAutoScroll = (
  element: HTMLElement,
  onAutoScrollChange: (enabled: boolean) => void,
  initialize = true,
): (() => void) => {
  const update = () => onAutoScrollChange(isNearCoworkBottom(element));
  const onWheel = (event: WheelEvent) => {
    if (event.deltaY < 0) onAutoScrollChange(false);
  };

  if (initialize) update();
  element.addEventListener('scroll', update, { passive: true });
  element.addEventListener('wheel', onWheel, { passive: true });
  return () => {
    element.removeEventListener('scroll', update);
    element.removeEventListener('wheel', onWheel);
  };
};

const taskSummaryFromEvent = (event: CoworkTaskEvent): CoworkTaskSummary => ({
  id: event.taskId,
  projectId: event.projectId,
  title: event.title,
  status: event.status,
  createdAt: event.createdAt,
  updatedAt: event.updatedAt,
  ...(event.startedAt !== undefined ? { startedAt: event.startedAt } : {}),
  ...(event.completedAt !== undefined
    ? { completedAt: event.completedAt }
    : {}),
});

const equalTaskSummary = (
  left: CoworkTaskSummary,
  right: CoworkTaskSummary,
): boolean =>
  left.id === right.id &&
  left.projectId === right.projectId &&
  left.title === right.title &&
  left.status === right.status &&
  left.createdAt === right.createdAt &&
  left.updatedAt === right.updatedAt &&
  left.startedAt === right.startedAt &&
  left.completedAt === right.completedAt;

/** Applies one summary delta without sorting the complete task rail. */
export const mergeCoworkTaskEvent = (
  current: CoworkTaskSummary[],
  event: CoworkTaskEvent,
): CoworkTaskSummary[] => {
  const existingIndex = current.findIndex((task) => task.id === event.taskId);
  const existing = existingIndex >= 0 ? current[existingIndex] : undefined;
  if (existing && existing.updatedAt > event.updatedAt) return current;

  const summary = existing
    ? { ...existing, ...taskSummaryFromEvent(event) }
    : taskSummaryFromEvent(event);
  if (existing && equalTaskSummary(existing, summary)) return current;

  const next =
    existingIndex >= 0
      ? current.filter((_task, index) => index !== existingIndex)
      : [...current];
  const insertionIndex = next.findIndex(
    (task) => task.updatedAt < summary.updatedAt,
  );
  if (insertionIndex < 0) next.push(summary);
  else next.splice(insertionIndex, 0, summary);
  return next;
};

type CoworkTaskEventBatch = {
  cancel: () => void;
  push: (event: CoworkTaskEvent) => void;
};

/** Coalesces event bursts and keeps only the newest summary for each task. */
export const createCoworkTaskEventBatch = (
  commit: (events: CoworkTaskEvent[]) => void,
  requestFrame: (callback: FrameRequestCallback) => number = (callback) =>
    window.requestAnimationFrame(callback),
  cancelFrame: (handle: number) => void = (handle) =>
    window.cancelAnimationFrame(handle),
): CoworkTaskEventBatch => {
  const pending = new Map<string, CoworkTaskEvent>();
  let frame: number | undefined;

  return {
    push(event) {
      const current = pending.get(event.taskId);
      if (!current || current.updatedAt <= event.updatedAt) {
        pending.set(event.taskId, event);
      }
      if (frame !== undefined) return;
      frame = requestFrame(() => {
        frame = undefined;
        if (!pending.size) return;
        const events = [...pending.values()];
        pending.clear();
        commit(events);
      });
    },
    cancel() {
      if (frame !== undefined) cancelFrame(frame);
      frame = undefined;
      pending.clear();
    },
  };
};

const STATUS_LABELS: Record<CoworkTaskStatus, string> = {
  draft: 'Draft',
  queued: 'Ready',
  running: 'Working',
  waiting_approval: 'Needs approval',
  paused: 'Paused',
  completed: 'Complete',
  failed: 'Failed',
  cancelled: 'Cancelled',
};

const APPROVAL_DESCRIPTIONS: Record<CoworkApprovalMode, string> = {
  manual: 'Ask before every file change',
  auto: 'Approve safe new files automatically',
  skip: 'Allow file changes without asking',
};

const DAY_NAMES = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
];

const localTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
const shortTimeFormatter = new Intl.DateTimeFormat(undefined, {
  hour: 'numeric',
  minute: '2-digit',
});

const formatTime = (value?: number): string => {
  if (!value) return '';
  return shortTimeFormatter.format(value);
};

const formatRelativeTime = (value: number): string => {
  const elapsed = Math.max(0, Date.now() - value);
  const minutes = Math.floor(elapsed / 60_000);
  if (minutes < 1) return 'now';
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  return `${Math.floor(hours / 24)}d`;
};

const formatScheduleTime = (value: number | undefined, timeZone: string) => {
  if (!value) return 'Not yet';
  try {
    return new Intl.DateTimeFormat(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short',
      timeZone,
    }).format(value);
  } catch {
    return new Date(value).toLocaleString();
  }
};

const clockLabel = (hour: number, minute: number): string => {
  const instant = new Date(2000, 0, 1, hour, minute);
  return shortTimeFormatter.format(instant);
};

const cadenceLabel = (cadence: CoworkScheduleCadence): string => {
  switch (cadence.kind) {
    case 'manual':
      return 'Manual only';
    case 'hourly':
      return `Hourly at :${String(cadence.minute).padStart(2, '0')}`;
    case 'daily':
      return `Daily at ${clockLabel(cadence.hour, cadence.minute)}`;
    case 'weekdays':
      return `Weekdays at ${clockLabel(cadence.hour, cadence.minute)}`;
    case 'weekly':
      return `${DAY_NAMES[cadence.dayOfWeek]}s at ${clockLabel(
        cadence.hour,
        cadence.minute,
      )}`;
  }
};

const cadenceFromForm = (
  kind: CoworkScheduleCadence['kind'],
  time: string,
  hourlyMinute: number,
  weekday: number,
): CoworkScheduleCadence => {
  if (kind === 'manual') return { kind };
  if (kind === 'hourly') return { kind, minute: hourlyMinute };
  const [hour, minute] = time.split(':').map(Number);
  if (kind === 'weekly') return { kind, dayOfWeek: weekday, hour, minute };
  return { kind, hour, minute };
};

const errorMessage = (error: unknown): string =>
  error instanceof Error ? error.message : 'Something went wrong.';

const modelKey = (
  model: Pick<CoworkModelOption, 'isLocal' | 'modelId' | 'sessionId'>,
): string =>
  [
    model.isLocal ? 'local' : 'remote',
    model.modelId,
    model.sessionId ?? '',
  ].join(':');

export const isActiveMarketplaceCoworkModel = (
  model: CoworkModelOption,
  now = Date.now(),
): boolean =>
  model.source === 'marketplace' &&
  !model.isLocal &&
  Boolean(model.sessionId) &&
  typeof model.sessionEndsAt === 'number' &&
  Number.isFinite(model.sessionEndsAt) &&
  model.sessionEndsAt > now;

export const selectCoworkModelKey = (
  models: CoworkModelOption[],
  current: string,
  requestedSessionId?: string,
): string => {
  const available = models.filter((model) =>
    isActiveMarketplaceCoworkModel(model),
  );
  const requested = requestedSessionId
    ? available.find((model) => model.sessionId === requestedSessionId)
    : undefined;
  if (requestedSessionId) return requested ? modelKey(requested) : '';
  if (current) {
    return available.some((model) => modelKey(model) === current)
      ? current
      : '';
  }
  return available[0] ? modelKey(available[0]) : '';
};

const parseToolArguments = (value: string): string => {
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value;
  }
};

const toolLabel = (name: string): string => {
  if (name === 'authorize_remote_model') {
    return 'Share task data with the selected model';
  }
  return name.replaceAll('_', ' ');
};

const SAFE_MARKDOWN_COMPONENTS = {
  a: ({ children, ...props }) => (
    <a {...props} rel="noreferrer" target="_blank">
      {children}
    </a>
  ),
  img: () => null,
} satisfies NonNullable<
  React.ComponentProps<typeof ReactMarkdown>['components']
>;

const SafeMarkdown = memo(({ children }: { children: string }) => (
  <ReactMarkdown
    className="cowork-markdown"
    components={SAFE_MARKDOWN_COMPONENTS}
    skipHtml
  >
    {children}
  </ReactMarkdown>
));

const CoworkMessageRow = memo(
  ({ message }: { message: CoworkDisplayMessage }) => {
    const modelAuthor =
      message.role === 'assistant' && message.author?.kind === 'model'
        ? message.author
        : null;
    const authorLabel =
      message.role === 'user'
        ? 'You'
        : modelAuthor?.modelName ||
          (message.author?.kind === 'model' ? 'Model' : 'Workspace');

    return (
      <article className={`cowork-message cowork-message-${message.role}`}>
        <div className="cowork-message-avatar">
          {message.role === 'assistant' ? <IconSparkles size={16} /> : 'You'}
        </div>
        <div>
          <div className="cowork-message-meta">
            <strong>{authorLabel}</strong>
            {modelAuthor && <span className="cowork-author-kind">Model</span>}
            <time dateTime={new Date(message.createdAt).toISOString()}>
              {formatTime(message.createdAt)}
            </time>
          </div>
          {message.role === 'assistant' ? (
            <SafeMarkdown>{message.content}</SafeMarkdown>
          ) : (
            <p>{message.content}</p>
          )}
        </div>
      </article>
    );
  },
  (previous, next) =>
    previous.message.id === next.message.id &&
    previous.message.role === next.message.role &&
    previous.message.content === next.message.content &&
    previous.message.createdAt === next.message.createdAt &&
    previous.message.sequence === next.message.sequence &&
    previous.message.author?.kind === next.message.author?.kind &&
    previous.message.author?.modelId === next.message.author?.modelId &&
    previous.message.author?.modelName === next.message.author?.modelName &&
    previous.message.author?.sessionId === next.message.author?.sessionId,
);

const PlanStatusIcon = ({ status }: Pick<CoworkPlanStep, 'status'>) => {
  if (status === 'completed') return <IconCircleCheck size={17} />;
  if (status === 'in_progress')
    return <IconLoader2 className="cowork-spin" size={17} />;
  return <IconCircleDashed size={17} />;
};

const TaskStatusIcon = ({ status }: Pick<CoworkTask, 'status'>) => {
  if (status === 'completed') return <IconCircleCheck size={16} />;
  if (status === 'running')
    return <IconLoader2 className="cowork-spin" size={16} />;
  if (status === 'waiting_approval') return <IconAlertTriangle size={16} />;
  if (status === 'paused') return <IconPlayerPause size={16} />;
  if (status === 'failed' || status === 'cancelled') return <IconX size={16} />;
  return <IconCircleDashed size={16} />;
};

const ScheduleStatusIcon = ({ schedule }: { schedule: CoworkSchedule }) =>
  schedule.runningSince ? (
    <IconLoader2 className="cowork-spin" size={16} />
  ) : schedule.status === 'active' ? (
    <IconCalendarTime size={16} />
  ) : (
    <IconPlayerPause size={16} />
  );

const EmptyPanel = ({
  icon,
  title,
  detail,
}: {
  icon: React.ReactNode;
  title: string;
  detail: string;
}) => (
  <div className="cowork-empty-panel">
    <span className="cowork-empty-icon">{icon}</span>
    <strong>{title}</strong>
    <span>{detail}</span>
  </div>
);

const ActivityIcon = ({ activity }: { activity: CoworkActivity }) => {
  if (activity.status === 'running')
    return <IconLoader2 className="cowork-spin" />;
  if (activity.status === 'success') return <IconCheck />;
  if (activity.status === 'error') return <IconX />;
  return <IconClock />;
};

function Cowork(): JSX.Element {
  const api = window.cowork;
  const location = useLocation();
  const navigate = useNavigate();
  const requestedSessionId = useMemo(
    () => new URLSearchParams(location.search).get('sessionId') ?? undefined,
    [location.search],
  );
  const [projects, setProjects] = useState<CoworkProject[]>([]);
  const [tasks, setTasks] = useState<CoworkTaskSummary[]>([]);
  const [schedules, setSchedules] = useState<CoworkSchedule[]>([]);
  const [extensions, setExtensions] = useState<CoworkExtensionCatalog | null>(
    null,
  );
  const [extensionsLoading, setExtensionsLoading] = useState(false);
  const [models, setModels] = useState<CoworkModelOption[]>([]);
  const [sessionNow, setSessionNow] = useState(() => Date.now());
  const [replacementModelKey, setReplacementModelKey] = useState('');
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(
    null,
  );
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [selectedScheduleId, setSelectedScheduleId] = useState<string | null>(
    null,
  );
  const [editingScheduleId, setEditingScheduleId] = useState<string | null>(
    null,
  );
  const [railMode, setRailMode] = useState<'tasks' | 'schedules'>('tasks');
  const [activeTask, setActiveTask] = useState<CoworkTask | null>(null);
  const [showProjectForm, setShowProjectForm] = useState(false);
  const [projectName, setProjectName] = useState('');
  const [projectInstructions, setProjectInstructions] = useState('');
  const [approvalPolicy, setApprovalPolicy] =
    useState<CoworkApprovalPolicy | null>(null);
  const [editingProjectId, setEditingProjectId] = useState<string | null>(null);
  const [projectSettingsName, setProjectSettingsName] = useState('');
  const [projectSettingsInstructions, setProjectSettingsInstructions] =
    useState('');
  const [taskTitle, setTaskTitle] = useState('');
  const [taskGoal, setTaskGoal] = useState('');
  const [selectedModelKey, setSelectedModelKey] = useState('');
  const [steeringMessage, setSteeringMessage] = useState('');
  const [scheduleName, setScheduleName] = useState('');
  const [scheduleTitle, setScheduleTitle] = useState('');
  const [scheduleGoal, setScheduleGoal] = useState('');
  const [scheduleCadence, setScheduleCadence] =
    useState<CoworkScheduleCadence['kind']>('weekdays');
  const [scheduleTime, setScheduleTime] = useState('09:00');
  const [scheduleHourlyMinute, setScheduleHourlyMinute] = useState(0);
  const [scheduleWeekday, setScheduleWeekday] = useState(1);
  const [scheduleTimeZone, setScheduleTimeZone] = useState(localTimeZone);
  const [scheduleStatus, setScheduleStatus] = useState<'active' | 'paused'>(
    'active',
  );
  const [busy, setBusy] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [preview, setPreview] = useState<CoworkArtifactPreview | null>(null);
  const [railOpen, setRailOpen] = useState(false);
  const [inspectorOpen, setInspectorOpen] = useState(false);
  const [earlierMessages, setEarlierMessages] = useState<
    CoworkDisplayMessage[]
  >([]);
  const [hasEarlierMessages, setHasEarlierMessages] = useState(false);
  const [nextBeforeSequence, setNextBeforeSequence] = useState<
    number | undefined
  >(undefined);
  const [loadingEarlierMessages, setLoadingEarlierMessages] = useState(false);
  const transcriptElementRef = useRef<HTMLDivElement | null>(null);
  const transcriptAutoScrollRef = useRef(true);
  const transcriptObserverCleanupRef = useRef<(() => void) | null>(null);
  const refreshTimerRef = useRef<number | undefined>(undefined);
  const taskListRequestRef = useRef(0);
  const activeTaskRequestRef = useRef(0);
  const selectedProjectIdRef = useRef<string | null>(selectedProjectId);
  const selectedTaskIdRef = useRef<string | null>(selectedTaskId);
  const transcriptScrollSnapshotRef = useRef<{
    scrollHeight: number;
    scrollTop: number;
  } | null>(null);
  const approvalSubmissionRef = useRef<string | null>(null);
  const activeTaskRefreshRef = useRef<{
    dirty: boolean;
    inFlight: boolean;
    taskId: string | null;
  }>({ dirty: false, inFlight: false, taskId: null });

  selectedProjectIdRef.current = selectedProjectId;
  selectedTaskIdRef.current = selectedTaskId;

  const attachTranscriptElement = useCallback(
    (element: HTMLDivElement | null) => {
      transcriptObserverCleanupRef.current?.();
      transcriptObserverCleanupRef.current = null;
      transcriptElementRef.current = element;
      if (!element) return;
      // A newly mounted transcript represents a newly opened task and should
      // start at its latest message. Subsequent user scrolling owns the policy.
      transcriptAutoScrollRef.current = true;
      scrollCoworkTranscriptToLatest(element);
      transcriptObserverCleanupRef.current = observeCoworkAutoScroll(
        element,
        (enabled) => {
          transcriptAutoScrollRef.current = enabled;
        },
        false,
      );
    },
    [],
  );

  const activeModels = useMemo(
    () =>
      models.filter((model) =>
        isActiveMarketplaceCoworkModel(model, sessionNow),
      ),
    [models, sessionNow],
  );

  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) ?? null,
    [projects, selectedProjectId],
  );
  const approvalMode = approvalPolicy?.mode ?? 'manual';

  const selectTask = useCallback((taskId: string | null) => {
    if (selectedTaskIdRef.current !== taskId) {
      activeTaskRequestRef.current += 1;
      setActiveTask(null);
    }
    selectedTaskIdRef.current = taskId;
    setSelectedTaskId(taskId);
  }, []);

  const selectedModel = useMemo(
    () =>
      activeModels.find((model) => modelKey(model) === selectedModelKey) ??
      null,
    [activeModels, selectedModelKey],
  );

  const activeTaskSessionModel = useMemo(
    () =>
      activeTask
        ? (activeModels.find(
            (model) => modelKey(model) === modelKey(activeTask.model),
          ) ?? null)
        : null,
    [activeModels, activeTask],
  );
  const exactTaskSessionActive = Boolean(activeTaskSessionModel);
  const taskNeedsRebind = Boolean(activeTask && !activeTaskSessionModel);

  const displayedMessages = useMemo(() => {
    const seen = new Set<string>();
    return [...earlierMessages, ...(activeTask?.messages ?? [])].filter(
      (message) => {
        if (seen.has(message.id)) return false;
        seen.add(message.id);
        return true;
      },
    );
  }, [activeTask?.messages, earlierMessages]);

  const selectedSchedule = useMemo(
    () =>
      schedules.find((schedule) => schedule.id === selectedScheduleId) ?? null,
    [schedules, selectedScheduleId],
  );

  const modelGroups = useMemo(() => {
    const capabilities = [
      { value: 'declared', label: 'Vision (declared)' },
      { value: 'detected', label: 'Possible vision (name match)' },
      { value: 'none', label: 'Text / vision not declared' },
    ] as const;
    return capabilities
      .map((capability) => ({
        label: `${capability.label} · active Morpheus sessions`,
        options: activeModels.filter(
          (model) => model.visionCapability === capability.value,
        ),
      }))
      .filter((group) => group.options.length > 0);
  }, [activeModels]);

  const loadProjects = useCallback(async () => {
    if (!api) return;
    const result = await api.listProjects();
    setProjects(result);
    setSelectedProjectId((current) =>
      current && result.some((project) => project.id === current)
        ? current
        : (result[0]?.id ?? null),
    );
  }, [api]);

  const loadApprovalPolicy = useCallback(async () => {
    if (!api) return;
    const policy = await api.getApprovalPolicy();
    setApprovalPolicy(policy);
  }, [api]);

  const loadModels = useCallback(
    async (force = false) => {
      if (!api) return;
      try {
        const result: CoworkModelOption[] = await api.listModelOptions(force);
        const activeMarketplaceModels = result.filter((model) =>
          isActiveMarketplaceCoworkModel(model),
        );
        setModels(activeMarketplaceModels);
        setSelectedModelKey((current) =>
          selectCoworkModelKey(
            activeMarketplaceModels,
            current,
            requestedSessionId,
          ),
        );
      } catch (loadError) {
        // Failing to revalidate must lock the workspace instead of retaining a
        // stale model/session selection from a previous successful request.
        setModels([]);
        setSelectedModelKey('');
        setReplacementModelKey('');
        throw loadError;
      }
    },
    [api, requestedSessionId],
  );

  const loadTasks = useCallback(
    async (projectId: string, preferredTaskId?: string | null) => {
      if (!api) return;
      const request = ++taskListRequestRef.current;
      const result = await api.listTasks(projectId);
      if (
        request !== taskListRequestRef.current ||
        selectedProjectIdRef.current !== projectId
      ) {
        return;
      }
      setTasks(result);
      const preferred = preferredTaskId ?? selectedTaskIdRef.current;
      const next =
        preferred && result.some((task) => task.id === preferred)
          ? preferred
          : (result[0]?.id ?? null);
      selectTask(next);
    },
    [api, selectTask],
  );

  const loadTask = useCallback(
    async (taskId: string) => {
      if (!api) return;
      const request = ++activeTaskRequestRef.current;
      const task = await api.getTask(taskId);
      if (
        request !== activeTaskRequestRef.current ||
        selectedTaskIdRef.current !== taskId
      ) {
        return;
      }
      setActiveTask(task);
    },
    [api],
  );

  const refreshActiveTask = useCallback(
    async (taskId: string) => {
      const state = activeTaskRefreshRef.current;
      state.taskId = taskId;
      if (state.inFlight) {
        state.dirty = true;
        return;
      }
      state.inFlight = true;
      try {
        do {
          state.dirty = false;
          const nextTaskId = state.taskId;
          if (!nextTaskId) return;
          await loadTask(nextTaskId);
        } while (state.dirty);
      } finally {
        state.inFlight = false;
      }
    },
    [loadTask],
  );

  const loadSchedules = useCallback(
    async (projectId: string, preferredScheduleId?: string | null) => {
      if (!api) return;
      const result: CoworkSchedule[] = await api.listSchedules(projectId);
      setSchedules(result);
      setSelectedScheduleId((current) => {
        const preferred = preferredScheduleId ?? current;
        return preferred && result.some((schedule) => schedule.id === preferred)
          ? preferred
          : (result[0]?.id ?? null);
      });
    },
    [api],
  );

  const loadExtensions = useCallback(
    async (projectId: string) => {
      if (!api) return;
      setExtensionsLoading(true);
      try {
        setExtensions(await api.listExtensions(projectId));
      } finally {
        setExtensionsLoading(false);
      }
    },
    [api],
  );

  useEffect(() => {
    const now = Date.now();
    const nextExpiry = models.reduce<number | undefined>((next, model) => {
      const endsAt = model.sessionEndsAt;
      if (!Number.isFinite(endsAt) || !endsAt || endsAt <= now) return next;
      return next === undefined || endsAt < next ? endsAt : next;
    }, undefined);
    if (nextExpiry === undefined) return;
    const timeout = window.setTimeout(
      () => setSessionNow(Date.now()),
      Math.max(0, nextExpiry - now),
    );
    return () => window.clearTimeout(timeout);
  }, [models, sessionNow]);

  useEffect(() => {
    if (!api) {
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
    Promise.allSettled([
      loadModels(Boolean(requestedSessionId)),
      loadProjects(),
      loadApprovalPolicy(),
    ])
      .then((results) => {
        if (!active) return;
        const failures = results.filter(
          (result): result is PromiseRejectedResult =>
            result.status === 'rejected',
        );
        if (failures.length > 0) {
          setError(
            failures.map((failure) => errorMessage(failure.reason)).join(' '),
          );
        }
      })
      .finally(() => active && setLoading(false));
    return () => {
      active = false;
    };
  }, [api, loadApprovalPolicy, loadModels, loadProjects, requestedSessionId]);

  useEffect(() => {
    if (!selectedProjectId) {
      taskListRequestRef.current += 1;
      activeTaskRequestRef.current += 1;
      setTasks([]);
      setSchedules([]);
      setExtensions(null);
      selectTask(null);
      setSelectedScheduleId(null);
      return;
    }
    setExtensions(null);
    Promise.all([
      loadTasks(selectedProjectId),
      loadSchedules(selectedProjectId),
      loadExtensions(selectedProjectId),
    ]).catch((loadError) => setError(errorMessage(loadError)));
  }, [loadExtensions, loadSchedules, loadTasks, selectTask, selectedProjectId]);

  useEffect(() => {
    if (!selectedTaskId) {
      activeTaskRequestRef.current += 1;
      setActiveTask(null);
      return;
    }
    loadTask(selectedTaskId).catch((loadError) =>
      setError(errorMessage(loadError)),
    );
  }, [loadTask, selectedTaskId]);

  useEffect(() => {
    if (!api) return;
    const taskEvents = createCoworkTaskEventBatch((events) => {
      setTasks((current) =>
        events.reduce(
          (next, event) =>
            event.projectId === selectedProjectIdRef.current
              ? mergeCoworkTaskEvent(next, event)
              : next,
          current,
        ),
      );
    });
    const unsubscribe = api.onTaskEvent((event) => {
      if (event.projectId !== selectedProjectIdRef.current) return;
      taskEvents.push(event);
      if (event.taskId !== selectedTaskIdRef.current) return;
      window.clearTimeout(refreshTimerRef.current);
      refreshTimerRef.current = window.setTimeout(() => {
        if (event.taskId === selectedTaskIdRef.current) {
          refreshActiveTask(event.taskId).catch((loadError) =>
            setError(errorMessage(loadError)),
          );
        }
      }, 80);
    });
    return () => {
      window.clearTimeout(refreshTimerRef.current);
      taskEvents.cancel();
      unsubscribe();
    };
  }, [api, refreshActiveTask]);

  useEffect(() => {
    if (!selectedProjectId || railMode !== 'schedules') return;
    const interval = window.setInterval(() => {
      loadSchedules(selectedProjectId, selectedScheduleId).catch((loadError) =>
        setError(errorMessage(loadError)),
      );
    }, 30_000);
    return () => window.clearInterval(interval);
  }, [loadSchedules, railMode, selectedProjectId, selectedScheduleId]);

  useEffect(() => {
    transcriptAutoScrollRef.current = true;
    setEarlierMessages([]);
    setHasEarlierMessages(false);
    setNextBeforeSequence(undefined);
    setLoadingEarlierMessages(false);
    transcriptScrollSnapshotRef.current = null;
  }, [selectedTaskId]);

  useEffect(() => {
    if (!activeTask || activeTask.id !== selectedTaskId) return;
    setHasEarlierMessages(Boolean(activeTask.hasEarlierMessages));
    setNextBeforeSequence((current) => {
      if (current !== undefined) return current;
      return activeTask.messages.reduce<number | undefined>(
        (oldest, message) =>
          message.sequence !== undefined &&
          (oldest === undefined || message.sequence < oldest)
            ? message.sequence
            : oldest,
        undefined,
      );
    });
  }, [activeTask?.hasEarlierMessages, activeTask?.id, selectedTaskId]);

  useLayoutEffect(() => {
    const snapshot = transcriptScrollSnapshotRef.current;
    const element = transcriptElementRef.current;
    if (!snapshot || !element) return;
    element.scrollTop =
      snapshot.scrollTop + (element.scrollHeight - snapshot.scrollHeight);
    transcriptScrollSnapshotRef.current = null;
  }, [earlierMessages]);

  useEffect(
    () => () => {
      transcriptObserverCleanupRef.current?.();
      transcriptObserverCleanupRef.current = null;
    },
    [],
  );

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key !== 'Escape') return;
      if (preview) {
        setPreview(null);
        return;
      }
      setRailOpen(false);
      setInspectorOpen(false);
    };
    window.addEventListener('keydown', handleEscape);
    return () => window.removeEventListener('keydown', handleEscape);
  }, [preview]);

  useEffect(() => {
    if (!transcriptAutoScrollRef.current) return;
    const frame = window.requestAnimationFrame(() => {
      const element = transcriptElementRef.current;
      if (
        !element ||
        !transcriptAutoScrollRef.current ||
        !activeTask ||
        activeTask.id !== selectedTaskId
      )
        return;
      scrollCoworkTranscriptToLatest(element);
    });
    return () => window.cancelAnimationFrame(frame);
  }, [
    activeTask?.id,
    activeTask?.error,
    activeTask?.messages.length,
    activeTask?.pendingApproval?.id,
    activeTask?.status,
    activeTask?.summary,
    selectedModelKey,
    selectedTaskId,
    taskNeedsRebind,
  ]);

  const runAction = async (name: string, action: () => Promise<void>) => {
    setBusy(name);
    setError(null);
    try {
      await action();
    } catch (actionError) {
      setError(errorMessage(actionError));
    } finally {
      setBusy(null);
    }
  };

  const handleCreateProject = (event: FormEvent) => {
    event.preventDefault();
    if (!api || !projectName.trim()) return;
    void runAction('create-project', async () => {
      const project = await api.createProject({
        name: projectName.trim(),
        instructions: projectInstructions.trim(),
      });
      if (!project) return;
      setProjectName('');
      setProjectInstructions('');
      setShowProjectForm(false);
      await loadProjects();
      setSelectedProjectId(project.id);
      selectTask(null);
      setEditingProjectId(null);
    });
  };

  const handleApprovalMode = (nextMode: CoworkApprovalMode) => {
    if (!api || !approvalPolicy || busy === 'approval-policy') return;
    const previous = approvalPolicy;
    setApprovalPolicy({ ...previous, mode: nextMode });
    setProjects((current) =>
      current.map((project) => ({ ...project, approvalMode: nextMode })),
    );
    void runAction('approval-policy', async () => {
      try {
        const updated = await api.updateApprovalPolicy(
          nextMode,
          previous.revision,
        );
        setApprovalPolicy(updated);
      } catch (updateError) {
        const latest = await api.getApprovalPolicy().catch(() => previous);
        setApprovalPolicy(latest);
        setProjects((current) =>
          current.map((project) => ({
            ...project,
            approvalMode: latest.mode,
          })),
        );
        throw updateError;
      }
    });
  };

  const handleEditProject = () => {
    if (!selectedProject) return;
    setEditingProjectId(selectedProject.id);
    setProjectSettingsName(selectedProject.name);
    setProjectSettingsInstructions(selectedProject.instructions);
  };

  const handleSaveProject = (event: FormEvent) => {
    event.preventDefault();
    if (
      !api ||
      !selectedProject ||
      editingProjectId !== selectedProject.id ||
      !projectSettingsName.trim()
    ) {
      return;
    }
    void runAction('update-project', async () => {
      const updated = await api.updateProject({
        id: selectedProject.id,
        name: projectSettingsName.trim(),
        instructions: projectSettingsInstructions.trim(),
      });
      setProjects((current) =>
        current.map((project) =>
          project.id === updated.id ? updated : project,
        ),
      );
      setEditingProjectId(null);
    });
  };

  const handleDeleteProject = () => {
    if (!api || !selectedProject) return;
    if (
      !window.confirm(
        `Archive “${selectedProject.name}”? Its task history will be kept.`,
      )
    ) {
      return;
    }
    void runAction('delete-project', async () => {
      await api.deleteProject(selectedProject.id);
      setSelectedProjectId(null);
      selectTask(null);
      setSelectedScheduleId(null);
      setEditingScheduleId(null);
      setEditingProjectId(null);
      await loadProjects();
    });
  };

  const handleCreateTask = (event: FormEvent) => {
    event.preventDefault();
    if (!api || !selectedProject || !selectedModel || !taskGoal.trim()) return;
    void runAction('create-task', async () => {
      const task = await api.createTask({
        projectId: selectedProject.id,
        title: taskTitle.trim(),
        goal: taskGoal.trim(),
        model: selectedModel,
      });
      setTaskTitle('');
      setTaskGoal('');
      selectTask(task.id);
      setActiveTask(task);
      await loadTasks(selectedProject.id, task.id);
      await api.startTask(task.id);
      await loadTask(task.id);
    });
  };

  const handleStartTask = () => {
    if (!api || !activeTask) return;
    if (!exactTaskSessionActive) {
      setError(
        'This task’s original session has ended. Choose an active session and continue the task before running it.',
      );
      return;
    }
    void runAction('start-task', async () => {
      await api.startTask(activeTask.id);
      await loadTask(activeTask.id);
    });
  };

  const handlePauseTask = () => {
    if (!api || !activeTask) return;
    void runAction('pause-task', async () => {
      await api.pauseTask(activeTask.id);
      await loadTask(activeTask.id);
    });
  };

  const handleCancelTask = () => {
    if (!api || !activeTask) return;
    if (
      !window.confirm(
        `Cancel “${activeTask.title}”? Its messages and completed work will remain available.`,
      )
    ) {
      return;
    }
    void runAction('cancel-task', async () => {
      await api.cancelTask(activeTask.id);
      await loadTask(activeTask.id);
    });
  };

  const handleDeleteTask = () => {
    if (!api || !activeTask || !selectedProject) return;
    if (!window.confirm(`Delete “${activeTask.title}” from task history?`))
      return;
    void runAction('delete-task', async () => {
      await api.deleteTask(activeTask.id);
      selectTask(null);
      setActiveTask(null);
      await loadTasks(selectedProject.id);
    });
  };

  const handleSteer = (event: FormEvent) => {
    event.preventDefault();
    if (
      !api ||
      !activeTask ||
      !exactTaskSessionActive ||
      !steeringMessage.trim()
    )
      return;
    const content = steeringMessage.trim();
    setSteeringMessage('');
    void runAction('steer-task', async () => {
      await api.steerTask(activeTask.id, content);
      await loadTask(activeTask.id);
    });
  };

  const handleApproval = (approved: boolean) => {
    if (
      !api ||
      !activeTask?.pendingApproval ||
      (approved && !exactTaskSessionActive)
    )
      return;
    const taskId = activeTask.id;
    const approvalId = activeTask.pendingApproval.id;
    const submissionKey = `${taskId}:${approvalId}`;
    if (approvalSubmissionRef.current === submissionKey) return;
    approvalSubmissionRef.current = submissionKey;
    void runAction('approval', async () => {
      try {
        const authoritativeTask = await api.resolveApproval(
          taskId,
          approvalId,
          approved,
        );
        if (selectedTaskIdRef.current === taskId) {
          setActiveTask(authoritativeTask);
        }
      } catch (approvalError) {
        // A transport failure must not leave a clickable stale card behind.
        await loadTask(taskId).catch(() => undefined);
        throw approvalError;
      } finally {
        if (approvalSubmissionRef.current === submissionKey) {
          approvalSubmissionRef.current = null;
        }
      }
    });
  };

  const handleRebindTask = () => {
    if (!api || !activeTask || !selectedModel) return;
    void runAction('rebind-task', async () => {
      await api.rebindTask(activeTask.id, selectedModel);
      await loadTask(activeTask.id);
      if (selectedProject) {
        await loadTasks(selectedProject.id, activeTask.id);
      }
    });
  };

  const handleLoadEarlierMessages = async () => {
    if (!api || !activeTask || loadingEarlierMessages) return;
    const taskId = activeTask.id;
    setLoadingEarlierMessages(true);
    setError(null);
    try {
      const page = await api.listTaskMessages(taskId, nextBeforeSequence, 50);
      if (selectedTaskIdRef.current !== taskId) return;

      const transcript = transcriptElementRef.current;
      if (transcript) {
        transcriptAutoScrollRef.current = false;
        transcriptScrollSnapshotRef.current = {
          scrollHeight: transcript.scrollHeight,
          scrollTop: transcript.scrollTop,
        };
      }

      setEarlierMessages((current) => {
        const knownIds = new Set([
          ...current.map((message) => message.id),
          ...(activeTask.messages ?? []).map((message) => message.id),
        ]);
        const additions = page.messages.filter(
          (message) => !knownIds.has(message.id),
        );
        return [...additions, ...current];
      });
      setHasEarlierMessages(page.hasMore);
      setNextBeforeSequence(page.nextBeforeSequence);
    } catch (loadError) {
      setError(errorMessage(loadError));
      transcriptScrollSnapshotRef.current = null;
    } finally {
      if (selectedTaskIdRef.current === taskId) {
        setLoadingEarlierMessages(false);
      }
    }
  };

  const handlePreviewArtifact = (artifact: CoworkArtifact) => {
    if (!api || !activeTask || artifact.kind !== 'file') return;
    void runAction(`preview:${artifact.path}`, async () => {
      setPreview(await api.previewArtifact(activeTask.id, artifact.path));
    });
  };

  const handleRevealArtifact = (artifact: CoworkArtifact) => {
    if (!api || !activeTask) return;
    void runAction(`reveal:${artifact.path}`, async () => {
      await api.revealArtifact(activeTask.id, artifact.path);
    });
  };

  const resetScheduleForm = () => {
    setEditingScheduleId(null);
    setScheduleName('');
    setScheduleTitle('');
    setScheduleGoal('');
    setScheduleCadence('weekdays');
    setScheduleTime('09:00');
    setScheduleHourlyMinute(0);
    setScheduleWeekday(1);
    setScheduleTimeZone(localTimeZone);
    setScheduleStatus('active');
  };

  const handleEditSchedule = () => {
    if (!selectedSchedule) return;
    const cadence = selectedSchedule.cadence;
    setEditingScheduleId(selectedSchedule.id);
    setScheduleName(selectedSchedule.name);
    setScheduleTitle(selectedSchedule.task.title);
    setScheduleGoal(selectedSchedule.task.goal);
    setScheduleCadence(cadence.kind);
    setScheduleTimeZone(selectedSchedule.timeZone);
    setScheduleStatus(selectedSchedule.status);
    if (cadence.kind === 'hourly') {
      setScheduleHourlyMinute(cadence.minute);
    } else if (cadence.kind !== 'manual') {
      setScheduleTime(
        `${String(cadence.hour).padStart(2, '0')}:${String(cadence.minute).padStart(2, '0')}`,
      );
      if (cadence.kind === 'weekly') {
        setScheduleWeekday(cadence.dayOfWeek);
      }
    }
    const currentModelKey = modelKey(selectedSchedule.task.model);
    setSelectedModelKey(
      activeModels.some((model) => modelKey(model) === currentModelKey)
        ? currentModelKey
        : '',
    );
  };

  const handleSaveSchedule = (event: FormEvent) => {
    event.preventDefault();
    if (
      !api ||
      !selectedProject ||
      !selectedModel ||
      !scheduleName.trim() ||
      !scheduleGoal.trim()
    ) {
      return;
    }
    const cadence = cadenceFromForm(
      scheduleCadence,
      scheduleTime,
      scheduleHourlyMinute,
      scheduleWeekday,
    );
    const action = editingScheduleId ? 'update-schedule' : 'create-schedule';
    void runAction(action, async () => {
      const schedule = editingScheduleId
        ? await api.updateSchedule({
            id: editingScheduleId,
            name: scheduleName.trim(),
            task: {
              title: scheduleTitle.trim(),
              goal: scheduleGoal.trim(),
              model: selectedModel,
            },
            cadence,
            timeZone: scheduleTimeZone.trim() || localTimeZone,
          })
        : await api.createSchedule({
            projectId: selectedProject.id,
            name: scheduleName.trim(),
            title: scheduleTitle.trim(),
            goal: scheduleGoal.trim(),
            model: selectedModel,
            cadence,
            timeZone: scheduleTimeZone.trim() || localTimeZone,
            status: scheduleStatus,
          });
      setSelectedScheduleId(schedule.id);
      resetScheduleForm();
      await loadSchedules(selectedProject.id, schedule.id);
    });
  };

  const handleScheduleStatus = () => {
    if (!api || !selectedSchedule || !selectedProject) return;
    const shouldResume = selectedSchedule.status === 'paused';
    if (shouldResume && !selectedModel) return;
    void runAction('schedule-status', async () => {
      if (shouldResume) await api.resumeSchedule(selectedSchedule.id);
      else await api.pauseSchedule(selectedSchedule.id);
      await loadSchedules(selectedProject.id, selectedSchedule.id);
    });
  };

  const handleRunSchedule = () => {
    if (!api || !selectedSchedule || !selectedProject || !selectedModel) return;
    void runAction('run-schedule', async () => {
      const updated = await api.runScheduleNow(selectedSchedule.id);
      await Promise.all([
        loadSchedules(selectedProject.id, selectedSchedule.id),
        loadTasks(selectedProject.id, updated.lastTaskId),
      ]);
    });
  };

  const handleDeleteSchedule = () => {
    if (!api || !selectedSchedule || !selectedProject) return;
    if (!window.confirm(`Delete schedule “${selectedSchedule.name}”?`)) return;
    void runAction('delete-schedule', async () => {
      await api.deleteSchedule(selectedSchedule.id);
      setSelectedScheduleId(null);
      setEditingScheduleId(null);
      await loadSchedules(selectedProject.id);
    });
  };

  const handleOpenScheduledTask = () => {
    if (!selectedSchedule?.lastTaskId) return;
    setRailMode('tasks');
    selectTask(selectedSchedule.lastTaskId);
  };

  const handleConfigureExtensions = (
    folderInstructionsEnabled: boolean,
    enabledSkillIds: string[],
  ) => {
    if (!api || !selectedProject) return;
    void runAction('configure-extensions', async () => {
      const catalog = await api.configureExtensions(selectedProject.id, {
        folderInstructionsEnabled,
        enabledSkillIds,
      });
      setExtensions(catalog);
      await loadProjects();
    });
  };

  const handleFolderInstructionsToggle = (enabled: boolean) => {
    if (!extensions) return;
    handleConfigureExtensions(
      enabled,
      extensions.skills
        .filter((skill) => skill.enabled)
        .map((skill) => skill.id),
    );
  };

  const handleSkillToggle = (skillId: string, enabled: boolean) => {
    if (!extensions) return;
    const enabledSkillIds = extensions.skills
      .filter((skill) => (skill.id === skillId ? enabled : skill.enabled))
      .map((skill) => skill.id);
    handleConfigureExtensions(
      extensions.projectInstructions?.enabled ?? false,
      enabledSkillIds,
    );
  };

  const chooseActiveModel = (nextModelKey: string) => {
    const nextModel = activeModels.find(
      (model) => modelKey(model) === nextModelKey,
    );
    if (!nextModel) return;
    setSelectedModelKey(nextModelKey);
    if (requestedSessionId && nextModel.sessionId !== requestedSessionId) {
      navigate('/workspace', { replace: true });
    }
  };

  const replacementIsActive = activeModels.some(
    (model) => modelKey(model) === replacementModelKey,
  );
  const requestedSessionUnavailable = Boolean(
    requestedSessionId &&
    !activeModels.some((model) => model.sessionId === requestedSessionId),
  );
  const endedSelection = Boolean(requestedSessionId || selectedModelKey);

  const projectSettingsBusy =
    busy === 'approval-policy' || busy === 'update-project';

  if (!api) {
    return (
      <main className="cowork-unavailable">
        <IconRobot size={34} />
        <h1>Workspace is not available</h1>
        <p>
          Restart Morpheus after the Workspace runtime has been installed. No
          file access is enabled in this window.
        </p>
      </main>
    );
  }

  if (loading) {
    return (
      <main
        className="cowork-unavailable"
        aria-label="Loading Morpheus Workspace"
      >
        <IconLoader2 className="cowork-spin" size={34} />
        <h1>Loading Morpheus Workspace</h1>
        <p>Restoring your projects, task history, and active sessions.</p>
      </main>
    );
  }

  return (
    <main
      aria-label="Morpheus Workspace"
      className="cowork-shell"
      data-inspector-open={inspectorOpen}
      data-rail-open={railOpen}
    >
      <button
        aria-label="Close Workspace panels"
        className="cowork-scrim"
        onClick={() => {
          setRailOpen(false);
          setInspectorOpen(false);
        }}
        type="button"
      />
      <aside
        aria-label="Workspace navigation"
        className="cowork-rail"
        id="workspace-navigation"
      >
        <div className="cowork-brand-row">
          <span className="cowork-brand-icon">
            <IconSparkles size={18} />
          </span>
          <div>
            <strong>Workspace</strong>
            <span>Projects and task history</span>
          </div>
          <button
            aria-label="Close project navigation"
            className="cowork-icon-button cowork-panel-close"
            onClick={() => setRailOpen(false)}
            type="button"
          >
            <IconX size={18} />
          </button>
        </div>

        <section className="cowork-rail-section cowork-projects-section">
          <div className="cowork-section-heading">
            <span>Projects</span>
            <button
              aria-label="Create project"
              className="cowork-icon-button"
              onClick={() => setShowProjectForm((current) => !current)}
              title="Connect a folder"
              type="button"
            >
              {showProjectForm ? (
                <IconX size={17} />
              ) : (
                <IconFolderPlus size={17} />
              )}
            </button>
          </div>

          {showProjectForm && (
            <form
              className="cowork-project-form"
              onSubmit={handleCreateProject}
            >
              <label>
                Project name
                <input
                  autoFocus
                  onChange={(event) => setProjectName(event.target.value)}
                  placeholder="Website redesign"
                  value={projectName}
                />
              </label>
              <label>
                Project instructions <span>optional</span>
                <textarea
                  onChange={(event) =>
                    setProjectInstructions(event.target.value)
                  }
                  placeholder="Conventions, constraints, and context…"
                  rows={3}
                  value={projectInstructions}
                />
              </label>
              <button
                className="cowork-primary-button cowork-full-button"
                disabled={busy === 'create-project' || !projectName.trim()}
                type="submit"
              >
                {busy === 'create-project' ? (
                  <IconLoader2 className="cowork-spin" size={16} />
                ) : (
                  <IconFolder size={16} />
                )}
                Choose folder & create
              </button>
              <p className="cowork-form-hint">
                Workspace can only access the folder you choose. File approval
                policy is shared across every project.
              </p>
            </form>
          )}

          <div className="cowork-project-list">
            {projects.map((project) => (
              <button
                aria-current={
                  project.id === selectedProjectId ? 'page' : undefined
                }
                className={`cowork-project-row ${
                  project.id === selectedProjectId ? 'is-active' : ''
                }`}
                key={project.id}
                onClick={() => {
                  setSelectedProjectId(project.id);
                  selectTask(null);
                  setSelectedScheduleId(null);
                  setEditingScheduleId(null);
                  setEditingProjectId(null);
                  setRailOpen(false);
                }}
                type="button"
              >
                <IconFolder size={17} />
                <span>
                  <strong>{project.name}</strong>
                  <small>{project.folderName || 'Connected folder'}</small>
                </span>
                <IconChevronRight className="cowork-row-arrow" size={15} />
              </button>
            ))}
            {!loading && projects.length === 0 && !showProjectForm && (
              <p className="cowork-rail-empty">
                Connect a folder to begin a project.
              </p>
            )}
          </div>
        </section>

        {selectedProject && (
          <section className="cowork-rail-section cowork-tasks-section">
            <div className="cowork-rail-switch" role="tablist">
              <button
                aria-controls="workspace-rail-panel"
                aria-selected={railMode === 'tasks'}
                className={railMode === 'tasks' ? 'is-active' : ''}
                id="workspace-tasks-tab"
                onClick={() => setRailMode('tasks')}
                role="tab"
                type="button"
              >
                Tasks
                <span>{tasks.length}</span>
              </button>
              <button
                aria-controls="workspace-rail-panel"
                aria-selected={railMode === 'schedules'}
                className={railMode === 'schedules' ? 'is-active' : ''}
                id="workspace-schedules-tab"
                onClick={() => setRailMode('schedules')}
                role="tab"
                type="button"
              >
                Schedules
                <span>{schedules.length}</span>
              </button>
            </div>
            <div className="cowork-section-heading">
              <span>
                {railMode === 'tasks' ? 'Recent tasks' : 'Automations'}
              </span>
              <div className="cowork-heading-actions">
                {railMode === 'schedules' && (
                  <button
                    aria-label="Refresh schedules"
                    className="cowork-icon-button"
                    disabled={busy === 'refresh-schedules'}
                    onClick={() =>
                      void runAction('refresh-schedules', async () =>
                        loadSchedules(selectedProject.id, selectedScheduleId),
                      )
                    }
                    title="Refresh schedules"
                    type="button"
                  >
                    <IconRefresh
                      className={
                        busy === 'refresh-schedules' ? 'cowork-spin' : ''
                      }
                      size={16}
                    />
                  </button>
                )}
                <button
                  aria-label={
                    railMode === 'tasks' ? 'New task' : 'New schedule'
                  }
                  className="cowork-icon-button"
                  onClick={() => {
                    if (railMode === 'tasks') {
                      selectTask(null);
                    } else {
                      resetScheduleForm();
                      setSelectedScheduleId(null);
                    }
                  }}
                  title={railMode === 'tasks' ? 'New task' : 'New schedule'}
                  type="button"
                >
                  <IconPlus size={17} />
                </button>
              </div>
            </div>
            <div
              aria-labelledby={
                railMode === 'tasks'
                  ? 'workspace-tasks-tab'
                  : 'workspace-schedules-tab'
              }
              className="cowork-task-list"
              id="workspace-rail-panel"
              role="tabpanel"
            >
              {railMode === 'tasks' ? (
                <>
                  {tasks.map((task) => (
                    <button
                      aria-current={
                        task.id === selectedTaskId ? 'true' : undefined
                      }
                      className={`cowork-task-row ${
                        task.id === selectedTaskId ? 'is-active' : ''
                      }`}
                      key={task.id}
                      onClick={() => {
                        selectTask(task.id);
                        setRailOpen(false);
                      }}
                      type="button"
                    >
                      <span
                        aria-hidden="true"
                        className={`cowork-task-status status-${task.status}`}
                      >
                        <TaskStatusIcon status={task.status} />
                      </span>
                      <span>
                        <strong>{task.title}</strong>
                        <small className="cowork-task-meta">
                          <span>{STATUS_LABELS[task.status]}</span>
                          <time
                            dateTime={new Date(task.updatedAt).toISOString()}
                          >
                            {formatRelativeTime(task.updatedAt)}
                          </time>
                        </small>
                      </span>
                    </button>
                  ))}
                  {tasks.length === 0 && (
                    <p className="cowork-rail-empty">No tasks yet.</p>
                  )}
                </>
              ) : (
                <>
                  {schedules.map((schedule) => (
                    <button
                      aria-current={
                        schedule.id === selectedScheduleId ? 'true' : undefined
                      }
                      className={`cowork-task-row ${
                        schedule.id === selectedScheduleId ? 'is-active' : ''
                      }`}
                      key={schedule.id}
                      onClick={() => {
                        setEditingScheduleId(null);
                        setSelectedScheduleId(schedule.id);
                        setRailOpen(false);
                      }}
                      type="button"
                    >
                      <span
                        aria-hidden="true"
                        className={`cowork-task-status schedule-${schedule.status}`}
                      >
                        <ScheduleStatusIcon schedule={schedule} />
                      </span>
                      <span>
                        <strong>{schedule.name}</strong>
                        <small className="cowork-task-meta">
                          <span>
                            {schedule.runningSince
                              ? 'Running'
                              : schedule.status === 'active'
                                ? 'Scheduled'
                                : 'Paused'}
                          </span>
                          <span>{cadenceLabel(schedule.cadence)}</span>
                        </small>
                      </span>
                    </button>
                  ))}
                  {schedules.length === 0 && (
                    <p className="cowork-rail-empty">No schedules yet.</p>
                  )}
                </>
              )}
            </div>
          </section>
        )}

        {selectedProject && (
          <div className="cowork-project-controls">
            <div className="cowork-project-settings-heading">
              <strong>Project settings</strong>
              <button
                aria-label={
                  editingProjectId === selectedProject.id
                    ? 'Close project editor'
                    : 'Edit project'
                }
                className="cowork-icon-button"
                disabled={projectSettingsBusy}
                onClick={() =>
                  editingProjectId === selectedProject.id
                    ? setEditingProjectId(null)
                    : handleEditProject()
                }
                title={
                  editingProjectId === selectedProject.id
                    ? 'Close project editor'
                    : 'Edit project'
                }
                type="button"
              >
                {editingProjectId === selectedProject.id ? (
                  <IconX size={15} />
                ) : (
                  <IconPencil size={15} />
                )}
              </button>
            </div>
            {editingProjectId === selectedProject.id && (
              <form
                className="cowork-project-settings-form"
                onSubmit={handleSaveProject}
              >
                <label>
                  Project name
                  <input
                    autoFocus
                    disabled={projectSettingsBusy}
                    onChange={(event) =>
                      setProjectSettingsName(event.target.value)
                    }
                    value={projectSettingsName}
                  />
                </label>
                <label>
                  Project instructions <span>optional</span>
                  <textarea
                    disabled={projectSettingsBusy}
                    onChange={(event) =>
                      setProjectSettingsInstructions(event.target.value)
                    }
                    placeholder="Conventions, constraints, and context…"
                    rows={4}
                    value={projectSettingsInstructions}
                  />
                </label>
                <button
                  className="cowork-primary-button cowork-full-button"
                  disabled={projectSettingsBusy || !projectSettingsName.trim()}
                  type="submit"
                >
                  {busy === 'update-project' ? (
                    <IconLoader2 className="cowork-spin" size={15} />
                  ) : (
                    <IconCheck size={15} />
                  )}
                  Save project
                </button>
              </form>
            )}
            <button
              className="cowork-text-button danger"
              disabled={busy === 'delete-project'}
              onClick={handleDeleteProject}
              type="button"
            >
              Archive project
            </button>
          </div>
        )}

        {approvalPolicy && (
          <div className="cowork-project-controls">
            <div className="cowork-project-settings-heading">
              <strong>Workspace settings</strong>
            </div>
            <label>
              File approval policy
              <select
                aria-label="Global Workspace approval policy"
                disabled={busy === 'approval-policy'}
                onChange={(event) =>
                  handleApprovalMode(event.target.value as CoworkApprovalMode)
                }
                value={approvalMode}
              >
                <option value="manual">Manual</option>
                <option value="auto">Auto</option>
                <option value="skip">Skip</option>
              </select>
            </label>
            <span>
              {APPROVAL_DESCRIPTIONS[approvalMode]}. Applies to every existing
              and future Workspace task.
            </span>
          </div>
        )}
      </aside>

      <section className="cowork-workspace">
        <header className="cowork-workspace-header">
          <div className="cowork-header-main">
            <button
              aria-controls="workspace-navigation"
              aria-expanded={railOpen}
              aria-label={
                railOpen
                  ? 'Close project navigation'
                  : 'Open project navigation'
              }
              className="cowork-icon-button cowork-rail-toggle"
              onClick={() => {
                setInspectorOpen(false);
                setRailOpen((current) => !current);
              }}
              type="button"
            >
              <IconMenu2 size={19} />
            </button>
            <div className="cowork-workspace-title">
              <span className="cowork-eyebrow">
                {selectedProject?.folderName || 'Local workspace'}
              </span>
              <h1>
                {railMode === 'schedules'
                  ? selectedSchedule?.name || 'New schedule'
                  : activeTask?.title || selectedProject?.name || 'Workspace'}
              </h1>
            </div>
          </div>
          <div className="cowork-header-actions">
            {railMode === 'schedules' && selectedSchedule && (
              <span
                className={`cowork-status schedule-${selectedSchedule.status}`}
              >
                {selectedSchedule.runningSince && (
                  <IconLoader2 className="cowork-spin" size={14} />
                )}
                {selectedSchedule.runningSince
                  ? 'Running'
                  : selectedSchedule.status === 'active'
                    ? 'Scheduled'
                    : 'Paused'}
              </span>
            )}
            {railMode === 'tasks' && activeTask && (
              <span className={`cowork-status status-${activeTask.status}`}>
                {activeTask.status === 'running' && (
                  <IconLoader2 className="cowork-spin" size={14} />
                )}
                {STATUS_LABELS[activeTask.status]}
              </span>
            )}
            {railMode === 'tasks' &&
              activeTask &&
              ['queued', 'paused', 'failed', 'cancelled'].includes(
                activeTask.status,
              ) && (
                <button
                  aria-label="Run task"
                  className="cowork-secondary-button"
                  disabled={busy === 'start-task' || !exactTaskSessionActive}
                  onClick={handleStartTask}
                  type="button"
                >
                  <IconPlayerPlay size={15} />
                  <span className="cowork-action-label">Run</span>
                </button>
              )}
            {railMode === 'tasks' &&
              activeTask &&
              ['running', 'waiting_approval'].includes(activeTask.status) && (
                <button
                  aria-label="Pause task"
                  className="cowork-secondary-button"
                  disabled={busy === 'pause-task'}
                  onClick={handlePauseTask}
                  type="button"
                >
                  <IconPlayerPause size={15} />
                  <span className="cowork-action-label">Pause</span>
                </button>
              )}
            {railMode === 'tasks' &&
              activeTask &&
              ['running', 'waiting_approval', 'paused'].includes(
                activeTask.status,
              ) && (
                <button
                  aria-label="Cancel task"
                  className="cowork-secondary-button cowork-danger-button"
                  disabled={busy === 'cancel-task'}
                  onClick={handleCancelTask}
                  type="button"
                >
                  <IconPlayerStop size={15} />
                  <span className="cowork-action-label">Cancel</span>
                </button>
              )}
            {railMode === 'tasks' && activeTask && (
              <button
                aria-label="Delete task"
                className="cowork-icon-button danger"
                disabled={busy === 'delete-task'}
                onClick={handleDeleteTask}
                title="Delete task"
                type="button"
              >
                <IconTrash size={16} />
              </button>
            )}
            <button
              aria-controls="workspace-inspector"
              aria-expanded={inspectorOpen}
              aria-label={`${inspectorOpen ? 'Close' : 'Open'} ${
                railMode === 'schedules' ? 'schedule' : 'task'
              } details`}
              className="cowork-icon-button cowork-inspector-toggle"
              onClick={() => {
                setRailOpen(false);
                setInspectorOpen((current) => !current);
              }}
              type="button"
            >
              <IconLayoutSidebarRight size={18} />
            </button>
          </div>
        </header>

        {!selectedModel && (
          <section className="cowork-session-banner" role="status">
            <IconCloud size={18} />
            <div className="cowork-session-copy">
              <strong>
                {requestedSessionUnavailable
                  ? 'That session is no longer active'
                  : endedSelection
                    ? 'Your Workspace session has ended'
                    : 'Open a session to start or continue work'}
              </strong>
              <span>
                Project history remains available. Session-bound actions stay
                paused until you choose an active Morpheus session.
              </span>
            </div>
            <div className="cowork-session-actions">
              {activeModels.length > 0 ? (
                <>
                  <select
                    aria-label="Choose another active Workspace session"
                    onChange={(event) =>
                      setReplacementModelKey(event.target.value)
                    }
                    value={replacementModelKey}
                  >
                    <option value="">Choose active session</option>
                    {modelGroups.map((group) => (
                      <optgroup key={group.label} label={group.label}>
                        {group.options.map((model) => (
                          <option key={modelKey(model)} value={modelKey(model)}>
                            {model.modelName}
                          </option>
                        ))}
                      </optgroup>
                    ))}
                  </select>
                  <button
                    className="cowork-primary-button"
                    disabled={!replacementIsActive}
                    onClick={() => {
                      chooseActiveModel(replacementModelKey);
                      setReplacementModelKey('');
                    }}
                    type="button"
                  >
                    Use session
                  </button>
                </>
              ) : (
                <button
                  className="cowork-primary-button"
                  onClick={() => navigate('/chat?setup=workspace')}
                  type="button"
                >
                  <IconExternalLink size={16} />
                  Open session in Chat
                </button>
              )}
            </div>
          </section>
        )}

        {error && (
          <div className="cowork-error-banner" role="alert">
            <IconAlertTriangle size={17} />
            <span>{error}</span>
            <button onClick={() => setError(null)} type="button">
              <IconX size={15} />
            </button>
          </div>
        )}

        {!selectedProject ? (
          <div className="cowork-stage cowork-welcome-stage">
            <div className="cowork-welcome-card">
              <span className="cowork-hero-icon">
                <IconSparkles size={30} />
              </span>
              <span className="cowork-eyebrow">Morpheus Workspace</span>
              <h2>Give an AI agent a goal, not a checklist.</h2>
              <p>
                Connect a project folder, choose one of your active Morpheus
                sessions, and review the agent’s plan and file changes as it
                works. Existing projects remain available between sessions.
              </p>
              <button
                className="cowork-primary-button"
                onClick={() => setShowProjectForm(true)}
                type="button"
              >
                <IconFolderPlus size={17} />
                Connect a folder
              </button>
              <div className="cowork-safety-row">
                <IconShieldCheck size={17} />
                Folder-scoped access · explicit approvals · visible activity
              </div>
            </div>
          </div>
        ) : railMode === 'schedules' ? (
          <div className="cowork-stage cowork-schedule-stage">
            {selectedSchedule && editingScheduleId !== selectedSchedule.id ? (
              <article className="cowork-schedule-card">
                <header className="cowork-schedule-card-header">
                  <span className="cowork-hero-icon compact">
                    <IconCalendarTime size={24} />
                  </span>
                  <div>
                    <span className="cowork-eyebrow">Schedule</span>
                    <h2>{selectedSchedule.name}</h2>
                    <p>{cadenceLabel(selectedSchedule.cadence)}</p>
                  </div>
                </header>

                <div className="cowork-schedule-actions">
                  <button
                    className="cowork-primary-button"
                    disabled={
                      busy === 'run-schedule' ||
                      !!selectedSchedule.runningSince ||
                      !selectedModel
                    }
                    onClick={handleRunSchedule}
                    type="button"
                  >
                    {busy === 'run-schedule' ||
                    selectedSchedule.runningSince ? (
                      <IconLoader2 className="cowork-spin" size={16} />
                    ) : (
                      <IconPlayerPlay size={16} />
                    )}
                    {selectedSchedule.runningSince ? 'Running' : 'Run now'}
                  </button>
                  <button
                    className="cowork-secondary-button"
                    disabled={
                      busy === 'schedule-status' ||
                      (selectedSchedule.status === 'paused' && !selectedModel)
                    }
                    onClick={handleScheduleStatus}
                    type="button"
                  >
                    {selectedSchedule.status === 'paused' ? (
                      <IconPlayerPlay size={16} />
                    ) : (
                      <IconPlayerPause size={16} />
                    )}
                    {selectedSchedule.status === 'paused' ? 'Resume' : 'Pause'}
                  </button>
                  <button
                    className="cowork-secondary-button"
                    disabled={busy === 'update-schedule'}
                    onClick={handleEditSchedule}
                    type="button"
                  >
                    <IconPencil size={15} />
                    Edit
                  </button>
                  <button
                    aria-label="Delete schedule"
                    className="cowork-icon-button danger"
                    disabled={busy === 'delete-schedule'}
                    onClick={handleDeleteSchedule}
                    title="Delete schedule"
                    type="button"
                  >
                    <IconTrash size={16} />
                  </button>
                </div>

                <dl className="cowork-schedule-stats">
                  <div>
                    <dt>Next run</dt>
                    <dd>
                      {selectedSchedule.cadence.kind === 'manual'
                        ? 'Manual only'
                        : selectedSchedule.status === 'paused'
                          ? 'Paused'
                          : formatScheduleTime(
                              selectedSchedule.nextRunAt,
                              selectedSchedule.timeZone,
                            )}
                    </dd>
                  </div>
                  <div>
                    <dt>Last run</dt>
                    <dd>
                      {formatScheduleTime(
                        selectedSchedule.lastRunAt,
                        selectedSchedule.timeZone,
                      )}
                    </dd>
                  </div>
                  <div>
                    <dt>Time zone</dt>
                    <dd>{selectedSchedule.timeZone}</dd>
                  </div>
                  <div>
                    <dt>Status</dt>
                    <dd>
                      {selectedSchedule.runningSince
                        ? 'Running now'
                        : selectedSchedule.status === 'active'
                          ? 'Active'
                          : 'Paused'}
                    </dd>
                  </div>
                </dl>

                <section className="cowork-schedule-task">
                  <span className="cowork-eyebrow">Task template</span>
                  <h3>
                    {selectedSchedule.task.title || selectedSchedule.name}
                  </h3>
                  <p>{selectedSchedule.task.goal}</p>
                  <span className="cowork-model-chip">
                    {selectedSchedule.task.model.dataBoundary ===
                    'on-device' ? (
                      <IconDeviceLaptop size={14} />
                    ) : (
                      <IconCloud size={14} />
                    )}
                    {selectedSchedule.task.model.modelName}
                  </span>
                </section>

                {selectedSchedule.task.model.dataBoundary !== 'on-device' && (
                  <div className="cowork-schedule-disclosure">
                    <IconShieldCheck size={18} />
                    <div>
                      <strong>Approval is still required</strong>
                      <p>
                        Each scheduled run creates a task waiting for your
                        data-sharing approval. Nothing is sent to the configured
                        endpoint or provider until you approve it.
                      </p>
                    </div>
                  </div>
                )}

                {selectedSchedule.lastError && (
                  <div className="cowork-inline-error">
                    <IconAlertTriangle size={17} />
                    <div>
                      <strong>Last run failed</strong>
                      <span>{selectedSchedule.lastError}</span>
                    </div>
                  </div>
                )}

                {selectedSchedule.lastTaskId && (
                  <button
                    className="cowork-schedule-task-link"
                    onClick={handleOpenScheduledTask}
                    type="button"
                  >
                    View the most recent task
                    <IconChevronRight size={16} />
                  </button>
                )}
              </article>
            ) : (
              <form
                className="cowork-new-task-card cowork-schedule-form"
                onSubmit={handleSaveSchedule}
              >
                <span className="cowork-hero-icon compact">
                  <IconCalendarTime size={24} />
                </span>
                <span className="cowork-eyebrow">
                  {editingScheduleId ? 'Edit schedule' : 'New schedule'}
                </span>
                <h2>
                  {editingScheduleId
                    ? 'Update this recurring task'
                    : 'Run a Workspace task automatically'}
                </h2>

                <div className="cowork-form-grid">
                  <label>
                    Schedule name
                    <input
                      autoFocus
                      onChange={(event) => setScheduleName(event.target.value)}
                      placeholder="Weekday project summary"
                      required
                      value={scheduleName}
                    />
                  </label>
                  <label>
                    Task title <span>optional</span>
                    <input
                      onChange={(event) => setScheduleTitle(event.target.value)}
                      placeholder="Prepare project summary"
                      value={scheduleTitle}
                    />
                  </label>
                </div>

                <label>
                  Goal
                  <textarea
                    onChange={(event) => setScheduleGoal(event.target.value)}
                    placeholder="Describe the recurring outcome and any constraints…"
                    required
                    rows={5}
                    value={scheduleGoal}
                  />
                </label>

                <div className="cowork-model-row">
                  <label>
                    Model
                    <select
                      disabled={activeModels.length === 0}
                      onChange={(event) =>
                        chooseActiveModel(event.target.value)
                      }
                      value={selectedModelKey}
                    >
                      {activeModels.length === 0 && (
                        <option value="">No models available</option>
                      )}
                      {activeModels.length > 0 && !selectedModelKey && (
                        <option value="">Choose an available model</option>
                      )}
                      {modelGroups.map((group) => (
                        <optgroup key={group.label} label={group.label}>
                          {group.options.map((model) => (
                            <option
                              key={modelKey(model)}
                              value={modelKey(model)}
                            >
                              {model.modelName}
                            </option>
                          ))}
                        </optgroup>
                      ))}
                    </select>
                  </label>
                  <button
                    aria-label="Refresh models"
                    className="cowork-icon-button"
                    disabled={busy === 'models'}
                    onClick={() =>
                      void runAction('models', async () => loadModels(true))
                    }
                    title="Refresh models"
                    type="button"
                  >
                    <IconRefresh
                      className={busy === 'models' ? 'cowork-spin' : ''}
                      size={17}
                    />
                  </button>
                </div>

                <div className="cowork-form-grid cowork-cadence-grid">
                  <label>
                    Cadence
                    <select
                      onChange={(event) =>
                        setScheduleCadence(
                          event.target.value as CoworkScheduleCadence['kind'],
                        )
                      }
                      value={scheduleCadence}
                    >
                      <option value="manual">Manual</option>
                      <option value="hourly">Hourly</option>
                      <option value="daily">Daily</option>
                      <option value="weekly">Weekly</option>
                      <option value="weekdays">Weekdays</option>
                    </select>
                  </label>

                  {scheduleCadence === 'hourly' ? (
                    <label>
                      Minute of hour
                      <input
                        max={59}
                        min={0}
                        onChange={(event) =>
                          setScheduleHourlyMinute(Number(event.target.value))
                        }
                        required
                        type="number"
                        value={scheduleHourlyMinute}
                      />
                    </label>
                  ) : scheduleCadence !== 'manual' ? (
                    <label>
                      Local time
                      <input
                        onChange={(event) =>
                          setScheduleTime(event.target.value)
                        }
                        required
                        type="time"
                        value={scheduleTime}
                      />
                    </label>
                  ) : (
                    <div className="cowork-manual-cadence">
                      Use <strong>Run now</strong> whenever you need this task.
                    </div>
                  )}

                  {scheduleCadence === 'weekly' && (
                    <label>
                      Day
                      <select
                        onChange={(event) =>
                          setScheduleWeekday(Number(event.target.value))
                        }
                        value={scheduleWeekday}
                      >
                        {DAY_NAMES.map((day, index) => (
                          <option key={day} value={index}>
                            {day}
                          </option>
                        ))}
                      </select>
                    </label>
                  )}
                </div>

                <div className="cowork-form-grid">
                  <label>
                    Time zone
                    <input
                      onChange={(event) =>
                        setScheduleTimeZone(event.target.value)
                      }
                      placeholder="Asia/Kuala_Lumpur"
                      required
                      value={scheduleTimeZone}
                    />
                  </label>
                  {editingScheduleId && selectedSchedule ? (
                    <label>
                      Current status
                      <input
                        disabled
                        value={
                          selectedSchedule.status === 'active'
                            ? 'Active'
                            : 'Paused'
                        }
                      />
                    </label>
                  ) : (
                    <label>
                      Initial status
                      <select
                        onChange={(event) =>
                          setScheduleStatus(
                            event.target.value as 'active' | 'paused',
                          )
                        }
                        value={scheduleStatus}
                      >
                        <option value="active">Active</option>
                        <option value="paused">Paused</option>
                      </select>
                    </label>
                  )}
                </div>

                {editingScheduleId && !selectedModel && (
                  <div className="cowork-inline-error cowork-model-required">
                    <IconAlertTriangle size={17} />
                    <div>
                      <strong>Choose an available model</strong>
                      <span>
                        The schedule’s previous model or marketplace session is
                        no longer available. Select a current model before
                        saving.
                      </span>
                    </div>
                  </div>
                )}

                {selectedModel &&
                  (selectedModel.dataBoundary === 'on-device' ? (
                    <div className="cowork-model-notice local">
                      <IconDeviceLaptop size={17} />
                      <span>
                        <strong>Runs on this device</strong>
                        Scheduled tasks keep project data on your computer.
                      </span>
                    </div>
                  ) : (
                    <div className="cowork-model-notice remote">
                      <IconShieldCheck size={17} />
                      <span>
                        <strong>Scheduled runs wait for your approval</strong>
                        <small>
                          A task is created on schedule, but no project content
                          is sent off-device until you approve data sharing.
                        </small>
                      </span>
                    </div>
                  ))}

                <div className="cowork-new-task-footer">
                  <span>
                    <IconClock size={16} />
                    Times use {scheduleTimeZone || localTimeZone}
                  </span>
                  <div className="cowork-form-actions">
                    {editingScheduleId && (
                      <button
                        className="cowork-secondary-button"
                        disabled={busy === 'update-schedule'}
                        onClick={() => {
                          setEditingScheduleId(null);
                          resetScheduleForm();
                        }}
                        type="button"
                      >
                        Cancel edit
                      </button>
                    )}
                    <button
                      className="cowork-primary-button"
                      disabled={
                        busy === 'create-schedule' ||
                        busy === 'update-schedule' ||
                        !scheduleName.trim() ||
                        !scheduleGoal.trim() ||
                        !selectedModel
                      }
                      type="submit"
                    >
                      {busy === 'create-schedule' ||
                      busy === 'update-schedule' ? (
                        <IconLoader2 className="cowork-spin" size={17} />
                      ) : editingScheduleId ? (
                        <IconPencil size={17} />
                      ) : (
                        <IconCalendarTime size={17} />
                      )}
                      {editingScheduleId ? 'Save changes' : 'Create schedule'}
                    </button>
                  </div>
                </div>
              </form>
            )}
          </div>
        ) : !activeTask ? (
          <div className="cowork-stage cowork-new-task-stage">
            <form className="cowork-new-task-card" onSubmit={handleCreateTask}>
              <span className="cowork-hero-icon compact">
                <IconRobot size={25} />
              </span>
              <span className="cowork-eyebrow">New task</span>
              <h2>What should Workspace accomplish?</h2>
              <label className="cowork-title-field">
                Task title <span>optional</span>
                <input
                  onChange={(event) => setTaskTitle(event.target.value)}
                  placeholder="Prepare the release notes"
                  value={taskTitle}
                />
              </label>
              <label>
                Goal
                <textarea
                  autoFocus
                  onChange={(event) => setTaskGoal(event.target.value)}
                  placeholder="Describe the outcome, relevant context, and anything Workspace must not change…"
                  rows={7}
                  value={taskGoal}
                />
              </label>
              <div className="cowork-model-row">
                <label>
                  Model
                  <select
                    disabled={activeModels.length === 0}
                    onChange={(event) => chooseActiveModel(event.target.value)}
                    value={selectedModelKey}
                  >
                    {activeModels.length === 0 && (
                      <option value="">No models available</option>
                    )}
                    {activeModels.length > 0 && !selectedModelKey && (
                      <option value="">Choose an active session</option>
                    )}
                    {modelGroups.map((group) => (
                      <optgroup key={group.label} label={group.label}>
                        {group.options.map((model) => (
                          <option key={modelKey(model)} value={modelKey(model)}>
                            {model.modelName}
                          </option>
                        ))}
                      </optgroup>
                    ))}
                  </select>
                </label>
                <button
                  aria-label="Refresh models"
                  className="cowork-icon-button"
                  disabled={busy === 'models'}
                  onClick={() =>
                    void runAction('models', async () => loadModels(true))
                  }
                  title="Refresh models"
                  type="button"
                >
                  <IconRefresh
                    className={busy === 'models' ? 'cowork-spin' : ''}
                    size={17}
                  />
                </button>
              </div>
              {selectedModel && (
                <div
                  className={`cowork-model-notice ${
                    selectedModel.dataBoundary === 'on-device'
                      ? 'local'
                      : 'remote'
                  }`}
                >
                  {selectedModel.dataBoundary === 'on-device' ? (
                    <IconDeviceLaptop size={17} />
                  ) : (
                    <IconCloud size={17} />
                  )}
                  <span>
                    <strong>
                      {selectedModel.dataBoundary === 'on-device'
                        ? 'Runs on this device'
                        : selectedModel.dataBoundary === 'configured-endpoint'
                          ? 'Uses the model’s configured endpoint'
                          : 'Uses an independent Morpheus provider'}
                    </strong>
                    {selectedModel.dataBoundary === 'on-device'
                      ? 'Project data stays local to your computer.'
                      : selectedModel.dataBoundary === 'configured-endpoint'
                        ? 'Workspace will ask before sending project content to that endpoint.'
                        : 'Workspace will ask before sharing project content with the remote provider.'}
                  </span>
                  {selectedModel.visionCapability !== 'none' && (
                    <b className="cowork-vision-badge">
                      {selectedModel.visionCapability === 'declared'
                        ? 'Vision'
                        : 'Likely vision'}
                    </b>
                  )}
                </div>
              )}
              <div className="cowork-new-task-footer">
                <span>
                  <IconShieldCheck size={16} />
                  {APPROVAL_DESCRIPTIONS[approvalMode]}
                </span>
                <button
                  className="cowork-primary-button"
                  disabled={
                    busy === 'create-task' || !taskGoal.trim() || !selectedModel
                  }
                  type="submit"
                >
                  {busy === 'create-task' ? (
                    <IconLoader2 className="cowork-spin" size={17} />
                  ) : (
                    <IconPlayerPlay size={17} />
                  )}
                  Start task
                </button>
              </div>
            </form>
          </div>
        ) : (
          <>
            <div
              key={activeTask.id}
              ref={attachTranscriptElement}
              className="cowork-transcript"
              aria-live="polite"
            >
              <div className="cowork-task-context">
                <span className="cowork-model-chip">
                  {activeTask.model.dataBoundary === 'on-device' ? (
                    <IconDeviceLaptop size={14} />
                  ) : (
                    <IconCloud size={14} />
                  )}
                  Created with {activeTask.model.modelName}
                </span>
                <span>
                  {formatTime(activeTask.startedAt || activeTask.createdAt)}
                </span>
              </div>

              {hasEarlierMessages && (
                <div className="cowork-transcript-history">
                  <button
                    className="cowork-history-button"
                    disabled={loadingEarlierMessages}
                    onClick={() => void handleLoadEarlierMessages()}
                    type="button"
                  >
                    {loadingEarlierMessages && (
                      <IconLoader2
                        aria-hidden="true"
                        className="cowork-spin"
                        size={15}
                      />
                    )}
                    {loadingEarlierMessages
                      ? 'Loading earlier messages…'
                      : 'Load earlier messages'}
                  </button>
                </div>
              )}

              {displayedMessages.map((message) => (
                <CoworkMessageRow key={message.id} message={message} />
              ))}

              {activeTask.status === 'running' && (
                <div className="cowork-thinking-row">
                  <IconLoader2 className="cowork-spin" size={17} />
                  Workspace is working through the plan…
                </div>
              )}

              {activeTask.status === 'waiting_approval' &&
                activeTask.pendingApproval && (
                  <section className="cowork-approval-card">
                    <div className="cowork-approval-heading">
                      <span
                        className={`risk-${activeTask.pendingApproval.risk}`}
                      >
                        <IconShieldCheck size={19} />
                      </span>
                      <div>
                        <span className="cowork-eyebrow">
                          {activeTask.pendingApproval.risk} approval
                        </span>
                        <h3>{activeTask.pendingApproval.reason}</h3>
                      </div>
                    </div>
                    <div className="cowork-tool-preview">
                      <strong>
                        {toolLabel(
                          activeTask.pendingApproval.toolCall.function.name,
                        )}
                      </strong>
                      <pre>
                        {parseToolArguments(
                          activeTask.pendingApproval.toolCall.function
                            .arguments,
                        )}
                      </pre>
                    </div>
                    <p>
                      Review the exact action above. Only approve it if it
                      matches this task’s goal.
                    </p>
                    <div className="cowork-approval-actions">
                      <button
                        className="cowork-secondary-button"
                        disabled={busy === 'approval'}
                        onClick={() => handleApproval(false)}
                        type="button"
                      >
                        <IconX size={16} />
                        Deny
                      </button>
                      <button
                        className="cowork-primary-button"
                        disabled={
                          busy === 'approval' || !exactTaskSessionActive
                        }
                        onClick={() => handleApproval(true)}
                        type="button"
                      >
                        {busy === 'approval' ? (
                          <IconLoader2 className="cowork-spin" size={16} />
                        ) : (
                          <IconCheck size={16} />
                        )}
                        Approve once
                      </button>
                    </div>
                  </section>
                )}

              {activeTask.error && (
                <div className="cowork-inline-error">
                  <IconAlertTriangle size={17} />
                  <div>
                    <strong>Task stopped</strong>
                    <span>{activeTask.error}</span>
                  </div>
                </div>
              )}

              {activeTask.status === 'completed' && activeTask.summary && (
                <div className="cowork-complete-card">
                  <IconCircleCheck size={19} />
                  <div>
                    <strong>Task complete</strong>
                    <SafeMarkdown>{activeTask.summary}</SafeMarkdown>
                  </div>
                </div>
              )}

              {taskNeedsRebind && (
                <section className="cowork-rebind-card" role="note">
                  <IconCloud size={20} />
                  <div>
                    <strong>The original task session has ended</strong>
                    <p>
                      This history was created with{' '}
                      <b>{activeTask.model.modelName}</b> and remains unchanged.
                      {selectedModel
                        ? ` Continue with ${selectedModel.modelName} only when you are ready; Workspace will preserve the model boundary in the task history.`
                        : ' Open a new session to continue without losing this project context.'}
                    </p>
                  </div>
                  {selectedModel ? (
                    <button
                      className="cowork-primary-button"
                      disabled={busy === 'rebind-task'}
                      onClick={handleRebindTask}
                      type="button"
                    >
                      {busy === 'rebind-task' ? (
                        <IconLoader2 className="cowork-spin" size={16} />
                      ) : (
                        <IconPlayerPlay size={16} />
                      )}
                      Continue with {selectedModel.modelName}
                    </button>
                  ) : (
                    <button
                      className="cowork-secondary-button"
                      onClick={() => navigate('/chat?setup=workspace')}
                      type="button"
                    >
                      Open a session in Chat
                    </button>
                  )}
                </section>
              )}
            </div>

            <form className="cowork-steer-bar" onSubmit={handleSteer}>
              <textarea
                aria-label="Message Workspace"
                disabled={
                  busy === 'steer-task' ||
                  !exactTaskSessionActive ||
                  activeTask.status === 'waiting_approval'
                }
                onChange={(event) => setSteeringMessage(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' && !event.shiftKey) {
                    event.preventDefault();
                    event.currentTarget.form?.requestSubmit();
                  }
                }}
                placeholder={
                  activeTask.status === 'waiting_approval'
                    ? 'Resolve the approval above to continue'
                    : !exactTaskSessionActive
                      ? 'Choose an active session and continue this task first'
                      : 'Give Workspace more context or change direction…'
                }
                rows={2}
                value={steeringMessage}
              />
              <button
                aria-label="Send direction"
                className="cowork-send-button"
                disabled={
                  busy === 'steer-task' ||
                  !steeringMessage.trim() ||
                  !exactTaskSessionActive ||
                  activeTask.status === 'waiting_approval'
                }
                type="submit"
              >
                {busy === 'steer-task' ? (
                  <IconLoader2 className="cowork-spin" size={18} />
                ) : (
                  <IconSend size={18} />
                )}
              </button>
              <span className="cowork-steer-hint">
                Enter to send · Shift+Enter for a new line
              </span>
            </form>
          </>
        )}
      </section>

      <aside
        aria-label={
          railMode === 'schedules' ? 'Schedule details' : 'Task details'
        }
        className="cowork-details"
        id="workspace-inspector"
      >
        <div className="cowork-panel-header">
          <strong>
            {railMode === 'schedules' ? 'Schedule details' : 'Task details'}
          </strong>
          <button
            aria-label={
              railMode === 'schedules'
                ? 'Close schedule details'
                : 'Close task details'
            }
            className="cowork-icon-button cowork-panel-close"
            onClick={() => setInspectorOpen(false)}
            type="button"
          >
            <IconX size={18} />
          </button>
        </div>
        {railMode === 'schedules' ? (
          <>
            <section className="cowork-detail-section">
              <div className="cowork-detail-heading">
                <IconCalendarTime size={17} />
                <h2>Timing</h2>
              </div>
              {selectedSchedule ? (
                <dl className="cowork-detail-list">
                  <div>
                    <dt>Cadence</dt>
                    <dd>{cadenceLabel(selectedSchedule.cadence)}</dd>
                  </div>
                  <div>
                    <dt>Next</dt>
                    <dd>
                      {selectedSchedule.cadence.kind === 'manual'
                        ? 'Run manually'
                        : selectedSchedule.status === 'paused'
                          ? 'Paused'
                          : formatScheduleTime(
                              selectedSchedule.nextRunAt,
                              selectedSchedule.timeZone,
                            )}
                    </dd>
                  </div>
                  <div>
                    <dt>Time zone</dt>
                    <dd>{selectedSchedule.timeZone}</dd>
                  </div>
                </dl>
              ) : (
                <EmptyPanel
                  detail="Choose a cadence and local time for the new schedule."
                  icon={<IconClock size={20} />}
                  title="Local time by default"
                />
              )}
            </section>

            <section className="cowork-detail-section">
              <div className="cowork-detail-heading">
                <IconActivity size={17} />
                <h2>Last run</h2>
              </div>
              {selectedSchedule?.lastRunAt ? (
                <div className="cowork-schedule-history">
                  <strong>
                    {formatScheduleTime(
                      selectedSchedule.lastRunAt,
                      selectedSchedule.timeZone,
                    )}
                  </strong>
                  <span>
                    {selectedSchedule.lastError
                      ? 'Failed'
                      : 'Task created successfully'}
                  </span>
                  {selectedSchedule.lastTaskId && (
                    <button onClick={handleOpenScheduledTask} type="button">
                      Open task
                      <IconChevronRight size={14} />
                    </button>
                  )}
                </div>
              ) : (
                <EmptyPanel
                  detail="Run history appears after the first occurrence."
                  icon={<IconActivity size={20} />}
                  title="Never run"
                />
              )}
            </section>

            <section className="cowork-detail-section">
              <div className="cowork-detail-heading">
                <IconShieldCheck size={17} />
                <h2>Data boundary</h2>
              </div>
              <div className="cowork-schedule-safety-note">
                <strong>
                  {selectedSchedule?.task.model.dataBoundary === 'on-device'
                    ? 'On-device model'
                    : 'Approval-gated sharing'}
                </strong>
                <p>
                  {selectedSchedule?.task.model.dataBoundary === 'on-device'
                    ? 'Project data remains on this computer during scheduled runs.'
                    : 'Off-device runs create a waiting task. Data is not shared until you approve it.'}
                </p>
              </div>
            </section>
          </>
        ) : (
          <>
            <section className="cowork-detail-section">
              <div className="cowork-detail-heading">
                <IconListCheck size={17} />
                <h2>Plan</h2>
                {activeTask?.plan.length ? (
                  <span>
                    {
                      activeTask.plan.filter(
                        (step) => step.status === 'completed',
                      ).length
                    }
                    /{activeTask.plan.length}
                  </span>
                ) : null}
              </div>
              {activeTask?.plan.length ? (
                <ol className="cowork-plan-list">
                  {activeTask.plan.map((step) => (
                    <li className={`plan-${step.status}`} key={step.id}>
                      <PlanStatusIcon status={step.status} />
                      <div>
                        <strong>{step.title}</strong>
                        {step.note && <span>{step.note}</span>}
                      </div>
                    </li>
                  ))}
                </ol>
              ) : (
                <EmptyPanel
                  detail="Workspace will outline its approach before changing files."
                  icon={<IconListCheck size={20} />}
                  title="No plan yet"
                />
              )}
            </section>

            <section className="cowork-detail-section cowork-artifact-section">
              <div className="cowork-detail-heading">
                <IconPaperclip size={17} />
                <h2>Artifacts</h2>
                {activeTask?.artifacts.length ? (
                  <span>{activeTask.artifacts.length}</span>
                ) : null}
              </div>
              {activeTask?.artifacts.length ? (
                <div className="cowork-artifact-list">
                  {activeTask.artifacts.map((artifact) => (
                    <div className="cowork-artifact-row" key={artifact.path}>
                      <button
                        disabled={
                          artifact.kind !== 'file' ||
                          busy === `preview:${artifact.path}`
                        }
                        onClick={() => handlePreviewArtifact(artifact)}
                        title={artifact.path}
                        type="button"
                      >
                        {busy === `preview:${artifact.path}` ? (
                          <IconLoader2 className="cowork-spin" size={16} />
                        ) : artifact.kind === 'folder' ? (
                          <IconFolder size={16} />
                        ) : (
                          <IconFile size={16} />
                        )}
                        <span>
                          <strong>{artifact.name}</strong>
                          <small>{artifact.path}</small>
                        </span>
                      </button>
                      <button
                        aria-label={`Reveal ${artifact.name}`}
                        className="cowork-icon-button"
                        onClick={() => handleRevealArtifact(artifact)}
                        title="Reveal in folder"
                        type="button"
                      >
                        <IconExternalLink size={15} />
                      </button>
                    </div>
                  ))}
                </div>
              ) : (
                <EmptyPanel
                  detail="Files created or updated by Workspace appear here."
                  icon={<IconPaperclip size={20} />}
                  title="No artifacts yet"
                />
              )}
            </section>

            <section className="cowork-detail-section cowork-activity-section">
              <div className="cowork-detail-heading">
                <IconActivity size={17} />
                <h2>Activity</h2>
              </div>
              {activeTask?.activities.length ? (
                <div className="cowork-activity-list">
                  {[...activeTask.activities].reverse().map((activity) => (
                    <div
                      className={`cowork-activity-row activity-${activity.status}`}
                      key={activity.id}
                    >
                      <span className="cowork-activity-icon">
                        <ActivityIcon activity={activity} />
                      </span>
                      <div>
                        <strong>{activity.label}</strong>
                        {activity.detail && <span>{activity.detail}</span>}
                        <small>{formatTime(activity.createdAt)}</small>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <EmptyPanel
                  detail="Tool calls and approvals will be recorded here."
                  icon={<IconActivity size={20} />}
                  title="Nothing to report"
                />
              )}
            </section>
          </>
        )}

        <section className="cowork-detail-section cowork-extension-section">
          <div className="cowork-detail-heading">
            <IconPuzzle size={17} />
            <h2>Extensions</h2>
            {selectedProject && (
              <button
                aria-label="Refresh detected extensions"
                className="cowork-icon-button"
                disabled={extensionsLoading || busy === 'configure-extensions'}
                onClick={() =>
                  void runAction('extensions', async () =>
                    loadExtensions(selectedProject.id),
                  )
                }
                title="Scan again"
                type="button"
              >
                <IconRefresh
                  className={extensionsLoading ? 'cowork-spin' : ''}
                  size={15}
                />
              </button>
            )}
          </div>
          <div className="cowork-extension-notice">
            <IconShieldCheck size={15} />
            <span>
              Enabled guidance is injected into new and resumed task prompts.
              Instruction-only skills cannot execute code, add tools or network
              access, or override approvals and security rules. Off-device data
              approval includes this guidance.
            </span>
          </div>
          {extensionsLoading && !extensions ? (
            <div className="cowork-extension-loading">
              <IconLoader2 className="cowork-spin" size={16} />
              Scanning declared extensions…
            </div>
          ) : extensions &&
            (extensions.projectInstructions ||
              extensions.skills.length ||
              extensions.connectors.length ||
              extensions.issues.length) ? (
            <div className="cowork-extension-list">
              {extensions.projectInstructions && (
                <div className="cowork-extension-row">
                  <span>Instructions</span>
                  <div>
                    <strong>Folder instructions</strong>
                    <small>{extensions.projectInstructions.source}</small>
                  </div>
                  <label
                    className="cowork-extension-toggle"
                    title="Inject this guidance into new and resumed task prompts"
                  >
                    <input
                      aria-label="Enable folder instructions"
                      checked={extensions.projectInstructions.enabled}
                      disabled={busy === 'configure-extensions'}
                      onChange={(event) =>
                        handleFolderInstructionsToggle(event.target.checked)
                      }
                      type="checkbox"
                    />
                    <span />
                  </label>
                </div>
              )}
              {extensions.skills.map((skill) => (
                <div className="cowork-extension-row" key={`skill:${skill.id}`}>
                  <span>Guidance</span>
                  <div>
                    <strong>{skill.name}</strong>
                    <small>
                      Instruction-only · {skill.description || skill.source}
                    </small>
                  </div>
                  <label
                    className="cowork-extension-toggle"
                    title="Inject this skill’s instructions into new and resumed task prompts"
                  >
                    <input
                      aria-label={`Enable ${skill.name} guidance`}
                      checked={skill.enabled}
                      disabled={
                        busy === 'configure-extensions' ||
                        (!skill.enabled &&
                          extensions.skills.filter((item) => item.enabled)
                            .length >= 8)
                      }
                      onChange={(event) =>
                        handleSkillToggle(skill.id, event.target.checked)
                      }
                      type="checkbox"
                    />
                    <span />
                  </label>
                </div>
              ))}
              {extensions.connectors.map((connector) => (
                <div
                  className="cowork-extension-row"
                  key={`connector:${connector.id}`}
                >
                  <span>Connector</span>
                  <div>
                    <strong>{connector.name}</strong>
                    <small>
                      Read-only declaration · not connected ·{' '}
                      {connector.transport}
                    </small>
                  </div>
                  <b>Inactive</b>
                </div>
              ))}
              {extensions.issues.map((issue, index) => (
                <div
                  className={`cowork-extension-issue issue-${issue.severity}`}
                  key={`${issue.code}:${issue.source}:${index}`}
                >
                  <IconAlertTriangle size={14} />
                  <span>{issue.message}</span>
                </div>
              ))}
            </div>
          ) : (
            <EmptyPanel
              detail="No project instructions, skills, or connectors were declared."
              icon={<IconPuzzle size={20} />}
              title="Nothing detected"
            />
          )}
        </section>
      </aside>

      {preview && (
        <div
          aria-label={`Preview ${preview.path}`}
          aria-modal="true"
          className="cowork-preview-backdrop"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) setPreview(null);
          }}
          role="dialog"
        >
          <section className="cowork-preview-modal">
            <header>
              <div>
                <IconFile size={18} />
                <span>
                  <strong>{preview.path.split('/').pop()}</strong>
                  <small>{preview.path}</small>
                </span>
              </div>
              <button
                aria-label="Close preview"
                className="cowork-icon-button"
                onClick={() => setPreview(null)}
                type="button"
              >
                <IconX size={18} />
              </button>
            </header>
            <pre>{preview.content}</pre>
            {preview.truncated && (
              <footer>Preview truncated for safety and performance.</footer>
            )}
          </section>
        </div>
      )}
    </main>
  );
}

export default Cowork;
