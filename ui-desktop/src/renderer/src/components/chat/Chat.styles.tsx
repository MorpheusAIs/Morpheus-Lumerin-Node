import styled from 'styled-components';
import TextareaAutosize from 'react-textarea-autosize';
import { BtnAccent } from '../dashboard/BalanceBlock.styles';

export const View = styled.div`
  display: flex;
  flex-direction: column;
  height: 100vh;
  max-width: 100%;
  min-width: 0;
  position: relative;
  width: 100%;
`;

export const Container = styled.div`
  max-width: 1120px;
  flex: 1 1 auto;
  min-height: 0;
  justify-content: space-between;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  padding: 20px 2.4rem 0;
  width: 100%;
`;

export const ChatBlock = styled.div`
  width: 100%;
  height: 100%;
  overflow-y: auto;
  margin-bottom: 20px;

  &.createSessionMode {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  &.createSessionMode .session-container {
    width: 450px;
    padding: 1rem;
    background-color: rgba(138, 43, 226, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.16);
  }

  &.createSessionMode .session-title {
    text-align: center;
    margin-bottom: 10px;
  }
`;

export const ChatHistoryContainer = styled.div`
  overflow-y: auto;
  width: 100%;
  height: 100%;
`;

export const ChatStartupState = styled.div`
  align-items: center;
  color: rgba(255, 255, 255, 0.72);
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 1rem;
  height: 100%;
  justify-content: center;
  padding: 3rem;
  text-align: center;

  svg {
    color: ${(p) => p.theme.colors.morMain};
  }

  strong {
    color: rgba(255, 255, 255, 0.95);
    font-size: 1.8rem;
    font-weight: 600;
  }

  span {
    font-size: 1.3rem;
    line-height: 1.55;
    max-width: 52ch;
  }

  button {
    margin-top: 0.8rem;
  }
`;

export const ChatIntroContainer = styled.div`
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow-y: auto;
  margin-bottom: 20px;
  padding: clamp(1.6rem, 5vh, 4rem) 0;
  display: flex;
  align-items: flex-start;
  justify-content: center;
`;

export const ChatIntroInner = styled.div`
  box-sizing: border-box;
  flex: 0 0 auto;
  margin: auto;
  max-width: 100%;
  padding: clamp(2.4rem, 6vw, 5.4rem);
  width: 48.6rem;
  background-color: ${(p) => p.theme.colors.primaryDark};
  border-radius: 15px;
`;

export const ChatIntroInnerTitle = styled.h2`
  font-size: 20px;
  font-weight: 600;
`;

export const ChatIntroButton = styled(BtnAccent)`
  width: 215px;
  font-size: 16px;
  padding: 0.5em;
  margin: 0;
  font-weight: 600;
  color: ${(p) => p.theme.colors.primaryDark};
`;

export const ChatIntroInnerText = styled.p`
  font-size: 14px;
  font-weight: 400;
  color: #ffffff;
  margin-top: 40px;
  margin-bottom: 25px;
`;

export const SessionDurationField = styled.label`
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 24px 0 8px;
  color: #ffffff;
  font-size: 14px;

  select {
    width: 100%;
    padding: 10px 12px;
    color: #ffffff;
    background: ${(p) => p.theme.colors.primary};
    border: 1px solid rgba(255, 255, 255, 0.22);
    border-radius: 8px;
    font: inherit;
    cursor: pointer;
  }

  select:focus-visible {
    outline: 2px solid ${(p) => p.theme.colors.active};
    outline-offset: 2px;
  }
`;

export const SessionCostSummary = styled.div`
  margin: 10px 0 24px;
  color: rgba(255, 255, 255, 0.72);
  font-size: 13px;
  line-height: 1.45;
`;

export const SessionSetupState = styled.div`
  min-height: 250px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 32px 0 12px;
  color: rgba(255, 255, 255, 0.7);
  text-align: center;

  strong {
    color: rgba(255, 255, 255, 0.95);
    font-size: 16px;
    font-weight: 600;
  }

  span {
    max-width: 38ch;
    font-size: 13px;
    line-height: 1.5;
  }

  .spinner-border {
    width: 32px;
    height: 32px;
  }
`;

export const SessionSetupActions = styled.div`
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
  margin-top: 8px;

  ${ChatIntroButton} {
    width: auto;
    min-width: 148px;
    min-height: 40px;
  }
`;

export const SessionHistoryNotice = styled.div`
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin: 18px 0 0;
  color: #e8a33d;
  font-size: 12px;
  line-height: 1.45;

  span {
    flex: 1;
  }

  button {
    min-height: 40px;
    padding: 0 10px;
    flex: 0 0 auto;
    border: 1px solid rgba(232, 163, 61, 0.45);
    border-radius: 7px;
    background: transparent;
    color: #f2bd6d;
    cursor: pointer;
    font: inherit;
  }

  button:focus-visible {
    outline: 2px solid ${(p) => p.theme.colors.active};
    outline-offset: 2px;
  }
`;

export const Control = styled.div`
  height: fit-content;
  position: relative;
  display: flex;
  flex-direction: column;

  textarea {
    resize: none;
    padding-right: 6rem;
  }

  textarea:focus,
  input:focus {
    outline: none !important;
  }
`;

export const SendBtnWrapper = styled.div`
  position: absolute;
  right: 16px;
  bottom: 12px;
  display: flex;
  gap: 10px;
`;

export const Btn = styled.button`
  padding: 2px 5px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  bottom: 12px;
  background: ${(p) => p.theme.colors.morMain};
  cursor: pointer;
  border: none;
  &[disabled] {
    opacity: 0.5;
  }
  border-radius: 5px;
`;

export const SendBtn = styled(Btn)`
  position: absolute;
  right: 16px;
  width: fit-content;
`;

