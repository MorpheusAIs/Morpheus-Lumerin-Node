import { useEffect, useState } from "react";
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

  const [duration, setDuration] = useState<number>(5);
  const [morStake, setMorStake] = useState<number>(0);

  useEffect(() => {
    setMorStake(calculateStake(5));
  }, [pricePerSecond])

  const calculateStake = (value) => {
    if(!pricePerSecond) {
      return 0;
    }
    const totalCost = pricePerSecond * value * 60;
    const stake = totalCost * +supply / +budget;
    return stake;
  }

  const isDisabled = +duration < 5 || +duration > 1440;

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
            value={duration?.toString()}
            onChange={(e) => {
              const value = Number(e.target.value);
              setDuration(value);
              setMorStake(calculateStake(value));
            }}
            type="number"
            min={5}
            max={1440}
            name="time"
            id="time"
          />
          <Sublabel>Session Length (min 5, max 1440)</Sublabel>
        </InputGroup>
      </Row>
      {
        duration && !isDisabled && morStake ? (
          <Row style={{ margin: "20px 0" }}>
            Funds to Stake: {(morStake / 10 ** 18).toFixed(2)} MOR
          </Row>
        ) : null
      }
      <InputGroup
        style={{
          textAlign: 'center',
          justifyContent: 'space-between',
          marginTop: '30px',
          height: '60px'
        }}
      >
        <Row style={{ justifyContent: 'center' }}>
          {/* <LeftBtn onClick={handleSaveDraft}>Save as Draft</LeftBtn> */}
          <RightBtn onClick={() => triggerOpen({ stake: morStake, duration })} disabled={isDisabled}>
            {"Open"}
          </RightBtn>
        </Row>
      </InputGroup>
    </Form>

  </Modal>)
}

export default OpenSessionModal;