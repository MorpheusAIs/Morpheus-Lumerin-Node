import {
  memo,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
// import component 👇
import Drawer from 'react-modern-drawer';
import {
  IconHistory,
  IconArrowUp,
  IconMessagePlus,
  IconShieldLock,
  IconUpload,
  IconMicrophone,
  IconPlayerStopFilled,
  IconPaperclip,
  IconSparkles,
} from '@tabler/icons-react';
import { useLocation, useNavigate } from 'react-router';
import {
  View,
  ContainerTitle,
  ChatTitleContainer,
  ChatAvatar,
  Avatar,
  TitleRow,
  AvatarHeader,
  MessageBody,
  Container,
  CustomTextArrea,
  Control,
  LoadingCover,
  ImageContainer,
  SubPriceLabel,
  VideoContainer,
  ChatIntroContainer,
  ChatHistoryContainer,
  ChatStartupState,
  ChatIntroInner,
  ChatIntroInnerTitle,
  ChatIntroInnerText,
  ChatIntroButton,
  SessionDurationField,
  SessionCostSummary,
  SessionSetupState,
  SessionSetupActions,
  SessionHistoryNotice,
  LoadingStatus,
  SendBtnWrapper,
  Btn,
  AudioInputZone,
  AudioActionBtn,
  AudioHint,
  TtsControlsRow,
  AudioPlayer,
} from './Chat.styles';
import { BtnAccent } from '../dashboard/BalanceBlock.styles';
import withChatState from '../../store/hocs/withChatState';
import { abbreviateAddress } from '../../utils';
import { ThinkingMessageBody } from './ThinkingMessageBody';

import 'react-modern-drawer/dist/index.css';
import './Chat.css';
import { ChatHistory } from './ChatHistory';
import Spinner from 'react-bootstrap/Spinner';
import ModelSelectionModal from './modals/ModelSelectionModal';
import {
  tryParseDataChunk,
  makeId,
  getColor,
  isClosed,
  generateHashId,
  isSecureModel,
  SECURE_BADGE_TOOLTIP,
  getModelModality,
  isCoworkCandidate,
  scheduleSessionExpiry,
} from './utils';
import { Cooldown } from './Cooldown';
import ImageViewer from 'react-simple-image-viewer';
import { ChatData, HistoryMessage } from './interfaces';
import { formatValue } from '../../utils/coinValue';
import { ApiGateway } from 'src/main/src/client/apiGateway';
import { queryKeys } from '../../store/queries';
import QueryError from '../common/QueryError';
import AttachmentBar from './AttachmentBar';
import {
  Attachment,
  buildUserMessage,
  classify,
  looksVisionCapable,
  readFile,
  validateFile,
} from '../../store/utils/attachments';
import { explainChainError } from '../../store/utils/chainErrors';
import {
  hasRejectedImages,
  isVisionRejection,
  rememberImageRejection,
} from '../../store/utils/visionMemory';
import {
  DEFAULT_SESSION_DURATION_SECONDS,
  SESSION_DURATION_OPTIONS,
  estimateSessionTokenAmount,
  toSessionRequestDuration,
} from '../../store/utils/sessionDuration';

const CHAT_BOTTOM_THRESHOLD_PX = 96;
const MAX_AUDIO_BYTES = 20 * 1024 * 1024;

let abort = false;
const userMessage = { user: 'Me', role: 'user', icon: 'M', color: '#20dc8e' };

type ScrollMetrics = Pick<
  HTMLElement,
  'clientHeight' | 'scrollHeight' | 'scrollTop'
>;

export const isNearChatBottom = (
  element: ScrollMetrics,
  threshold = CHAT_BOTTOM_THRESHOLD_PX,
): boolean =>
  element.scrollHeight - element.scrollTop - element.clientHeight <= threshold;

export const observeChatAutoScroll = (
  element: HTMLElement,
  onAutoScrollChange: (enabled: boolean) => void,
): (() => void) => {
  const updateAutoScroll = () => {
    onAutoScrollChange(isNearChatBottom(element));
  };
  const handleWheel = (event: WheelEvent) => {
    const legacyDelta = (event as WheelEvent & { wheelDelta?: number })
      .wheelDelta;
    if (event.deltaY < 0 || (legacyDelta !== undefined && legacyDelta > 0)) {
      onAutoScrollChange(false);
    }
  };

  element.addEventListener('scroll', updateAutoScroll, { passive: true });
  element.addEventListener('wheel', handleWheel, { passive: true });
  return () => {
    element.removeEventListener('scroll', updateAutoScroll);
    element.removeEventListener('wheel', handleWheel);
  };
};

type AnimationFrameBatch<T> = {
  cancel: () => void;
  flush: () => void;
  schedule: (value: T) => void;
};

type MutableRef<T> = { current: T };

/** Invalidates queued chat work and propagates cancellation to the IPC stream. */
export const disposeActiveChatStream = async (
  mountedRef: MutableRef<boolean>,
  generationRef: MutableRef<number>,
  activeReaderRef: MutableRef<ReadableStreamDefaultReader<Uint8Array> | null>,
): Promise<void> => {
  mountedRef.current = false;
  generationRef.current += 1;
  const reader = activeReaderRef.current;
  activeReaderRef.current = null;
  if (reader) {
    await reader.cancel().catch(() => undefined);
  }
};

/** Keeps the latest stream state and commits it at most once per paint. */
export const createAnimationFrameBatch = <T,>(
  commit: (value: T) => void,
  requestFrame: (callback: FrameRequestCallback) => number = (callback) =>
    window.requestAnimationFrame(callback),
  cancelFrame: (handle: number) => void = (handle) =>
    window.cancelAnimationFrame(handle),
): AnimationFrameBatch<T> => {
  let frame: number | undefined;
  let hasPendingValue = false;
  let pendingValue: T;

  const commitPending = () => {
    frame = undefined;
    if (!hasPendingValue) return;
    hasPendingValue = false;
    commit(pendingValue);
  };

  return {
    schedule(value) {
      pendingValue = value;
      hasPendingValue = true;
      frame ??= requestFrame(commitPending);
    },
    flush() {
      if (frame !== undefined) cancelFrame(frame);
      commitPending();
    },
    cancel() {
      if (frame !== undefined) cancelFrame(frame);
      frame = undefined;
      hasPendingValue = false;
    },
  };
};

export const revokeInactiveObjectUrls = (
  ownedUrls: Set<string>,
  activeUrls: ReadonlySet<string>,
  revoke: (url: string) => void = (url) => URL.revokeObjectURL(url),
): void => {
  for (const url of ownedUrls) {
    if (activeUrls.has(url)) continue;
    revoke(url);
    ownedUrls.delete(url);
  }
};

/** A disabled/paused no-data query is pending, but it is not doing any work. */
export const isInitialQueryFetchActive = (
  data: unknown,
  fetchStatus: string,
): boolean => data === undefined && fetchStatus === 'fetching';

type SessionOpenResult =
  | { kind: 'opened'; sessionId: string }
  | { kind: 'existing'; sessionId: string };

/** Normalizes both the legacy string result and the duplicate-session sentinel. */
export const resolveSessionOpenResult = (
  value: unknown,
): SessionOpenResult | null => {
  if (typeof value === 'string' && value) {
    return { kind: 'opened', sessionId: value };
  }
  if (!value || typeof value !== 'object') {
    return null;
  }

  const result = value as {
    existingSessionID?: unknown;
    sessionID?: unknown;
  };
  if (
    typeof result.existingSessionID === 'string' &&
    result.existingSessionID
  ) {
    return { kind: 'existing', sessionId: result.existingSessionID };
  }
  if (typeof result.sessionID === 'string' && result.sessionID) {
    return { kind: 'opened', sessionId: result.sessionID };
  }
  return null;
};

// Common TTS voice presets. Names are backend-specific (Kokoro `af_*`,
// OpenAI `alloy`/`nova`/...), so the field also accepts free-text input.
const TTS_VOICES = [
  'af_bella',
  'af_alloy',
  'af_sky',
  'af_nicole',
  'am_adam',
  'am_michael',
  'alloy',
  'nova',
  'shimmer',
];

type ChatProps = {
  client: ApiGateway;
  address: string;
  symbol: string;
  config: any;
  toasts: {
    toast: (
      type: string,
      message: string,
      options?: { autoClose?: number },
    ) => void;
  };
  getAllModels: () => Promise<any[]>;
  getLocalModels: () => Promise<any[]>;
  getMetaInfo: () => Promise<{ budget: number; supply: number }>;
  getBalances: () => Promise<{ eth: number; mor: number }>;
  getSessionsByUser: (address: string) => Promise<any>;
  getBidInfo: (id: string) => Promise<any>;
  getBidsByModelId: (id: string) => Promise<any>;
  onOpenSession: (props: {
    modelId: string;
    duration: number;
    isDirectPay: boolean;
  }) => Promise<any>;
  closeSession: (sessionId: string) => Promise<any>;
};

const Chat = (props: ChatProps) => {
  const location = useLocation();
  const navigate = useNavigate();
  const chatBlockRef = useRef<null | HTMLDivElement>(null);
  const autoScrollRef = useRef(true);
  const scrollFrameRef = useRef<number | undefined>(undefined);
  const ownedAudioUrlsRef = useRef(new Set<string>());
  const mountedRef = useRef(true);
  const chatGenerationRef = useRef(0);
  const activeReaderRef =
    useRef<ReadableStreamDefaultReader<Uint8Array> | null>(null);
  const queryClient = useQueryClient();
  const initializedRef = useRef(false);

  const [promptInput, setPromptInput] = useState('');
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const attachInputRef = useRef<HTMLInputElement | null>(null);
  // Overlay shown during user-triggered actions (open/close/reopen session,
  // manual session refresh). The *initial* page load no longer uses this — it
  // is gated on the react-query cache so revisiting the tab is instant.
  const [actionStatus, setActionStatus] = useState<string | null>(null);
  const sessionActionInFlightRef = useRef(false);
  const [messages, setMessages] = useState<any>([]);
  const [chatScrollElement, setChatScrollElement] =
    useState<HTMLDivElement | null>(null);
  const [isOpen, setIsOpen] = useState(false);

  const [isSpinning, setIsSpinning] = useState(false);

  const [imagePreview, setImagePreview] = useState<string>();
  const [activeSession, setActiveSession] = useState<any>(undefined);
  const [, setSessionValidityVersion] = useState(0);

  const [chatData, setChatsData] = useState<ChatData[]>([]);

  const [openChangeModal, setOpenChangeModal] = useState(false);
  const [coworkModelSelection, setCoworkModelSelection] = useState(false);
  const [isReadonly, setIsReadonly] = useState(false);

  const [selectedBid, setSelectedBid] = useState<any>(null);
  const [selectedModel, setSelectedModel] = useState<any>(undefined);
  const [sessionDuration, setSessionDuration] = useState(
    DEFAULT_SESSION_DURATION_SECONDS,
  );

  const [chat, setChat] = useState<ChatData | undefined>(undefined);

  const attachChatScrollElement = useCallback(
    (element: HTMLDivElement | null) => {
      chatBlockRef.current = element;
      setChatScrollElement(element);
    },
    [],
  );

  // --- Cached data layer (stale-while-revalidate via react-query) ---------
  // These queries live in the app-level QueryClient, so navigating away from
  // and back to /chat serves cached data instantly and revalidates silently
  // instead of blocking behind a full-screen spinner.

  const marketplaceModelsQuery = useQuery({
    queryKey: queryKeys.allModels,
    queryFn: () => props.getAllModels(),
    select: (models) => (models ?? []).filter((model: any) => !model.IsDeleted),
  });

  // Local models and session-funding data are useful, but neither should hold
  // the model registry hostage. Active bids are intentionally absent here:
  // loading them for every registered model produced hundreds of requests and
  // kept every picker row disabled until the entire sweep finished.
  const localModelsQuery = useQuery({
    queryKey: queryKeys.localModels,
    queryFn: () => props.getLocalModels(),
  });

  const fundingQuery = useQuery({
    queryKey: queryKeys.chatFunding(props.address),
    enabled: !!props.address,
    queryFn: async () => {
      const [meta, userBalances] = await Promise.all([
        props.getMetaInfo(),
        props.getBalances(),
      ]);
      return { meta, userBalances };
    },
  });

  const sessionsQuery = useQuery({
    queryKey: queryKeys.sessions(props.address),
    queryFn: () => props.getSessionsByUser(props.address),
    enabled: !!props.address,
  });

  // One cached active-bids request for the model the user actually selected.
  // A raw registry row remains useful/selectable while this is idle; choosing
  // it moves to an honest price-loading state instead of starting a 738-model
  // background sweep.
  const selectedModelBidsQuery = useQuery({
    queryKey: queryKeys.modelBids(props.address, selectedModel?.Id),
    enabled: !!props.address && !!selectedModel?.Id && !selectedModel?.isLocal,
    staleTime: 60_000,
    queryFn: async () =>
      (await props.getBidsByModelId(selectedModel?.Id)) ?? [],
  });

  const chatTitlesQuery = useQuery({
    queryKey: queryKeys.chatTitles,
    queryFn: () => props.client.getChatHistoryTitles(),
  });

  // Full (unfiltered) model list — local + every marketplace model, no bids.
  // Used for mapping sessions/chats by id, matching the original mount logic.
  const localModels = useMemo(
    () =>
      (localModelsQuery.data ?? []).map((model: any) => ({
        ...model,
        isLocal: true,
      })),
    [localModelsQuery.data],
  );
  const allModels: any[] | undefined = useMemo(
    () =>
      marketplaceModelsQuery.data === undefined
        ? undefined
        : [...localModels, ...marketplaceModelsQuery.data],
    [localModels, marketplaceModelsQuery.data],
  );
  const allModelsById = useMemo(
    () =>
      new Map(
        (allModels ?? [])
          .filter((model: any) => !model.isLocal)
          .map((model: any) => [model.Id, model]),
      ),
    [allModels],
  );

  // Model browsing and history only need registry metadata. The selected model
  // receives its bid data separately through selectedModelBidsQuery.
  const chainData = useMemo(() => {
    if (marketplaceModelsQuery.data === undefined) {
      return null;
    }
    return {
      models: [...localModels, ...marketplaceModelsQuery.data],
      meta: fundingQuery.data?.meta,
      userBalances: fundingQuery.data?.userBalances,
    };
  }, [fundingQuery.data, localModels, marketplaceModelsQuery.data]);

  const meta = fundingQuery.data?.meta ?? { budget: 0, supply: 0 };
  const balances = fundingQuery.data?.userBalances ?? { eth: 0, mor: 0 };
  const modelsLoading =
    marketplaceModelsQuery.isPending &&
    marketplaceModelsQuery.data === undefined;

  const sessions = useMemo(() => {
    const raw = sessionsQuery.data;
    if (!raw || !allModels) {
      return [];
    }
    return raw.reduce((res: any[], item: any) => {
      const sessionModel = allModelsById.get(item.ModelAgentId);
      if (sessionModel) {
        res.push({ ...item, ModelName: sessionModel.Name });
      }
      return res;
    }, []);
  }, [sessionsQuery.data, allModels, allModelsById]);

  // The blocking overlay is reserved for user-triggered mutations. Startup
  // reads now render honest inline states, so a slow session scan cannot make
  // the entire Chat route look frozen or hide the New Chat control.
  const isLoading = Boolean(actionStatus);

  // TTS controls + STT recording state
  const [ttsVoice, setTtsVoice] = useState('af_bella');
  const [ttsSpeed, setTtsSpeed] = useState(1);
  const [recording, setRecording] = useState(false);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const peakLevelRef = useRef<number>(0);
  const levelRafRef = useRef<number | null>(null);

  const modelName = selectedModel?.Name || 'Model';
  const isLocal = chat?.isLocal;
  const isSecure = isSecureModel(selectedModel);
  const modality = getModelModality(selectedModel);

  const providerAddress = isLocal
    ? '(local)'
    : selectedBid?.Provider
      ? abbreviateAddress(selectedBid?.Provider, 6)
      : 'Unknown';
  const marketplaceSessionUnavailable = !isLocal && isClosed(activeSession);
  const isDisabled = marketplaceSessionUnavailable || isReadonly;
  const isCreateSessionMode =
    Boolean(selectedModel) &&
    !messages?.length &&
    !isLocal &&
    !activeSession &&
    !isLoading;
  const stakedFunds = activeSession
    ? (
        ((activeSession.EndsAt - activeSession.OpenedAt) *
          activeSession.PricePerSecond) /
        10 ** 18
      ).toFixed(2)
    : 0;

  useEffect(() => {
    if (isLocal || !activeSession) return;
    return scheduleSessionExpiry(activeSession, () => {
      // Cooldown owns its own countdown state. Bump the Chat parent as well so
      // submit, attachments, recording, and Cowork handoff all close together.
      setSessionValidityVersion((current) => current + 1);
      setIsReadonly(true);
    });
  }, [activeSession, isLocal]);

  // A user can pick a model while the slower session inventory is still in
  // flight. Reconcile that choice when the inventory arrives so an existing
  // session is resumed instead of offering to stake twice.
  useEffect(() => {
    if (
      !selectedModel ||
      selectedModel.isLocal ||
      activeSession ||
      !sessionsQuery.isSuccess
    )
      return;
    const existingSession = sessions.find(
      (session) =>
        !isClosed(session) && session.ModelAgentId == selectedModel.Id,
    );
    if (!existingSession) return;
    setActiveSession(existingSession);
  }, [activeSession, selectedModel, sessions, sessionsQuery.isSuccess]);

  useEffect(() => {
    if (!activeSession || !selectedModelBidsQuery.data) {
      return;
    }
    const matchingBid = selectedModelBidsQuery.data.find(
      (bid: any) => bid.Id == activeSession.BidID,
    );
    if (matchingBid) {
      setSelectedBid(matchingBid);
    }
  }, [activeSession, selectedModelBidsQuery.data]);

  // One-time selection of the default chat once the (possibly cached) model and
  // session data is available. Runs in a layout effect so that on a warm cache
  // the selection is committed before paint — no flash of the empty/intro state
  // and no transient spinner on tab revisits.
  useLayoutEffect(() => {
    if (initializedRef.current) {
      return;
    }
    const rawSessions = sessionsQuery.data;
    if (!allModels || !rawSessions) {
      return;
    }
    initializedRef.current = true;

    const models = allModels;

    const requireMarketplaceSelection = () => {
      setSelectedModel(undefined);
      setSelectedBid(undefined);
      setActiveSession(undefined);
      setChat(undefined);
      setCoworkModelSelection(false);
      setOpenChangeModal(true);
    };

    const mappedSessions = rawSessions.reduce((res: any[], item: any) => {
      const sessionModel = models.find(
        (x) => !x.isLocal && x.Id == item.ModelAgentId,
      );
      if (sessionModel) {
        res.push({ ...item, ModelName: sessionModel.Name });
      }
      return res;
    }, []);
    const openSessions = mappedSessions.filter((s) => !isClosed(s));

    if (!openSessions.length) {
      requireMarketplaceSelection();
      return;
    }

    const latestSession = openSessions[0];
    const latestSessionModel = models.find(
      (m: any) => !m.isLocal && m.Id == latestSession.ModelAgentId,
    );

    if (!latestSessionModel) {
      requireMarketplaceSelection();
      return;
    }

    // Commit the session selection synchronously (before paint), then fetch the
    // bid details in the background.
    setSelectedModel(latestSessionModel);
    setActiveSession(latestSession);
    setChat({
      id: generateHashId(),
      createdAt: new Date(),
      modelId: latestSessionModel.Id,
    });
    props
      .getBidInfo(latestSession.BidID)
      .then((openBid) => {
        if (openBid) setSelectedBid(openBid);
      })
      .catch((e) => console.error('Failed to load open bid', e));
  }, [allModels, sessionsQuery.data]);

  // Workspace routes users here when they do not yet have an active marketplace
  // session. Open the normal model picker in marketplace-only mode; after the
  // session opens, the header offers an explicit choice between Chat and
  // Workspace. The legacy query value remains supported for existing links.
  useEffect(() => {
    const setup = new URLSearchParams(location.search).get('setup');
    if (setup !== 'workspace' && setup !== 'cowork') return;
    setCoworkModelSelection(true);
    setOpenChangeModal(true);
    navigate('/chat', { replace: true });
  }, [location.search, navigate]);

  // Keep the chat-history drawer list in sync with the cached titles + models.
  useEffect(() => {
    const titles = chatTitlesQuery.data as
      | Array<{
          chatId: string;
          title: string;
          modelId: string;
          createdAt: number;
          isLocal: boolean;
        }>
      | undefined;
    if (!titles || !allModels) {
      return;
    }
    const mappedChatData = titles.reduce<ChatData[]>((res, item) => {
      const chatModel = allModels.find((x) => x.Id == item.modelId);
      if (chatModel) {
        res.push({
          id: item.chatId,
          title: item.title,
          createdAt: new Date(item.createdAt * 1000),
          modelId: item.modelId,
          isLocal: item.isLocal,
        });
      }
      return res;
    }, []);
    setChatsData(mappedChatData);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chatTitlesQuery.data, allModels]);

  const toggleDrawer = () => {
    setIsOpen((prevState) => !prevState);
  };

  const scrollToBottom = (behavior: ScrollBehavior = 'auto', force = false) => {
    const element = chatBlockRef.current;
    if (!element || (!force && !autoScrollRef.current)) return;
    autoScrollRef.current = true;
    element.scroll({ top: element.scrollHeight, behavior });
  };

  const scheduleScrollToBottom = (
    behavior: ScrollBehavior = 'auto',
    force = false,
  ) => {
    if (!force && !autoScrollRef.current) return;
    if (scrollFrameRef.current !== undefined) {
      window.cancelAnimationFrame(scrollFrameRef.current);
    }
    scrollFrameRef.current = window.requestAnimationFrame(() => {
      scrollFrameRef.current = undefined;
      scrollToBottom(behavior, force);
    });
  };

  useEffect(() => {
    const activeUrls = new Set<string>();
    if (Array.isArray(messages)) {
      for (const message of messages) {
        if (message?.isAudioContent && typeof message.text === 'string') {
          activeUrls.add(message.text);
        }
      }
    }
    revokeInactiveObjectUrls(ownedAudioUrlsRef.current, activeUrls);
  }, [messages]);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      void disposeActiveChatStream(
        mountedRef,
        chatGenerationRef,
        activeReaderRef,
      );
      if (scrollFrameRef.current !== undefined) {
        window.cancelAnimationFrame(scrollFrameRef.current);
        scrollFrameRef.current = undefined;
      }
      revokeInactiveObjectUrls(ownedAudioUrlsRef.current, new Set());
    };
  }, []);

  useEffect(() => {
    const element = chatScrollElement;
    if (!element) return;

    if (autoScrollRef.current) {
      element.scroll({ top: element.scrollHeight, behavior: 'auto' });
    } else {
      autoScrollRef.current = isNearChatBottom(element);
    }
    return observeChatAutoScroll(element, (enabled) => {
      autoScrollRef.current = enabled;
    });
  }, [chatScrollElement]);

  // A session that was just opened on-chain does not always show up in the very
  // next indexer read. Poll briefly instead of assuming the first response
  // contains it — previously a miss meant `targetSessionData` was undefined and
  // the next line threw, which aborted the handler and left the UI wedged
  // (and the user re-staking into a second session they didn't need).
  const findSessionWithRetry = async (
    sessionId,
    attempts = 5,
    delayMs = 1200,
  ) => {
    for (let i = 0; i < attempts; i++) {
      const allSessions = await refreshSessions();
      const match = allSessions.find((x) => x.Id == sessionId);
      if (match) {
        return match;
      }
      if (i < attempts - 1) {
        await new Promise((r) => setTimeout(r, delayMs));
      }
    }
    return undefined;
  };

  const setSessionData = async (sessionId) => {
    const targetSessionData = await findSessionWithRetry(sessionId);

    if (!targetSessionData) {
      // The stake did go through — we just can't see it yet. Say so explicitly
      // rather than silently falling back to the "open a session" screen, which
      // is what led people to stake twice.
      props.toasts.toast(
        'info',
        'Session created, but not visible yet. It will appear in Sessions shortly — please do not stake again.',
        { autoClose: 12000 },
      );
      return;
    }

    if (isClosed(targetSessionData)) {
      props.toasts.toast(
        'info',
        'The new session is already closed or expired. Refresh Sessions before trying again.',
      );
      return;
    }

    setActiveSession({ ...targetSessionData, sessionId });

    const targetBid = selectedModelBidsQuery.data?.find(
      (x) => x.Id == targetSessionData.BidID,
    );
    if (targetBid) {
      setSelectedBid(targetBid);
    } else if (targetSessionData.BidID) {
      // A live session can outlast the provider's currently-active bid. Keep
      // the resumed chat usable immediately, then recover its provider label
      // from the bid record without holding the checkout overlay open.
      void props
        .getBidInfo(targetSessionData.BidID)
        .then((openBid) => {
          if (openBid) setSelectedBid(openBid);
        })
        .catch((error) =>
          console.error('Failed to load resumed session bid', error),
        );
    }
  };

  const onOpenSession = async (isReopen: boolean, isDirectPay: boolean) => {
    if (!selectedModel?.Id || !selectedModelBidsQuery.data?.length) {
      props.toasts.toast(
        'error',
        'This model has no active provider price. Retry the price check or choose another model.',
      );
      return;
    }

    // State updates are asynchronous, so use a ref to close the same-tick
    // double-click window before the first network await. The proxy-router's
    // rejectExisting guard remains the cross-process source of truth.
    if (sessionActionInFlightRef.current) {
      return;
    }
    sessionActionInFlightRef.current = true;

    let resolvedResult: SessionOpenResult | null = null;
    setActionStatus('Checking and opening session…');
    try {
      if (!isReopen) {
        setChat({
          id: generateHashId(),
          createdAt: new Date(),
          modelId: selectedModel.Id,
        });
      }

      const duration = toSessionRequestDuration(
        sessionDuration,
        isDirectPay,
        meta,
      );

      const rawResult = await props.onOpenSession({
        modelId: selectedModel.Id,
        duration,
        isDirectPay,
      });
      resolvedResult = resolveSessionOpenResult(rawResult);
      if (!resolvedResult) {
        return;
      }

      setActionStatus(
        resolvedResult.kind === 'existing'
          ? 'Existing session found — syncing it…'
          : 'Session opened — syncing it…',
      );

      // Invalidate the shared caches *before* touching local component state.
      // These run against the app-level QueryClient, which outlives this
      // component — so even if the user navigates to Wallet mid-stake and this
      // component unmounts, the sessions and balances caches are already marked
      // stale and the new session shows up on return. Previously the only
      // record of the new session was local state that died with the unmount,
      // and the 30s-stale cache kept serving the pre-stake session list.
      queryClient.invalidateQueries({
        queryKey: queryKeys.sessions(props.address),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.balances(props.address),
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.chatFunding(props.address),
      });

      await setSessionData(resolvedResult.sessionId);
      return resolvedResult.sessionId;
    } catch (e: any) {
      // Never let a post-open failure escape as an unhandled rejection — that
      // used to take the whole view down with it.
      console.error('Failed to finalize opened session', e);
      props.toasts.toast(
        'error',
        resolvedResult?.kind === 'existing'
          ? 'An existing session was found, but the app could not load it. Refresh Sessions before trying again.'
          : resolvedResult?.kind === 'opened'
            ? 'Session opened, but the app could not load it. Check the Sessions list before staking again.'
            : 'The session could not be opened. Retry when your node connection is stable.',
        { autoClose: 12000 },
      );
      return;
    } finally {
      sessionActionInFlightRef.current = false;
      setActionStatus(null);
    }
  };

  const loadChatHistory = async (chatId: string) => {
    try {
      const history = await props.client.getChatHistory(chatId);
      const messages: HistoryMessage[] = [];
      if (!history) {
        return;
      }

      const model = chainData?.models.find((m) => m.Id == history.modelId);
      const modelName = model?.Name || 'Model';
      const aiIcon = modelName.toUpperCase()[0];
      const aiColor = getColor(aiIcon);

      (history.messages || []).forEach((m: any) => {
        const prompt = m?.prompt || {};
        // Prompt shape differs by modality:
        //  - LLM/chat: { messages: [{ content }] }
        //  - TTS:      { input: '...' } (audio response is not stored replayably)
        //  - STT:      audio request (no messages); response is the transcript,
        //              flagged with isAudioContent on the stored message.
        const isChatPrompt =
          Array.isArray(prompt.messages) && prompt.messages.length > 0;
        const isTtsPrompt = !isChatPrompt && typeof prompt.input === 'string';
        const isSttMessage =
          !isChatPrompt && !isTtsPrompt && !!m.isAudioContent;

        let userText: string;
        if (isChatPrompt) {
          userText = prompt.messages[0]?.content ?? '';
        } else if (isTtsPrompt) {
          userText = prompt.input;
        } else if (isSttMessage) {
          // The uploaded/recorded audio is not retained in a replayable form.
          userText = prompt.Prompt || prompt.prompt || '🎤 Audio input';
        } else {
          userText = '';
        }

        messages.push({
          id: makeId(16),
          text: userText,
          user: userMessage.user,
          role: userMessage.role,
          icon: userMessage.icon,
          color: userMessage.color,
        });

        const assistant: HistoryMessage = {
          id: makeId(16),
          text: m.response,
          user: modelName,
          role: 'assistant',
          icon: aiIcon,
          color: aiColor,
        };
        if (isTtsPrompt) {
          // Synthesized audio is not persisted in a replayable form.
          assistant.text =
            '[Audio response — replay is not available from history]';
        } else if (!isSttMessage) {
          assistant.isImageContent = m.isImageContent;
          assistant.isVideoRawContent = m.isVideoRawContent;
        }
        messages.push(assistant);
      });
      setMessages(messages);
    } catch (e) {
      console.error('Failed to load chat history', e);
      props.toasts.toast('error', 'Failed to load chat history');
    }
  };

  // Refetch sessions through react-query so the shared cache (and every derived
  // `sessions` consumer) updates, and return the freshly-mapped list for callers
  // that need it synchronously (e.g. setSessionData).
  const refreshSessions = async () => {
    const fresh = await queryClient.fetchQuery({
      queryKey: queryKeys.sessions(props.address),
      queryFn: () => props.getSessionsByUser(props.address),
      // Reconciliation is an explicit post-transaction poll. It must bypass
      // the normal app-level stale window on every retry.
      staleTime: 0,
    });
    const models = allModels ?? [];
    return (fresh || []).reduce((res, item) => {
      const sessionModel = models.find(
        (x) => !x.isLocal && x.Id == item.ModelAgentId,
      );
      if (sessionModel) {
        res.push({ ...item, ModelName: sessionModel.Name });
      }
      return res;
    }, []);
  };

  const closeSession = async (sessionId: string) => {
    setActionStatus('Closing session…');
    try {
      await props.closeSession(sessionId);
      await refreshSessions();
    } finally {
      setActionStatus(null);
    }

    if (activeSession?.Id == sessionId) {
      setActiveSession(undefined);
      setSelectedBid(undefined);
      setSelectedModel(undefined);
      setChat(undefined);
      setMessages([]);
      setCoworkModelSelection(false);
      setOpenChangeModal(true);
    }
  };

  const selectChat = async (chatData: ChatData) => {
    abort = true;
    chatGenerationRef.current += 1;
    const modelId = chatData.modelId;
    if (!modelId) {
      console.warn('Model ID is missed');
      return;
    }

    const availableModels = chainData?.models ?? [];
    const selectedModel = availableModels.find((m: any) => m.Id == modelId);
    setSelectedModel(selectedModel);
    setIsReadonly(false);

    setChat({ ...chatData });

    if (chatData.isLocal) {
      // Local TinyLlama was historically a development demo. Keep its saved
      // transcript readable, but do not let it bypass the production
      // marketplace-session gate.
      setActiveSession(undefined);
      setSelectedBid(undefined);
      setIsReadonly(true);
      await loadChatHistory(chatData.id);
      return;
    }

    const openSessions = sessions.filter((s) => !isClosed(s));
    // search open session by model ID
    const openSession = openSessions.find((s) => s.ModelAgentId == modelId);
    setIsReadonly(!openSession);

    if (openSession) {
      setActiveSession(openSession);
      // The selected-model query resolves the provider bid without a global
      // marketplace sweep. Clear any bid from the previous chat meanwhile.
      setSelectedBid(undefined);
    } else {
      setActiveSession(undefined);
      setSelectedBid(undefined);
    }

    autoScrollRef.current = true;
    await loadChatHistory(chatData.id);
    setTimeout(() => scheduleScrollToBottom('smooth', true), 400);
  };

  const handleReopen = async (isDirectPay: boolean) => {
    await onOpenSession(true, isDirectPay);
    setIsReadonly(false);
  };

  const call = async (message, callAttachments: Attachment[] = []) => {
    const chatGeneration = chatGenerationRef.current;
    let memoState = [
      ...messages,
      {
        id: makeId(16),
        text: promptInput,
        // Shown as chips under the user's bubble so the transcript reflects
        // what was actually sent.
        attachments: callAttachments.map((a) => ({
          name: a.name,
          kind: a.kind,
          dataUrl: a.kind === 'image' ? a.dataUrl : undefined,
        })),
        ...userMessage,
      },
    ];
    setMessages(memoState);
    scheduleScrollToBottom();

    // Plain string content when there are no images, so the text-only path
    // keeps exactly the shape it had before attachments existed.
    const incommingMessage = buildUserMessage(message, callAttachments);
    const proxyResponse = await props.client
      .chatCompletion({
        target: isLocal
          ? { modelId: selectedModel.Id, chatId: chat?.id }
          : { sessionId: activeSession.Id, chatId: chat?.id },
        messages: [incommingMessage],
      })
      .catch((e) => {
        console.log('Failed to send request', e);
        return null;
      });

    if (!proxyResponse) {
      return;
    }

    if (!mountedRef.current || chatGenerationRef.current !== chatGeneration) {
      await proxyResponse.body?.cancel().catch(() => undefined);
      return memoState;
    }

    const response = new Response(proxyResponse.body, {
      status: proxyResponse.status,
      headers: { 'Content-Type': proxyResponse.contentType },
    });

    if (!response.ok) {
      // The provider's reason arrives wrapped in several layers of JSON. Show
      // the actionable part rather than a generic "Failed to send prompt",
      // which threw away the one piece of information the user needed.
      const body = await response.json().catch(() => null);
      if (!mountedRef.current || chatGenerationRef.current !== chatGeneration) {
        return memoState;
      }
      const detail = body?.error ?? body?.message ?? `HTTP ${response.status}`;
      console.error('Prompt failed:', detail);

      // A refusal is a definitive answer to "does this model do vision?" —
      // record it so next time the warning is a fact, not a guess.
      if (isVisionRejection(detail)) {
        rememberImageRejection(selectedModel?.Id);
      }

      const { message, hint } = explainChainError(detail);
      props.toasts.toast('error', hint ? `${message} ${hint}` : message, {
        autoClose: 15000,
      });
      return;
    }

    if (!response.body) {
      console.error('Body is missed');
      return;
    }

    const textDecoder = new TextDecoder();
    const reader = response.body.getReader();
    activeReaderRef.current = reader;

    const icon = modelName.toUpperCase()[0];
    const iconProps = {
      icon,
      color: getColor(icon),
      user: modelName,
      role: 'assistant',
    };
    const messageBatch = createAnimationFrameBatch<any[]>((nextMessages) => {
      if (!mountedRef.current || chatGenerationRef.current !== chatGeneration) {
        return;
      }
      setMessages(nextMessages);
      scheduleScrollToBottom();
    });
    try {
      let chunksBuffer = '';
      while (true) {
        if (abort) {
          await reader.cancel();
          abort = false;
        }

        const { value, done } = await reader.read();
        if (done) {
          if (
            mountedRef.current &&
            chatGenerationRef.current === chatGeneration
          ) {
            setIsSpinning(false);
          }
          break;
        }

        const decodedString = textDecoder.decode(value, { stream: true });

        chunksBuffer = chunksBuffer + decodedString;

        const { data: parts, isChunkIncomplete } =
          tryParseDataChunk(chunksBuffer);

        if (isChunkIncomplete) {
          continue;
        } else {
          chunksBuffer = '';
        }

        parts.forEach((part) => {
          if (!part) {
            return;
          }

          if (part.error) {
            // Mid-stream failures were only console.warn'd, so the chat just
            // stopped producing text with no explanation.
            console.error('Stream error:', part.error);
            if (isVisionRejection(part.error)) {
              rememberImageRejection(selectedModel?.Id);
            }
            const { message, hint } = explainChainError(part.error);
            props.toasts.toast('error', hint ? `${message} ${hint}` : message, {
              autoClose: 15000,
            });
            return;
          }

          if (typeof part === 'string') {
            handleSystemMessage(part);
            return;
          }

          const imageContent = part.imageUrl;
          const imageRawContent = part.imageRawContent;
          const videoRawContent = part.videoRawContent;

          if (
            !part?.id &&
            !imageContent &&
            !videoRawContent &&
            !imageRawContent
          ) {
            return;
          }

          let result: any[] = [];
          const message = memoState.find((m) => m.id == part.id);
          const otherMessages = memoState.filter((m) => m.id != part.id);

          if (imageRawContent) {
            result = [
              ...otherMessages,
              {
                id: makeId(16),
                text: imageRawContent,
                isImageContent: true,
                ...iconProps,
              },
            ];
          } else if (imageContent) {
            result = [
              ...otherMessages,
              {
                id: part.job,
                text: imageContent,
                isImageContent: true,
                ...iconProps,
              },
            ];
          } else if (videoRawContent) {
            result = [
              ...otherMessages,
              {
                id: part.job,
                text: videoRawContent,
                isVideoRawContent: true,
                ...iconProps,
              },
            ];
          } else {
            const text =
              `${message?.text || ''}${part?.choices[0]?.delta?.content || ''}`
                .replace('<|im_start|>', '')
                .replace('<|im_end|>', '');
            result = [
              ...otherMessages,
              { id: part.id, text: text, ...iconProps },
            ];
          }
          memoState = result;
          messageBatch.schedule(result);
        });
      }
    } catch (e) {
      if (mountedRef.current && chatGenerationRef.current === chatGeneration) {
        props.toasts.toast('error', 'Something goes wrong. Try later.');
        console.error(e);
      }
    } finally {
      if (activeReaderRef.current === reader) {
        activeReaderRef.current = null;
      }
      // requestAnimationFrame can be throttled while the window is hidden. A
      // synchronous final flush keeps persisted history and the visible state
      // aligned on completion, cancellation, and error paths.
      messageBatch.flush();
    }

    return memoState;
  };

  const buildInferenceTarget = () =>
    isLocal
      ? { modelId: selectedModel.Id, chatId: chat?.id }
      : { sessionId: activeSession.Id, chatId: chat?.id };

  const audioIconProps = () => {
    const icon = modelName.toUpperCase()[0];
    return {
      icon,
      color: getColor(icon),
      user: modelName,
      role: 'assistant',
    };
  };

  // TTS: text in -> synthesized audio out
  const callSpeech = async (text: string) => {
    const chatGeneration = chatGenerationRef.current;
    const userText = { id: makeId(16), text, ...userMessage };
    let memoState = [...messages, userText];
    setMessages(memoState);
    scheduleScrollToBottom();

    try {
      const response = await props.client.synthesizeSpeech({
        target: buildInferenceTarget(),
        text,
        voice: ttsVoice,
        speed: Number(ttsSpeed),
      });

      if (!response || !response.ok) {
        props.toasts.toast(
          'error',
          response?.error || 'Failed to synthesize speech',
        );
        return memoState;
      }

      if (!mountedRef.current || chatGenerationRef.current !== chatGeneration) {
        return memoState;
      }
      const blob = new Blob([response.data], {
        type: response.mimeType || 'audio/mpeg',
      });
      const url = URL.createObjectURL(blob);
      ownedAudioUrlsRef.current.add(url);
      memoState = [
        ...memoState,
        {
          id: makeId(16),
          text: url,
          isAudioContent: true,
          ...audioIconProps(),
        },
      ];
      setMessages(memoState);
      scheduleScrollToBottom();
    } catch (e) {
      props.toasts.toast('error', 'Something goes wrong. Try later.');
      console.error(e);
    }
    return memoState;
  };

  // STT: audio in -> transcription text out
  const callTranscription = async (file: File) => {
    const chatGeneration = chatGenerationRef.current;
    const userAudioUrl = URL.createObjectURL(file);
    ownedAudioUrlsRef.current.add(userAudioUrl);
    let memoState = [
      ...messages,
      {
        id: makeId(16),
        text: userAudioUrl,
        isAudioContent: true,
        ...userMessage,
      },
    ];
    setMessages(memoState);
    scheduleScrollToBottom();

    if (messages.length === 0 && chat) {
      setChatsData([
        ...chatData,
        { ...chat, title: file.name || 'Transcription' },
      ]);
    }

    try {
      const response = await props.client.transcribeAudio({
        target: buildInferenceTarget(),
        fileName: file.name,
        mimeType: file.type || 'application/octet-stream',
        data: await file.arrayBuffer(),
      });

      if (!response || !response.ok) {
        props.toasts.toast('error', 'Failed to transcribe audio');
        return memoState;
      }

      const contentType = response.contentType || '';
      let transcript = '';
      if (contentType.includes('application/json')) {
        const data = JSON.parse(response.body);
        transcript = data?.text ?? JSON.stringify(data);
      } else {
        transcript = response.body;
      }

      if (!mountedRef.current || chatGenerationRef.current !== chatGeneration) {
        return memoState;
      }

      memoState = [
        ...memoState,
        { id: makeId(16), text: transcript, ...audioIconProps() },
      ];
      setMessages(memoState);
      scheduleScrollToBottom();
    } catch (e) {
      props.toasts.toast('error', 'Something goes wrong. Try later.');
      console.error(e);
    }
    return memoState;
  };

  const handleAudioFile = (file?: File | null) => {
    if (!file || isDisabled) {
      return;
    }
    if (!file.size || file.size > MAX_AUDIO_BYTES) {
      props.toasts.toast('error', 'Audio must be between 1 byte and 20 MB.');
      return;
    }
    setIsSpinning(true);
    callTranscription(file).finally(() => setIsSpinning(false));
  };

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

      // Detect a silent/muted input (e.g. macOS handing us a denied mic track):
      // sample the peak amplitude while recording so we can warn the user
      // instead of submitting silence that transcribes to garbage.
      peakLevelRef.current = 0;
      try {
        const AudioCtx =
          window.AudioContext || (window as any).webkitAudioContext;
        const audioContext = new AudioCtx();
        audioContextRef.current = audioContext;
        const source = audioContext.createMediaStreamSource(stream);
        const analyser = audioContext.createAnalyser();
        analyser.fftSize = 2048;
        source.connect(analyser);
        const data = new Uint8Array(analyser.fftSize);
        const sampleLevel = () => {
          analyser.getByteTimeDomainData(data);
          let peak = 0;
          for (let i = 0; i < data.length; i++) {
            peak = Math.max(peak, Math.abs(data[i] - 128));
          }
          peakLevelRef.current = Math.max(peakLevelRef.current, peak);
          levelRafRef.current = requestAnimationFrame(sampleLevel);
        };
        sampleLevel();
      } catch (levelErr) {
        console.warn('Could not set up audio level monitoring', levelErr);
      }

      audioChunksRef.current = [];
      const recorder = new MediaRecorder(stream);
      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) {
          audioChunksRef.current.push(e.data);
        }
      };
      recorder.onstop = () => {
        stream.getTracks().forEach((t) => t.stop());
        if (levelRafRef.current != null) {
          cancelAnimationFrame(levelRafRef.current);
          levelRafRef.current = null;
        }
        audioContextRef.current?.close().catch(() => {});
        audioContextRef.current = null;

        // Peak is 0..127 (deviation from the 128 silence midpoint). A few
        // counts of jitter is still effectively silence.
        if (peakLevelRef.current <= 2) {
          props.toasts.toast(
            'error',
            'No sound was captured. Check microphone permissions and that the correct input device is selected.',
          );
          return;
        }

        const blob = new Blob(audioChunksRef.current, { type: 'audio/webm' });
        const file = new File([blob], `recording-${Date.now()}.webm`, {
          type: 'audio/webm',
        });
        handleAudioFile(file);
      };
      mediaRecorderRef.current = recorder;
      recorder.start();
      setRecording(true);
    } catch (e) {
      props.toasts.toast('error', 'Microphone access was denied');
      console.error(e);
    }
  };

  const stopRecording = () => {
    mediaRecorderRef.current?.stop();
    mediaRecorderRef.current = null;
    setRecording(false);
  };

  const handleSystemMessage = (message) => {
    const openSessionEventMessage = 'new session opened';
    const failoverTurnOnMessage = 'provider failed, failover enabled';

    const renderMessage = (value) => {
      props.toasts.toast('info', value, {
        autoClose: 1500,
      });
    };

    if (message.includes(openSessionEventMessage)) {
      const sessionId = message.split(':')[1].trim(); // new session opened: 0x123456
      setSessionData(sessionId).catch((err) =>
        renderMessage(`Failed to load session data: ${err.message}`),
      );
      renderMessage('Opening session with available provider...');
      return;
    }
    if (message.includes(failoverTurnOnMessage)) {
      renderMessage('Target provider unavailable. Applying failover policy...');
      return;
    }
    renderMessage(message);
    return;
  };

  /**
   * Adds files to the pending attachment list.
   *
   * Documents cross IPC as bounded binary and are handed to the main process
   * for hardened text extraction. Images are kept as a data URI and sent as an
   * image_url part. Each file lands in the list immediately with status
   * 'parsing' so a slow document doesn't look like nothing happened.
   */
  const addFiles = async (files: File[]) => {
    for (const file of files) {
      const reason = validateFile(file, attachments.length);
      if (reason) {
        props.toasts.toast('error', reason, { autoClose: 8000 });
        continue;
      }

      const kind = classify(file);
      const id = makeId(12);

      setAttachments((prev) => [
        ...prev,
        {
          id,
          name: file.name,
          mime: file.type,
          size: file.size,
          kind,
          status: 'parsing',
        },
      ]);

      try {
        if (kind === 'image') {
          const { dataUrl } = await readFile(file);
          setAttachments((prev) =>
            prev.map((a) =>
              a.id === id ? { ...a, dataUrl, status: 'ready' as const } : a,
            ),
          );
          continue;
        }

        const parsed = await props.client.parseAttachment({
          name: file.name,
          mime: file.type,
          data: await file.arrayBuffer(),
        });

        setAttachments((prev) =>
          prev.map((a) =>
            a.id === id
              ? {
                  ...a,
                  text: parsed.text,
                  note: parsed.note,
                  empty: parsed.empty,
                  status: 'ready' as const,
                }
              : a,
          ),
        );
      } catch (e: any) {
        setAttachments((prev) =>
          prev.map((a) =>
            a.id === id
              ? { ...a, status: 'error' as const, error: e?.message }
              : a,
          ),
        );
        props.toasts.toast(
          'error',
          e?.message || `Could not read ${file.name}`,
        );
      }
    }
  };

  const removeAttachment = (id: string) =>
    setAttachments((prev) => prev.filter((a) => a.id !== id));

  // Attachments only make sense for text/vision chat. TTS synthesises the text
  // you type, and STT has its own audio input.
  const attachmentsSupported =
    Boolean(selectedModel) && modality === 'llm' && !isReadonly;

  const handlePaste = (e: React.ClipboardEvent) => {
    if (!attachmentsSupported) return;
    const files = Array.from(e.clipboardData?.files ?? []);
    if (files.length) {
      e.preventDefault();
      addFiles(files);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    if (!attachmentsSupported) return;
    const files = Array.from(e.dataTransfer?.files ?? []);
    if (files.length) {
      e.preventDefault();
      addFiles(files);
    }
  };

  const handleSubmit = () => {
    if (abort) {
      abort = false;
    }

    if (isSpinning) {
      abort = true;
      setIsSpinning(false);
      return;
    }

    if (isDisabled) {
      if (!isLocal) {
        props.toasts.toast(
          'info',
          'This session is closed or expired. Open a new session before sending another message.',
        );
      }
      return;
    }

    const readyAttachments = attachments.filter((a) => a.status === 'ready');

    // A message with only attachments is meaningful ("summarise this"), but a
    // completely empty one is not.
    if (!promptInput && !readyAttachments.length) {
      return;
    }

    // Sending while a PDF is still being read would silently drop it.
    if (attachments.some((a) => a.status === 'parsing')) {
      props.toasts.toast(
        'info',
        'Still reading your attachments — one moment.',
      );
      return;
    }

    if (messages.length === 0 && chat) {
      const title = {
        ...chat,
        title: promptInput || readyAttachments[0]?.name,
      };
      setChatsData([...chatData, title]);
    }

    setIsSpinning(true);
    const requestGeneration = chatGenerationRef.current;
    const request =
      modality === 'tts'
        ? callSpeech(promptInput)
        : call(promptInput, readyAttachments);
    request.finally(() => {
      if (
        mountedRef.current &&
        chatGenerationRef.current === requestGeneration
      ) {
        setIsSpinning(false);
      }
    });
    setPromptInput('');
    setAttachments([]);
  };

  const deleteChatEntry = (id: string) => {
    props.client
      .deleteChatHistory(id)
      .then(() => {
        const newChats = chatData.filter((x) => x.id != id);
        setChatsData(newChats);
      })
      .catch(console.error);
  };

  const onCreateNewChat = ({ modelId, isLocal }) => {
    if (isLocal) {
      props.toasts.toast(
        'info',
        'The local model is a read-only legacy demo. Choose a Morpheus marketplace model and open a session to chat.',
      );
      return;
    }

    const model = (allModels ?? chainData?.models ?? []).find(
      (candidate: any) => !candidate.isLocal && candidate.Id == modelId,
    );
    if (!model) {
      props.toasts.toast(
        'info',
        'Model details are still loading. Please try again in a moment.',
      );
      return;
    }

    // A deliberate model choice owns the screen. If the slower session query
    // completes afterwards, the one-time bootstrap must not replace it with a
    // different historical session.
    initializedRef.current = true;
    abort = true;
    chatGenerationRef.current += 1;
    autoScrollRef.current = true;
    setMessages([]);
    setActiveSession(undefined);
    setSelectedBid(undefined);
    setIsReadonly(false);
    setChat({ id: generateHashId(), createdAt: new Date(), modelId });
    // Selecting registry metadata starts exactly one cached active-bids query
    // for this model. It does not wait for every model in the marketplace.
    setSelectedModel(model);

    const openSessions = sessions.filter((s) => !isClosed(s));
    const openModelSession = openSessions.find(
      (s) => s.ModelAgentId == modelId,
    );

    if (openModelSession) {
      setActiveSession(openModelSession);
    }
  };

  const wrapChangeTitle = async (data: { id; title }) => {
    await props.client.updateChatHistoryTitle(data);
  };

  const renderChatBlock = () => {
    if (!selectedModel) {
      const sessionsAreLoading = isInitialQueryFetchActive(
        sessionsQuery.data,
        sessionsQuery.fetchStatus,
      );
      return (
        <ChatStartupState role="status" aria-live="polite">
          <IconMessagePlus size={30} stroke={1.7} aria-hidden="true" />
          <strong>
            {modelsLoading
              ? 'Loading available models…'
              : !props.address
                ? 'Wallet connection isn’t ready'
                : sessionsAreLoading
                  ? 'Checking your active sessions…'
                  : 'Choose a model to start a chat'}
          </strong>
          <span>
            {modelsLoading
              ? 'The Chat screen is ready. Models will appear as soon as your node responds.'
              : !props.address
                ? 'You can browse models now. Open Wallet to finish connecting before starting a paid session.'
                : sessionsAreLoading
                  ? 'You can browse models now while the session list finishes loading.'
                  : 'Open New chat to browse current Morpheus marketplace models.'}
          </span>
          <ChatIntroButton
            onClick={() => {
              setCoworkModelSelection(false);
              setOpenChangeModal(true);
            }}
            type="button"
          >
            Browse models
          </ChatIntroButton>
        </ChatStartupState>
      );
    }

    // `meta` falls back to { budget: 0, supply: 0 } while the models query is
    // loading or has failed. Dividing by a zero budget produced NaN, and every
    // `x > NaN` comparison is false — which silently disabled *both* payment
    // buttons with no explanation. Treat unknown pricing as "not ready yet" and
    // say so, rather than rendering a dead screen.
    const selectedModelBids = selectedModelBidsQuery.data ?? [];
    const prices = selectedModelBids.map((x: any) => Number(x.PricePerSecond));
    const maxPrice = prices.length ? Math.max(...prices) : Number.NaN;
    const isPricingReady =
      fundingQuery.data !== undefined &&
      Number(meta.budget) > 0 &&
      Number(meta.supply) > 0 &&
      Number.isFinite(maxPrice);

    const requiredStake = isPricingReady
      ? {
          min: estimateSessionTokenAmount(
            maxPrice,
            SESSION_DURATION_OPTIONS[0].seconds,
            false,
            meta,
          ),
          max: estimateSessionTokenAmount(
            maxPrice,
            SESSION_DURATION_OPTIONS[SESSION_DURATION_OPTIONS.length - 1]
              .seconds,
            false,
            meta,
          ),
        }
      : null;

    // for stake mode
    const selectedStake = isPricingReady
      ? estimateSessionTokenAmount(maxPrice, sessionDuration, false, meta)
      : Number.POSITIVE_INFINITY;
    const requiredStakeForDirectPay = isPricingReady
      ? estimateSessionTokenAmount(maxPrice, sessionDuration, true, meta)
      : Number.POSITIVE_INFINITY;
    const hasSelectedStakeFunds =
      isPricingReady && Number(balances.mor) >= selectedStake;
    const isEnoughFundsForDirectPay =
      isPricingReady && Number(balances.mor) >= requiredStakeForDirectPay;

    const bidsAreLoading = isInitialQueryFetchActive(
      selectedModelBidsQuery.data,
      selectedModelBidsQuery.fetchStatus,
    );
    const bidsFailed =
      selectedModelBidsQuery.isError && selectedModelBids.length === 0;
    const bidsCouldNotStart =
      Boolean(props.address) &&
      selectedModelBidsQuery.data === undefined &&
      !bidsAreLoading &&
      !bidsFailed;
    const noActiveProviders =
      selectedModelBidsQuery.data !== undefined &&
      !selectedModelBidsQuery.isError &&
      selectedModelBids.length === 0;
    const fundingIsLoading = isInitialQueryFetchActive(
      fundingQuery.data,
      fundingQuery.fetchStatus,
    );
    const fundingFailed =
      fundingQuery.data === undefined && fundingQuery.isError;
    const fundingCouldNotStart =
      Boolean(props.address) &&
      fundingQuery.data === undefined &&
      !fundingIsLoading &&
      !fundingFailed;
    const pricingUsesCachedData =
      (selectedModelBidsQuery.isError && selectedModelBids.length > 0) ||
      (fundingQuery.isError && fundingQuery.data !== undefined);
    const showModelPicker = () => {
      setCoworkModelSelection(false);
      setOpenChangeModal(true);
    };

    // The user may already hold an open session for this model. Surfacing it
    // here is what stops people staking a second time when the first session
    // simply hadn't been re-selected yet.
    const openSessionsForModel = (sessions || []).filter(
      (s: any) => !isClosed(s) && s.ModelAgentId == selectedModel?.Id,
    );

    return (
      <>
        {isCreateSessionMode ? (
          <ChatIntroContainer>
            <ChatIntroInner>
              <ChatIntroInnerTitle>Select payment method</ChatIntroInnerTitle>
              {!props.address ? (
                <SessionSetupState role="alert">
                  <strong>Wallet connection isn’t ready</strong>
                  <span>
                    Pricing and payment will become available after the app
                    finishes connecting to your wallet.
                  </span>
                  <SessionSetupActions>
                    <ChatIntroButton
                      type="button"
                      onClick={() => navigate('/wallet')}
                    >
                      Open Wallet
                    </ChatIntroButton>
                    <ChatIntroButton type="button" onClick={showModelPicker}>
                      Change model
                    </ChatIntroButton>
                  </SessionSetupActions>
                </SessionSetupState>
              ) : bidsAreLoading ? (
                <SessionSetupState
                  role="status"
                  aria-live="polite"
                  aria-busy="true"
                >
                  <Spinner animation="border" variant="success" />
                  <strong>Checking current price</strong>
                  <span>
                    Looking for an active provider for {selectedModel.Name}.
                  </span>
                  <SessionSetupActions>
                    <ChatIntroButton type="button" onClick={showModelPicker}>
                      Change model
                    </ChatIntroButton>
                  </SessionSetupActions>
                </SessionSetupState>
              ) : bidsFailed || bidsCouldNotStart ? (
                <SessionSetupState role="alert">
                  <strong>Couldn’t load this model’s price</strong>
                  <span>
                    {bidsCouldNotStart
                      ? 'The provider check is waiting for your node connection.'
                      : 'The provider check failed. Your balance has not been charged.'}
                  </span>
                  <SessionSetupActions>
                    <ChatIntroButton
                      type="button"
                      onClick={() => selectedModelBidsQuery.refetch()}
                    >
                      Retry
                    </ChatIntroButton>
                    <ChatIntroButton type="button" onClick={showModelPicker}>
                      Change model
                    </ChatIntroButton>
                  </SessionSetupActions>
                </SessionSetupState>
              ) : noActiveProviders ? (
                <SessionSetupState role="status" aria-live="polite">
                  <strong>No provider currently offers this model</strong>
                  <span>
                    Availability can change. Retry this model or choose another
                    one.
                  </span>
                  <SessionSetupActions>
                    <ChatIntroButton
                      type="button"
                      onClick={() => selectedModelBidsQuery.refetch()}
                    >
                      Retry
                    </ChatIntroButton>
                    <ChatIntroButton type="button" onClick={showModelPicker}>
                      Change model
                    </ChatIntroButton>
                  </SessionSetupActions>
                </SessionSetupState>
              ) : fundingIsLoading ? (
                <SessionSetupState
                  role="status"
                  aria-live="polite"
                  aria-busy="true"
                >
                  <Spinner animation="border" variant="success" />
                  <strong>Loading balance and session pricing</strong>
                  <span>
                    The model price is ready. Finishing the MOR estimate.
                  </span>
                  <SessionSetupActions>
                    <ChatIntroButton type="button" onClick={showModelPicker}>
                      Change model
                    </ChatIntroButton>
                  </SessionSetupActions>
                </SessionSetupState>
              ) : fundingFailed || fundingCouldNotStart ? (
                <SessionSetupState role="alert">
                  <strong>Couldn’t load balance and session pricing</strong>
                  <span>
                    {fundingCouldNotStart
                      ? 'The estimate is waiting for your node connection.'
                      : 'Retry the estimate before choosing a payment method.'}
                  </span>
                  <SessionSetupActions>
                    <ChatIntroButton
                      type="button"
                      onClick={() => fundingQuery.refetch()}
                    >
                      Retry
                    </ChatIntroButton>
                    <ChatIntroButton type="button" onClick={showModelPicker}>
                      Change model
                    </ChatIntroButton>
                  </SessionSetupActions>
                </SessionSetupState>
              ) : (
                <>
                  <SessionDurationField>
                    Session length
                    <select
                      aria-label="Session length"
                      value={sessionDuration}
                      onChange={(event) =>
                        setSessionDuration(Number(event.target.value))
                      }
                    >
                      {SESSION_DURATION_OPTIONS.map((option) => (
                        <option key={option.seconds} value={option.seconds}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </SessionDurationField>

                  {openSessionsForModel.length > 0 && (
                    <ChatIntroInnerText style={{ color: '#20dc8e' }}>
                      You already have {openSessionsForModel.length} open
                      session{openSessionsForModel.length > 1 ? 's' : ''} for
                      this model. Open it from the Sessions list in the sidebar
                      instead of staking again — staking again locks additional
                      MOR.
                    </ChatIntroInnerText>
                  )}

                  {sessionsQuery.isError && (
                    <SessionHistoryNotice role="status">
                      <span>
                        Session history couldn’t refresh. You can still
                        continue; your node will check for an existing session
                        before opening another.
                      </span>
                      <button
                        type="button"
                        onClick={() => sessionsQuery.refetch()}
                      >
                        Retry history
                      </button>
                    </SessionHistoryNotice>
                  )}
                  {pricingUsesCachedData && (
                    <SessionHistoryNotice role="status">
                      <span>
                        The latest price or balance refresh failed. This
                        estimate uses recently cached data; the node will still
                        validate the session when you continue.
                      </span>
                      <button
                        type="button"
                        onClick={() => {
                          void selectedModelBidsQuery.refetch();
                          void fundingQuery.refetch();
                        }}
                      >
                        Retry estimate
                      </button>
                    </SessionHistoryNotice>
                  )}
                  <ChatIntroInnerText>
                    Stake MOR to reserve compute for the session length selected
                    above (min:{' '}
                    {requiredStake
                      ? `${formatValue(requiredStake.min, 18)} MOR`
                      : 'calculating…'}
                    , max:{' '}
                    {requiredStake
                      ? `${formatValue(requiredStake.max, 18)} MOR`
                      : 'calculating…'}
                    ). The MOR is escrowed, and unused stake returns when the
                    session closes.
                  </ChatIntroInnerText>
                  <SessionSetupActions>
                    <ChatIntroButton
                      onClick={() => onOpenSession(false, false)}
                      disabled={!hasSelectedStakeFunds}
                    >
                      Stake MOR
                    </ChatIntroButton>
                  </SessionSetupActions>
                  <SessionCostSummary>
                    Estimated stake for this length:{' '}
                    {Number.isFinite(selectedStake)
                      ? `${formatValue(selectedStake, 18)} MOR`
                      : 'calculating…'}
                    {!hasSelectedStakeFunds && isPricingReady
                      ? ' — insufficient balance'
                      : ''}
                  </SessionCostSummary>
                  <ChatIntroInnerText>
                    Pay with your MOR tokens directly for the session length
                    selected above.
                  </ChatIntroInnerText>
                  <SessionSetupActions>
                    <ChatIntroButton
                      onClick={() => onOpenSession(false, true)}
                      disabled={!isEnoughFundsForDirectPay}
                    >
                      Direct Pay
                    </ChatIntroButton>
                  </SessionSetupActions>
                  <SessionCostSummary>
                    Estimated direct payment:{' '}
                    {Number.isFinite(requiredStakeForDirectPay)
                      ? `${formatValue(requiredStakeForDirectPay, 18)} MOR`
                      : 'calculating…'}
                    {!isEnoughFundsForDirectPay && isPricingReady
                      ? ' — insufficient balance'
                      : ''}
                  </SessionCostSummary>
                </>
              )}
            </ChatIntroInner>
          </ChatIntroContainer>
        ) : (
          <ChatHistoryContainer ref={attachChatScrollElement}>
            {messages?.map((x, index) => (
              <Message
                key={x.id ?? index}
                message={x}
                onOpenImage={setImagePreview}
              />
            ))}
          </ChatHistoryContainer>
        )}
      </>
    );
  };

  // If the models query failed there is no marketplace to render at all, so
  // show the reason instead of an empty shell with dead buttons. Previously the
  // main-process handler swallowed the error and returned [], which made a
  // down proxy-router look identical to "no models exist".
  if (
    marketplaceModelsQuery.isError &&
    marketplaceModelsQuery.data === undefined
  ) {
    return (
      <View data-testid="chat-container">
        <QueryError
          error={marketplaceModelsQuery.error}
          what="models"
          onRetry={() => marketplaceModelsQuery.refetch()}
        />
      </View>
    );
  }

  return (
    <>
      {isLoading && (
        <LoadingCover role="status" aria-live="polite" aria-busy="true">
          <LoadingStatus>
            <Spinner
              style={{ width: '4rem', height: '4rem' }}
              animation="border"
              variant="success"
            />
            <strong>{actionStatus}</strong>
            <span>Please keep the app open while this finishes.</span>
          </LoadingStatus>
        </LoadingCover>
      )}

      {/* Non-fatal: models loaded from cache but the latest refresh failed. */}
      {marketplaceModelsQuery.isError &&
        marketplaceModelsQuery.data !== undefined && (
          <QueryError
            error={marketplaceModelsQuery.error}
            what="the latest model data"
            onRetry={() => marketplaceModelsQuery.refetch()}
          />
        )}
      <Drawer
        open={isOpen}
        onClose={toggleDrawer}
        direction="right"
        className="history-drawer"
      >
        <ChatHistory
          activeChat={chat}
          open={isOpen}
          chatData={chatData}
          sessions={sessions}
          deleteHistory={deleteChatEntry}
          models={chainData?.models || []}
          onSelectChat={selectChat}
          refreshSessions={async () => {
            setActionStatus('Refreshing sessions…');
            try {
              await refreshSessions();
            } finally {
              setActionStatus(null);
            }
          }}
          onChangeTitle={wrapChangeTitle}
          onCloseSession={closeSession}
        />
      </Drawer>
      <View>
        <ContainerTitle>
          <TitleRow>
            {/* <Title>Chat</Title> */}
            <div className="d-flex" style={{ alignItems: 'center' }}>
              <div className="d-flex model-selector">
                <div className="model-selector__info">
                  <h3>{isLocal ? '(local)' : providerAddress}</h3>
                  {isLocal ? (
                    <>
                      <span>0 MOR</span>
                    </>
                  ) : (
                    <>
                      <SubPriceLabel>{stakedFunds} MOR</SubPriceLabel>
                    </>
                  )}
                </div>
                {!isLocal && activeSession?.EndsAt && (
                  <div className="model-selector__icons">
                    <Cooldown endDate={activeSession?.EndsAt} />
                  </div>
                )}
              </div>
              <BtnAccent
                className="change-modal"
                onClick={() => {
                  setCoworkModelSelection(false);
                  setOpenChangeModal(true);
                }}
              >
                <IconMessagePlus></IconMessagePlus> New chat
              </BtnAccent>
              {activeSession?.Id &&
                !marketplaceSessionUnavailable &&
                isCoworkCandidate(selectedModel) && (
                  <BtnAccent
                    className="change-modal"
                    onClick={() =>
                      navigate(
                        `/workspace?sessionId=${encodeURIComponent(activeSession.Id)}`,
                      )
                    }
                  >
                    <IconSparkles size={18} /> Use this session in Workspace
                  </BtnAccent>
                )}
            </div>
          </TitleRow>
        </ContainerTitle>
        <ChatTitleContainer>
          <ChatAvatar>
            <Avatar
              style={{ color: 'white' }}
              color={getColor(modelName.toUpperCase()[0])}
            >
              {modelName.toUpperCase()[0]}
            </Avatar>
            <div style={{ marginLeft: '10px' }}>{modelName}</div>
            {isSecure && (
              <span
                title={SECURE_BADGE_TOOLTIP}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '4px',
                  marginLeft: '10px',
                  padding: '2px 8px 2px 6px',
                  fontSize: '1.1rem',
                  fontWeight: 600,
                  letterSpacing: '0.3px',
                  color: 'rgba(173, 211, 255, 0.95)',
                  background: 'rgba(125, 188, 255, 0.14)',
                  borderRadius: '6px',
                  cursor: 'default',
                }}
              >
                <IconShieldLock size={13} stroke={2.2} /> Secure
              </span>
            )}
          </ChatAvatar>
          {/* {
                        (selectedBid || isLocal) && <div>
                            <span style={{ color: 'white' }}>Provider:</span> {isLocal ? "(local)" : providerAddress}
                        </div>
                    } */}
          <div>
            <div onClick={toggleDrawer}>
              <IconHistory size={'2.4rem'}></IconHistory>
            </div>
          </div>
        </ChatTitleContainer>

        {imagePreview && (
          <ImageViewer
            src={[imagePreview]}
            onClose={() => setImagePreview('')}
            disableScroll={false}
            backgroundStyle={{
              backgroundColor: 'rgba(0,0,0,0.9)',
              zIndex: 1000,
            }}
            closeOnClickOutside={true}
          />
        )}

        <Container>
          {renderChatBlock()}
          <Control>
            {modality === 'stt' && !isReadonly ? (
              <AudioInputZone data-disabled={isDisabled}>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="audio/*"
                  style={{ display: 'none' }}
                  onChange={(e) => {
                    handleAudioFile(e.target.files?.[0]);
                    e.target.value = '';
                  }}
                />
                <AudioActionBtn
                  type="button"
                  disabled={isDisabled || isSpinning || recording}
                  onClick={() => fileInputRef.current?.click()}
                >
                  <IconUpload size={16} /> Upload audio
                </AudioActionBtn>
                <AudioActionBtn
                  type="button"
                  data-recording={recording}
                  disabled={isDisabled || isSpinning}
                  onClick={() =>
                    recording ? stopRecording() : startRecording()
                  }
                >
                  {recording ? (
                    <>
                      <IconPlayerStopFilled size={16} /> Stop
                    </>
                  ) : (
                    <>
                      <IconMicrophone size={16} /> Record
                    </>
                  )}
                </AudioActionBtn>
                {isSpinning && <Spinner animation="border" size="sm" />}
                <AudioHint>
                  {recording
                    ? 'Recording… click Stop to transcribe.'
                    : 'Upload or record audio to transcribe.'}
                </AudioHint>
              </AudioInputZone>
            ) : (
              <>
                {modality === 'tts' && !isReadonly && (
                  <TtsControlsRow>
                    <label>
                      Voice
                      <input
                        type="text"
                        list="tts-voices"
                        value={ttsVoice}
                        onChange={(e) => setTtsVoice(e.target.value)}
                      />
                      <datalist id="tts-voices">
                        {TTS_VOICES.map((v) => (
                          <option key={v} value={v} />
                        ))}
                      </datalist>
                    </label>
                    <label>
                      Speed
                      <input
                        type="range"
                        min={0.5}
                        max={2}
                        step={0.25}
                        value={ttsSpeed}
                        onChange={(e) => setTtsSpeed(Number(e.target.value))}
                      />
                      {ttsSpeed}x
                    </label>
                  </TtsControlsRow>
                )}
                {attachmentsSupported && (
                  <>
                    <AttachmentBar
                      attachments={attachments}
                      prompt={promptInput}
                      onRemove={removeAttachment}
                      visionWarning={!looksVisionCapable(selectedModel)}
                      visionRejected={hasRejectedImages(selectedModel?.Id)}
                      modelName={selectedModel?.Name}
                    />
                    <input
                      ref={attachInputRef}
                      type="file"
                      multiple
                      accept="image/*,.pdf,.docx,.svg,.txt,.md,.csv,.json,.xml,.yaml,.yml,.log,.ts,.tsx,.js,.jsx,.py,.go,.rs,.java,.c,.h,.cpp,.sh,.sql,.html,.css"
                      style={{ display: 'none' }}
                      onChange={(e) => {
                        addFiles(Array.from(e.target.files ?? []));
                        e.target.value = '';
                      }}
                    />
                  </>
                )}
                <CustomTextArrea
                  disabled={isDisabled}
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleSubmit();
                    }
                  }}
                  onPaste={handlePaste}
                  onDrop={handleDrop}
                  onDragOver={(e) => attachmentsSupported && e.preventDefault()}
                  value={promptInput}
                  onChange={(ev) => setPromptInput(ev.target.value)}
                  placeholder={
                    !selectedModel
                      ? 'Choose a model to start a chat'
                      : isReadonly
                        ? 'Session is closed. Chat in ReadOnly Mode'
                        : modality === 'tts'
                          ? 'Enter text to synthesize...'
                          : 'Ask me anything, or drop in a file...'
                  }
                  minRows={1}
                  maxRows={6}
                />
                <SendBtnWrapper>
                  {isReadonly && isLocal ? (
                    <AudioHint>
                      Legacy local-demo history is read-only. Choose a
                      marketplace model and open a session to continue.
                    </AudioHint>
                  ) : isReadonly ? (
                    <>
                      <Btn onClick={() => handleReopen(false)}>
                        {isSpinning ? (
                          <Spinner animation="border" />
                        ) : (
                          <span>Staking</span>
                        )}
                      </Btn>
                      <Btn onClick={() => handleReopen(true)}>
                        {isSpinning ? (
                          <Spinner animation="border" />
                        ) : (
                          <span>Direct Pay</span>
                        )}
                      </Btn>
                    </>
                  ) : (
                    <>
                      {attachmentsSupported && (
                        <Btn
                          disabled={isDisabled || isSpinning}
                          onClick={() => attachInputRef.current?.click()}
                          title="Attach images or documents"
                        >
                          <IconPaperclip size={'22px'} />
                        </Btn>
                      )}
                      <Btn disabled={isDisabled} onClick={handleSubmit}>
                        {isSpinning ? (
                          <Spinner animation="border" />
                        ) : (
                          <IconArrowUp size={'26px'}></IconArrowUp>
                        )}
                      </Btn>
                    </>
                  )}
                </SendBtnWrapper>
              </>
            )}
          </Control>
        </Container>
      </View>
      <ModelSelectionModal
        models={(chainData as any)?.models}
        modelsLoading={modelsLoading}
        isActive={openChangeModal}
        marketplaceOnly
        coworkSetup={coworkModelSelection}
        symbol={props.symbol}
        onChangeModel={(eventData) => {
          onCreateNewChat(eventData);
        }}
        handleClose={() => {
          setOpenChangeModal(false);
          setCoworkModelSelection(false);
        }}
      />
    </>
  );
};

