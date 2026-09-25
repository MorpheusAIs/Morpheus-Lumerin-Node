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

  it('gives the user a labelled way back out of the picker', () => {
    const handleClose = vi.fn();
    render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal
          isActive
          handleClose={handleClose}
          onChangeModel={vi.fn()}
          symbol="MOR"
          models={[marketplaceModel({ Id: 'chat', Name: 'Chat Model' })]}
        />
      </ThemeProvider>,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Back' }));
    expect(handleClose).toHaveBeenCalledOnce();
  });

  it('remembers where the user was instead of resetting on every close', () => {
    const props = {
      handleClose: vi.fn(),
      onChangeModel: vi.fn(),
      symbol: 'MOR',
      models: [
        marketplaceModel({ Id: 'chat', Name: 'Chat Model' }),
        marketplaceModel({
          Id: 'speech',
          Name: 'Speech Model',
          ModelType: 'tts',
          Tags: ['tts'],
        }),
      ],
    };
    const { rerender } = render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal {...props} isActive />
      </ThemeProvider>,
    );

    fireEvent.click(
      screen.getByRole('button', { name: 'Show Text-to-Speech models (1)' }),
    );
    fireEvent.change(screen.getByRole('textbox', { name: 'Search models' }), {
      target: { value: 'Speech' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Speech Model' }));

    rerender(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal {...props} isActive={false} />
      </ThemeProvider>,
    );
    rerender(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal {...props} isActive />
      </ThemeProvider>,
    );

    expect(
      screen.getByRole('textbox', { name: 'Search models' }),
    ).toHaveValue('Speech');
    expect(
      screen
        .getByRole('button', { name: 'Show Text-to-Speech models (1)' })
        .getAttribute('aria-pressed'),
    ).toBe('true');
  });

  it('drops a filter that the other mode does not even show', () => {
    const props = {
      handleClose: vi.fn(),
      onChangeModel: vi.fn(),
      symbol: 'MOR',
      models: [
        marketplaceModel({ Id: 'chat', Name: 'Chat Model' }),
        marketplaceModel({
          Id: 'speech',
          Name: 'Speech Model',
          ModelType: 'tts',
          Tags: ['tts'],
        }),
      ],
    };
    const { rerender } = render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal {...props} isActive />
      </ThemeProvider>,
    );

    fireEvent.click(
      screen.getByRole('button', { name: /Workspace candidate/ }),
    );

    // Workspace setup hides the candidate pill entirely. Carrying the filter
    // over would leave no visible control explaining the shortened list.
    rerender(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal {...props} isActive coworkSetup />
      </ThemeProvider>,
    );

    expect(
      screen.getByRole('button', { name: /Show All models/ }).getAttribute(
        'aria-pressed',
      ),
    ).toBe('true');
  });

  it('keeps the catalog usable when model tags have inconsistent shapes', () => {
    render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal
          isActive
          handleClose={vi.fn()}
          onChangeModel={vi.fn()}
          symbol="MOR"
          models={[
            marketplaceModel({
              Id: 'null-tags',
              Name: 'Null tags model',
              Tags: null,
            }),
            marketplaceModel({
              Id: 'string-tags',
              Name: 'String tags model',
              Tags: 'llm, tee',
            }),
            marketplaceModel({
              Id: 'object-tags',
              Name: 'Object tags model',
              Tags: { unexpected: 'shape' },
            }),
            marketplaceModel({
              Id: 'mixed-tags',
              Name: 'Mixed tags model',
              Tags: ['llm', null, { nested: true }, 'vision'],
            }),
            null,
            'not-a-model',
          ]}
        />
      </ThemeProvider>,
    );

    expect(screen.getByText('Null tags model')).toBeTruthy();
    expect(screen.getByText('Object tags model')).toBeTruthy();
    expect(
      screen.getByRole('button', { name: 'Show All models (4)' }),
    ).toBeTruthy();

    fireEvent.click(
      screen.getByRole('button', { name: 'Show Secure models (1)' }),
    );
    expect(screen.getByText('String tags model')).toBeTruthy();
    expect(screen.queryByText('Null tags model')).toBeNull();

    fireEvent.click(
      screen.getByRole('button', { name: 'Show All models (4)' }),
    );
    fireEvent.change(screen.getByRole('textbox', { name: 'Search models' }), {
      target: { value: 'vision' },
    });

    expect(screen.getByText('Mixed tags model')).toBeTruthy();
    expect(screen.queryByText('Object tags model')).toBeNull();
  });
});

