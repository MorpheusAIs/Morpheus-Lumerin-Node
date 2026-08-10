import { useEffect, useState } from 'react';
import styled from 'styled-components';
import { IconAlertTriangle } from '@tabler/icons-react';

/**
 * Confirmation for the wallet reset.
 *
 * This guards the single most destructive action in the app. Both entry points
 * — "Or setup new wallet" on the login screen and "Reset" in Settings — called
 * logout() directly, which does DELETE /wallet and wipes the private key or
 * mnemonic from the proxy-router's keychain, then relaunches. No confirmation,
 * no warning, and a label ("setup new wallet") that reads as additive rather
 * than destructive. Clicking it on the login screen because you forgot your
 * password destroyed the wallet.
 *
 * Requires typing the confirmation word: a plain OK/Cancel is too easy to
 * dismiss reflexively for something irreversible.
 */

const CONFIRM_WORD = 'DELETE';

const Overlay = styled.div`
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
`;

const Panel = styled.div`
  width: 100%;
  max-width: 480px;
  padding: 2.4rem;
  border-radius: 16px;
  background: #0d1f18;
  border: 1px solid rgba(255, 107, 107, 0.35);
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.6);
  color: #fff;
`;

const Head = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 1.6rem;
`;

const Title = styled.h2`
  margin: 0;
  font-size: 1.9rem;
  color: #ff6b6b;
`;

const Body = styled.div`
  font-size: 1.35rem;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.8);
`;

const List = styled.ul`
  margin: 1.2rem 0;
  padding-left: 1.8rem;

  li {
    margin-bottom: 0.6rem;
  }
`;

const Callout = styled.div`
  padding: 1.2rem 1.4rem;
  margin: 1.6rem 0;
  border-radius: 10px;
  background: rgba(255, 107, 107, 0.1);
  border: 1px solid rgba(255, 107, 107, 0.3);
  font-size: 1.3rem;
  color: #fff;
`;

const Input = styled.input`
  width: 100%;
  padding: 1rem 1.2rem;
  margin-top: 0.8rem;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: #03160e;
  color: #fff;
  font-size: 1.4rem;
  letter-spacing: 1px;
`;

const Actions = styled.div`
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
`;

const Btn = styled.button`
  flex: 1;
  padding: 1.1rem;
  border-radius: 999px;
  font-size: 1.35rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid transparent;
`;

const CancelBtn = styled(Btn)`
  background: transparent;
  border-color: rgba(255, 255, 255, 0.25);
  color: #fff;

  &:hover {
    background: rgba(255, 255, 255, 0.08);
  }
`;

const DangerBtn = styled(Btn)`
  background: ${(p) => (p.disabled ? 'rgba(255,107,107,0.2)' : '#ff6b6b')};
  color: ${(p) => (p.disabled ? 'rgba(255,255,255,0.4)' : '#1a0000')};
  cursor: ${(p) => (p.disabled ? 'not-allowed' : 'pointer')};
`;

type Props = {
  open: boolean;
  onCancel: () => void;
  onConfirm: () => void;
  /** Shown when the user reached this from the login screen. */
  fromLogin?: boolean;
};

export function ConfirmWalletReset({
  open,
  onCancel,
  onConfirm,
  fromLogin,
}: Props) {
  const [typed, setTyped] = useState('');

  useEffect(() => {
    if (!open) {
      setTyped('');
    }
  }, [open]);

  useEffect(() => {
    if (!open) return undefined;
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onCancel();
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open, onCancel]);

  if (!open) {
    return null;
  }

  const armed = typed.trim().toUpperCase() === CONFIRM_WORD;

  return (
    <Overlay onClick={onCancel}>
      <Panel onClick={(e) => e.stopPropagation()} role="alertdialog">
        <Head>
          <IconAlertTriangle size={26} color="#ff6b6b" />
          <Title>This deletes your wallet</Title>
        </Head>

        <Body>
          This does not create an additional wallet. It erases the current one
          from this computer:

          <List>
            <li>Your private key / recovery phrase is deleted from the keychain</li>
            <li>Your password and custom RPC settings are cleared</li>
            <li>The app restarts into first-time setup</li>
          </List>

          <Callout>
            <strong>Your funds are not destroyed</strong> — they stay at the
            address on-chain. But you can only reach them again with your
            recovery phrase or private key. If you have not saved those
            somewhere, this is irreversible.
          </Callout>

          {fromLogin && (
            <p style={{ color: 'rgba(255,255,255,0.6)' }}>
              Forgot your password? This is the only way to regain access — but
              you will need your recovery phrase to restore this wallet.
            </p>
          )}

          <p style={{ color: 'rgba(255,255,255,0.6)' }}>
            To add another wallet without losing this one, cancel and use the
            wallet switcher on the Wallet tab instead.
          </p>

          <label>
            Type <strong>{CONFIRM_WORD}</strong> to confirm:
            <Input
              autoFocus
              value={typed}
              spellCheck={false}
              placeholder={CONFIRM_WORD}
              onChange={(e) => setTyped(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && armed) onConfirm();
              }}
            />
          </label>
        </Body>

        <Actions>
          <CancelBtn onClick={onCancel}>Cancel</CancelBtn>
          <DangerBtn disabled={!armed} onClick={() => armed && onConfirm()}>
            Delete wallet
          </DangerBtn>
        </Actions>
      </Panel>
    </Overlay>
  );
}

export default ConfirmWalletReset;