export const Avatar = styled.div`
  height: 36px;
  min-width: 36px;
  width: 36px;
  display: flex;
  justify-content: center;
  align-items: center;
  /* border: 1px solid; */
  background: ${(p) => p.color};
  font-weight: 400;
  font-size: 15px;
  border-radius: 4px;
`;

export const AvatarHeader = styled.div`
  color: ${(p) => p.theme.colors.morMain};
  font-weight: 900;
  padding: 0 8px;
  font-size: 18px;
  line-height: 18px;
  margin-bottom: 5px;
`;

export const MessageBody = styled.div`
  font-weight: 400;
  padding: 0 8px;
  font-size: 18px;
  max-width: calc(100vw - 165px);

  code {
    color: ${(p) => p.theme.colors.morMain};
  }

  @media (min-width: 800px) {
    max-width: calc(100vw - 310px);
  }
`;

export const ChatTitleContainer = styled.div`
  color: ${(p) => p.theme.colors.morMain};
  font-weight: 900;
  padding: 0 8px;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.16);
`;

export const ChatAvatar = styled.div`
  display: flex;
  align-items: center;
`;

export const CustomTextArrea = styled(TextareaAutosize)`
  background: transparent;
  box-sizing: border-box;
  width: 100%;
  font-size: 18px;
  border-radius: 12px;
  color: white;
  padding: 12px 16px;

  &::focus {
    outline: none !important;
  }

  textarea:focus,
  input:focus {
    outline: none !important;
  }
`;

export const ContainerTitle = styled.div`
  display: flex;
  flex: 0 0 auto;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  position: sticky;
  width: 100%;
  padding: 0 2.4rem;
  z-index: 2;
  right: 0;
  left: 0;
  top: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.16);
`;

export const TitleRow = styled.div`
  width: 100%;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
`;

export const ChatHeaderControls = styled.div`
  align-items: center;
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 35rem) max-content;
  justify-content: start;
  min-width: 0;
  width: 100%;
`;

export const ChatHeaderActions = styled.div`
  align-items: center;
  display: flex;
  gap: 1rem;
  min-width: 0;
`;

export const ChatHeaderActionButton = styled(BtnAccent)`
  align-items: center;
  box-sizing: border-box;
  display: inline-flex;
  flex: 0 0 14rem;
  gap: 0.8rem;
  height: auto;
  justify-content: center;
  line-height: 1.2;
  margin: 0;
  min-height: 6rem;
  padding: 1.2rem;
  white-space: nowrap;
  width: 14rem;

  svg {
    flex: 0 0 auto;
  }

  &:focus-visible {
    outline: 2px solid ${(p) => p.theme.colors.active};
    outline-offset: 2px;
  }
`;

export const Title = styled.label`
  font-size: 2.4rem;
  line-height: 3rem;
  white-space: nowrap;
  margin: 0;
  max-width: 1120px;
  font-weight: 600;
  color: ${(p) => p.theme.colors.morMain};
  margin-bottom: 4.8px;
  margin-right: 2.4rem;
  cursor: default;
  /* width: 100%; */

  @media (min-width: 1140px) {
  }

  @media (min-width: 1200px) {
  }
`;

export const LoadingCover = styled.div`
  width: 100%;
  height: 100%;
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);

  z-index: 5;
`;

export const LoadingStatus = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  max-width: 36rem;
  padding: 24px;
  text-align: center;

  strong {
    color: rgba(255, 255, 255, 0.95);
    font-size: 16px;
  }

  span {
    color: rgba(255, 255, 255, 0.72);
    font-size: 13px;
  }
`;

export const ImageContainer = styled.img`
  cursor: pointer;
  padding: 0.25rem;
  background-color: ${(p) => p.theme.colors.morMain}B3;
  border: var(--bs-border-width) solid var(--bs-highlight-color);
  border-radius: var(--bs-border-radius);
  max-width: 100%;
  height: 256px;

  @media (min-height: 700px) {
    height: 320px;
  }
`;

export const VideoContainer = styled.div`
  cursor: pointer;
  padding: 0.25rem;
  max-width: 100%;
  height: 256px;

  @media (min-height: 700px) {
    height: 320px;
  }
`;

export const SubPriceLabel = styled.span`
  color: ${(p) => p.theme.colors.morMain};
`;

export const AudioInputZone = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding: 14px 16px;
  border: 1px dashed rgba(255, 255, 255, 0.2);
  border-radius: 12px;

  &[data-disabled='true'] {
    opacity: 0.5;
    pointer-events: none;
  }
`;

export const AudioActionBtn = styled.button`
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid ${(p) => p.theme.colors.morMain};
  background: transparent;
  color: ${(p) => p.theme.colors.morMain};
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition:
    background 0.12s ease,
    color 0.12s ease;

  &:hover {
    background: rgba(32, 220, 142, 0.12);
  }

  &[disabled] {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &[data-recording='true'] {
    border-color: #ff5b5b;
    color: #ff5b5b;
    background: rgba(255, 91, 91, 0.12);
  }
`;

export const AudioHint = styled.span`
  font-size: 13px;
  color: rgba(255, 255, 255, 0.55);
`;

export const TtsControlsRow = styled.div`
  display: flex;
  align-items: center;
  gap: 18px;
  flex-wrap: wrap;
  padding: 0 16px 10px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);

  label {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin: 0;
  }

  select,
  input[type='text'] {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.16);
    border-radius: 6px;
    color: white;
    padding: 4px 8px;
    font-size: 13px;
  }

  select:focus,
  input:focus {
    outline: none !important;
    border-color: ${(p) => p.theme.colors.morMain};
  }
`;

export const AudioPlayer = styled.audio`
  width: 320px;
  max-width: 100%;
  margin-top: 4px;
`;
