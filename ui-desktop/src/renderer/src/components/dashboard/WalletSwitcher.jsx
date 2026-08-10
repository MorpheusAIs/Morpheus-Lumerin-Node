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
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.92);
  font-size: 1.3rem;
  cursor: pointer;
  font-variant-numeric: tabular-nums;

  &:hover {
    border-color: rgba(32, 220, 142, 0.35);
  }
`;

const Dot = styled.div`
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: ${(p) => p.theme.colors.morMain};
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
  background: #0d1f18;
  border: 1px solid rgba(255, 255, 255, 0.1);
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
  color: #fff;
  background: ${(p) => (p.$active ? 'rgba(32,220,142,0.10)' : 'transparent')};

  &:hover {
    background: ${(p) =>
      p.$disabled ? 'transparent' : 'rgba(255,255,255,0.06)'};
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
  color: rgba(255, 255, 255, 0.45);
  font-variant-numeric: tabular-nums;
`;

const IconBtn = styled.button`
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.35);
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  display: inline-flex;

  &:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.08);
  }
`;

const Divider = styled.div`
  height: 1px;
  margin: 0.5rem 0.4rem;
  background: rgba(255, 255, 255, 0.08);
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
  color: ${(p) => p.theme.colors.morMain};
  font-size: 1.25rem;
  cursor: pointer;
  text-align: left;

  &:hover:not(:disabled) {
    background: rgba(32, 220, 142, 0.1);
  }
  &:disabled {
    color: rgba(255, 255, 255, 0.25);
    cursor: not-allowed;
  }
`;

const Warn = styled.div`
  display: flex;
  gap: 8px;
  padding: 0.9rem 1rem;
  margin: 0.4rem;
  border-radius: 10px;
  background: rgba(232, 163, 61, 0.1);
  border: 1px solid rgba(232, 163, 61, 0.3);
  color: rgba(255, 255, 255, 0.85);
  font-size: 1.15rem;
  line-height: 1.5;
`;

const Input = styled.input`
  width: 100%;
  padding: 0.8rem 1rem;
  margin: 0.4rem 0;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: #03160e;
  color: #fff;
  font-size: 1.2rem;
`;

function WalletSwitcher({ client, activeAddress, openSessionCount = 0 }) {
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(null);
  const [importing, setImporting] = useState(false);
  const [importKey, setImportKey] = useState('');
  const [pendingSwitch, setPendingSwitch] = useState(null);
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
      }
    };
    const onKey = (e) => e.key === 'Escape' && setOpen(false);
    document.addEventListener('mousedown', onDown);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDown);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

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

  const onRename = async (e, w) => {
    e.stopPropagation();
    const label = window.prompt('Wallet name', w.label);
    if (!label) return;
    try {
      await client.renameWallet({ walletId: w.id, label });
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
                <div style={{ marginTop: 8, display: 'flex', gap: 8 }}>
                  <Action
                    style={{ padding: '0.5rem 0.9rem' }}
                    onClick={() => doSwitch(pendingSwitch.id)}
                  >
                    Switch anyway
                  </Action>
                  <Action
                    style={{ padding: '0.5rem 0.9rem', color: '#fff' }}
                    onClick={() => setPendingSwitch(null)}
                  >
                    Cancel
                  </Action>
                </div>
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
                <IconCheck size={16} color="#20dc8e" />
              ) : (
                <span style={{ width: 16 }} />
              )}
              <RowText>
                <Label>{w.label}</Label>
                <Sub>
                  {abbreviateAddress(w.address, 6)}
                  {w.kind === 'hd' ? ` · account ${w.derivationPath}` : ' · imported'}
                </Sub>
              </RowText>
              {busy === w.id ? (
                <Sub>switching…</Sub>
              ) : (
                <>
                  <IconBtn title="Rename" onClick={(e) => onRename(e, w)}>
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
            <div style={{ padding: '0.4rem' }}>
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
              <div style={{ display: 'flex', gap: 8, marginTop: 6 }}>
                <Action onClick={onImport} disabled={busy === 'import'}>
                  {busy === 'import' ? 'Importing…' : 'Import'}
                </Action>
                <Action
                  style={{ color: '#fff' }}
                  onClick={() => {
                    setImporting(false);
                    setImportKey('');
                  }}
                >
                  Cancel
                </Action>
              </div>
            </div>
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
