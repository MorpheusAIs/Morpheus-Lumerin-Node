import { useMemo } from 'react';
import styled from 'styled-components';
import {
  IconMessage,
  IconMicrophone,
  IconHeadphones,
  IconVector,
  IconPhoto,
  IconEye,
  IconPlugConnectedX,
  IconChevronRight,
  IconHome,
  IconShieldLock,
} from '@tabler/icons-react';
import { formatSmallNumber } from '../utils';

type IconCmp = React.ComponentType<any>;

// Modality tags drive the leading icon + a single canonical badge.
// Any other tags get rendered as muted family/provider chips.
const MODALITY: Record<string, { label: string; Icon: IconCmp }> = {
  llm: { label: 'LLM', Icon: IconMessage },
  chat: { label: 'LLM', Icon: IconMessage },
  tts: { label: 'TTS', Icon: IconHeadphones },
  stt: { label: 'STT', Icon: IconMicrophone },
  embeddings: { label: 'Embeddings', Icon: IconVector },
  embedding: { label: 'Embeddings', Icon: IconVector },
  image: { label: 'Image', Icon: IconPhoto },
  vision: { label: 'Vision', Icon: IconEye },
  multimodal: { label: 'Multimodal', Icon: IconEye },
};

const RowContainer = styled.button<{ $online: boolean }>`
  width: 100%;
  display: grid;
  grid-template-columns: 36px 1fr auto auto;
  gap: 1rem;
  align-items: center;
  padding: 1.2rem 1.4rem;
  margin: 0;
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  color: rgba(255, 255, 255, 0.92);
  cursor: ${(p) => (p.$online ? 'pointer' : 'not-allowed')};
  text-align: left;
  font: inherit;
  transition: background 0.12s ease, border-color 0.12s ease, transform 0.06s ease;
  opacity: ${(p) => (p.$online ? 1 : 0.55)};

  &:hover {
    background: ${(p) => (p.$online ? 'rgba(32, 220, 142, 0.08)' : 'rgba(255, 255, 255, 0.04)')};
    border-color: ${(p) => (p.$online ? 'rgba(32, 220, 142, 0.4)' : 'rgba(255, 255, 255, 0.08)')};
  }

  &:active:not(:disabled) {
    transform: scale(0.997);
  }

  &:focus-visible {
    outline: 2px solid rgba(32, 220, 142, 0.6);
    outline-offset: 2px;
  }

  &:disabled {
    pointer-events: none;
  }
`;

const IconWrap = styled.div`
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: rgba(32, 220, 142, 0.12);
  color: ${(p) => p.theme.colors.morMain};
  display: flex;
  align-items: center;
  justify-content: center;
`;

const NameStack = styled.div`
  min-width: 0; /* allow truncation inside grid cell */
`;

const NameLine = styled.div`
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 1.4rem;
  font-weight: 600;
  letter-spacing: 0.2px;
  color: ${(p) => p.theme.colors.morMain};
`;

const NameText = styled.span`
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
`;

const StatusDot = styled.span<{ $online: boolean }>`
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
  background: ${(p) => (p.$online ? '#20dc8e' : 'rgba(255, 255, 255, 0.25)')};
  box-shadow: ${(p) =>
    p.$online ? '0 0 0 3px rgba(32, 220, 142, 0.18)' : 'none'};
`;

const MetaLine = styled.div`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 3px;
  font-size: 1.1rem;
  color: rgba(255, 255, 255, 0.5);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
`;

const Pill = styled.span<{ $accent?: boolean }>`
  display: inline-flex;
  align-items: center;
  padding: 1px 7px;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: 0.3px;
  text-transform: uppercase;
  background: ${(p) =>
    p.$accent ? 'rgba(32, 220, 142, 0.16)' : 'rgba(255, 255, 255, 0.07)'};
  color: ${(p) =>
    p.$accent ? p.theme.colors.morMain : 'rgba(255, 255, 255, 0.65)'};
`;

/* Distinct accent for the TEE chip so the security attribute reads at a
   glance, even when the row is rendered outside the TEE section (e.g. when
   the user filters to a specific modality). */
const TeePill = styled.span`
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 1px 7px 1px 5px;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: 0.3px;
  background: rgba(125, 188, 255, 0.14);
  color: rgba(173, 211, 255, 0.95);
`;

const Dot = styled.span`
  color: rgba(255, 255, 255, 0.25);
  padding: 0 2px;
`;

const PriceBlock = styled.div`
  text-align: right;
  white-space: nowrap;
`;

const PriceValue = styled.div`
  font-variant-numeric: tabular-nums;
  font-size: 1.25rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.92);
`;

const PriceUnit = styled.div`
  font-size: 0.95rem;
  color: rgba(255, 255, 255, 0.4);
  margin-top: 1px;
`;

const LocalBadge = styled.div`
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px 3px 6px;
  border-radius: 6px;
  background: rgba(32, 220, 142, 0.16);
  color: ${(p) => p.theme.colors.morMain};
  font-size: 1.1rem;
  font-weight: 600;
  letter-spacing: 0.3px;
`;

const OfflineBadge = styled.div`
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px 3px 6px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.55);
  font-size: 1.1rem;
  font-weight: 600;
  letter-spacing: 0.3px;
`;

const Caret = styled.div`
  color: rgba(255, 255, 255, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  ${RowContainer}:hover & {
    color: ${(p) => p.theme.colors.morMain};
  }
`;

