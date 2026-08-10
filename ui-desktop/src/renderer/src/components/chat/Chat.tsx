import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
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
} from '@tabler/icons-react';
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
  ChatIntroInner,
  ChatIntroInnerTitle,
  ChatIntroInnerText,
  ChatIntroButton,
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
} from './utils';
import { Cooldown } from './Cooldown';
import ImageViewer from 'react-simple-image-viewer';
import { ChatData, HistoryMessage } from './interfaces';
import { formatValue } from '../../utils/coinValue';
import { ApiGateway } from 'src/main/src/client/apiGateway';
import { queryKeys } from '../../store/queries';
import { pooledMapSettled } from '../../store/utils/concurrency';
import QueryError from '../common/QueryError';

// Max simultaneous per-model bid requests. See store/utils/concurrency.ts.
const BIDS_CONCURRENCY = 6;

let abort = false;
let cancelScroll = false;
const userMessage = { user: 'Me', role: 'user', icon: 'M', color: '#20dc8e' };

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
  getModelsData: () => Promise<any>;
  getSessionsByUser: (address: string) => Promise<any>;
  getProvidersAvailability: (providers: any[]) => Promise<any[]>;
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
  const chatBlockRef = useRef<null | HTMLDivElement>(null);
  const queryClient = useQueryClient();
  const initializedRef = useRef(false);

  const [promptInput, setPromptInput] = useState('');
  // Overlay shown during user-triggered actions (open/close/reopen session,
  // manual session refresh). The *initial* page load no longer uses this — it
  // is gated on the react-query cache so revisiting the tab is instant.
  const [isActionLoading, setIsActionLoading] = useState(false);
  const [initialized, setInitialized] = useState(false);
  const [messages, setMessages] = useState<any>([]);
  const [isOpen, setIsOpen] = useState(false);

  const [isSpinning, setIsSpinning] = useState(false);

  const [imagePreview, setImagePreview] = useState<string>();
  const [activeSession, setActiveSession] = useState<any>(undefined);

  const [chatData, setChatsData] = useState<ChatData[]>([]);

  const [openChangeModal, setOpenChangeModal] = useState(false);
  const [isReadonly, setIsReadonly] = useState(false);

  const [selectedBid, setSelectedBid] = useState<any>(null);
  const [selectedModel, setSelectedModel] = useState<any>(undefined);
  const [requiredStake, setRequiredStake] = useState<{
    min: number;
    max: number;
  }>({ min: 0, max: 0 });

  const [chat, setChat] = useState<ChatData | undefined>(undefined);

  // --- Cached data layer (stale-while-revalidate via react-query) ---------
  // These queries live in the app-level QueryClient, so navigating away from
  // and back to /chat serves cached data instantly and revalidates silently
  // instead of blocking behind a full-screen spinner.

  const modelsDataQuery = useQuery({
    queryKey: queryKeys.modelsData,
    queryFn: () => props.getModelsData(),
  });

  const sessionsQuery = useQuery({
    queryKey: queryKeys.sessions(props.address),
    queryFn: () => props.getSessionsByUser(props.address),
    enabled: !!props.address,
  });

  const chatTitlesQuery = useQuery({
    queryKey: queryKeys.chatTitles,
    queryFn: () => props.client.getChatHistoryTitles(),
  });

  // Bid fan-out for every marketplace model. Runs in the background after the
  // base model list is available; does NOT gate the initial render. Mirrors the
  // previous "effect #2" merge logic but cached across visits.
  const modelsWithBidsQuery = useQuery({
    queryKey: queryKeys.modelsWithBids,
    enabled: !!modelsDataQuery.data,
    queryFn: async () => {
      const md = modelsDataQuery.data;
      const providersMap = md.providers.reduce(
        (a: any, b: any) => ({ ...a, [b.Address.toLowerCase()]: b }),
        {},
      );
      // Bounded fan-out: one bid request per model, but at most
      // BIDS_CONCURRENCY in flight. The previous unbounded Promise.all fired
      // hundreds of simultaneous IPC calls, which saturated the proxy-router
      // and blocked the renderer while the Chat tab loaded.
      const merged = (
        await pooledMapSettled(
          md.models,
          async (m: any) => {
            const id = m.Id;
            if (m.isLocal) {
              return { id };
            }
            const bids = ((await props.getBidsByModelId(id)) ?? [])
              .map((b: any) => ({
                ...b,
                ProviderData: providersMap[b.Provider.toLowerCase()],
                Model: m,
              }))
              .filter((b: any) => b.ProviderData);

            if (!bids.length) {
              return null;
            }

            return { id, bids };
          },
          // One failing model must not blank the whole marketplace list.
          () => null,
          BIDS_CONCURRENCY,
        )
      ).reduce((acc: any[], next: any) => {
        if (!next) {
          return acc;
        }
        const model = md.models.find((m: any) => m.Id == next.id);
        return [...acc, { ...model, bids: next.bids }];
      }, []);
      return merged;
    },
  });

  const availabilityQuery = useQuery({
    queryKey: queryKeys.providersAvailability,
    enabled: !!modelsDataQuery.data?.providers?.length,
    staleTime: 5 * 60_000,
    queryFn: () => props.getProvidersAvailability(modelsDataQuery.data.providers),
  });

  // Full (unfiltered) model list — local + every marketplace model, no bids.
  // Used for mapping sessions/chats by id, matching the original mount logic.
  const allModels: any[] | undefined = modelsDataQuery.data?.models;

  // chainData.models prefers the bid-enriched (and bid-filtered) list once it
  // is available, otherwise falls back to the raw list so the UI can render.
  const chainData = useMemo(() => {
    const md = modelsDataQuery.data;
    if (!md) {
      return null;
    }
    return { ...md, models: modelsWithBidsQuery.data ?? md.models };
  }, [modelsDataQuery.data, modelsWithBidsQuery.data]);

  const meta = modelsDataQuery.data?.meta ?? { budget: 0, supply: 0 };
  const balances = modelsDataQuery.data?.userBalances ?? { eth: 0, mor: 0 };
  const providersAvailability = availabilityQuery.data ?? [];
  const bidsLoading = modelsWithBidsQuery.isFetching;

  const sessions = useMemo(() => {
    const raw = sessionsQuery.data;
    if (!raw || !allModels) {
      return [];
    }
    return raw.reduce((res: any[], item: any) => {
      const sessionModel = allModels.find((x) => x.Id == item.ModelAgentId);
      if (sessionModel) {
        res.push({ ...item, ModelName: sessionModel.Name });
      }
      return res;
    }, []);
  }, [sessionsQuery.data, allModels]);

  // Initial-load overlay: only while there is no cached data yet. On revisits
  // every query resolves synchronously from cache, so this is false and the
  // spinner never appears.
  const isLoading =
    isActionLoading ||
    !modelsDataQuery.data ||
    sessionsQuery.isLoading ||
    !initialized;

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
  const isDisabled = (!activeSession && !isLocal) || isReadonly;
  const stakedFunds = activeSession
    ? (
        ((activeSession.EndsAt - activeSession.OpenedAt) *
          activeSession.PricePerSecond) /
        10 ** 18
      ).toFixed(2)
    : 0;

  // One-time selection of the default chat once the (possibly cached) model and
  // session data is available. Runs in a layout effect so that on a warm cache
  // the selection is committed before paint — no flash of the empty/intro state
  // and no transient spinner on tab revisits.
  useLayoutEffect(() => {
    if (initializedRef.current) {
      return;
    }
    const md = modelsDataQuery.data;
    const rawSessions = sessionsQuery.data;
    if (!md || !rawSessions) {
      return;
    }
    initializedRef.current = true;

    const models: any[] = md.models;

    const useLocalModelChat = () => {
      const localModel = models.find((m: any) => m.isLocal);
      if (localModel) {
        setSelectedModel(localModel);
        setChat({
          id: generateHashId(),
          createdAt: new Date(),
          modelId: localModel.Id,
          isLocal: true,
        });
      }
    };

    const mappedSessions = rawSessions.reduce((res: any[], item: any) => {
      const sessionModel = models.find((x) => x.Id == item.ModelAgentId);
      if (sessionModel) {
        res.push({ ...item, ModelName: sessionModel.Name });
      }
      return res;
    }, []);
    const openSessions = mappedSessions.filter((s) => !isClosed(s));

    if (!openSessions.length) {
      useLocalModelChat();
      setInitialized(true);
      return;
    }

    const latestSession = openSessions[0];
    const latestSessionModel = models.find(
      (m: any) => m.Id == latestSession.ModelAgentId,
    );

    if (!latestSessionModel) {
      useLocalModelChat();
      setInitialized(true);
      return;
    }

    // Commit the session selection synchronously (before paint), then fetch the
    // bid details in the background.
    setSelectedModel(latestSessionModel);
    setActiveSession(latestSession);
    setChat({
      id: generateHashId(),
      createdAt: new Date(),
      modelId: latestSessionModel.ModelAgentId,
    });
    setInitialized(true);

    props
      .getBidInfo(latestSession.BidID)
      .then((openBid) => {
        if (!openBid) {
          useLocalModelChat();
          return;
        }
        setSelectedBid(openBid);
      })
      .catch((e) => console.error('Failed to load open bid', e));
  }, [modelsDataQuery.data, sessionsQuery.data]);

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

  const scrollToBottom = (behavior: ScrollBehavior = 'instant') => {
    if (!cancelScroll) {
      chatBlockRef.current?.scroll({
        top: chatBlockRef.current.scrollHeight,
        behavior: behavior,
      });
    }
  };

  const calculateAcceptableDuration = (
    pricePerSecond: number,
    balance: number,
    stakingInfo: { budget: number; supply: number },
  ) => {
    const delta = 60; // 1 minute

    if (balance > requiredStake.max) {
      return 24 * 60 * 60; // 1 day in seconds
    }

    const targetDuration = Math.round(
      (balance * Number(stakingInfo.budget)) /
        (Number(stakingInfo.supply) * pricePerSecond),
    );

    if (targetDuration - delta < 5 * 60) {
      return 5 * 60;
    }

    return targetDuration - (targetDuration % 60) - delta;
  };

  const calculateAcceptableDurationForDirectPay = (stakingInfo: {
    budget: number;
    supply: number;
  }) => {
    // there is a bug in the contract, that incorrectly validates the duration when using direct pay, (as if user would stake)
    // so we calculate which duration is equivalent to amount of stake for minimum stake session duration (5 minutes)
    return Math.round((5 * 60 * stakingInfo.supply) / stakingInfo.budget) + 1;
  };

  // A session that was just opened on-chain does not always show up in the very
  // next indexer read. Poll briefly instead of assuming the first response
  // contains it — previously a miss meant `targetSessionData` was undefined and
  // the next line threw, which aborted the handler and left the UI wedged
  // (and the user re-staking into a second session they didn't need).
  const findSessionWithRetry = async (sessionId, attempts = 5, delayMs = 1200) => {
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

    setActiveSession({ ...targetSessionData, sessionId });

    const targetModel = chainData?.models?.find(
      (x) => x.Id == targetSessionData.ModelAgentId,
    );
    const targetBid = targetModel?.bids?.find(
      (x) => x.Id == targetSessionData.BidID,
    );
    if (targetBid) {
      setSelectedBid(targetBid);
    }
  };

  const onOpenSession = async (isReopen: boolean, isDirectPay: boolean) => {
    // Everything below used to run before the try block, so a missing
    // selectedModel threw an unhandled promise rejection *after*
    // setIsActionLoading(true) — leaving the spinner stuck on with no error
    // shown and no way to recover short of switching tabs.
    if (!selectedModel?.Id || !selectedModel?.bids?.length) {
      props.toasts.toast(
        'error',
        'No model selected, or its pricing has not loaded yet. Pick a model and try again.',
      );
      return;
    }

    setIsActionLoading(true);
    try {
      if (!isReopen) {
        setChat({
          id: generateHashId(),
          createdAt: new Date(),
          modelId: selectedModel.Id,
        });
      }

      const prices = selectedModel.bids.map((x) => Number(x.PricePerSecond));
      const maxPrice = Math.max(...prices);
      const duration = isDirectPay
        ? calculateAcceptableDurationForDirectPay(meta)
        : calculateAcceptableDuration(maxPrice, Number(balances.mor), meta);

      const openedSession = await props.onOpenSession({
        modelId: selectedModel.Id,
        duration,
        isDirectPay,
      });
      if (!openedSession) {
        return;
      }

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

      await setSessionData(openedSession);
      return openedSession;
    } catch (e: any) {
      // Never let a post-open failure escape as an unhandled rejection — that
      // used to take the whole view down with it.
      console.error('Failed to finalize opened session', e);
      props.toasts.toast(
        'error',
        'Session opened, but the app could not load it. Check the Sessions list before staking again.',
        { autoClose: 12000 },
      );
      return;
    } finally {
      setIsActionLoading(false);
    }
  };

  const loadChatHistory = async (chatId: string) => {
    try {
      const history = await props.client.getChatHistory(chatId);
      const messages: HistoryMessage[] = [];
      if (!history) {
        return;
      }

      const model = chainData.models.find((m) => m.Id == history.modelId);
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
        const isTtsPrompt =
          !isChatPrompt && typeof prompt.input === 'string';
        const isSttMessage = !isChatPrompt && !isTtsPrompt && !!m.isAudioContent;

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
          assistant.text = '[Audio response — replay is not available from history]';
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
    });
    const models = allModels ?? [];
    return (fresh || []).reduce((res, item) => {
      const sessionModel = models.find((x) => x.Id == item.ModelAgentId);
      if (sessionModel) {
        res.push({ ...item, ModelName: sessionModel.Name });
      }
      return res;
    }, []);
  };

  const closeSession = async (sessionId: string) => {
    setIsActionLoading(true);
    await props.closeSession(sessionId);
    await refreshSessions();
    setIsActionLoading(false);

    if (activeSession.Id == sessionId) {
      const localModel = chainData?.models?.find((m: any) => m.isLocal);
      if (localModel) {
        setSelectedModel(localModel);
        setChat({
          id: generateHashId(),
          createdAt: new Date(),
          modelId: localModel.Id,
          isLocal: true,
        });
      }
      setMessages([]);
    }
  };

  const selectChat = async (chatData: ChatData) => {
    const modelId = chatData.modelId;
    if (!modelId) {
      console.warn('Model ID is missed');
      return;
    }

    const selectedModel = chainData.isLocal
      ? chainData.models.find((m: any) => m.Id == modelId)
      : chainData.models.find((m: any) => m.Id == modelId && m.bids);
    setSelectedModel(selectedModel);
    setIsReadonly(false);

    setChat({ ...chatData });

    if (chatData.isLocal) {
      await loadChatHistory(chatData.id);
      return;
    }

    const openSessions = sessions.filter((s) => !isClosed(s));
    // search open session by model ID
    const openSession = openSessions.find((s) => s.ModelAgentId == modelId);
    setIsReadonly(!openSession);

    if (openSession) {
      setActiveSession(openSession);
      const activeBid = selectedModel.bids.find(
        (b) => b.Id == openSession.BidID,
      );
      setSelectedBid(activeBid);
    } else {
      setActiveSession(undefined);
      setSelectedBid(undefined);
    }

    await loadChatHistory(chatData.id);
    setTimeout(() => scrollToBottom('smooth'), 400);
  };

  const handleReopen = async (isDirectPay: boolean) => {
    await onOpenSession(true, isDirectPay);
    setIsReadonly(false);
  };

  const registerScrollEvent = (register) => {
    cancelScroll = false;
    const handler = (event: any) => {
      const isUp = event.wheelDelta ? event.wheelDelta > 0 : event.deltaY < 0;
      if (isUp) {
        cancelScroll = true;
      } else {
        if (!chatBlockRef?.current || !cancelScroll) {
          return;
        }
        // Return scrolling if scrolled to div end
        if (
          chatBlockRef.current.offsetHeight + chatBlockRef.current.scrollTop >=
          chatBlockRef.current.scrollHeight
        ) {
          cancelScroll = false;
        }
      }
    };

    if (register) {
      chatBlockRef?.current?.addEventListener('wheel', handler);
    } else {
      chatBlockRef?.current?.removeEventListener('wheel', handler);
    }
  };

  const call = async (message) => {
    let memoState = [
      ...messages,
      { id: makeId(16), text: promptInput, ...userMessage },
    ];
    setMessages(memoState);
    scrollToBottom();

    const headers = {
      Accept: 'application/json',
    };
    if (isLocal) {
      headers['model_id'] = selectedModel.Id;
    } else {
      headers['session_id'] = activeSession.Id;
    }
    headers['chat_id'] = chat?.id;

    const incommingMessage = { role: 'user', content: message };
    const payload = {
      stream: true,
      messages: [incommingMessage],
    };

    const authHeaders = await props.client.getAuthHeaders();
    // If image take only last message
    const response = await fetch(
      `${props.config.chain.localProxyRouterUrl}/v1/chat/completions`,
      {
        method: 'POST',
        headers: {
          ...headers,
          ...authHeaders,
        },
        body: JSON.stringify(payload),
      },
    ).catch((e) => {
      console.log('Failed to send request', e);
      return null;
    });

    if (!response) {
      return;
    }

    if (!response.ok) {
      console.log('Failed', await response.json());
      props.toasts.toast('error', 'Failed to send prompt');
      return;
    }

    if (!response.body) {
      console.error('Body is missed');
      return;
    }

    registerScrollEvent(true);

    const textDecoder = new TextDecoder();
    const reader = response.body.getReader();

    const icon = modelName.toUpperCase()[0];
    const iconProps = {
      icon,
      color: getColor(icon),
      user: modelName,
      role: 'assistant',
    };
    try {
      let chunksBuffer = '';
      while (true) {
        if (abort) {
          await reader.cancel();
          abort = false;
        }

        const { value, done } = await reader.read();
        if (done) {
          setIsSpinning(false);
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
            console.warn(part.error);
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
          setMessages(result);
          scrollToBottom();
        });
      }
    } catch (e) {
      props.toasts.toast('error', 'Something goes wrong. Try later.');
      console.error(e);
    }

    registerScrollEvent(false);
    return memoState;
  };

  const buildAudioHeaders = async () => {
    const headers: Record<string, string> = {};
    if (isLocal) {
      headers['model_id'] = selectedModel.Id;
    } else {
      headers['session_id'] = activeSession.Id;
    }
    if (chat?.id) {
      headers['chat_id'] = chat.id;
    }
    const authHeaders = await props.client.getAuthHeaders();
    return { ...headers, ...authHeaders };
  };

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
    const userText = { id: makeId(16), text, ...userMessage };
    let memoState = [...messages, userText];
    setMessages(memoState);
    scrollToBottom();

    try {
      const headers = await buildAudioHeaders();
      const response = await fetch(
        `${props.config.chain.localProxyRouterUrl}/v1/audio/speech`,
        {
          method: 'POST',
          headers: { ...headers, 'Content-Type': 'application/json' },
          body: JSON.stringify({
            input: text,
            voice: ttsVoice,
            response_format: 'mp3',
            speed: Number(ttsSpeed),
          }),
        },
      );

      if (!response || !response.ok) {
        props.toasts.toast('error', 'Failed to synthesize speech');
        return memoState;
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      memoState = [
        ...memoState,
        { id: makeId(16), text: url, isAudioContent: true, ...audioIconProps() },
      ];
      setMessages(memoState);
      scrollToBottom();
    } catch (e) {
      props.toasts.toast('error', 'Something goes wrong. Try later.');
      console.error(e);
    }
    return memoState;
  };

  // STT: audio in -> transcription text out
  const callTranscription = async (file: File) => {
    const userAudioUrl = URL.createObjectURL(file);
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
    scrollToBottom();

    if (messages.length === 0 && chat) {
      setChatsData([...chatData, { ...chat, title: file.name || 'Transcription' }]);
    }

    try {
      const headers = await buildAudioHeaders();
      const form = new FormData();
      form.append('file', file);
      form.append('response_format', 'json');

      // NB: do not set Content-Type; the browser adds the multipart boundary.
      const response = await fetch(
        `${props.config.chain.localProxyRouterUrl}/v1/audio/transcriptions`,
        {
          method: 'POST',
          headers,
          body: form,
        },
      );

      if (!response || !response.ok) {
        props.toasts.toast('error', 'Failed to transcribe audio');
        return memoState;
      }

      const contentType = response.headers.get('content-type') || '';
      let transcript = '';
      if (contentType.includes('application/json')) {
        const data = await response.json();
        transcript = data?.text ?? JSON.stringify(data);
      } else {
        transcript = await response.text();
      }

      memoState = [
        ...memoState,
        { id: makeId(16), text: transcript, ...audioIconProps() },
      ];
      setMessages(memoState);
      scrollToBottom();
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

  const handleSubmit = () => {
    if (abort) {
      abort = false;
    }

    if (isSpinning) {
      abort = true;
      setIsSpinning(false);
      return;
    }

    if (!promptInput) {
      return;
    }

    if (messages.length === 0 && chat) {
      const title = { ...chat, title: promptInput };
      setChatsData([...chatData, title]);
    }

    setIsSpinning(true);
    const request = modality === 'tts' ? callSpeech(promptInput) : call(promptInput);
    request.finally(() => setIsSpinning(false));
    setPromptInput('');
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

  const calculateStake = (pricePerSecond, durationInMin) => {
    const totalCost = pricePerSecond * durationInMin * 60;
    const stake = (totalCost * Number(meta.supply)) / Number(meta.budget);
    return stake;
  };

  const onCreateNewChat = ({ modelId, isLocal }) => {
    abort = true;
    setMessages([]);
    setActiveSession(undefined);
    setSelectedBid(undefined);
    setIsReadonly(false);
    setChat({ id: generateHashId(), createdAt: new Date(), modelId, isLocal });

    const selectedModel = isLocal
      ? chainData.models.find((m: any) => m.Id == modelId)
      : chainData.models.find((m: any) => m.Id == modelId && m.bids);

    // Marketplace selection needs the bid list, which may still be loading on a
    // cold first visit. Guard instead of dereferencing undefined bids.
    if (!isLocal && !selectedModel) {
      props.toasts.toast(
        'info',
        'Model options are still loading. Please try again in a moment.',
      );
      return;
    }

    setSelectedModel(selectedModel);

    if (isLocal) {
      setActiveSession(undefined);
      setSelectedBid(undefined);
      return;
    }

    const openSessions = sessions.filter((s) => !isClosed(s));
    const openModelSession = openSessions.find(
      (s) => s.ModelAgentId == modelId,
    );

    if (openModelSession) {
      const selectedBid = selectedModel.bids.find(
        (b) => b.Id == openModelSession.BidID && b.bids,
      );
      setSelectedBid(selectedBid);
      setActiveSession(openModelSession);
      return;
    }

    const prices = selectedModel.bids.map((x) => Number(x.PricePerSecond));
    const maxPrice = Math.max(...prices);

    setRequiredStake({
      min: calculateStake(maxPrice, 5),
      max: calculateStake(maxPrice, 24 * 60),
    });
  };

  const wrapChangeTitle = async (data: { id; title }) => {
    await props.client.updateChatHistoryTitle(data);
  };

  const renderChatBlock = () => {
    const isNewChat = !messages?.length;
    const isCreateSessionMode =
      isNewChat && !isLocal && !activeSession && !isLoading;

    // `meta` falls back to { budget: 0, supply: 0 } while the models query is
    // loading or has failed. Dividing by a zero budget produced NaN, and every
    // `x > NaN` comparison is false — which silently disabled *both* payment
    // buttons with no explanation. Treat unknown pricing as "not ready yet" and
    // say so, rather than rendering a dead screen.
    const isPricingReady =
      Number(meta.budget) > 0 &&
      Number(meta.supply) > 0 &&
      Number.isFinite(Number(requiredStake.min));

    // for stake mode
    const isEnoughFunds =
      isPricingReady && Number(balances.mor) > Number(requiredStake.min);

    const requiredStakeForDirectPay = isPricingReady
      ? (5 * 3600 * Number(meta.supply)) / Number(meta.budget)
      : Number.POSITIVE_INFINITY;
    const isEnoughFundsForDirectPay =
      isPricingReady && Number(balances.mor) > requiredStakeForDirectPay;

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

              {openSessionsForModel.length > 0 && (
                <ChatIntroInnerText style={{ color: '#20dc8e' }}>
                  You already have {openSessionsForModel.length} open session
                  {openSessionsForModel.length > 1 ? 's' : ''} for this model.
                  Open it from the Sessions list in the sidebar instead of
                  staking again — staking again locks additional MOR.
                </ChatIntroInnerText>
              )}

              {!isPricingReady && (
                <ChatIntroInnerText style={{ color: '#e8a33d' }}>
                  Pricing data hasn't loaded yet, so staking is temporarily
                  unavailable. If this persists, check that the proxy-router is
                  running in Settings.
                </ChatIntroInnerText>
              )}
              <ChatIntroInnerText>
                Stake MOR to get a free compute. Session will last from 5 mins
                up to 24 hours depending on the amount you stake (min:{' '}
                {formatValue(requiredStake.min, 18)} MOR, max:{' '}
                {formatValue(requiredStake.max, 18)} MOR). You can claim your
                stake in 24h.
              </ChatIntroInnerText>
              <div style={{ display: 'flex', justifyContent: 'center' }}>
                <ChatIntroButton
                  onClick={() => onOpenSession(false, false)}
                  disabled={!isEnoughFunds}
                >
                  Stake MOR
                </ChatIntroButton>
              </div>
              <ChatIntroInnerText>
                Pay with your MOR tokens directly. The duration of the session
                is limited only with your MOR balance.
              </ChatIntroInnerText>
              <div style={{ display: 'flex', justifyContent: 'center' }}>
                <ChatIntroButton
                  onClick={() => onOpenSession(false, true)}
                  disabled={!isEnoughFundsForDirectPay}
                >
                  Direct Pay
                </ChatIntroButton>
              </div>
            </ChatIntroInner>
          </ChatIntroContainer>
        ) : (
          <ChatHistoryContainer>
            {messages?.map((x, index) => (
              <Message key={index} message={x} onOpenImage={setImagePreview} />
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
  if (modelsDataQuery.isError && !modelsDataQuery.data) {
    return (
      <View data-testid="chat-container">
        <QueryError
          error={modelsDataQuery.error}
          what="models"
          onRetry={() => modelsDataQuery.refetch()}
        />
      </View>
    );
  }

  return (
    <>
      {isLoading && (
        <LoadingCover>
          <Spinner
            style={{ width: '5rem', height: '5rem' }}
            animation="border"
            variant="success"
          />
        </LoadingCover>
      )}

      {/* Non-fatal: models loaded from cache but the latest refresh failed. */}
      {modelsDataQuery.isError && !!modelsDataQuery.data && (
        <QueryError
          error={modelsDataQuery.error}
          what="the latest model data"
          onRetry={() => modelsDataQuery.refetch()}
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
            setIsActionLoading(true);
            await refreshSessions();
            setIsActionLoading(false);
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
                onClick={() => setOpenChangeModal(true)}
              >
                <IconMessagePlus></IconMessagePlus> New chat
              </BtnAccent>
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
                  onClick={() => (recording ? stopRecording() : startRecording())}
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
                <CustomTextArrea
                  disabled={isDisabled}
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleSubmit();
                    }
                  }}
                  value={promptInput}
                  onChange={(ev) => setPromptInput(ev.target.value)}
                  placeholder={
                    isReadonly
                      ? 'Session is closed. Chat in ReadOnly Mode'
                      : modality === 'tts'
                        ? 'Enter text to synthesize...'
                        : 'Ask me anything...'
                  }
                  minRows={1}
                  maxRows={6}
                />
                <SendBtnWrapper>
                  {isReadonly ? (
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
                    <Btn disabled={isDisabled} onClick={handleSubmit}>
                      {isSpinning ? (
                        <Spinner animation="border" />
                      ) : (
                        <IconArrowUp size={'26px'}></IconArrowUp>
                      )}
                    </Btn>
                  )}
                </SendBtnWrapper>
              </>
            )}
          </Control>
        </Container>
      </View>
      <ModelSelectionModal
        models={(chainData as any)?.models}
        isActive={openChangeModal}
        symbol={props.symbol}
        bidsLoading={bidsLoading}
        providersAvailability={providersAvailability}
        onChangeModel={(eventData) => {
          onCreateNewChat(eventData);
        }}
        handleClose={() => setOpenChangeModal(false)}
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
      <ThinkingMessageBody text={message.text} />
    </MessageBody>
  );
};

const Message = ({ message, onOpenImage }) => {
  return (
    <div style={{ display: 'flex', margin: '12px 0 28px 0' }}>
      <Avatar color={message.color}>{message.icon}</Avatar>
      <div>
        <AvatarHeader>{message.user}</AvatarHeader>
        {renderMessage(message, onOpenImage)}
      </div>
    </div>
  );
};

// withChatState injects props that are loosely typed in its HOC signature;
// cast to suppress the HOC-vs-component prop mismatch.
export default withChatState(Chat as React.ComponentType<any>);
