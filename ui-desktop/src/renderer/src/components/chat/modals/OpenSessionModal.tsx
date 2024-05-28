import React, { useState } from "react";
import Modal from '../../contracts/modals/Modal';
import {
  TitleWrapper,
  Title,
  Subtitle,
  Form,
  InputGroup,
  Row,
  Input,
  Label,
  Sublabel,
  RightBtn,
} from '../../contracts/modals/CreateContractModal.styles';

const OpenSessionModal = ({ isActive, handleClose, budget, supply, pricePerSecond, triggerOpen }) => {
console.log("🚀 ~ OpenSessionModal ~ budget, supply, pricePerSecond,:", budget, supply, pricePerSecond,)

  const [duration, setDuration] = useState<number | undefined>(undefined);
  const [morStake, setMorStake] = useState<number | undefined>(undefined);


  if (!isActive) {
    return <></>;
  }
  return (<Modal onClose={handleClose}>
    <TitleWrapper>
      <Title>Open Session</Title>
      <Subtitle>MOR stake amount changes based on duration</Subtitle>
    </TitleWrapper>
    <Form>
      <Row>
        <InputGroup>
          <Label htmlFor="time">Session Duration (in minutes)</Label>
          <Input
            placeholder="# of minutes"
            value={duration}
            onChange={(e) => {
              const value = Number(e.target.value);
              const totalCost = (pricePerSecond * 10 ** 18) * value * 60;
              const stake = totalCost * supply / budget;
              setMorStake(stake);
              setDuration(value);
            }}
            type="number"
            min={5}
            max={1440}
            name="time"
            id="time"
          />
          <Sublabel>Session Length (min 5, max 1440)</Sublabel>
          {/* {formState?.errors?.time?.type === 'required' && (
                  <ErrorLabel>Duration is required</ErrorLabel>
                )}
                {formState?.errors?.time?.type === 'min' && (
                  <ErrorLabel>{'Minimum 24 hours'}</ErrorLabel>
                )}
                {formState?.errors?.time?.type === 'max' && (
                  <ErrorLabel>{'Maximum 48 hours'}</ErrorLabel>
                )} */}
        </InputGroup>
      </Row>
      {
        duration && morStake && (
          <Row style={{ margin: "20px 0" }}>
            Funds to Stake: {(morStake / 10 ** 18).toFixed(2)} MOR
          </Row>
        )
      }
      <InputGroup
        style={{
          textAlign: 'center',
          justifyContent: 'space-between',
          height: '60px'
        }}
      >
        <Row style={{ justifyContent: 'center' }}>
          {/* <LeftBtn onClick={handleSaveDraft}>Save as Draft</LeftBtn> */}
          <RightBtn onClick={() => triggerOpen({ stake: morStake, duration })}>
            {"Open"}
          </RightBtn>
        </Row>
      </InputGroup>
    </Form>

  </Modal>)
}

export default OpenSessionModal;