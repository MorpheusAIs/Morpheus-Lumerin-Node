import React, { useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import Modal from './Modal';
import theme from '../../../ui/theme';

function DialogExample({ onClose = () => {}, bodyProps }) {
  const [open, setOpen] = useState(false);
  return (
    <ThemeProvider theme={theme}>
      <>
        <button onClick={() => setOpen(true)}>Choose a model</button>
        <button>Background action</button>
        {open && (
          <Modal
            ariaLabel="Model selection"
            bodyProps={bodyProps}
            onClose={() => {
              onClose();
              setOpen(false);
            }}
          >
            <h2>Models</h2>
            <input aria-label="Search models" />
            <button disabled>Unavailable model</button>
            <button>Choose DeepSeek</button>
          </Modal>
        )}
      </>
    </ThemeProvider>
  );
}

describe('desktop custom dialog shell', () => {
  it('portals outside the route container and starts focus in its form', async () => {
    const user = userEvent.setup();
    const { container } = render(<DialogExample />);
    await user.click(screen.getByRole('button', { name: 'Choose a model' }));

    const dialog = screen.getByRole('dialog', { name: 'Model selection' });
    expect(container.contains(dialog)).toBe(false);
    expect(dialog).toHaveAttribute('aria-modal', 'true');
    expect(
      screen.getByRole('textbox', { name: 'Search models' }),
    ).toHaveFocus();
  });

  it('traps Tab and Shift+Tab without landing on disabled controls', async () => {
    const user = userEvent.setup();
    render(<DialogExample />);
    await user.click(screen.getByRole('button', { name: 'Choose a model' }));

    await user.tab({ shift: true });
    expect(screen.getByRole('button', { name: 'Close' })).toHaveFocus();
    await user.tab({ shift: true });
    expect(
      screen.getByRole('button', { name: 'Choose DeepSeek' }),
    ).toHaveFocus();
    await user.tab();
    expect(screen.getByRole('button', { name: 'Close' })).toHaveFocus();

    screen.getByRole('button', { name: 'Background action' }).focus();
    expect(
      screen.getByRole('textbox', { name: 'Search models' }),
    ).toHaveFocus();
  });

  it('closes on Escape and returns focus and scroll behavior to the opener', async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    const previousOverflow = document.body.style.overflow;
    render(<DialogExample onClose={onClose} />);
    const opener = screen.getByRole('button', { name: 'Choose a model' });
    await user.click(opener);
    expect(document.body.style.overflow).toBe('hidden');

    await user.keyboard('{Escape}');
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(opener).toHaveFocus();
    expect(document.body.style.overflow).toBe(previousOverflow);
  });

  it('does not dismiss when selecting text inside and releasing over the backdrop', async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    const onBodyClick = vi.fn();
    render(
      <DialogExample
        onClose={onClose}
        bodyProps={{ onClick: onBodyClick, style: { padding: 0 } }}
      />,
    );
    await user.click(screen.getByRole('button', { name: 'Choose a model' }));
    const dialog = screen.getByRole('dialog');
    const backdrop = dialog.parentElement;

    await user.click(screen.getByRole('heading', { name: 'Models' }));
    expect(onBodyClick).toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
    expect(dialog.style.padding).toBe('0px');

    fireEvent.mouseDown(screen.getByRole('textbox'));
    fireEvent.mouseUp(backdrop);
    fireEvent.click(backdrop);
    expect(onClose).not.toHaveBeenCalled();
    fireEvent.click(backdrop);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('always honors the explicit close button after a backdrop drag', async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    render(<DialogExample onClose={onClose} />);
    await user.click(screen.getByRole('button', { name: 'Choose a model' }));
    fireEvent.mouseDown(screen.getByRole('textbox'));
    fireEvent.mouseUp(screen.getByRole('dialog').parentElement);

    await user.click(screen.getByRole('button', { name: 'Close' }));
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('dialog')).toBeNull();
  });
});
