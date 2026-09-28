import { useMemo } from 'react';
import styled from 'styled-components';
import {
  IconMessage,
  IconMicrophone,
  IconHeadphones,
  IconVector,
  IconPhoto,
  IconEye,
  IconChevronRight,
  IconHome,
  IconShieldLock,
} from '@tabler/icons-react';
import { formatSmallNumber, SECURE_TAG, SECURE_BADGE_TOOLTIP } from '../utils';
import { getVisionCapability } from '../../../store/utils/attachments';
import {
  normalizeModelName,
  normalizeModelTags,
} from '../../../store/utils/modelMetadata';
import {
  ModelPriceEntry,
  weiToMorPerSecond,
} from '../../../store/utils/modelPrices';

type IconCmp = React.ComponentType<any>;

// Modality tags drive the leading icon + a single canonical badge.
// Any other tags get rendered as muted family/provider chips.
const MODALITY: Record<string, { label: string; Icon: IconCmp }> = {
  llm: { label: 'LLM', Icon: IconMessage },
  chat: { label: 'LLM', Icon: IconMessage },
  tts: { label: 'Text-to-Speech', Icon: IconHeadphones },
  stt: { label: 'Speech-to-Text', Icon: IconMicrophone },
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
  content-visibility: auto;
  contain-intrinsic-size: auto 68px;
  cursor: ${(p) => (p.$online ? 'pointer' : 'not-allowed')};
  text-align: left;
  font: inherit;
  transition:
    background 0.12s ease,
    border-color 0.12s ease,
    transform 0.06s ease;
  opacity: ${(p) => (p.$online ? 1 : 0.55)};

  &:hover {
    background: ${(p) =>
      p.$online ? 'rgba(32, 220, 142, 0.08)' : 'rgba(255, 255, 255, 0.04)'};
    border-color: ${(p) =>
      p.$online ? 'rgba(32, 220, 142, 0.4)' : 'rgba(255, 255, 255, 0.08)'};
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

const VisionPill = styled.span<{ $declared: boolean }>`
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 1px 7px 1px 5px;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: 0.3px;
  background: ${(p) =>
    p.$declared ? 'rgba(190, 125, 255, 0.18)' : 'rgba(190, 125, 255, 0.1)'};
  color: ${(p) =>
    p.$declared ? 'rgba(224, 190, 255, 1)' : 'rgba(210, 180, 240, 0.85)'};
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

const CheckPriceBadge = styled(OfflineBadge)`
  color: ${(p) => p.theme.colors.morMain};
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

function classifyTags(rawTags: unknown, modelName: string = '') {
  const modalityKeys: string[] = [];
  const familyTags: string[] = [];
  const seenModality = new Set<string>();
  const normalisedName = modelName.toLowerCase();
  let hasTee = false;

  for (const tag of normalizeModelTags(rawTags)) {
    const lower = tag.toLowerCase().trim();
    if (!lower) continue;
    // TEE is a security attribute, not a family tag — surface separately.
    if (lower === SECURE_TAG) {
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
  /** Nobody is serving this model, so there is no price to quote. */
  | { kind: 'none' }
  /** Not looked up yet, or the lookup failed. Distinct from 'none' on purpose. */
  | { kind: 'unknown' }
  | { kind: 'single'; perSec: number; providers: number }
  | { kind: 'range'; minPerSec: number; maxPerSec: number; providers: number };

/**
 * What this model costs per second, from whichever source knows.
 *
 * Two sources, in order. `model.bids` is the per-model active-bid read, which
 * only the selected model has and which is the more current of the two.
 * `priceEntry` is the router's marketplace-wide sweep, which every row has once
 * the picker has loaded it and which is what makes ordering by cost possible at
 * all. They agree on substance: both exclude this wallet's own bids and both
 * exclude withdrawn ones, the sweep on the router side and the bid list here.
 *
 * A zero price is dropped rather than shown as free. getSessionEnd divides by
 * price per second, so a zero-priced bid is malformed, not a bargain, and
 * showing it would sort the model to the top of a cheapest-first list and then
 * fail at the point of opening.
 */
function computePrice(
  model: any,
  priceEntry: ModelPriceEntry | undefined,
): PriceInfo {
  if (model?.isLocal) return { kind: 'local' };

  const bids = Array.isArray(model?.bids)
    ? model.bids.filter((b: any) => b?.Id && Number(b.DeletedAt ?? 0) === 0)
    : null;
  if (bids) {
    const prices = bids
      .map((b: any) => Number(b.PricePerSecond))
      .filter((n: number) => Number.isFinite(n) && n > 0);
    if (prices.length === 0) return { kind: 'none' };
    const min = Math.min(...prices) / 1e18;
    const max = Math.max(...prices) / 1e18;
    return min === max
      ? { kind: 'single', perSec: min, providers: prices.length }
      : {
          kind: 'range',
          minPerSec: min,
          maxPerSec: max,
          providers: prices.length,
        };
  }

  if (priceEntry) {
    const min = weiToMorPerSecond(priceEntry.min_price_per_second_wei);
    const max = weiToMorPerSecond(priceEntry.max_price_per_second_wei);
    if (min === undefined) return { kind: 'none' };
    const providers = priceEntry.bid_count;
    return max === undefined || max === min
      ? { kind: 'single', perSec: min, providers }
      : { kind: 'range', minPerSec: min, maxPerSec: max, providers };
  }

  // Neither source has spoken for this model. That is not the same as having no
  // providers, and saying so would tell the user a model is dead when all that
  // happened is that the sweep has not answered yet or could not be read.
  return { kind: 'unknown' };
}

function ModelRow(props: {
  model: any;
  symbol: string;
  /**
   * This model's row in the router's marketplace-wide price sweep.
   *
   * Undefined means the sweep has not answered for this model — not loaded yet,
   * or it failed. An entry with an empty price is the opposite: the sweep did
   * answer, and the answer is that nobody is serving this model.
   */
  priceEntry?: ModelPriceEntry;
  onChangeModel: (data: {
    modelId: string;
    bidId?: string;
    isLocal?: boolean;
  }) => void;
}) {
  const model = props.model || {};
  const modelId = model.Id || '';
  const modelName = normalizeModelName(model.Name);
  const modelTags = useMemo(() => normalizeModelTags(model.Tags), [model.Tags]);
  const isLocal = !!model.isLocal;
  const hasBidData = Array.isArray(model?.bids);
  const providerCount = hasBidData
    ? model.bids.filter((bid: any) => bid?.Id).length
    : (props.priceEntry?.bid_count ?? 0);
  // Neither the per-model bid read nor the marketplace sweep has spoken for this
  // row yet. Saying "offline since" on that basis would be an assertion the row
  // has not earned.
  const availabilityUnknown = !isLocal && !hasBidData && !props.priceEntry;
  // The price sweep says how many providers are live, but it is a cached read
  // taken up to a minute ago and it excludes this wallet's own bids. Letting it
  // disable a row would mean a model the user could open being greyed out on
  // stale data, so it informs the price column and nothing else. Only the
  // per-model bid read, which is fetched for the model actually chosen, still
  // gates selection.
  const isOnline =
    isLocal ||
    !hasBidData ||
    (providerCount > 0 && model.isOnline !== false);
  const symbol = props.symbol || 'MOR';
  const lastCheck: Date | undefined = model.lastCheck
    ? new Date(model.lastCheck)
    : undefined;

  const { modalityKeys, familyTags, hasTee } = useMemo(
    () => classifyTags(modelTags, modelName),
    [modelName, modelTags],
  );

  const primaryModalityKey = modalityKeys[0] || 'llm';
  const ModalityIcon = MODALITY[primaryModalityKey]?.Icon || IconMessage;

  const price = useMemo(
    () => computePrice(model, props.priceEntry),
    [model, props.priceEntry],
  );
  const visionCapability = getVisionCapability(model);

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
  const tooltip = `${modelName}${modelTags.length ? ` — ${modelTags.join(', ')}` : ''}`;

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
          <StatusDot $online={isLocal || providerCount > 0} />
          <NameText>{modelName}</NameText>
        </NameLine>
        <MetaLine>
          {modalityKeys.slice(0, 1).map((key) => (
            <Pill key={key} $accent>
              {MODALITY[key].label}
            </Pill>
          ))}
          {hasTee && (
            <TeePill title={SECURE_BADGE_TOOLTIP}>
              <IconShieldLock size={11} stroke={2.2} />
              Secure
            </TeePill>
          )}
          {visionCapability !== 'none' && (
            <VisionPill
              $declared={visionCapability === 'declared'}
              title={
                visionCapability === 'declared'
                  ? 'Image input support is declared by this model’s tags.'
                  : 'Likely supports image input based on its recognised model family; the provider has not declared a vision tag.'
              }
            >
              <IconEye size={11} stroke={2.2} />
              {visionCapability === 'declared' ? 'Vision' : 'Likely vision'}
            </VisionPill>
          )}
          {/* The provider count now sits under the price, where it qualifies
              the figure it belongs to. Repeating it here said the same thing
              twice in one row. */}
          {familyTags.slice(0, 2).map((t) => (
            <Pill key={t}>{t}</Pill>
          ))}
          {!availabilityUnknown && !isOnline && lastCheck && (
            <>
              <Dot>·</Dot>
              <span>Offline since {lastCheck.toLocaleTimeString()}</span>
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
        {price.kind === 'unknown' && (
          <CheckPriceBadge>Check price</CheckPriceBadge>
        )}
        {price.kind === 'none' && <OfflineBadge>No providers</OfflineBadge>}
        {/* The unit says what the number is per, and the provider count says
            what it is one of. A single figure with neither reads like the price
            of the model, when it is the cheapest of several offers for it. */}
        {price.kind === 'single' && (
          <>
            <PriceValue>{formatSmallNumber(price.perSec)}</PriceValue>
            <PriceUnit>
              {symbol}/s
              {price.providers > 1 ? ` · ${price.providers} providers` : ''}
            </PriceUnit>
          </>
        )}
        {price.kind === 'range' && (
          <>
            <PriceValue>
              {formatSmallNumber(price.minPerSec)} –{' '}
              {formatSmallNumber(price.maxPerSec)}
            </PriceValue>
            <PriceUnit>
              {symbol}/s
              {price.providers > 1 ? ` · ${price.providers} providers` : ''}
            </PriceUnit>
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
