import { useState, useEffect } from 'react';
import { withRouter } from 'react-router-dom';
import withProvidersState from "../../store/hocs/withProvidersState";
import styled from 'styled-components';
import Accordion from 'react-bootstrap/Accordion';

import Card from 'react-bootstrap/Card';
import { abbreviateAddress } from '../../utils';
import { Btn } from '../../components/common'
import Table from 'react-bootstrap/Table';
import Button from 'react-bootstrap/Button';
import './Providers.css'

const ClaimBtn = styled(Btn)`
  background-color: ${p => p.theme.colors.morMain};
  color: black;
  font-weight: 600;
  border-radius: 5px;
`;

const CustomCard = styled(Card)`
  background: #244a47!important;
  color: #21dc8f!important;
  border: 0.5px solid!important;
  cursor: pointer!important;

  p {
    color: white!important;
  }

  .gap-20 {
    gap: 20px!important;
  }
`

const Container = styled.div`
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  gap: 24px;
`

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


function renderTable({ onClaim, sessions }) {
  return (
      <BidTable striped bordered hover size="sm">
          <thead>
              <tr>
                  <th>Session</th>
                  <th>Bid</th>
                  <th>Status</th>
                  <th>Balance</th>
                  <th></th>
              </tr>
          </thead>
          <tbody>
              {sessions?.length
                  ? sessions.map(b => {
                      // const provider = providers?.find(x => x.Address.toLowerCase() === b.Provider.toLowerCase());
                      return (
                      <tr key={b.Id}>
                          <td>{abbreviateAddress(b.Id, 5)}</td>
                          <td>{abbreviateAddress(b.BidID, 5)}</td>
                          <td>CLOSED</td>
                          <td>4 MOR</td>
                          <td><StartBtn>Claim</StartBtn></td>
                      </tr>)
                  }) : null}
          </tbody>
      </BidTable>
  );
}


function AllCollapseExample() {
  return (
    <Accordion>
      <Accordion.Item eventKey="0">
        <Accordion.Header className='model-header'>Llama 3.0</Accordion.Header>
        <Accordion.Body>
          {renderTable({ onClaim: null,     "sessions": [
        {
            "Id": "0xf7294224df37b4f3b72e5cd00e5c8ccf4a6f4f2987563f07ff63f63a039ed433",
            "User": "0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc",
            "Provider": "0x70997970c51812dc3a010c7d01b50e0d17dc79c8",
            "ModelAgentId": "0x21731a519f4467e105a19b27db37e96bb915bd866e93140874a0d3206404f4e9",
            "BidID": "0xfab7ad9b96cf4ced3a025b6ea3605f17343c8f9630bdf25fbfd51137525ce099",
            "Stake": 38092532819383690000000,
            "PricePerSecond": 100000000000000,
            "CloseoutReceipt": "",
            "CloseoutType": 0,
            "ProviderWithdrawnAmount": 0,
            "OpenedAt": 1716879674,
            "EndsAt": 1716879981,
            "ClosedAt": 0
        }
    ]})}
        </Accordion.Body>
      </Accordion.Item>
    </Accordion>
  );
}

function ProvidersList({
  getAllModels,
  history,
  setSelectedModel
} : any) {
  const [models, setModels] = useState<any[]>([]);
  console.log("🚀 ~ models:", models)

  useEffect(() => {
    getAllModels().then(data => {
      setModels(data.filter(d => !d.IsDeleted));
    });
  }, [])

  // Accordiaon
  // with table
  // Claim each

  // filter models by User
  // Show accordion
  // Get Claimable balance

//   {

// }
  
  return (<div>
      {AllCollapseExample()}
    </div>)
}

export default withRouter(withProvidersState(ProvidersList));
