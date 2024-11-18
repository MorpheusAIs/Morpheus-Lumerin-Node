//import Modal from '../../contracts/modals/Modal';
import { List as RVList, AutoSizer } from 'react-virtualized';
import Modal from '../../contracts/modals/Modal';
import styled from 'styled-components';
import {
    TitleWrapper,
    Title,
    SearchContainer
} from '../../contracts/modals/CreateContractModal.styles';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import { IconSearch } from '@tabler/icons-react';

import ModelRow from './ModelRow';
import { useState } from 'react';

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
    height: '550px',
    width: '70%',
    maxWidth: '100%',
    onClick: e => e.stopPropagation()
}
const RVContainer = styled(RVList)`
 .ReactVirtualized__Grid__innerScrollContainer {
   overflow: visible !important;
  }`

const ModelSelectionModal = ({ isActive, handleClose, models, onChangeModel }) => {
    const [search, setSearch] = useState<string | undefined>();

    if (!isActive) {
        return <></>;
    }

    const changeModelHandler = (data) => {
        onChangeModel(data);
        handleClose();
    }

    const sortedModels = models
        .map(m => ({ ...m, isOnline: m.isLocal || m.bids.some(b => b.ProviderData?.availabilityStatus != "disconnected") }))
        .sort((a, b) => b.isOnline - a.isOnline);

    const filterdModels = search ? sortedModels.filter(m => m.Name.toLowerCase().includes(search.toLowerCase())) : sortedModels;

    return (
        <Modal 
            onClose={() => {
                setSearch(undefined);
                handleClose();
            }}
            bodyProps={bodyProps}
        >
            <TitleWrapper>
                <Title>Select Model To Create Chat</Title>
            </TitleWrapper>
            <SearchContainer>
                <InputGroup style={{ marginBottom: '15px' }}>
                    <InputGroup.Text>
                        <IconSearch />
                    </InputGroup.Text>
                    <Form.Control
                        type="text"
                        placeholder="Search..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </InputGroup>
            </SearchContainer>
            { filterdModels.length == 0 && <div>No models found</div> }
            <AutoSizer width={400} height={385}>
                {({ width }) => (
                    <RVContainer
                        rowRenderer={rowRenderer(filterdModels, changeModelHandler)}
                        rowHeight={45}
                        rowCount={filterdModels.length}
                        height={385} // defaults for tests
                        width={width || 500} // defaults for tests
                    />
                )}
            </AutoSizer>
        </Modal>
    );
}

export default ModelSelectionModal;