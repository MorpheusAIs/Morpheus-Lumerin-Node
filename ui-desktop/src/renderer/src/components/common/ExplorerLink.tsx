import React from 'react';
import styled, { css } from 'styled-components';
import { IconExternalLink } from '@tabler/icons-react';

/**
 * Every "view this on the block explorer" affordance in the app used to be
 * built from scratch at the call site: a whole invisible clickable row here, a
 * bare underlined span there, a full-width accent button somewhere else, and a
 * raw `<a target="_blank">` in the agents modal. Same destination, four
 * different-looking controls, and two of them gave no hint they were clickable
 * at all. This is the one component they all go through now.
 */

export type ExplorerLinkKind = 'transaction' | 'account' | 'contract';
export type ExplorerLinkVariant = 'inline' | 'button' | 'icon' | 'row';

/**
 * The host is worth showing: it tells the user they are about to leave the app
 * and exactly where to. Falls back to neutral wording rather than throwing on a
 * malformed or missing URL, since these come from remote chain config.
 */
export function explorerHost(url?: string | null): string {
  if (!url) return 'the block explorer';
  try {
    return new URL(url).hostname.replace(/^www\./, '');
  } catch {
    return 'the block explorer';
  }
}

const focusRing = css`
  &:focus-visible {
    outline: 2px solid ${(p) => p.theme.colors.morMain};
    outline-offset: 2px;
  }
`;

const disabled = css`
  &:disabled {
    color: rgba(255, 255, 255, 0.35);
    cursor: not-allowed;
    text-decoration: none;
  }
`;

const Inline = styled.button`
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  margin: 0;
  padding: 0;
  border: none;
  background: none;
  color: ${(p) => p.theme.colors.morMain};
  font: inherit;
  font-weight: 500;
  text-align: left;
  overflow-wrap: anywhere;
  cursor: pointer;
  border-radius: 4px;

  svg {
    flex: 0 0 auto;
    opacity: 0.75;
  }

  &:hover:not(:disabled) {
    text-decoration: underline;

    svg {
      opacity: 1;
    }
  }

  ${focusRing}
  ${disabled}
`;

const Accent = styled.button`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 1.2rem 1.6rem;
  border: 1px solid rgba(32, 220, 142, 0.4);
  border-radius: 0.6rem;
  background: rgba(32, 220, 142, 0.12);
  color: ${(p) => p.theme.colors.morMain};
  font-size: 1.5rem;
  font-weight: 600;
  letter-spacing: 0.2px;
  cursor: pointer;
  transition:
    background 0.15s ease,
    border-color 0.15s ease;

  &:hover:not(:disabled) {
    background: rgba(32, 220, 142, 0.2);
    border-color: rgba(32, 220, 142, 0.6);
  }

  &:disabled {
    background: rgba(255, 255, 255, 0.04);
    border-color: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 0.35);
    cursor: not-allowed;
  }

  ${focusRing}
`;

/* Sized to match the other circular icon buttons that sit beside it (copy,
   etc.) so it reads as part of the same control group. */
const IconOnly = styled.button`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  transition:
    background 0.15s ease,
    color 0.15s ease;

  &:hover:not(:disabled) {
    background: rgba(32, 220, 142, 0.14);
    color: ${(p) => p.theme.colors.morMain};
  }

  &:disabled {
    color: rgba(255, 255, 255, 0.25);
    cursor: not-allowed;
  }

  ${focusRing}
`;

/* A whole list row that happens to be a link. It carries its own content, so
   the shared part is the hover state, the cursor and the trailing icon that
   marks it as leaving the app. */
const Row = styled.button`
  display: flex;
  align-items: center;
  gap: 1rem;
  width: 100%;
  margin: 0;
  padding: 0;
  border: none;
  background: none;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s ease;

  &:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.03);
  }

  &:disabled {
    cursor: default;
  }

  ${focusRing}
`;

const RowAffordance = styled.span`
  display: inline-flex;
  align-items: center;
  flex: 0 0 auto;
  margin-left: auto;
  color: rgba(255, 255, 255, 0.35);
  transition: color 0.15s ease;

  ${Row}:hover & {
    color: ${(p) => p.theme.colors.morMain};
  }
`;

const VARIANTS = {
  inline: Inline,
  button: Accent,
  icon: IconOnly,
  row: Row,
} as const;

const ICON_SIZE = {
  inline: 15,
  button: 18,
  icon: 16,
  row: 16,
} as const;

export interface ExplorerLinkProps {
  /** Fully-resolved explorer URL. Missing or malformed renders disabled. */
  url?: string | null;
  /** What sits at the other end; drives the default label and tooltip. */
  kind?: ExplorerLinkKind;
  variant?: ExplorerLinkVariant;
  /** Overrides the generated label. The `icon` variant ignores it. */
  children?: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
  title?: string;
  'aria-label'?: string;
  'data-testid'?: string;
}

export const ExplorerLink: React.FC<ExplorerLinkProps> = ({
  url,
  kind = 'transaction',
  variant = 'inline',
  children,
  className,
  style,
  title,
  'aria-label': ariaLabel,
  'data-testid': testId,
}) => {
  const host = explorerHost(url);
  const label = `View ${kind} on ${host}`;
  const Control = VARIANTS[variant];
  const iconSize = ICON_SIZE[variant];

  const open = () => {
    if (!url) return;
    window.openLink(url);
  };

  // When the caller supplies its own visible text — a hash, an address — that
  // text is the better accessible name. Overriding it with a generic label
  // would hide from screen readers the one detail that identifies the link.
  // The host is still announced via the tooltip.
  const hasOwnText =
    (variant === 'inline' || variant === 'button') && children != null;
  const resolvedAriaLabel = ariaLabel ?? (hasOwnText ? undefined : label);

  return (
    <Control
      type="button"
      className={className}
      style={style}
      disabled={!url}
      onClick={open}
      title={title ?? label}
      aria-label={resolvedAriaLabel}
      data-testid={testId}
    >
      {variant === 'icon' ? (
        <IconExternalLink size={iconSize} aria-hidden="true" />
      ) : variant === 'row' ? (
        <>
          {children}
          <RowAffordance>
            <IconExternalLink size={iconSize} aria-hidden="true" />
          </RowAffordance>
        </>
      ) : (
        <>
          {children ?? label}
          <IconExternalLink size={iconSize} aria-hidden="true" />
        </>
      )}
    </Control>
  );
};

export default ExplorerLink;
