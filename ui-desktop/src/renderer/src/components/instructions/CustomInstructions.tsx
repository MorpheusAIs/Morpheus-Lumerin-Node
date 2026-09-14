import { useEffect, useState } from 'react';
import styled from 'styled-components';
import { IconCheck, IconTrash } from '@tabler/icons-react';
import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';
import {
  CUSTOM_INSTRUCTIONS_MAX_LENGTH,
  loadCustomInstructions,
  saveCustomInstructions,
} from '../../lib/customInstructions';

const Intro = styled.p`
  max-width: 68ch;
  margin: 0 0 2.4rem;
  color: rgba(255, 255, 255, 0.62);
  font-size: 1.5rem;
  line-height: 1.6;
`;

const Label = styled.label`
  display: block;
  margin-bottom: 0.8rem;
  color: rgba(255, 255, 255, 0.82);
  font-size: 1.4rem;
  font-weight: 600;
`;

const Editor = styled.textarea`
  display: block;
  width: 100%;
  max-width: 88ch;
  min-height: 26rem;
  padding: 1.4rem 1.6rem;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.04);
  color: #fff;
  font: inherit;
  font-size: 1.5rem;
  line-height: 1.6;
  resize: vertical;

  &::placeholder {
    color: rgba(255, 255, 255, 0.35);
  }

  &:focus {
    outline: none;
    border-color: ${(p) => p.theme.colors.morMain};
  }
`;

const Meta = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.2rem;
  max-width: 88ch;
  margin-top: 0.8rem;
  color: rgba(255, 255, 255, 0.4);
  font-size: 1.3rem;
`;

const Counter = styled.span<{ $over: boolean }>`
  font-variant-numeric: tabular-nums;
  color: ${(p) =>
    p.$over ? p.theme.colors.danger : 'rgba(255, 255, 255, 0.4)'};
`;

const Actions = styled.div`
  display: flex;
  align-items: center;
  gap: 1.2rem;
  margin-top: 2rem;
`;

const SaveBtn = styled.button.attrs({ type: 'button' })`
  display: inline-flex;
  align-items: center;
  gap: 0.8rem;
  padding: 1rem 2rem;
  border: none;
  border-radius: 8px;
  background: ${(p) => p.theme.colors.morMain};
  color: ${(p) => p.theme.colors.primaryDark};
  font: inherit;
  font-size: 1.4rem;
  font-weight: 650;
  cursor: pointer;

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
`;

const ClearBtn = styled.button.attrs({ type: 'button' })`
  display: inline-flex;
  align-items: center;
  gap: 0.8rem;
  padding: 1rem 1.6rem;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 8px;
  background: transparent;
  color: rgba(255, 255, 255, 0.62);
  font: inherit;
  font-size: 1.4rem;
  font-weight: 550;
  cursor: pointer;

  &:hover:not(:disabled) {
    color: ${(p) => p.theme.colors.danger};
    border-color: ${(p) => p.theme.colors.danger};
  }

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
`;

const Saved = styled.span`
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  color: ${(p) => p.theme.colors.morMain};
  font-size: 1.35rem;
`;

const PLACEHOLDER = [
  'Examples:',
  '',
  'Answer in British English and keep replies under 200 words.',
  'I am a backend engineer; skip the basics and show code first.',
  'Always show your reasoning before the conclusion.',
].join('\n');

export function CustomInstructions() {
  const [value, setValue] = useState('');
  const [savedValue, setSavedValue] = useState('');
  const [justSaved, setJustSaved] = useState(false);

  useEffect(() => {
    const stored = loadCustomInstructions();
    setValue(stored);
    setSavedValue(stored);
  }, []);

  useEffect(() => {
    if (!justSaved) return undefined;
    const id = setTimeout(() => setJustSaved(false), 2200);
    return () => clearTimeout(id);
  }, [justSaved]);

  const over = value.length > CUSTOM_INSTRUCTIONS_MAX_LENGTH;
  const dirty = value.trim() !== savedValue.trim();

  const onSave = () => {
    const next = value.trim();
    saveCustomInstructions(next);
    setSavedValue(next);
    setValue(next);
    setJustSaved(true);
  };

  const onClear = () => {
    saveCustomInstructions('');
    setSavedValue('');
    setValue('');
    setJustSaved(true);
  };

  return (
    <View data-testid="custom-instructions-container">
      <LayoutHeader title="Custom instructions" />
      <Intro>
        Tell every model how you want it to respond — a tone, a format, a
        language, or anything it should always know about you. These
        instructions are sent with each message in Chat, and apply to every chat
        and every model. Leave the box empty to turn the feature off entirely.
      </Intro>

      <Label htmlFor="custom-instructions-input">Your instructions</Label>
      <Editor
        id="custom-instructions-input"
        value={value}
        placeholder={PLACEHOLDER}
        spellCheck
        onChange={(event) => setValue(event.target.value)}
      />
      <Meta>
        <span>Stored on this computer only. Never sent anywhere else.</span>
        <Counter $over={over} aria-live="polite">
          {value.length} / {CUSTOM_INSTRUCTIONS_MAX_LENGTH}
        </Counter>
      </Meta>

      <Actions>
        <SaveBtn onClick={onSave} disabled={!dirty || over}>
          Save instructions
        </SaveBtn>
        <ClearBtn onClick={onClear} disabled={!value && !savedValue}>
          <IconTrash size={16} stroke={1.8} />
          Clear
        </ClearBtn>
        {justSaved && (
          <Saved role="status">
            <IconCheck size={16} stroke={2.2} />
            Saved
          </Saved>
        )}
      </Actions>
    </View>
  );
}

export default CustomInstructions;
