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
    ErrorLabel,
    ApplyBtn,
    ProfitMessageLabel,
    ProfitLabel
  } from '../../contracts/modals/CreateContractModal.styles';

const OpenSessionModal = ({ isActive, handleClose, budget, supply, pricePerSecond }) => {

    const [duration, setDuration] = useState<number | undefined>(undefined);
    const [morStake, setMorStake] = useState<number | undefined>(undefined);

    console.log("🚀 ~ OpenSessionModal ~ duration:", duration)
    

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
                <Label htmlFor="time">Session Duration</Label>
                <Input
                  placeholder="# of minutes"
                  value={duration}
                  onChange={(e) => {
                    const value = Number(e.target.value);
                    const totalCost = pricePerSecond * value * 60;
                    const stake = totalCost * supply / budget;
                    console.log("STAKE", stake)
                    setMorStake(stake);
                    setDuration(value);
                  } }
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
                duration && (
                    <Row style={{ margin: "20px 0"}}>
                        MOR To Stake: 5 MOR
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
                <RightBtn type="submit">
                  {"Open"}
                </RightBtn>
              </Row>
            </InputGroup>
          </Form>
        
    </Modal>)
}

export default OpenSessionModal;