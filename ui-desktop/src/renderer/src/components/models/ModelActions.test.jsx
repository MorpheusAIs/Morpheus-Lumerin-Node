import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import ModelsTable from './ModelsTable';
import PinnedFilesTable from './PinnedFilesTable';
import theme from '../../ui/theme';

describe('model card keyboard actions', () => {
  it('copies and downloads from native buttons without selecting the model card', async () => {
    const user = userEvent.setup();
    const select = vi.fn();
    const openFolder = vi.fn().mockResolvedValue({ canceled: true });
    render(
      <ThemeProvider theme={theme}>
        <ModelsTable
          models={[
            { Id: 'model-id', Name: 'Test model', IpfsCID: 'model-cid' },
          ]}
          setSelectedModel={select}
          openSelectDownloadFolder={openFolder}
          toasts={{ toast: vi.fn() }}
          client={{}}
        />
      </ThemeProvider>,
    );

    screen
      .getByRole('button', { name: 'Copy model ID for Test model' })
      .focus();
    await user.keyboard('{Enter}');
    expect(window.copyToClipboard).toHaveBeenCalledWith('model-id');
    screen.getByRole('button', { name: 'Download Test model' }).focus();
    await user.keyboard(' ');
    expect(openFolder).toHaveBeenCalledTimes(1);
    expect(select).not.toHaveBeenCalled();
  });

  it('makes pinned file identifiers keyboard copyable and does not claim unpin success early', async () => {
    const user = userEvent.setup();
    const toast = vi.fn();
    const unpinFile = vi.fn();
    render(
      <ThemeProvider theme={theme}>
        <PinnedFilesTable
          pinnedFiles={[
            {
              fileName: 'model.gguf',
              fileCIDHash: 'file-hash',
              metadataCIDHash: 'meta-hash',
              metadataCID: 'meta-cid',
              id: 'model-id',
            },
          ]}
          unpinFile={unpinFile}
          toasts={{ toast }}
        />
      </ThemeProvider>,
    );

    screen.getByRole('button', { name: 'Copy CID for model.gguf' }).focus();
    await user.keyboard('{Enter}');
    expect(window.copyToClipboard).toHaveBeenCalledWith('meta-cid');
    toast.mockClear();
    screen.getByRole('button', { name: 'Unpin model.gguf' }).focus();
    await user.keyboard(' ');
    expect(unpinFile).toHaveBeenCalledWith('file-hash');
    expect(unpinFile).toHaveBeenCalledWith('meta-hash');
    expect(toast).not.toHaveBeenCalled();
  });
});
