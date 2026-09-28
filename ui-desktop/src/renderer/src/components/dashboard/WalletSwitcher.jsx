import React, { useContext, useEffect, useRef, useState } from 'react';
import styled from 'styled-components';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  IconChevronDown,
  IconCheck,
  IconPlus,
  IconDownload,
  IconTrash,
  IconPencil,
  IconAlertTriangle,
} from '@tabler/icons-react';

import { ToastsContext } from '../toasts';
import { withClient } from '../../store/hocs/clientContext';
import { queryKeys } from '../../store/queries';
import { abbreviateAddress } from '../../utils';

const Wrap = styled.div`
  position: relative;
  display: inline-flex;
`;

const Trigger = styled.button`
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.6rem 1rem;
  border-radius: 999px;
  background: var(--surface-raised);
  border: 1px solid var(--border-subtle);
  color: var(--text-primary);
  font-size: 1.3rem;
  cursor: pointer;
  font-variant-numeric: tabular-nums;

  &:hover {
    border-color: var(--border-strong);
  }
`;

const Dot = styled.div`
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: 0 0 0 3px rgba(32, 220, 142, 0.18);
  flex-shrink: 0;
`;

const Menu = styled.div`
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  z-index: 40;
  min-width: 320px;
  max-height: 420px;
  overflow-y: auto;
  padding: 0.6rem;
  border-radius: 14px;
  background: var(--surface-raised);
  border: 1px solid var(--border-strong);
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.5);
`;

const Row = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 0.9rem 1rem;
  border-radius: 10px;
  cursor: ${(p) => (p.$disabled ? 'not-allowed' : 'pointer')};
  opacity: ${(p) => (p.$disabled ? 0.5 : 1)};
  color: var(--text-primary);
  background: ${(p) => (p.$active ? 'var(--surface-hover)' : 'transparent')};

  &:hover {
    background: ${(p) => (p.$disabled ? 'transparent' : 'var(--surface-hover)')};
  }
`;

const RowText = styled.div`
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
`;

const Label = styled.span`
  font-size: 1.3rem;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const Sub = styled.span`
  font-size: 1.1rem;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
`;

const IconBtn = styled.button`
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  display: inline-flex;

  &:hover {
    color: var(--text-primary);
    background: var(--surface-hover);
  }
`;

const Divider = styled.div`
  height: 1px;
  margin: 0.5rem 0.4rem;
  background: var(--border-subtle);
`;

const Action = styled.button`
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 0.85rem 1rem;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: var(--accent);
  font-size: 1.25rem;
  cursor: pointer;
  text-align: left;

  &:hover:not(:disabled) {
    background: var(--surface-hover);
  }

  &:disabled {
    color: var(--text-muted);
    cursor: not-allowed;
  }
`;

/** The same row, for the choice that backs out rather than the one that acts. */
const SecondaryAction = styled(Action)`
  color: var(--text-primary);
`;

/** Sits inside a Warn, where the two choices need to read as one control pair. */
const InlineActions = styled.div`
  display: flex;
  gap: 8px;
  margin-top: 8px;

  ${Action} {
    padding: 0.5rem 0.9rem;
    width: auto;
  }
`;

const Warn = styled.div`
  display: flex;
  gap: 8px;
  padding: 0.9rem 1rem;
  margin: 0.4rem;
  border-radius: 10px;
  background: rgba(246, 191, 98, 0.1);
  border: 1px solid rgba(246, 191, 98, 0.3);
  color: var(--text-primary);
  font-size: 1.15rem;
  line-height: 1.5;
`;

const Input = styled.input`
  width: 100%;
  padding: 0.8rem 1rem;
  margin: 0.4rem 0;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  background: var(--surface-base);
  color: var(--text-primary);
  font-size: 1.2rem;
`;

/** The rename field replaces a label in place, so it carries no outer margin. */
const InlineInput = styled(Input)`
  margin: 0;
`;

/** The private key form and the row of buttons under it. */
const ImportPanel = styled.div`
  padding: 0.4rem;
`;

const ImportActions = styled.div`
  display: flex;
  gap: 8px;
  margin-top: 6px;
`;

/** Keeps a row without a check mark aligned with the rows that have one. */
const CheckSpacer = styled.span`
  width: 16px;
  flex-shrink: 0;
`;

