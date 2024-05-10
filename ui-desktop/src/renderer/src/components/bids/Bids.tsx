import React, { useEffect, useState } from "react";
import { withRouter } from 'react-router-dom';
import withBidsState from '../../store/hocs/withBidsState';

import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import Table from 'react-bootstrap/Table';
import Button from 'react-bootstrap/Button';
import { abbreviateAddress } from '../../utils';
import styled from 'styled-components';

const BidTable = styled(Table)`
    text-align: center!important;
    border: 0.5px solid#21dc8f!important;

    th {
        background: #244a47!important;
        color: #21dc8f!important;
    }

    td {
        background: #244a47!important;
        color: #21dc8f!important;
        padding: 12px 0!important
    }
`

const StartBtn = styled(Button)`    
    background: rgba(0,0,0, 0.9)!important;
    border-radius: 0!important;
    border: 1px solid #21dc8f!important;
`

function renderTable({ onStart, bids, providers }) {
    return (
        <BidTable striped bordered hover size="sm">
            <thead>
                <tr>
                    <th>Bid</th>
                    <th>Provider</th>
                    <th>Provider Endpoint</th>
                    <th>Price Per Second</th>
                    <th></th>
                </tr>
            </thead>
            <tbody>
                {bids?.length
                    ? bids.map(b => {
                        const provider = providers?.find(x => x.Address.toLowerCase() === b.Provider.toLowerCase());
                        return (
                        <tr key={b.Id}>
                            <td>{abbreviateAddress(b.Id, 5)}</td>
                            <td>{abbreviateAddress(b.Provider, 5)}</td>
                            <td>{provider?.Endpoint}</td>
                            <td>{b.PricePerSecond / 10 ** 18} MOR</td>
                            <td><StartBtn onClick={() => onStart(b.Id, provider)}>Start</StartBtn></td>
                        </tr>)
                    }) : null}
            </tbody>
        </BidTable>
    );
}

const Bids = ({ history, getProviders, selectedModel, getBitsByModels, setBid }) => {
    const [bids, setBids] = useState([]);
    const [providers, setProviders] = useState([]);

    useEffect(() => {
        if (!selectedModel) {
            history.push("/models");
            return;
        }

        getProviders().then(data => {
            const providersMap = data.filter(d => !d.IsDeleted)
            setProviders(providersMap);
        })

        getBitsByModels(selectedModel.Id).then(data => {
            setBids(data.filter(d => !d.DeletedAt));
        })
    }, [])

    const onStart = (bidId, provider) => {
        setBid({ bidId, provider }).then((isSuccess) => {
            if(!isSuccess) {
                return;
            }
            history.push("/chat");
        })
    }

    return (
        <View data-testid="bids-container">
            <LayoutHeader title="Bids" />
            <div>{renderTable({ onStart, bids, providers })}</div>
        </View>)
}

export default withRouter(withBidsState((Bids)));