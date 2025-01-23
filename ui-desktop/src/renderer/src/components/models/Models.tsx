import { useState } from 'react';
import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import { BtnAccent } from '../dashboard/BalanceBlock.styles';
import ModelsTable from './ModelsTable';
import Tab from 'react-bootstrap/Tab';
import Tabs from 'react-bootstrap/Tabs';
import styled from 'styled-components'
import FileSelectionModal from './FileSelectionModal';


const Container = styled.div`
    overflow-y: auto;
    
    .nav-link {
        color: ${p => p.theme.colors.morMain}
    }

    .nav-link.active {
        color: ${p => p.theme.colors.morMain}
        border-color: ${p => p.theme.colors.morMain}
        background-color: rgba(0,0,0,0.4);
    }
`

const Models = () => {

    const [openChangeModal, setOpenChangeModal] = useState(false);
    
    return (
        <View data-testid="models-container">
            <LayoutHeader title="Models">
                <BtnAccent style={{ padding: '1.5rem' }} onClick={() => setOpenChangeModal(true)}>Pin Model</BtnAccent>
            </LayoutHeader>
            <Container>
                <Tabs
                    defaultActiveKey="registry"
                    id="tab-models"
                    className="mb-3"
                >
                    <Tab eventKey="registry" title="Registry">
                        <ModelsTable />
                    </Tab>
                    <Tab eventKey="pinned" title="Pinned Models">

                    </Tab>
                </Tabs>

            </Container>
            <FileSelectionModal
                isActive={openChangeModal}
                handleClose={() => setOpenChangeModal(false)} />
        </View>)

}

export default Models;