function WalletSwitcher({ client, activeAddress, openSessionCount = 0 }) {
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(null);
  const [importing, setImporting] = useState(false);
  const [importKey, setImportKey] = useState('');
  const [pendingSwitch, setPendingSwitch] = useState(null);
  // `{ id, draft }` while a wallet is being renamed in place. Electron's
  // Chromium does not implement `window.prompt()` — it throws — so the rename
  // has to happen inside the menu.
  const [renaming, setRenaming] = useState(null);
  const renameInFlight = useRef(false);
  const ref = useRef(null);
  const context = useContext(ToastsContext);
  const queryClient = useQueryClient();

  const walletsQuery = useQuery({
    queryKey: queryKeys.wallets,
    queryFn: () => client.getWallets(),
    staleTime: 10_000,
  });

  // Close on outside click / Escape.
  useEffect(() => {
    if (!open) return undefined;
    const onDown = (e) => {
      if (ref.current && !ref.current.contains(e.target)) {
        setOpen(false);
        setImporting(false);
        setPendingSwitch(null);
        setRenaming(null);
      }
    };
    // Escape backs out of a rename first, so a mistyped name does not also
    // close the whole menu.
    const onKey = (e) => {
      if (e.key !== 'Escape') return;
      if (renaming) {
        setRenaming(null);
        return;
      }
      setOpen(false);
    };
    document.addEventListener('mousedown', onDown);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDown);
      document.removeEventListener('keydown', onKey);
    };
  }, [open, renaming]);

  const data = walletsQuery.data;
  const wallets = data?.wallets ?? [];
  const activeId = data?.activeId;
  const active = wallets.find((w) => w.id === activeId);

  /**
   * Switching swaps the key the proxy-router is using and restarts its session
   * machinery, so every address-scoped cache is stale afterwards. Clearing the
   * whole cache is blunt but correct — leaving one stale entry behind means
   * showing another wallet's balance or sessions, which is exactly the class of
   * bug that made people think their funds had moved.
   */
  const doSwitch = async (walletId) => {
    setBusy(walletId);
    setPendingSwitch(null);
    try {
      const res = await client.switchWallet({ walletId });
      await queryClient.invalidateQueries();
      context.toast('success', `Switched to ${abbreviateAddress(res.address, 6)}`);
      setOpen(false);
    } catch (e) {
      context.toast('error', e?.message || 'Failed to switch wallet', {
        autoClose: 12000,
      });
    } finally {
      setBusy(null);
    }
  };

  const onPick = (w) => {
    if (w.id === activeId || busy) return;
    // An open session belongs to the address that opened it. Switching does not
    // close or lose it, but any in-flight chat stops working, so make that a
    // deliberate choice rather than a surprise.
    if (openSessionCount > 0) {
      setPendingSwitch(w);
      return;
    }
    doSwitch(w.id);
  };

  const onAddHd = async () => {
    setBusy('add');
    try {
      const w = await client.addHdWallet({});
      await walletsQuery.refetch();
      context.toast('success', `Added ${w.label}`);
    } catch (e) {
      context.toast('error', e?.message || 'Failed to add account', {
        autoClose: 12000,
      });
    } finally {
      setBusy(null);
    }
  };

  const onImport = async () => {
    setBusy('import');
    try {
      const w = await client.importWallet({ privateKey: importKey });
      setImportKey('');
      setImporting(false);
      await walletsQuery.refetch();
      context.toast('success', `Imported ${w.label}`);
    } catch (e) {
      context.toast('error', e?.message || 'Failed to import wallet', {
        autoClose: 12000,
      });
    } finally {
      setBusy(null);
    }
  };

  const onRemove = async (e, w) => {
    e.stopPropagation();
    if (!window.confirm(`Remove "${w.label}" from this app?`)) return;
    try {
      await client.removeWallet({ walletId: w.id });
      await walletsQuery.refetch();
      context.toast('info', `Removed ${w.label}`);
    } catch (err) {
      context.toast('error', err?.message || 'Failed to remove wallet');
    }
  };

  const startRename = (e, w) => {
    e.stopPropagation();
    setPendingSwitch(null);
    renameInFlight.current = false;
    setRenaming({ id: w.id, draft: w.label });
  };

  const commitRename = async () => {
    // Enter and blur can both land on the same edit. The ref makes the commit
    // idempotent without waiting for a re-render to clear `renaming`.
    if (!renaming || renameInFlight.current) return;
    renameInFlight.current = true;
    const label = renaming.draft.trim();
    const original = wallets.find((w) => w.id === renaming.id);
    // Nothing to save, and an empty name would be rejected by the main process
    // anyway — just back out rather than showing an error for a no-op.
    if (!label || label === original?.label) {
      setRenaming(null);
      return;
    }
    setRenaming(null);
    try {
      await client.renameWallet({ walletId: renaming.id, label });
      await walletsQuery.refetch();
    } catch (err) {
      context.toast('error', err?.message || 'Failed to rename wallet');
    }
  };

  const shown = active?.label || abbreviateAddress(activeAddress, 6) || 'Wallet';

  return (
    <Wrap ref={ref}>
      <Trigger onClick={() => setOpen((o) => !o)} title="Switch wallet">
        <Dot />
        {shown}
        <IconChevronDown size={14} />
      </Trigger>

      {open && (
        <Menu>
          {walletsQuery.isError && (
            <Warn>
              <IconAlertTriangle size={16} />
              <span>
                {walletsQuery.error?.message || 'Could not load your wallets.'}
              </span>
            </Warn>
          )}

          {pendingSwitch && (
            <Warn>
              <IconAlertTriangle size={16} />
              <div>
                You have {openSessionCount} open session
                {openSessionCount > 1 ? 's' : ''} on this wallet. Switching
                keeps {openSessionCount > 1 ? 'them' : 'it'} open on-chain and
                your stake is unaffected, but any chat in progress will stop.
                <InlineActions>
                  <Action onClick={() => doSwitch(pendingSwitch.id)}>
                    Switch anyway
                  </Action>
                  <SecondaryAction onClick={() => setPendingSwitch(null)}>
                    Cancel
                  </SecondaryAction>
                </InlineActions>
              </div>
            </Warn>
          )}

          {wallets.map((w) => (
            <Row
              key={w.id}
              $active={w.id === activeId}
              $disabled={!!busy}
              onClick={() => onPick(w)}
            >
              {w.id === activeId ? (
                <IconCheck size={16} color="var(--accent)" />
              ) : (
                <CheckSpacer />
              )}
              <RowText>
                {renaming?.id === w.id ? (
                  <InlineInput
                    autoFocus
                    aria-label={`Rename ${w.label}`}
                    maxLength={64}
                    value={renaming.draft}
                    onClick={(e) => e.stopPropagation()}
                    onChange={(e) =>
                      setRenaming((r) => r && { ...r, draft: e.target.value })
                    }
                    onKeyDown={(e) => {
                      e.stopPropagation();
                      if (e.key === 'Enter') commitRename();
                      if (e.key === 'Escape') setRenaming(null);
                    }}
                    onBlur={commitRename}
                  />
                ) : (
                  <Label>{w.label}</Label>
                )}
                <Sub>
                  {abbreviateAddress(w.address, 6)}
                  {w.kind === 'hd' ? ` · account ${w.derivationPath}` : ' · imported'}
                </Sub>
              </RowText>
              {busy === w.id ? (
                <Sub>switching…</Sub>
              ) : renaming?.id === w.id ? (
                <Sub>enter to save</Sub>
              ) : (
                <>
                  <IconBtn title="Rename" onClick={(e) => startRename(e, w)}>
                    <IconPencil size={14} />
                  </IconBtn>
                  {w.id !== activeId && wallets.length > 1 && (
                    <IconBtn title="Remove" onClick={(e) => onRemove(e, w)}>
                      <IconTrash size={14} />
                    </IconBtn>
                  )}
                </>
              )}
            </Row>
          ))}

          <Divider />

          {importing ? (
            <ImportPanel>
              <Input
                autoFocus
                type="password"
                spellCheck={false}
                placeholder="Private key (64 hex characters)"
                value={importKey}
                onChange={(e) => setImportKey(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && onImport()}
              />
              <Sub>
                Stored in your operating system's keychain, not in app settings.
              </Sub>
              <ImportActions>
                <Action onClick={onImport} disabled={busy === 'import'}>
                  {busy === 'import' ? 'Importing…' : 'Import'}
                </Action>
                <SecondaryAction
                  onClick={() => {
                    setImporting(false);
                    setImportKey('');
                  }}
                >
                  Cancel
                </SecondaryAction>
              </ImportActions>
            </ImportPanel>
          ) : (
            <>
              <Action
                onClick={onAddHd}
                disabled={!data?.canAddHd || !!busy}
                title={
                  data?.canAddHd
                    ? `Adds account ${data?.nextHdIndex} from your recovery phrase`
                    : 'This wallet was imported from a private key, so it has no recovery phrase to derive accounts from'
                }
              >
                <IconPlus size={15} />
                {busy === 'add'
                  ? 'Adding…'
                  : `Add account${data?.canAddHd ? ` ${data.nextHdIndex}` : ''}`}
              </Action>
              <Action onClick={() => setImporting(true)} disabled={!!busy}>
                <IconDownload size={15} />
                Import a private key
              </Action>
            </>
          )}
        </Menu>
      )}
    </Wrap>
  );
}

export default withClient(WalletSwitcher);
