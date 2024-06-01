import { useState, useEffect } from 'react';
import { withRouter } from 'react-router-dom';
import withModelsState from "../../store/hocs/withModelsState";
import styled from 'styled-components';

import Card from 'react-bootstrap/Card';
import { abbreviateAddress } from '../../utils';
import { IconDownload } from '@tabler/icons-react';

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
        <Card.Title as={"div"} style={{ display: 'flex', justifyContent: "space-between"}}>
          {model.Name}
          <IconDownload></IconDownload>
        </Card.Title>
        <Card.Subtitle className="mb-2">{abbreviateAddress(model.Id, 6)}</Card.Subtitle>
        <Card.Text>
          {/* <div>
          Fee: {model.Fee}
          </div>
          <div>
          Stake: {model.Stake}
          </div> */}
        </Card.Text>
        <Card.Footer className='d-flex gap-20'>
          {
            model.Tags.map(t => (<div key={t}>{t}</div>)) 
          }
        </Card.Footer>
      </Card.Body>
    </CustomCard>
  );
}


function ModelsTable({
  getAllModels,
  history,
  setSelectedModel
} : any) {
  const [models, setModels] = useState<any[]>([{
    "Id": "0x0557d796a4490cb847efa225c610e56921e1aee2cefcd6e3577c5d470b5bbf80",
    "IpfsCID": "0x0000000000000000000000000000697066733a2f2f6970667361646472657373",
    "Fee": 100,
    "Stake": 100,
    "Owner": "0xb4b12a69fdbb70b31214d4d3c063752c186ff8de",
    "Name": "Llama 3.0",
    "Tags": [
        "llama",
        "llm",
        "chat"
    ],
    "Timestamp": 1715336698,
    "IsDeleted": false
}]);

  const onSelect = (id) => {
    console.log("selected", id);
    setSelectedModel(models.find(x => x.Id == id));
    history.push("/bids");
  }

  useEffect(() => {
    // getAllModels().then(data => {
    //   console.log("🚀 ~ getAllModels ~ data:", data)
    //   setModels(data.filter(d => !d.IsDeleted));
    // });
  }, [])

  return (<Container>
     {
      models.length ? models.map((x => (<div>{ModelCard({ onSelect, model: x})}</div>))) : null
     }
    </Container>)
}

export default withRouter(withModelsState(ModelsTable));
