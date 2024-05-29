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
                          <td>{b.ClosedAt ? "CLOSED" : "OPEN"}</td>
                          <td>{b.Balance / 10 ** 18} MOR</td>
                          <td><StartBtn onClick={() => onClaim(b.Id)}>Claim</StartBtn></td>
                      </tr>)
                  }) : null}
          </tbody>
      </BidTable>
  );
}


function ProvidersList({ data, claimFunds }) {
  
  return (<div>
      {data?.modelsNames && Object.keys(data?.modelsNames).map(model => {
        const modelSessions = data.results.filter(r => r.ModelAgentId.toLowerCase() == model.toLowerCase());
        
        return (
          <Accordion>
          <Accordion.Item eventKey="0">
            <Accordion.Header className='model-header'>{data?.modelsNames[model]}</Accordion.Header>
            <Accordion.Body>
              {renderTable({ onClaim: claimFunds, sessions: modelSessions})}
            </Accordion.Body>
          </Accordion.Item>
        </Accordion>
        )
      })}
    </div>)
}

export default withRouter(withProvidersState(ProvidersList));
