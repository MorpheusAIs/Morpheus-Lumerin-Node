import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import FileSelectionModal from './FileSelectionModal';
import theme from '../../ui/theme';

vi.mock('../contracts/modals/Modal', () => ({
  default: ({ children, ariaLabel }) => (
    <div role="dialog" aria-label={ariaLabel}>
      {children}
    </div>
  ),
}));

const selectedFile = () => {
  const file = new File(['model data'], 'model.gguf');
  Object.defineProperty(file, 'path', { value: '/selected/model.gguf' });
  return file;
};
const response = { metadataCIDHash: 'metadata-hash', fileCIDHash: 'file-hash' };

function mountModal(overrides = {}) {
  const props = {
    isActive: true,
    handleClose: vi.fn(),
    addFileToIpfs: vi.fn().mockResolvedValue(response),
    pinFile: vi.fn().mockResolvedValue({ result: true }),
    toasts: { toast: vi.fn() },
    ...overrides,
  };
  const element = (next = {}) => (
    <ThemeProvider theme={theme}>
      <FileSelectionModal {...props} {...next} />
    </ThemeProvider>
  );
  const view = render(element());
  return { props, ...view, update: (next) => view.rerender(element(next)) };
}

describe('pin one model file', () => {
  it('keeps hooks stable across visibility changes and preserves draft fields', () => {
    const { update } = mountModal({ isActive: false });
    expect(screen.queryByRole('dialog')).toBeNull();
    update({ isActive: true });
    fireEvent.change(screen.getByLabelText('Model name (optional)'), {
      target: { value: 'My model' },
    });
    update({ isActive: false });
    update({ isActive: true });
    expect(screen.getByLabelText('Model name (optional)')).toHaveValue(
      'My model',
    );
  });

  it('clearly accepts one file and refuses an empty form submission', () => {
    const { props } = mountModal();
    expect(screen.getByLabelText('Model file')).not.toHaveAttribute('multiple');
    expect(
      screen.getByText('Choose one file per pin, such as a .gguf model file.'),
    ).toBeVisible();
    expect(
      fireEvent.submit(screen.getByRole('form', { name: 'Pin model file' })),
    ).toBe(false);
    expect(props.addFileToIpfs).not.toHaveBeenCalled();
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Choose one model file',
    );
  });

  it('clears an old ID error when the optional field is emptied', () => {
    mountModal();
    const input = screen.getByLabelText('Model ID (optional)');
    fireEvent.change(input, { target: { value: 'invalid' } });
    expect(screen.getByRole('alert')).toHaveTextContent('must start with');
    fireEvent.change(input, { target: { value: '' } });
    expect(screen.queryByRole('alert')).toBeNull();
    expect(input).not.toHaveClass('is-invalid');
  });

  it('guards duplicate clicks and submits the one file with trimmed metadata', async () => {
    let finishUpload;
    const addFileToIpfs = vi.fn(
      () =>
        new Promise((resolve) => {
          finishUpload = resolve;
        }),
    );
    const { props } = mountModal({ addFileToIpfs });
    fireEvent.change(screen.getByLabelText('Model name (optional)'), {
      target: { value: '  My model  ' },
    });
    fireEvent.change(screen.getByLabelText('Tags (optional)'), {
      target: { value: 'llm, , test ' },
    });
    fireEvent.change(screen.getByLabelText('Model file'), {
      target: { files: [selectedFile()] },
    });
    const button = screen.getByRole('button', { name: 'Pin model file' });
    fireEvent.click(button);
    fireEvent.click(button);
    fireEvent.submit(screen.getByRole('form', { name: 'Pin model file' }));

    expect(addFileToIpfs).toHaveBeenCalledTimes(1);
    expect(addFileToIpfs).toHaveBeenCalledWith(
      '/selected/model.gguf',
      '',
      'My model',
      ['llm', 'test'],
    );
    expect(
      screen.getByRole('button', { name: 'Pinning file…' }),
    ).toBeDisabled();
    expect(screen.getByLabelText('Model file')).toBeDisabled();

    finishUpload(response);
    await waitFor(() => expect(props.handleClose).toHaveBeenCalledTimes(1));
    expect(props.pinFile).toHaveBeenCalledWith('metadata-hash');
    expect(props.pinFile).toHaveBeenCalledWith('file-hash');
  });

  it('preserves the file and metadata after failure so retry can complete', async () => {
    const user = userEvent.setup();
    const addFileToIpfs = vi
      .fn()
      .mockRejectedValueOnce(new Error('IPFS disconnected'))
      .mockResolvedValue(response);
    const { props } = mountModal({ addFileToIpfs });
    await user.type(screen.getByLabelText('Model name (optional)'), 'My model');
    await user.upload(screen.getByLabelText('Model file'), selectedFile());
    await user.click(screen.getByRole('button', { name: 'Pin model file' }));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Check your IPFS connection and try again',
    );
    expect(screen.getByLabelText('Model name (optional)')).toHaveValue(
      'My model',
    );
    expect(screen.getByText('model.gguf')).toBeVisible();
    expect(props.handleClose).not.toHaveBeenCalled();
    expect(props.toasts.toast).not.toHaveBeenCalledWith(
      'success',
      expect.anything(),
    );

    await user.click(screen.getByRole('button', { name: 'Try pinning again' }));
    await waitFor(() => expect(props.handleClose).toHaveBeenCalledTimes(1));
    expect(addFileToIpfs).toHaveBeenCalledTimes(2);
  });

  it('keeps the dialog open if either pin is not confirmed', async () => {
    const { props } = mountModal({
      pinFile: vi
        .fn()
        .mockResolvedValueOnce({ result: true })
        .mockResolvedValueOnce({ result: false }),
    });
    fireEvent.change(screen.getByLabelText('Model file'), {
      target: { files: [selectedFile()] },
    });
    fireEvent.submit(screen.getByRole('form', { name: 'Pin model file' }));
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Could not pin this model',
    );
    expect(props.handleClose).not.toHaveBeenCalled();
  });
});
