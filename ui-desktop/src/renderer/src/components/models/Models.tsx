import { useState, useEffect } from 'react';
import withModelsState from "../../store/hocs/withModelsState";

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

const IpfsStatus = styled.div`
    color: ${p => p.theme.colors.morMain};
    font-size: 1.2rem;
`

const Models = ({
    setSelectedModel,
    getIpfsVersion,
    getAllModels,
    openSelectDonwloadFolder,
    downloadModelFromIpfs,
    addFileToIpfs,
    pinFile,
    unpinFile,
    toasts,
}: any) => {

    const [openChangeModal, setOpenChangeModal] = useState(false);
    const [ipfsVersion, setIpfsVersion] = useState(null);
    const [isIpfsConnected, setIsIpfsConnected] = useState(false);
    const [models, setModels] = useState([]);

    useEffect(() => {
        getIpfsVersion().then((response) => {
            if (response?.version) {
                setIpfsVersion(response?.version);
                setIsIpfsConnected(true);
            } else {
                setIsIpfsConnected(false);
            }
        }).catch((error) => {
            console.error("Error", error);
            setIsIpfsConnected(false);
        });

        getAllModels().then((response) => {
            setModels(response);
        }).catch((error) => {
            console.error("Error", error);
        });
    }, []);

    return (
        <View data-testid="models-container">
                        {isIpfsConnected ? (
                    <IpfsStatus>
                        <span>IPFS Connected. Version: {ipfsVersion}</span>
                    </IpfsStatus>
                ) : (
                    <IpfsStatus>
                        <span>IPFS is not connected</span>
                    </IpfsStatus>
                )}
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
                        <ModelsTable setSelectedModel={setSelectedModel} models={models} openSelectDonwloadFolder={openSelectDonwloadFolder} downloadModelFromIpfs={downloadModelFromIpfs} toasts={toasts} />
                    </Tab>
                    <Tab eventKey="pinned" title="Pinned Models">

                    </Tab>
                </Tabs>

            </Container>
            <FileSelectionModal
                isActive={openChangeModal}
                addFileToIpfs={addFileToIpfs}
                pinFile={pinFile}
                unpinFile={unpinFile}
                toasts={toasts}
                handleClose={() => setOpenChangeModal(false)}
                 />
        </View>)

}

export default withModelsState(Models);