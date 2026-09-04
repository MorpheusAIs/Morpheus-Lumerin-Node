import { fireEvent, render, screen } from '@testing-library/react';
import type { FC, PropsWithChildren, ReactNode } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import theme from '../../../ui/theme';
import ModelSelectionModal from './ModelSelectionModal';

vi.mock('../../contracts/modals/Modal', () => ({
  default: ({ children }: { children: ReactNode }) => (
    <div role="dialog">{children}</div>
  ),
}));

vi.mock('./ModelRow', () => ({
  default: ({
    model,
    onChangeModel,
  }: {
    model: { Id: string; Name: string };
    onChangeModel: (value: { modelId: string }) => void;
  }) => (
    <button type="button" onClick={() => onChangeModel({ modelId: model.Id })}>
      {model.Name}
    </button>
  ),
}));

const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

const marketplaceModel = (overrides: Record<string, unknown>) => ({
  Id: 'model-id',
  Name: 'Model',
  ModelType: 'llm',
  Tags: ['llm'],
  bids: [{ Id: 'bid-id', PricePerSecond: '1000000000000000' }],
  ...overrides,
});

describe('ModelSelectionModal Workspace filter', () => {
  it('lets a raw marketplace model start its one-model price check', () => {
    const onChangeModel = vi.fn();
    const handleClose = vi.fn();

    render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal
          isActive
          handleClose={handleClose}
          onChangeModel={onChangeModel}
          symbol="MOR"
          models={[
            marketplaceModel({
              Id: 'raw-model',
              Name: 'Unchecked model',
              bids: undefined,
            }),
          ]}
        />
      </ThemeProvider>,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Unchecked model' }));

    expect(onChangeModel).toHaveBeenCalledOnce();
    expect(onChangeModel).toHaveBeenCalledWith({ modelId: 'raw-model' });
    expect(handleClose).toHaveBeenCalledOnce();
  });

  it('distinguishes a catalog still loading from a genuinely empty catalog', () => {
    const props = {
      isActive: true,
      handleClose: vi.fn(),
      onChangeModel: vi.fn(),
      symbol: 'MOR',
      models: undefined,
    };
    const { rerender } = render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal {...props} modelsLoading />
      </ThemeProvider>,
    );

    expect(screen.getByRole('status').textContent).toContain(
      'Loading models from your node',
    );
    expect(
      screen.queryByText('No models are currently available from this node.'),
    ).toBeNull();

    rerender(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal {...props} models={[]} modelsLoading={false} />
      </ThemeProvider>,
    );

    expect(
      screen.getByText('No models are currently available from this node.'),
    ).toBeTruthy();
  });

  it('shows only marketplace chat models as unverified Workspace candidates', () => {
    render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal
          isActive
          handleClose={vi.fn()}
          onChangeModel={vi.fn()}
          symbol="MOR"
          models={[
            marketplaceModel({ Id: 'chat', Name: 'Chat Model' }),
            marketplaceModel({
              Id: 'legacy-chat',
              Name: 'Legacy Chat Model',
              ModelType: 'UNKNOWN',
              Tags: ['chat'],
            }),
            marketplaceModel({
              Id: 'speech',
              Name: 'Speech Model',
              ModelType: 'tts',
              Tags: ['tts'],
            }),
            marketplaceModel({
              Id: 'embedding',
              Name: 'Embedding Model',
              ModelType: 'embedding',
              Tags: ['embedding'],
            }),
            marketplaceModel({
              Id: 'local',
              Name: 'Local Chat Model',
              isLocal: true,
            }),
            marketplaceModel({
              Id: 'deleted',
              Name: 'Deleted Chat Model',
              IsDeleted: true,
            }),
          ]}
        />
      </ThemeProvider>,
    );

    const coworkFilter = screen.getByRole('button', {
      name: 'Show Workspace candidate models (2)',
    });
    expect(coworkFilter.getAttribute('aria-pressed')).toBe('false');

    fireEvent.click(coworkFilter);

    expect(coworkFilter.getAttribute('aria-pressed')).toBe('true');
    expect(screen.getByRole('note').textContent).toContain(
      'Candidates are not verified',
    );
    expect(screen.getByText('Chat Model')).toBeTruthy();
    expect(screen.getByText('Legacy Chat Model')).toBeTruthy();
    expect(screen.queryByText('Speech Model')).toBeNull();
    expect(screen.queryByText('Embedding Model')).toBeNull();
    expect(screen.queryByText('Local Chat Model')).toBeNull();
    expect(screen.queryByText('Deleted Chat Model')).toBeNull();

    fireEvent.change(screen.getByRole('textbox', { name: 'Search models' }), {
      target: { value: 'Legacy' },
    });

    expect(
      screen.getByRole('button', {
        name: 'Show Workspace candidate models (1)',
      }),
    ).toBeTruthy();
    expect(screen.queryByText('Chat Model')).toBeNull();
    expect(screen.getByText('Legacy Chat Model')).toBeTruthy();
  });

  it('keeps Workspace setup gated and hides the redundant filter', () => {
    render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal
          isActive
          coworkSetup
          handleClose={vi.fn()}
          onChangeModel={vi.fn()}
          symbol="MOR"
          models={[
            marketplaceModel({ Id: 'chat', Name: 'Chat Model' }),
            marketplaceModel({
              Id: 'speech',
              Name: 'Speech Model',
              ModelType: 'tts',
              Tags: ['tts'],
            }),
            marketplaceModel({
              Id: 'local',
              Name: 'Local Chat Model',
              isLocal: true,
            }),
            marketplaceModel({
              Id: 'deleted',
              Name: 'Deleted Chat Model',
              IsDeleted: true,
            }),
          ]}
        />
      </ThemeProvider>,
    );

    expect(
      screen.queryByRole('button', { name: /Workspace candidate/ }),
    ).toBeNull();
    expect(screen.getByText('Chat Model')).toBeTruthy();
    expect(screen.queryByText('Speech Model')).toBeNull();
    expect(screen.queryByText('Local Chat Model')).toBeNull();
    expect(screen.queryByText('Deleted Chat Model')).toBeNull();
  });
});
