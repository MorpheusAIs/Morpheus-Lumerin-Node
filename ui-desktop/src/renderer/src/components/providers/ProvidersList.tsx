import { useState, useEffect } from 'react';
import { withRouter } from 'react-router-dom';
import withModelsState from "../../store/hocs/withModelsState";
import styled from 'styled-components';

import Card from 'react-bootstrap/Card';
import { abbreviateAddress } from '../../utils';
import { Btn } from '../../components/common'

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

function ModelCard({ onSelect, model }) {
  return (
    <CustomCard style={{ width: '36rem' }} onClick={() => onSelect(model.Id)}>
      <Card.Body>
        <Card.Title as={"h2"}>{model.Name}</Card.Title>
        <Card.Subtitle className="mb-2">{abbreviateAddress(model.Id, 6)}</Card.Subtitle>
        <Card.Text>
          <div>
          Fee: {model.Fee}
          </div>
          <div>
          Stake: {model.Stake}
          </div>
        </Card.Text>
        <Card.Footer>
            <ClaimBtn block>Claim</ClaimBtn>
        </Card.Footer>
      </Card.Body>
    </CustomCard>
  );
}


function ProvidersList({
  getAllModels,
  history,
  setSelectedModel
} : any) {
  const [models, setModels] = useState<any[]>([]);

  const onSelect = (id) => {
    console.log("selected", id);
    setSelectedModel(models.find(x => x.Id == id));
    history.push("/bids");
  }

  useEffect(() => {
    getAllModels().then(data => {
      setModels(data.filter(d => !d.IsDeleted));
    });
  }, [])

  return (<Container>
     {
      models.length ? models.map((x => (<div>{ModelCard({ onSelect, model: x})}</div>))) : null
     }
    </Container>)
}

export default withRouter(withModelsState(ProvidersList));
