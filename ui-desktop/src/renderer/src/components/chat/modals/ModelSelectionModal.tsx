//import Modal from '../../contracts/modals/Modal';
import { List as RVList, AutoSizer } from 'react-virtualized';
import Modal from '../../contracts/modals/Modal';
import styled from 'styled-components';
import {
    TitleWrapper,
    Title
} from '../../contracts/modals/CreateContractModal.styles';

import ModelRow from './ModelRow';

const rowRenderer = (models, onChangeModel) => ({ index, style }) => (
    <div style={style}>
        <ModelRow
            onChangeModel={onChangeModel}
            key={models[index].Id}
            model={models[index]}
        />
    </div>
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

    const changeModelHandler = (data) => {
        onChangeModel(data);
        handleClose();
    }

    return (
        <Modal onClose={handleClose} bodyProps={bodyProps}
        >
            <TitleWrapper>
                <Title>Change Model</Title>
            </TitleWrapper>
            <AutoSizer width={400}>
                {({ width, height }) => (
                    <RVContainer
                        rowRenderer={rowRenderer(models, changeModelHandler)}
                        rowHeight={75}
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