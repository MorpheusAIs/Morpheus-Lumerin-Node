import React, { useLayoutEffect, useRef } from 'react';
import { createPortal } from 'react-dom';

import {
  Modal as ModalBase,
  Body,
  CloseModal,
} from './CreateContractModal.styles';

const openDialogs = [];
let originalBodyOverflow = '';

const focusableElements = (dialog) =>
  Array.from(
    dialog.querySelectorAll(
      'a[href], button, input, select, textarea, [tabindex], [contenteditable="true"]',
    ),
  ).filter((element) => {
    const style = window.getComputedStyle(element);
    return (
      element.tabIndex >= 0 &&
      !element.matches(':disabled, input[type="hidden"]') &&
      !element.closest('[hidden], [inert], [aria-hidden="true"]') &&
      style.display !== 'none' &&
      style.visibility !== 'hidden'
    );
  });

function Modal({ children, onClose, bodyProps, ariaLabel = 'Dialog' }) {
  const waitingForMouseUpRef = useRef(false);
  const ignoreBackdropClickRef = useRef(false);
  const modalRef = useRef(null);
  const dialogRef = useRef(null);
  const returnFocusRef = useRef(document.activeElement);
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  useLayoutEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (!openDialogs.length) {
      originalBodyOverflow = document.body.style.overflow;
      document.body.style.overflow = 'hidden';
    }
    openDialogs.push(dialog);
    const isTopDialog = () => openDialogs[openDialogs.length - 1] === dialog;

    const focusFirst = () => {
      const fields = focusableElements(dialog);
      const input = fields.find((field) =>
        field.matches('input, textarea, select'),
      );
      (input || fields[0] || dialog).focus({ preventScroll: true });
    };

    // Preserve an explicit autoFocus already applied by the child form.
    if (!dialog.contains(document.activeElement)) focusFirst();

    const onKeyDown = (event) => {
      if (!isTopDialog()) return;
      if (event.key === 'Escape' && !event.isComposing) {
        event.preventDefault();
        event.stopPropagation();
        onCloseRef.current();
        return;
      }
      if (event.key !== 'Tab') return;
      const fields = focusableElements(dialog);
      const first = fields[0];
      const last = fields[fields.length - 1];
      const active = document.activeElement;
      if (!first) {
        event.preventDefault();
        dialog.focus();
      } else if (
        event.shiftKey &&
        (active === first || !fields.includes(active))
      ) {
        event.preventDefault();
        last.focus();
      } else if (
        !event.shiftKey &&
        (active === last || !fields.includes(active))
      ) {
        event.preventDefault();
        first.focus();
      }
    };

    const onFocusIn = (event) => {
      if (isTopDialog() && !dialog.contains(event.target)) focusFirst();
    };

    document.addEventListener('keydown', onKeyDown, true);
    document.addEventListener('focusin', onFocusIn);
    return () => {
      const wasTopDialog = isTopDialog();
      document.removeEventListener('keydown', onKeyDown, true);
      document.removeEventListener('focusin', onFocusIn);
      openDialogs.splice(openDialogs.indexOf(dialog), 1);
      if (!openDialogs.length)
        document.body.style.overflow = originalBodyOverflow;
      if (wasTopDialog && returnFocusRef.current?.isConnected) {
        returnFocusRef.current.focus({ preventScroll: true });
      }
    };
  }, []);

  const handleDialogMouseDown = () => {
    waitingForMouseUpRef.current = true;
  };
  const handleMouseUp = (e) => {
    if (waitingForMouseUpRef.current && e.target == modalRef.current) {
      ignoreBackdropClickRef.current = true;
    }
    waitingForMouseUpRef.current = false;
  };

  const wrapClose = (e, force) => {
    // `force` is the explicit close button. It must always close, regardless
    // of the backdrop guards below (the click target is the inner X icon, not
    // the button itself, so the `e.target !== e.currentTarget` check would
    // otherwise swallow it).
    if (force) {
      ignoreBackdropClickRef.current = false;
      onClose();
      return;
    }
    if (ignoreBackdropClickRef.current || e.target !== e.currentTarget) {
      ignoreBackdropClickRef.current = false;
      return;
    }
    onClose();
  };

  return createPortal(
    <ModalBase onClick={wrapClose} onMouseUp={handleMouseUp} ref={modalRef}>
      <Body
        {...bodyProps}
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel}
        tabIndex={-1}
        onClick={(event) => {
          event.stopPropagation();
          bodyProps?.onClick?.(event);
        }}
        onMouseDown={(event) => {
          handleDialogMouseDown();
          bodyProps?.onMouseDown?.(event);
        }}
      >
        {CloseModal((e) => wrapClose(e, true))}
        {children}
      </Body>
    </ModalBase>,
    document.body,
  );
}

export default Modal;
