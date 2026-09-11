import styled from 'styled-components';
import { BaseBtn } from '../common';

export const Container = styled.div`
  margin: 1.6rem 0 1.6rem;
  background-color: #fff;
  padding: 6px 1.6rem 6px 1.6rem;
  border-radius: 0.375rem;
  color: white;
  max-width: 720px;

  background: rgba(255, 255, 255, 0.04);
  border-width: 1px;
  border: 1px solid rgba(255, 255, 255, 0.04);
  color: white;
`;

export const SecondaryContainer = styled.div`
  display: flex;
  min-height: 90px;
  justify-content: space-between;
  align-items: center;
`;

export const WalletBalanceHeader = styled.div`
  font-size: 1.4rem;
  text-align: center;
  color: ${(p) => p.theme.colors.primary};
  margin: 0 0 0.3em;
`;

export const Primary = styled.div`
  display: flex;
  align-items: center;
  line-height: 1.5;
  font-weight: 500;
  letter-spacing: -1px;
  color: ${(p) => p.theme.colors.primary};
  margin: 0 2rem 0 0;
  flex-grow: 1;
  font-size: min(max(20px, 4vw), 24px);
  min-width: 20px;
  overflow: scroll;
  font-size: 2.8rem;
  -ms-overflow-style: none;
  scrollbar-width: none;
  ::-webkit-scrollbar {
    display: none;
  }
`;

export const Btn = styled(BaseBtn)`
  font-size: 1.4rem;
  min-height: 4.4rem;
  margin-left: 0;
  padding: 1rem 1.6rem;
  border-radius: 10px;
  border: 1px solid var(--border-strong, #456a57);
  background-color: var(--surface-raised, #10271e);
  color: var(--text-primary, #edf7f0);
  &:hover:not(:disabled) {
    background-color: var(--surface-hover, #18372a);
  }

  &[data-disabled='true'],
  &[disabled] {
    border: 0.0625rem solid #494949 !important;
    color: #fffc !important;
    background: transparent !important;
  }
`;

export const BtnAccent = styled(Btn)`
  background-color: var(--accent, ${(p) => p.theme.colors.morMain});
  border-color: transparent;
  color: #032117;
  font-weight: 600;
  &:hover:not(:disabled) {
    background-color: #48e4ae;
  }
`;

export const BtnRow = styled.div`
  display: flex;
  flex-direction: row;
  width: 180px;
  height: 100%;
  align-items: center;
  justify-content: space-between;
`;

export const BalanceContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
`;

export const CoinsRow = styled.div`
  display: flex;
`;

export const GlobalContainer = styled.div`
  display: flex;
  align-items: center;
`;
