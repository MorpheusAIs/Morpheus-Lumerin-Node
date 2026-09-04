import { useMemo, useState } from 'react';
import styled, { keyframes } from 'styled-components';

const spin = keyframes`
  to { transform: rotate(360deg); }
`;
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import {
  IconSearch,
  IconMessage,
  IconMicrophone,
  IconHeadphones,
  IconVector,
  IconHome,
  IconWorld,
  IconShieldLock,
  IconInfoCircle,
  IconEye,
  IconSparkles,
  IconLoader2,
} from '@tabler/icons-react';
import Modal from '../../contracts/modals/Modal';
import ModelRow from './ModelRow';
import { isCoworkCandidate, isSecureModel, SECURE_MODE_INFO } from '../utils';
import { getVisionCapability } from '../../../store/utils/attachments';

/* The shared outer modal `Body` (in CreateContractModal.styles) bakes in
   `padding: 5rem` and never sets `overflow: hidden`, so an `auto`-height box
   with `max-height: 78vh` still lets children visually spill past its
   bottom edge.
   We override via inline `style` (beats the styled-component CSS) to:
     - give the box a definite height so child `height: 100%` resolves,
     - clip overflow so the inner scroll region is the real scroll boundary,
     - zero out the 5rem padding so we control spacing inside the Layout. */
const bodyProps = {
  width: '640px',
  maxWidth: '90%',
  onClick: (e: React.MouseEvent) => e.stopPropagation(),
  style: {
    height: 'min(78vh, 760px)',
    maxHeight: '78vh',
    padding: 0,
    overflow: 'hidden',
  },
};

const Layout = styled.div`
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
`;

/* Right padding leaves room for the absolute-positioned close X
   (32px button at top: 12px / right: 12px → clears ~52px from the right). */
const Header = styled.div`
  padding: 1.8rem 5.5rem 1.4rem 2.4rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
`;

const TitleRow = styled.div`
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1.4rem;
  margin-bottom: 1.4rem;
`;

const Title = styled.h2`
  margin: 0;
  font-size: 1.9rem;
  font-weight: 600;
  letter-spacing: 0.2px;
  color: ${(p) => p.theme.colors.morMain};
`;

const ResultCount = styled.div`
  font-size: 1.15rem;
  color: rgba(255, 255, 255, 0.45);
  font-variant-numeric: tabular-nums;
`;

const SearchWrapper = styled.div`
  .input-group {
    background: rgba(255, 255, 255, 0.04);
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.06);
    transition:
      border-color 0.15s ease,
      background 0.15s ease;
  }

  .input-group:focus-within {
    border-color: ${(p) => p.theme.colors.morMain};
    background: rgba(255, 255, 255, 0.06);
  }

  .input-group-text {
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.8);
    padding-right: 0;
  }

  /* Bright placeholder so the prompt reads clearly against the dark surface. */
  .form-control::placeholder,
  input::placeholder {
    color: rgba(255, 255, 255, 0.7) !important;
    opacity: 1; /* Firefox dims placeholders by default; reset. */
  }
`;

const FilterRow = styled.div`
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 1.2rem;
`;

const CoworkCandidateHint = styled.div`
  display: flex;
  align-items: flex-start;
  gap: 7px;
  margin-top: 0.9rem;
  color: rgba(255, 255, 255, 0.72);
  font-size: 1.05rem;
  line-height: 1.45;

  svg {
    flex: 0 0 auto;
    margin-top: 1px;
    color: ${(p) => p.theme.colors.morMain};
  }
`;