function classifyTags(rawTags: string[] = [], modelName: string = '') {
  const modalityKeys: string[] = [];
  const familyTags: string[] = [];
  const seenModality = new Set<string>();
  const normalisedName = modelName.toLowerCase();
  let hasTee = false;

  for (const tag of rawTags) {
    const lower = tag.toLowerCase().trim();
    if (!lower) continue;
    // TEE is a security attribute, not a family tag — surface separately.
    if (lower === 'tee') {
      hasTee = true;
      continue;
    }
    if (MODALITY[lower]) {
      if (!seenModality.has(MODALITY[lower].label)) {
        seenModality.add(MODALITY[lower].label);
        modalityKeys.push(lower);
      }
      continue;
    }
    // Skip tags that are just a prefix of the model name — they duplicate
    // information already shown (e.g. `qwen3-c` tag on `qwen3-coder-…`).
    if (normalisedName.includes(lower) || lower.includes(normalisedName)) {
      continue;
    }
    familyTags.push(tag);
  }

  return { modalityKeys, familyTags, hasTee };
}

type PriceInfo =
  | { kind: 'local' }
  | { kind: 'offline' }
  | { kind: 'single'; perSec: number }
  | { kind: 'range'; minPerSec: number; maxPerSec: number };

function computePrice(model: any): PriceInfo {
  if (model?.isLocal) return { kind: 'local' };
  const bids = (model?.bids || []).filter((b: any) => b?.Id);
  if (bids.length === 0) return { kind: 'offline' };
  const prices = bids
    .map((b: any) => Number(b.PricePerSecond))
    .filter((n: number) => Number.isFinite(n));
  if (prices.length === 0) return { kind: 'offline' };
  const min = Math.min(...prices) / 1e18;
  const max = Math.max(...prices) / 1e18;
  if (min === max) return { kind: 'single', perSec: min };
  return { kind: 'range', minPerSec: min, maxPerSec: max };
}

function ModelRow(props: {
  model: any;
  symbol: string;
  onChangeModel: (data: { modelId: string; bidId?: string; isLocal?: boolean }) => void;
}) {
  const model = props.model || {};
  const modelId = model.Id || '';
  const isLocal = !!model.isLocal;
  const isOnline = isLocal || model.isOnline !== false;
  const symbol = props.symbol || 'MOR';
  const lastCheck: Date | undefined = model.lastCheck
    ? new Date(model.lastCheck)
    : undefined;

  const { modalityKeys, familyTags, hasTee } = useMemo(
    () => classifyTags(model.Tags, model.Name),
    [model.Tags, model.Name],
  );

  const primaryModalityKey = modalityKeys[0] || 'llm';
  const ModalityIcon =
    MODALITY[primaryModalityKey]?.Icon || IconMessage;

  const price = useMemo(() => computePrice(model), [model]);
  const providerCount = (model?.bids || []).filter((b: any) => b?.Id).length;

  const handleSelect = () => {
    if (!isOnline) return;
    if (isLocal) {
      props.onChangeModel({ modelId, isLocal: true });
    } else {
      props.onChangeModel({ modelId });
    }
  };

  // Title tooltip surfaces the full model name + all original tags for
  // discoverability when the row is truncated.
  const tooltip = `${model.Name}${
    model.Tags?.length ? ' — ' + model.Tags.join(', ') : ''
  }`;

  return (
    <RowContainer
      type="button"
      $online={isOnline}
      disabled={!isOnline}
      onClick={handleSelect}
      title={tooltip}
    >
      <IconWrap>
        <ModalityIcon size={20} stroke={1.8} />
      </IconWrap>

      <NameStack>
        <NameLine>
          <StatusDot $online={isOnline} />
          <NameText>{model.Name}</NameText>
        </NameLine>
        <MetaLine>
          {modalityKeys.slice(0, 1).map((key) => (
            <Pill key={key} $accent>
              {MODALITY[key].label}
            </Pill>
          ))}
          {hasTee && (
            <TeePill title="Runs in a Trusted Execution Environment">
              <IconShieldLock size={11} stroke={2.2} />
              TEE
            </TeePill>
          )}
          {!isLocal && providerCount > 1 && (
            <>
              <Dot>·</Dot>
              <span>{providerCount} providers</span>
            </>
          )}
          {familyTags.slice(0, 2).map((t) => (
            <Pill key={t}>{t}</Pill>
          ))}
          {!isOnline && lastCheck && (
            <>
              <Dot>·</Dot>
              <span>
                <IconPlugConnectedX
                  size={12}
                  style={{ verticalAlign: '-2px', marginRight: 3 }}
                />
                Offline since {lastCheck.toLocaleTimeString()}
              </span>
            </>
          )}
        </MetaLine>
      </NameStack>

      <PriceBlock>
        {price.kind === 'local' && (
          <LocalBadge>
            <IconHome size={13} stroke={2} />
            Local
          </LocalBadge>
        )}
        {price.kind === 'offline' && <OfflineBadge>Unavailable</OfflineBadge>}
        {price.kind === 'single' && (
          <>
            <PriceValue>{formatSmallNumber(price.perSec)}</PriceValue>
            <PriceUnit>{symbol}/s</PriceUnit>
          </>
        )}
        {price.kind === 'range' && (
          <>
            <PriceValue>
              {formatSmallNumber(price.minPerSec)} – {formatSmallNumber(price.maxPerSec)}
            </PriceValue>
            <PriceUnit>{symbol}/s</PriceUnit>
          </>
        )}
      </PriceBlock>

      <Caret>
        <IconChevronRight size={18} stroke={2} />
      </Caret>
    </RowContainer>
  );
}

export default ModelRow;
