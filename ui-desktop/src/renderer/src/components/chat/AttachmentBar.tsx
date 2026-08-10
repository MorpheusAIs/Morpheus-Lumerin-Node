import styled from 'styled-components';
import {
  IconFileText,
  IconPhoto,
  IconX,
  IconAlertTriangle,
  IconLoader2,
} from '@tabler/icons-react';

import {
  Attachment,
  estimateTokens,
  formatBytes,
  formatTokens,
} from '../../store/utils/attachments';

const Bar = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0.8rem 1rem;
  margin-bottom: 0.6rem;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
`;

const Chips = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
`;

const Chip = styled.div<{ $error?: boolean }>`
  display: inline-flex;
  align-items: center;
  gap: 8px;
  max-width: 280px;
  padding: 0.5rem 0.7rem;
  border-radius: 8px;
  font-size: 1.15rem;
  color: #fff;
  background: ${(p) =>
    p.$error ? 'rgba(255,107,107,0.12)' : 'rgba(255,255,255,0.06)'};
  border: 1px solid
    ${(p) => (p.$error ? 'rgba(255,107,107,0.35)' : 'rgba(255,255,255,0.1)')};
`;

const Thumb = styled.img`
  width: 26px;
  height: 26px;
  border-radius: 4px;
  object-fit: cover;
  flex-shrink: 0;
`;

const Name = styled.span`
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const Meta = styled.span`
  color: rgba(255, 255, 255, 0.45);
  font-size: 1.05rem;
  white-space: nowrap;
`;

const Remove = styled.button`
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.4);
  cursor: pointer;
  display: inline-flex;
  padding: 2px;
  border-radius: 4px;

  &:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }
`;

const Footer = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 1.1rem;
  color: rgba(255, 255, 255, 0.45);
`;

const Warning = styled.div`
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 0.7rem 0.9rem;
  border-radius: 8px;
  background: rgba(232, 163, 61, 0.1);
  border: 1px solid rgba(232, 163, 61, 0.3);
  color: rgba(255, 255, 255, 0.85);
  font-size: 1.1rem;
  line-height: 1.5;
`;

const spin = `
  @keyframes spin { to { transform: rotate(360deg); } }
`;

const Spinner = styled(IconLoader2)`
  ${spin}
  animation: spin 0.9s linear infinite;
`;

type Props = {
  attachments: Attachment[];
  prompt: string;
  onRemove: (id: string) => void;
  /** True when the selected model is not recognised as vision-capable. */
  visionWarning?: boolean;
  /** True when this model has actually refused an image before — a fact, not a guess. */
  visionRejected?: boolean;
  modelName?: string;
};

export function AttachmentBar({
  attachments,
  prompt,
  onRemove,
  visionWarning,
  visionRejected,
  modelName,
}: Props) {
  if (!attachments.length) {
    return null;
  }

  const tokens = estimateTokens(attachments, prompt);
  const hasImages = attachments.some((a) => a.kind === 'image');
  const emptyOnes = attachments.filter((a) => a.empty);

  return (
    <Bar data-testid="attachment-bar">
      <Chips>
        {attachments.map((a) => (
          <Chip key={a.id} $error={a.status === 'error'} title={a.error || a.note || a.name}>
            {a.status === 'parsing' ? (
              <Spinner size={16} />
            ) : a.kind === 'image' && a.dataUrl ? (
              <Thumb src={a.dataUrl} alt="" />
            ) : a.kind === 'image' ? (
              <IconPhoto size={16} />
            ) : (
              <IconFileText size={16} />
            )}

            <Name>{a.name}</Name>
            <Meta>
              {a.status === 'parsing'
                ? 'reading…'
                : a.status === 'error'
                  ? 'failed'
                  : formatBytes(a.size)}
            </Meta>

            <Remove onClick={() => onRemove(a.id)} title="Remove">
              <IconX size={14} />
            </Remove>
          </Chip>
        ))}
      </Chips>

      {/* Two levels, because they carry very different confidence.
          `visionRejected` means this model has actually refused an image
          before — a fact. `visionWarning` is only a guess from the model name,
          since vision capability is not published on-chain. */}
      {hasImages && visionRejected ? (
        <Warning style={{ background: 'rgba(255,107,107,0.1)', borderColor: 'rgba(255,107,107,0.35)' }}>
          <IconAlertTriangle size={15} />
          <span>
            {modelName ? `"${modelName}"` : 'This model'} has already rejected an
            image once — it cannot read pictures. Sending will fail again.
            Remove the image, or pick a vision-capable model. Documents still work.
          </span>
        </Warning>
      ) : (
        hasImages &&
        visionWarning && (
          <Warning>
            <IconAlertTriangle size={15} />
            <span>
              {modelName ? `"${modelName}"` : 'This model'} isn&apos;t recognised
              as accepting images. You can still send — it may work, or the
              provider may return an error or ignore the picture. Documents are
              unaffected.
            </span>
          </Warning>
        )
      )}

      {/* A scanned PDF parses "successfully" and yields nothing. Without this
          the user sends an empty context block and wonders why the model has
          not read their document. */}
      {emptyOnes.length > 0 && (
        <Warning>
          <IconAlertTriangle size={15} />
          <span>
            No readable text in {emptyOnes.map((a) => a.name).join(', ')}
            {emptyOnes[0].note ? ` (${emptyOnes[0].note})` : ''}. Nothing from{' '}
            {emptyOnes.length > 1 ? 'these files' : 'this file'} will reach the
            model.
          </span>
        </Warning>
      )}

      <Footer>
        <span>
          {attachments.length} attachment{attachments.length > 1 ? 's' : ''}
        </span>
        <span title="Rough estimate — actual usage depends on the model's tokenizer">
          {formatTokens(tokens)} incl. your message
        </span>
      </Footer>
    </Bar>
  );
}

export default AttachmentBar;
