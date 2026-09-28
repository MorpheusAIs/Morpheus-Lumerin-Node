import { useCallback, useEffect, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router';
import {
  IconArrowRight,
  IconCheck,
  IconCompass,
  IconMessage,
  IconMinus,
  IconSettings,
  IconSparkles,
  IconWallet,
  IconX,
} from '@tabler/icons-react';
import './QuickStartGuide.css';

export const QUICK_START_STORAGE_KEY = 'morpheus.quick-start.v1';
export const QUICK_START_OPEN_EVENT = 'morpheus:open-quick-start';

/** A UI-only invitation. Never opens a compute session or changes preferences. */
export function openQuickStartGuide() {
  window.dispatchEvent(new Event(QUICK_START_OPEN_EVENT));
}

type GuideMode = 'nudge' | 'guide' | 'minimized' | 'closed';

const steps = [
  {
    target: 'wallet',
    route: '/wallet',
    page: 'Wallet',
    title: 'Your wallet, your address',
    description:
      'Your public wallet address is safe to share for receiving funds. Use the copy button beside your address to copy the full address.',
    note: 'Keep your private key and recovery phrase private. The guide never asks for them.',
    icon: IconWallet,
  },
  {
    target: 'chat',
    route: '/chat',
    page: 'Chat',
    title: 'Choose your model & session',
    description:
      'Start a new chat, choose a model and session length, then review your payment option. Use the Workspace candidates filter for multi-step tasks; provider support is checked when you connect.',
    note: 'No subscriptions. With Stake MOR, unused stake returns when the session closes.',
    icon: IconMessage,
  },
  {
    target: 'workspace',
    route: '/workspace',
    page: 'Workspace',
    title: 'Give your task a home',
    description:
      'Connect a folder to create a project, then use an active Chat session to work on files. Workspace’s file tools stay inside the folder you choose.',
    note: 'Your project history stays when a session ends. Open another session in Chat, then continue the same task in Workspace.',
    icon: IconSparkles,
  },
  {
    target: 'settings',
    route: '/settings',
    page: 'Settings',
    title: 'Know where to check',
    description:
      'If the app cannot connect, check the local proxy-router in Settings. A connection problem is not a balance of zero.',
    note: 'You can replay this tour from the sidebar whenever you need a refresher.',
    icon: IconSettings,
  },
] as const;

function initialMode(): GuideMode {
  try {
    const preference = window.localStorage.getItem(QUICK_START_STORAGE_KEY);
    if (preference === 'dismissed' || preference === 'completed')
      return 'closed';
  } catch {
    // Onboarding is optional, including when storage is unavailable.
  }
  return 'nudge';
}

function remember(value: 'dismissed' | 'completed') {
  try {
    window.localStorage.setItem(QUICK_START_STORAGE_KEY, value);
  } catch {
    // Do not turn an unavailable preference store into an application error.
  }
}

/**
 * Mount once inside the authenticated router. A sidebar button may call
 * openQuickStartGuide() and use data-guide-launcher for focus restoration.
 * Optional data-guide="wallet|chat|workspace|settings" nav targets are outlined
 * during the tour; missing targets do not prevent using the guide.
 */
export default function QuickStartGuide() {
  const [mode, setMode] = useState<GuideMode>(initialMode);
  const [stepIndex, setStepIndex] = useState(0);
  const panelRef = useRef<HTMLElement>(null);
  const titleRef = useRef<HTMLHeadingElement>(null);
  const resumeRef = useRef<HTMLButtonElement>(null);
  const returnFocusRef = useRef<HTMLElement | null>(null);
  const location = useLocation();
  const navigate = useNavigate();
  const step = steps[stepIndex];
  const StepIcon = step.icon;
  const onStepPage = location.pathname === step.route;

  const start = useCallback(() => {
    if (
      document.activeElement instanceof HTMLElement &&
      !panelRef.current?.contains(document.activeElement)
    ) {
      returnFocusRef.current = document.activeElement;
    }
    remember('dismissed');
    setStepIndex(0);
    setMode('guide');
    titleRef.current?.focus({ preventScroll: true });
  }, []);

  const dismiss = useCallback((completed = false) => {
    remember(completed ? 'completed' : 'dismissed');
    setMode('closed');
    const previous = returnFocusRef.current;
    const launcher = document.querySelector<HTMLElement>(
      '[data-guide-launcher]',
    );
    if (previous?.isConnected && previous !== document.body) {
      previous.focus({ preventScroll: true });
    } else {
      launcher?.focus({ preventScroll: true });
    }
  }, []);

  useEffect(() => {
    window.addEventListener(QUICK_START_OPEN_EVENT, start);
    return () => window.removeEventListener(QUICK_START_OPEN_EVENT, start);
  }, [start]);

  useEffect(() => {
    if (mode === 'guide') titleRef.current?.focus({ preventScroll: true });
    if (mode === 'minimized') resumeRef.current?.focus({ preventScroll: true });
  }, [mode, stepIndex]);

  useEffect(() => {
    if (mode !== 'guide') return;
    const target = document.querySelector<HTMLElement>(
      `[data-guide="${step.target}"]`,
    );
    if (!target) return;
    const previous = target.getAttribute('data-guide-active');
    target.setAttribute('data-guide-active', 'true');
    return () => {
      if (previous === null) target.removeAttribute('data-guide-active');
      else target.setAttribute('data-guide-active', previous);
    };
  }, [mode, step.target]);

  if (mode === 'closed') return null;

  if (mode === 'minimized') {
    return (
      <aside
        className="quick-start quick-start--compact"
        aria-label="Quick start guide"
      >
        <button
          className="quick-start__resume"
          onClick={() => setMode('guide')}
          ref={resumeRef}
          type="button"
        >
          <IconCompass aria-hidden="true" size={18} />
          Resume tour
          <span className="quick-start__count">
            {stepIndex + 1} / {steps.length}
          </span>
        </button>
        <button
          aria-label="Close quick start guide"
          className="quick-start__icon-button"
          onClick={() => dismiss()}
          type="button"
        >
          <IconX aria-hidden="true" size={18} />
        </button>
      </aside>
    );
  }

  if (mode === 'nudge') {
    return (
      <aside
        className="quick-start quick-start--nudge"
        aria-label="Getting started"
      >
        <IconCompass
          aria-hidden="true"
          className="quick-start__nudge-icon"
          size={24}
        />
        <div className="quick-start__nudge-content">
          <h2>New to Morpheus?</h2>
          <p>Find your way from wallet to Workspace.</p>
          <button
            className="quick-start__text-button"
            onClick={start}
            type="button"
          >
            Take a quick tour <IconArrowRight aria-hidden="true" size={16} />
          </button>
        </div>
        <button
          aria-label="Dismiss quick tour invitation"
          className="quick-start__icon-button"
          onClick={() => dismiss()}
          type="button"
        >
          <IconX aria-hidden="true" size={18} />
        </button>
      </aside>
    );
  }

  return (
    <aside
      aria-label="Quick start guide"
      className="quick-start quick-start--guide"
      onKeyDown={(event) => {
        // Non-modal: leave shortcuts elsewhere in the app alone.
        if (event.key === 'Escape' && !event.defaultPrevented) {
          event.preventDefault();
          event.stopPropagation();
          dismiss();
        }
      }}
      ref={panelRef}
    >
      <div className="quick-start__heading-row">
        <StepIcon
          aria-hidden="true"
          className="quick-start__step-icon"
          size={22}
        />
        <h2 ref={titleRef} tabIndex={-1}>
          {step.title}
        </h2>
        <button
          aria-label="Minimize quick start guide"
          className="quick-start__icon-button"
          onClick={() => setMode('minimized')}
          type="button"
        >
          <IconMinus aria-hidden="true" size={18} />
        </button>
      </div>
      <div className="quick-start__copy" key={step.target}>
        <p>{step.description}</p>
        <p className="quick-start__note">{step.note}</p>
      </div>
      <div className="quick-start__page-action">
        {onStepPage ? (
          <span className="quick-start__current-page">
            <IconCheck aria-hidden="true" size={16} /> You’re in {step.page}
          </span>
        ) : (
          <button
            className="quick-start__text-button"
            onClick={() => {
              navigate(step.route);
              titleRef.current?.focus({ preventScroll: true });
            }}
            type="button"
          >
            Open {step.page} <IconArrowRight aria-hidden="true" size={16} />
          </button>
        )}
      </div>
      <div
        className="quick-start__progress"
        aria-label={`Step ${stepIndex + 1} of ${steps.length}`}
      >
        {steps.map((item, index) => (
          <span
            className={index <= stepIndex ? 'is-complete' : undefined}
            key={item.target}
          />
        ))}
      </div>
      <div className="quick-start__footer">
        <button
          className="quick-start__skip"
          onClick={() => dismiss()}
          type="button"
        >
          Skip tour
        </button>
        <span className="quick-start__count" aria-hidden="true">
          {stepIndex + 1} / {steps.length}
        </span>
        <div className="quick-start__navigation">
          {stepIndex > 0 && (
            <button
              className="quick-start__button quick-start__button--secondary"
              onClick={() => setStepIndex((index) => index - 1)}
              type="button"
            >
              Back
            </button>
          )}
          <button
            className="quick-start__button quick-start__button--primary"
            onClick={() => {
              if (stepIndex === steps.length - 1) dismiss(true);
              else setStepIndex((index) => index + 1);
            }}
            type="button"
          >
            {stepIndex === steps.length - 1 ? 'Done' : 'Next'}
            {stepIndex === steps.length - 1 ? (
              <IconCheck aria-hidden="true" size={16} />
            ) : (
              <IconArrowRight aria-hidden="true" size={16} />
            )}
          </button>
        </div>
      </div>
    </aside>
  );
}