const FilterPill = styled.button<{ $active: boolean }>`
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 10px;
  border-radius: 999px;
  border: 1px solid
    ${(p) =>
      p.$active ? 'rgba(32, 220, 142, 0.5)' : 'rgba(255, 255, 255, 0.08)'};
  background: ${(p) =>
    p.$active ? 'rgba(32, 220, 142, 0.14)' : 'rgba(255, 255, 255, 0.03)'};
  color: ${(p) =>
    p.$active ? p.theme.colors.morMain : 'rgba(255, 255, 255, 0.7)'};
  font-size: 1.15rem;
  font-weight: 500;
  letter-spacing: 0.2px;
  cursor: pointer;
  transition:
    background 0.12s ease,
    border-color 0.12s ease,
    color 0.12s ease;

  &:hover {
    background: ${(p) =>
      p.$active ? 'rgba(32, 220, 142, 0.2)' : 'rgba(255, 255, 255, 0.06)'};
    color: ${(p) =>
      p.$active ? p.theme.colors.morMain : 'rgba(255, 255, 255, 0.9)'};
  }

  &:focus-visible {
    outline: 2px solid rgba(32, 220, 142, 0.5);
    outline-offset: 2px;
  }
`;

const FilterCount = styled.span<{ $active: boolean }>`
  font-size: 0.95rem;
  padding: 1px 6px;
  border-radius: 8px;
  background: ${(p) =>
    p.$active ? 'rgba(32, 220, 142, 0.18)' : 'rgba(255, 255, 255, 0.06)'};
  color: ${(p) =>
    p.$active ? p.theme.colors.morMain : 'rgba(255, 255, 255, 0.55)'};
`;

const Body = styled.div`
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding: 1.4rem 2.4rem 2rem;

  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.12) transparent;
  &::-webkit-scrollbar {
    width: 6px;
  }
  &::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.12);
    border-radius: 3px;
  }
`;

const Section = styled.section`
  & + & {
    margin-top: 1.8rem;
  }
`;

const SectionLabel = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.05rem;
  font-weight: 500;
  letter-spacing: 0.4px;
  color: rgba(255, 255, 255, 0.4);
  margin-bottom: 0.8rem;
  padding-left: 0.2rem;
`;

const SectionList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 6px;
`;

const SectionHint = styled.span`
  color: rgba(255, 255, 255, 0.3);
  font-weight: 400;
  font-size: 0.95rem;
  letter-spacing: 0.2px;
  text-transform: none;
`;

const InfoToggle = styled.button`
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
  padding: 2px 6px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: rgba(173, 211, 255, 0.95);
  font-size: 1rem;
  font-weight: 500;
  letter-spacing: 0.2px;
  text-transform: none;
  cursor: pointer;

  &:hover {
    background: rgba(125, 188, 255, 0.12);
  }

  &:focus-visible {
    outline: 2px solid rgba(125, 188, 255, 0.5);
    outline-offset: 2px;
  }
`;

const InfoPanel = styled.div`
  margin-bottom: 0.9rem;
  padding: 1rem 1.2rem;
  border: 1px solid rgba(125, 188, 255, 0.25);
  border-radius: 8px;
  background: rgba(125, 188, 255, 0.08);
  color: rgba(214, 232, 255, 0.92);
  font-size: 1.2rem;
  line-height: 1.5;
`;

const EmptyState = styled.div`
  padding: 5rem 2rem;
  text-align: center;
  color: rgba(255, 255, 255, 0.45);
  font-size: 1.35rem;
  line-height: 1.5;

  svg {
    opacity: 0.4;
    margin-bottom: 1rem;
  }
`;

const LoadingState = styled(EmptyState)`
  color: rgba(255, 255, 255, 0.72);

  svg {
    animation: ${spin} 0.7s linear infinite;
    color: ${(p) => p.theme.colors.morMain};
    opacity: 0.9;
  }
`;

type FilterId =
  | 'all'
  | 'llm'
  | 'cowork'
  | 'vision'
  | 'embeddings'
  | 'tts'
  | 'stt'
  | 'local'
  | 'tee';

const FILTERS: { id: FilterId; label: string; modality?: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'llm', label: 'LLM', modality: 'llm' },
  { id: 'cowork', label: 'Workspace candidates' },
  { id: 'vision', label: 'Vision' },
  { id: 'embeddings', label: 'Embeddings', modality: 'embeddings' },
  { id: 'tts', label: 'Text-to-Speech', modality: 'tts' },
  { id: 'stt', label: 'Speech-to-Text', modality: 'stt' },
  { id: 'tee', label: 'Secure' },
  { id: 'local', label: 'Local' },
];

