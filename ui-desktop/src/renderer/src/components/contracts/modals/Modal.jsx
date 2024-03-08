import React, { useRef } from 'react';

import {
  Modal as ModalBase,
  Body,
  CloseModal
} from './CreateContractModal.styles';

function Modal({ children, onClose }) {
  const waitingForMouseUpRef = useRef(false);
  const ignoreBackdropClickRef = useRef(false);
  const modalRef = useRef(false);
  const handleDialogMouseDown = () => {
    waitingForMouseUpRef.current = true;
  };
  const handleMouseUp = e => {
    if (waitingForMouseUpRef.current && e.target == modalRef.current) {
      ignoreBackdropClickRef.current = true;
    }
    waitingForMouseUpRef.current = false;
  };

  const wrapClose = (e, force) => {
    if (
      (!force && ignoreBackdropClickRef.current) ||
      e.target !== e.currentTarget
    ) {
      ignoreBackdropClickRef.current = false;
      return;
    }
    onClose();
  };

  return (
    <ModalBase onClick={wrapClose} onMouseUp={handleMouseUp} ref={modalRef}>
      <Body
        onClick={e => e.stopPropagation()}
        onMouseDown={handleDialogMouseDown}
      >
        {CloseModal(e => wrapClose(e, true))}
        {children}
      </Body>
    </ModalBase>
  );
}

export default Modal;
