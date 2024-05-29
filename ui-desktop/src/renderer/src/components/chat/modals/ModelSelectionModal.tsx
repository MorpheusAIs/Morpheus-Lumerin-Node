//import Modal from '../../contracts/modals/Modal';
import React, { useEffect, useState } from 'react';
import { List as RVList, AutoSizer } from 'react-virtualized';
import Modal from '../../contracts/modals/Modal';
import styled from 'styled-components';
import {
    TitleWrapper,
    Title,
    Subtitle,
    Form,
    InputGroup,
    Row,
    Input,
    Label,
    Sublabel,
    // Modal,
    Body,
    CloseModal
} from '../../contracts/modals/CreateContractModal.styles';

import ModelRow from './ModelRow';

const rowRenderer = (models, onChangeModel) => ({ key, index, style }) => (
    <ModelRow
        onChangeModel={onChangeModel}
        key={models[index].Id}
        model={models[index]}
    />
);

const bodyProps = {
    height: '500px',
    width: '70%',
    maxWidth: '100%',
    onClick: e => e.stopPropagation()
}
const RVContainer = styled(RVList)`
 .ReactVirtualized__Grid__innerScrollContainer {
   overflow: visible !important;
  }`

const ModelSelectionModal = ({ isActive, handleClose, models, onChangeModel }) => {

    if (!isActive) {
        return <></>;
    }

    return (
        <Modal onClose={handleClose} bodyProps={bodyProps}
        >
            <TitleWrapper>
                <Title>Change Model</Title>
            </TitleWrapper>
            <AutoSizer width={400} height={500}>
                {({ width, height }) => (
                    <RVContainer
                        rowRenderer={rowRenderer(models, (id) => {
                            onChangeModel(id);
                            handleClose();
                        }
                        )}
                        rowHeight={100}
                        rowCount={models.length}
                        height={height || 500} // defaults for tests
                        width={width || 500} // defaults for tests
                    />
                )}
            </AutoSizer>
        </Modal>
    );
}

export default ModelSelectionModal;