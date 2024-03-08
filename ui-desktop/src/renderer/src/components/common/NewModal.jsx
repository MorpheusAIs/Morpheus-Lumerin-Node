import React, { useState } from 'react';
import styled from 'styled-components';

import withNewContractModalState from '../../store/hocs/withNewContractModalState';

import { BaseBtn } from '.';

const Modal = styled.div`
  display: flex;
  flex-direction: column;
  position: fixed;
  z-index: 10;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  overflow: auto;
  background-color: rgb(0, 0, 0);
  background-color: rgba(0, 0, 0, 0.4);
  align-items: center;
  justify-content: center;
`;

const Body = styled.div`
  position: fixed;
  z-index: 20;
  background-color: ${p => p.theme.colors.light};
  width: 50%;
  height: 80%;
  border-radius: 5px;
  padding: 3rem 5%;

  @media (min-height: 700px) {
    padding: 6.4rem 1.6rem;
  }
`;

const TitleWrapper = styled.div`
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 10%;
`;
const Title = styled.div`
  display: block;
  line-height: 2.4rem;
  font-size: 2rem;
  font-weight: 900;
  color: ${p => p.theme.colors.dark};
  cursor: default;
`;

const Subtitle = styled.div`
  display: block;
  line-height: 1.4rem;
  font-size: 1rem;
  font-weight: 400;
  color: ${p => p.theme.colors.dark};
  cursor: default;
`;

const Form = styled.form`
  display: flex;
  height: 85%;
  margin: 1rem 0 0;
  flex-direction: column;
  justify-content: space-between;
  color: ${p => p.theme.colors.dark};
`;

const InputGroup = styled.div`
  margin: 2rem 0 0;
  display: flex;
  flex-direction: column;
  height: 100%;
`;

const Row = styled.div`
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  width: 100%;
`;

const Input = styled.input`
  height: 40%;
  width: 80%;
  padding: 4px 8px;
  margin: 0.4rem 0 0.2rem 0;
  border-radius: 3px;
  border-style: solid;
  border-color: ${p => p.theme.colors.lightBG};
  border-width: 1px;
  ::placeholder {
    color: ${p => p.theme.colors.lightBG};
  }
  ::focus {
    outline: none !important;
  }
`;

const Label = styled.label`
  line-height: 1.4rem;
  font-size: 1.2rem;
  font-weight: 900;
  color: ${p => p.theme.colors.dark};
  cursor: default;
`;

const Sublabel = styled.label`
  line-height: 1.4rem;
  font-size: 0.8rem;
  font-weight: 400;
  color: ${p => p.theme.colors.dark};
  cursor: default;
  margin-bottom: 0.4rem;
`;

const SublabelGreen = styled(Sublabel)`
  color: ${p => p.theme.colors.primary};
  font-weight: 800;
`;

const LeftBtn = styled(BaseBtn)`
  width: 45%;
  height: 40px;
  font-size: 1rem;
  border-radius: 5px;
  border: 1px solid ${p => p.theme.colors.primary};
  background-color: ${p => p.theme.colors.light};
  color: ${p => p.theme.colors.primary};

  @media (min-width: 1040px) {
    margin-left: 0;
  }
`;

const RightBtn = styled(BaseBtn)`
  width: 45%;
  height: 40px;
  font-size: 1rem;
  border-radius: 5px;
  background-color: ${p => p.theme.colors.primary};
  color: ${p => p.theme.colors.light};

  @media (min-width: 1040px) {
    margin-left: 0;
  }
`;

function NewContractModal({ isActive, save, deploy, close }) {
  const [inputs, setInputs] = useState({
    address: '',
    time: '',
    date: '',
    price: ''
  });

  const handleInputs = e => {
    e.preventDefault();

    setInputs({ ...inputs, [e.target.name]: e.target.value });
  };

  const handleDeploy = e => {
    e.preventDefault();
    deploy(e);
  };

  const handleSaveDraft = e => {
    e.preventDefault();
    save(e);
  };

  const handleClose = e => {
    close(e);
  };
  const handlePropagation = e => e.stopPropagation();

  if (!isActive) {
    return <></>;
  }

  return (
    <Modal onClick={handleClose}>
      <Body onClick={handlePropagation}>
        <TitleWrapper>
          <Title>Create new contract</Title>
          <Subtitle>Sell your hashpower to the Lumerin Marketplace</Subtitle>
        </TitleWrapper>
        <Form onSubmit={handleDeploy}>
          <Row>
            <InputGroup>
              <Label htmlFor="address">Ethereum Address</Label>
              <Input
                value={inputs.address}
                onChange={handleInputs}
                placeholder="0x0c34..."
                style={{ width: '100%' }}
                type="text"
                name="address"
                id="address"
              />
              <Sublabel>
                Finds will be paid into this account once the contract is
                fulfilled.
              </Sublabel>
            </InputGroup>
          </Row>
          <Row style={{ maxWidth: '70%' }}>
            <InputGroup>
              <Label htmlFor="time">Time</Label>
              <Input
                value={inputs.time}
                onChange={handleInputs}
                placeholder="# of hours"
                type="text"
                name="time"
                id="time"
              />
              <Sublabel>Contract Length (min 1 hour)</Sublabel>
            </InputGroup>
            <InputGroup>
              <Label htmlFor="date">End Date</Label>
              <Input
                value={inputs.date}
                onChange={handleInputs}
                placeholder="12/05/21"
                type="text"
                name="date"
                id="date"
              />
              <Sublabel>Contract Expiry</Sublabel>
            </InputGroup>
          </Row>
          <Row>
            <InputGroup>
              <div>
                <Label htmlFor="price">List Price: </Label>
                <Sublabel>LMR Per TH/s</Sublabel>
              </div>
              <Input
                value={inputs.price}
                onChange={handleInputs}
                placeholder="cost per terahash"
                type="text"
                name="price"
                id="price"
              />
              <Sublabel>
                This is the price you will deploy your contract to the
                marketplace.
              </Sublabel>
            </InputGroup>
          </Row>
          <InputGroup
            style={{
              textAlign: 'center',
              justifyContent: 'space-between',
              height: '60px'
            }}
          >
            <SublabelGreen style={{}}>
              This is the price you will deploy your contract to the
              marketplace.
            </SublabelGreen>
            <Row>
              <LeftBtn onClick={handleSaveDraft}>Save as Draft</LeftBtn>
              <RightBtn type="submit">Create New Contract</RightBtn>
            </Row>
          </InputGroup>
        </Form>
      </Body>
    </Modal>
  );
}

export default withNewContractModalState(NewContractModal);