const renderMessage = (message, onOpenImage) => {
  if (message.isAudioContent) {
    return (
      <MessageBody>
        <AudioPlayer controls src={message.text} />
      </MessageBody>
    );
  }

  if (message.isImageContent) {
    return (
      <MessageBody>
        {
          <ImageContainer
            src={message.text}
            onClick={() => onOpenImage(message.text)}
          />
        }
      </MessageBody>
    );
  }

  if (message.isVideoRawContent) {
    return (
      <MessageBody>
        <VideoContainer>
          <video controls src={`${message.text}`} />
        </VideoContainer>
      </MessageBody>
    );
  }

  return (
    <MessageBody>
      {/* Attachments the user sent with this message. Shown so the transcript
          reflects what was actually submitted — the extracted document text is
          deliberately not rendered, since a 40-page PDF pasted into the log
          would bury the conversation. */}
      {Array.isArray(message.attachments) && message.attachments.length > 0 && (
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: '8px',
            marginBottom: '10px',
          }}
        >
          {message.attachments.map((a, i) =>
            a.dataUrl ? (
              <img
                key={i}
                src={a.dataUrl}
                alt={a.name}
                title={a.name}
                onClick={() => onOpenImage(a.dataUrl)}
                style={{
                  maxWidth: '160px',
                  maxHeight: '160px',
                  borderRadius: '8px',
                  cursor: 'pointer',
                  border: '1px solid rgba(255,255,255,0.12)',
                }}
              />
            ) : (
              <span
                key={i}
                title={a.name}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '6px',
                  padding: '4px 10px',
                  borderRadius: '6px',
                  fontSize: '1.1rem',
                  background: 'rgba(255,255,255,0.06)',
                  border: '1px solid rgba(255,255,255,0.1)',
                  maxWidth: '220px',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                📎 {a.name}
              </span>
            ),
          )}
        </div>
      )}
      <ThinkingMessageBody text={message.text} />
    </MessageBody>
  );
};

const Message = memo(
  ({
    message,
    onOpenImage,
  }: {
    message: any;
    onOpenImage: (url: string) => void;
  }) => {
    return (
      <div style={{ display: 'flex', margin: '12px 0 28px 0' }}>
        <Avatar color={message.color}>{message.icon}</Avatar>
        <div>
          <AvatarHeader>{message.user}</AvatarHeader>
          {renderMessage(message, onOpenImage)}
        </div>
      </div>
    );
  },
);

// withChatState injects props that are loosely typed in its HOC signature;
// cast to suppress the HOC-vs-component prop mismatch.
export default withChatState(Chat as React.ComponentType<any>);