describe('ModelSelectionModal price sort', () => {
  const priced = (modelId: string, minWei: string, maxWei = minWei) => ({
    model_id: modelId,
    min_price_per_second_wei: minWei,
    max_price_per_second_wei: maxWei,
    bid_count: 1,
  });

  // Four models covering every price state the sort has to order: two with real
  // prices, one the sweep answered for but found no provider on, and one the
  // sweep never reached.
  const sortModels = [
    marketplaceModel({ Id: 'expensive', Name: 'Expensive model' }),
    marketplaceModel({ Id: 'cheap', Name: 'Cheap model' }),
    marketplaceModel({ Id: 'unpriced', Name: 'Unpriced model' }),
    marketplaceModel({ Id: 'unswept', Name: 'Unswept model' }),
  ];

  const priceIndex = {
    byModelId: new Map([
      // Deliberately past Number.MAX_SAFE_INTEGER: a float comparison would
      // call these two equal and leave the list in registry order.
      ['cheap', priced('cheap', '9007199254740993')],
      ['expensive', priced('expensive', '9007199254740994')],
      ['unpriced', priced('unpriced', '')],
    ]),
    failedModelIds: new Set(['unswept']),
  };

  const renderPicker = (extra: Record<string, unknown> = {}) =>
    render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal
          isActive
          handleClose={vi.fn()}
          onChangeModel={vi.fn()}
          symbol="MOR"
          models={sortModels}
          priceIndex={priceIndex}
          {...extra}
        />
      </ThemeProvider>,
    );

  /** The picker's rows, in the order they appear in the document. */
  const rowOrder = () =>
    screen
      .getAllByRole('button')
      .map((button) => button.textContent || '')
      .filter((label) => label.endsWith(' model'));

  it('orders by price in both directions and sinks unpriced models in each', () => {
    renderPicker();

    fireEvent.click(screen.getByRole('button', { name: /Cheapest first/ }));
    expect(rowOrder()).toEqual([
      'Cheap model',
      'Expensive model',
      // Neither of these has a price. They are not free, so they must not lead
      // the cheap list, and they are not dear, so they must not lead the other
      // one either. Alphabetical between themselves keeps the order stable.
      'Unpriced model',
      'Unswept model',
    ]);

    fireEvent.click(screen.getByRole('button', { name: /Most expensive first/ }));
    expect(rowOrder()).toEqual([
      'Expensive model',
      'Cheap model',
      'Unpriced model',
      'Unswept model',
    ]);
  });

  it('flattens the capability sections into one ordered list', () => {
    renderPicker();

    // Recommended keeps the capability grouping, which is what makes a price
    // order impossible: five sections means five separate cheapest-first lists.
    expect(
      screen.queryByText('(models with no live provider are listed last)'),
    ).toBeNull();
    expect(screen.getByText('Marketplace')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: /Cheapest first/ }));

    expect(screen.queryByText('Marketplace')).toBeNull();
    expect(
      screen.getByText('(models with no live provider are listed last)'),
    ).toBeTruthy();
  });

  it('says which of the three price states the list is in', () => {
    const { rerender } = renderPicker({ pricesLoading: true });

    // Nothing is claimed about prices until the list is actually ordered by
    // them.
    expect(screen.queryByRole('status')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: /Cheapest first/ }));
    expect(screen.getByRole('status').textContent).toContain('Checking prices');

    const props = {
      isActive: true,
      handleClose: vi.fn(),
      onChangeModel: vi.fn(),
      symbol: 'MOR',
      models: sortModels,
      priceIndex,
    };
    rerender(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal {...props} pricesFailed />
      </ThemeProvider>,
    );
    expect(screen.getByRole('status').textContent).toContain(
      'could not be read',
    );

    rerender(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal {...props} />
      </ThemeProvider>,
    );
    expect(screen.getByText('Price per second of compute.')).toBeTruthy();
  });

  it('still orders the list when no prices arrived at all', () => {
    render(
      <ThemeProvider theme={theme}>
        <ModelSelectionModal
          isActive
          handleClose={vi.fn()}
          onChangeModel={vi.fn()}
          symbol="MOR"
          models={sortModels}
          pricesFailed
        />
      </ThemeProvider>,
    );

    fireEvent.click(screen.getByRole('button', { name: /Cheapest first/ }));

    // Every model is unpriced, so the price comparison ties throughout and the
    // name tie-break decides. A sort that threw or shuffled here would make a
    // failed price read look like a broken picker.
    expect(rowOrder()).toEqual([
      'Cheap model',
      'Expensive model',
      'Unpriced model',
      'Unswept model',
    ]);
  });
});