const isTee = (m: any) => isSecureModel(m);
const isVision = (m: any) => getVisionCapability(m) !== 'none';
const isCoworkCandidateModel = (m: any) => isCoworkCandidate(m);

function hasModality(tags: any[] = [], modality: string) {
  return tags.some((t: any) => String(t).toLowerCase() === modality);
}

function matchesQuery(model: any, q: string) {
  const needle = q.trim().toLowerCase();
  if (!needle) return true;
  if ((model.Name || '').toLowerCase().includes(needle)) return true;
  return (model.Tags || []).some((t: any) =>
    String(t).toLowerCase().includes(needle),
  );
}

const ModelSelectionModal = ({
  isActive,
  handleClose,
  models,
  onChangeModel,
  symbol,
  modelsLoading = false,
  marketplaceOnly = false,
  coworkSetup = false,
}: any) => {
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<FilterId>('all');
  const [showTeeInfo, setShowTeeInfo] = useState(false);

  // NB: hooks must run on every render — keep `useMemo` BEFORE the
  // `isActive` early-return, otherwise the hook count changes between
  // renders and React throws "Rendered more hooks than during the previous
  // render".
  const enriched = useMemo(
    () =>
      (models || [])
        .filter((m: any) => !marketplaceOnly || !m.isLocal)
        .filter((m: any) => !coworkSetup || isCoworkCandidate(m)),
    [coworkSetup, marketplaceOnly, models],
  );

  // Count results per filter (using the current search query) so the pills
  // can show live counts and disabled-look for empty filters.
  const counts: Record<FilterId, number> = useMemo(() => {
    const c: Record<FilterId, number> = {
      all: 0,
      llm: 0,
      cowork: 0,
      vision: 0,
      embeddings: 0,
      tts: 0,
      stt: 0,
      tee: 0,
      local: 0,
    };
    for (const m of enriched) {
      if (!matchesQuery(m, search)) continue;
      c.all++;
      if (m.isLocal) c.local++;
      if (isTee(m)) c.tee++;
      if (isVision(m)) c.vision++;
      if (isCoworkCandidateModel(m)) c.cowork++;
      const tags = m.Tags || [];
      if (hasModality(tags, 'llm') || hasModality(tags, 'chat')) c.llm++;
      if (hasModality(tags, 'embeddings') || hasModality(tags, 'embedding'))
        c.embeddings++;
      if (hasModality(tags, 'tts')) c.tts++;
      if (hasModality(tags, 'stt')) c.stt++;
    }
    return c;
  }, [enriched, search]);

  const visible = useMemo(() => {
    const filtered = enriched.filter((m: any) => {
      if (!matchesQuery(m, search)) return false;
      const tags = m.Tags || [];
      switch (filter) {
        case 'all':
          return true;
        case 'local':
          return !!m.isLocal;
        case 'tee':
          return isTee(m);
        case 'vision':
          return isVision(m);
        case 'cowork':
          return isCoworkCandidateModel(m);
        case 'llm':
          return hasModality(tags, 'llm') || hasModality(tags, 'chat');
        case 'embeddings':
          return (
            hasModality(tags, 'embeddings') || hasModality(tags, 'embedding')
          );
        case 'tts':
          return hasModality(tags, 'tts');
        case 'stt':
          return hasModality(tags, 'stt');
      }
    });

    // Stable sort: online first, then local first inside online group,
    // then alphabetical by name.
    return filtered.sort((a: any, b: any) => {
      if (!!b.isOnline !== !!a.isOnline) return b.isOnline ? 1 : -1;
      if (!!b.isLocal !== !!a.isLocal) return b.isLocal ? 1 : -1;
      return (a.Name || '').localeCompare(b.Name || '');
    });
  }, [enriched, search, filter]);

  // Bail out *after* all hooks have run.
  if (!isActive) return null;

  const resetAndClose = () => {
    setSearch('');
    setFilter('all');
    setShowTeeInfo(false);
    handleClose();
  };

  const handlePick = (data: any) => {
    onChangeModel(data);
    resetAndClose();
  };

  // A provider-declared capability and a name-family heuristic are materially
  // different confidence levels. Keep them in separate buckets so users never
  // have to infer which models actually advertise image input.
  const declaredVisionModels = visible.filter(
    (m: any) => getVisionCapability(m) === 'declared',
  );
  const possibleVisionModels = visible.filter(
    (m: any) => getVisionCapability(m) === 'detected',
  );
  const localModels = visible.filter(
    (m: any) => m.isLocal && getVisionCapability(m) === 'none',
  );
  const teeModels = visible.filter(
    (m: any) => !m.isLocal && getVisionCapability(m) === 'none' && isTee(m),
  );
  const remoteModels = visible.filter(
    (m: any) => !m.isLocal && getVisionCapability(m) === 'none' && !isTee(m),
  );

  const filterIconFor = (id: FilterId) => {
    switch (id) {
      case 'llm':
        return <IconMessage size={13} stroke={2} />;
      case 'vision':
        return <IconEye size={13} stroke={2} />;
      case 'cowork':
        return <IconSparkles size={13} stroke={2} />;
      case 'embeddings':
        return <IconVector size={13} stroke={2} />;
      case 'tts':
        return <IconHeadphones size={13} stroke={2} />;
      case 'stt':
        return <IconMicrophone size={13} stroke={2} />;
      case 'tee':
        return <IconShieldLock size={13} stroke={2} />;
      case 'local':
        return <IconHome size={13} stroke={2} />;
      default:
        return null;
    }
  };

  return (
    <Modal onClose={resetAndClose} bodyProps={bodyProps}>
      <Layout>
        <Header>
          <TitleRow>
            <Title>
              {coworkSetup ? 'Choose a text model for Workspace' : 'New chat'}
            </Title>
            {/* Only surface the counter when filtering/search actually hides
                models — otherwise "N of N" is noise. */}
            {visible.length !== enriched.length && (
              <ResultCount role="status" aria-live="polite">
                {visible.length} of {enriched.length}{' '}
                {enriched.length === 1 ? 'model' : 'models'}
              </ResultCount>
            )}
          </TitleRow>
          <SearchWrapper>
            <InputGroup>
              <InputGroup.Text>
                <IconSearch size={18} />
              </InputGroup.Text>
              <Form.Control
                type="text"
                aria-label="Search models"
                placeholder="Search models or tags…"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                autoFocus
                style={{
                  background: 'transparent',
                  color: 'rgba(255, 255, 255, 0.95)',
                  border: 'none',
                  boxShadow: 'none',
                  outline: 'none',
                  fontSize: '1.35rem',
                }}
              />
            </InputGroup>
          </SearchWrapper>
          <FilterRow role="group" aria-label="Filter models by capability">
            {FILTERS.filter(
              (item) =>
                (!marketplaceOnly || item.id !== 'local') &&
                (!coworkSetup || item.id !== 'cowork'),
            ).map((f) => {
              const active = filter === f.id;
              const count = counts[f.id];
              return (
                <FilterPill
                  key={f.id}
                  $active={active}
                  type="button"
                  onClick={() => setFilter(f.id)}
                  aria-pressed={active}
                  aria-label={
                    f.id === 'cowork'
                      ? `Show Workspace candidate models (${count})`
                      : `Show ${f.label} models (${count})`
                  }
                  title={
                    f.id === 'cowork'
                      ? 'Marketplace text/chat models that can be tried with Workspace. Provider tool support is not published and is checked only after a session opens.'
                      : undefined
                  }
                >
                  {filterIconFor(f.id)}
                  {f.label}
                  <FilterCount $active={active} aria-hidden="true">
                    {count}
                  </FilterCount>
                </FilterPill>
              );
            })}
          </FilterRow>
          {filter === 'cowork' && (
            <CoworkCandidateHint role="note">
              <IconInfoCircle size={15} stroke={2} aria-hidden="true" />
              <span>
                Candidates are not verified. Providers do not publish tool
                support; Workspace checks it after a session opens and uses
                compatibility mode when possible.
              </span>
            </CoworkCandidateHint>
          )}
        </Header>

        <Body aria-busy={modelsLoading}>
          {modelsLoading ? (
            <LoadingState role="status" aria-live="polite">
              <IconLoader2 size={32} stroke={1.8} aria-hidden="true" />
              <div>Loading models from your node…</div>
            </LoadingState>
          ) : visible.length === 0 ? (
            <EmptyState>
              <IconWorld size={36} stroke={1.5} />
              <div>
                {search.trim()
                  ? 'No models match your search.'
                  : filter === 'cowork'
                    ? 'No marketplace text/chat models are available to try with Workspace.'
                    : filter === 'all'
                      ? 'No models are currently available from this node.'
                      : 'No models available for this filter.'}
              </div>
            </EmptyState>
          ) : null}

          {localModels.length > 0 && (
            <Section>
              <SectionLabel>
                <IconHome size={13} stroke={2} />
                Local
              </SectionLabel>
              <SectionList>
                {localModels.map((m: any) => (
                  <ModelRow
                    key={m.Id}
                    model={m}
                    symbol={symbol}
                    onChangeModel={handlePick}
                  />
                ))}
              </SectionList>
            </Section>
          )}

          {teeModels.length > 0 && (
            <Section>
              <SectionLabel>
                <IconShieldLock size={13} stroke={2} />
                Secure&nbsp;
                <SectionHint>(Trusted Execution Environment)</SectionHint>
                <InfoToggle
                  type="button"
                  aria-expanded={showTeeInfo}
                  onClick={() => setShowTeeInfo((v) => !v)}
                >
                  <IconInfoCircle size={13} stroke={2} />
                  What is this?
                </InfoToggle>
              </SectionLabel>
              {showTeeInfo && <InfoPanel>{SECURE_MODE_INFO}</InfoPanel>}
              <SectionList>
                {teeModels.map((m: any) => (
                  <ModelRow
                    key={m.Id}
                    model={m}
                    symbol={symbol}
                    onChangeModel={handlePick}
                  />
                ))}
              </SectionList>
            </Section>
          )}

          {declaredVisionModels.length > 0 && (
            <Section>
              <SectionLabel>
                <IconEye size={13} stroke={2} />
                Vision (declared)
                <SectionHint>
                  (image input advertised by model metadata)
                </SectionHint>
              </SectionLabel>
              <SectionList>
                {declaredVisionModels.map((m: any) => (
                  <ModelRow
                    key={m.Id}
                    model={m}
                    symbol={symbol}
                    onChangeModel={handlePick}
                  />
                ))}
              </SectionList>
            </Section>
          )}

          {possibleVisionModels.length > 0 && (
            <Section>
              <SectionLabel>
                <IconEye size={13} stroke={2} />
                Possible vision (name match)
                <SectionHint>
                  (recognised family; provider did not declare it)
                </SectionHint>
              </SectionLabel>
              <SectionList>
                {possibleVisionModels.map((m: any) => (
                  <ModelRow
                    key={m.Id}
                    model={m}
                    symbol={symbol}
                    onChangeModel={handlePick}
                  />
                ))}
              </SectionList>
            </Section>
          )}

          {remoteModels.length > 0 && (
            <Section>
              <SectionLabel>
                <IconWorld size={13} stroke={2} />
                Marketplace
              </SectionLabel>
              <SectionList>
                {remoteModels.map((m: any) => (
                  <ModelRow
                    key={m.Id}
                    model={m}
                    symbol={symbol}
                    onChangeModel={handlePick}
                  />
                ))}
              </SectionList>
            </Section>
          )}
        </Body>
      </Layout>
    </Modal>
  );
};

export default ModelSelectionModal;
