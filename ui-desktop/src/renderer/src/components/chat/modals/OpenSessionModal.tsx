import React from "react";
import Modal from '../../contracts/modals/Modal';

const OpenSessionModal = ({ isActive, handleClose }) => {

    if (!isActive) {
        return <></>;
      }
    return (<Modal onClose={handleClose}>
        Open Session
    </Modal>)
}

export default OpenSessionModal;