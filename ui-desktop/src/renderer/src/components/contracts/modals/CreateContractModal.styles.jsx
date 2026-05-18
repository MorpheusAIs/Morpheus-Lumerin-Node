import React, { useState } from 'react';
import styled from 'styled-components';
import { BaseBtn } from '../../common';
import { IconX } from '@tabler/icons-react';

const CloseButton = styled.button`
  position: absolute;
  top: 12px;
  right: 12px;
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  padding: 0;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.7);
  transition: background 0.12s ease, color 0.12s ease;
  z-index: 21;

  &:hover {
    background: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 1);
  }

  &:focus-visible {
    outline: 2px solid rgba(32, 220, 142, 0.5);
    outline-offset: 2px;
  }
`;

export const CloseModal = onClose => (
  <CloseButton type="button" aria-label="Close" onClick={onClose}>
    <IconX size={18} stroke={2} />
  </CloseButton>
);

export const Modal = styled.div`
  display: flex;
  flex-direction: column;
  position: fixed;
  z-index: 10;
  left: 0;
  top: 0;
  width: 100%;
  min-width: 330px;
  height: 100%;
  overflow: auto;
  background-color: rgb(0, 0, 0);
  background-color: rgba(0, 0, 0, 0.4);
  align-items: center;
  justify-content: center;
  color: ${p => p.theme.colors.primaryDark};
`;

export const Body = styled.div`
  position: fixed;
  z-index: 20;
  background-color: ${p => p.theme.colors.light};
  width: ${p => p.width || '45%'};
  height: ${p => p.height || 'fit-content'};
  border-radius: 5px;
  padding: 3rem 5%;
  max-width: ${p => p.maxWidth || '600px'};
  max-height: ${p => p.maxHeight || '800px'};
  background-color: #173629;
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.04);

  @media (min-height: 700px) {
    padding: 5rem;
  }
`;

export const TitleWrapper = styled.div`
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 10%;
  margin-bottom: 10px;
`;
export const Title = styled.div`
  display: block;
  line-height: 2.4rem;
  font-size: 2.5rem;
  font-weight: 900;
  cursor: default;
  margin-bottom: 30px;
`;

export const Subtitle = styled.div`
  display: block;
  font-size: 1.5rem;
  font-weight: 400;
  margin-bottom: 10px;
`;

export const ContractLink = styled.div`
  line-height: 1.4rem;
  font-size: 1.4rem;
  font-weight: 100;
  cursor: pointer;
  display: flex;
  align-items: center;
  color: #014353;
`;

export const InstructionLink = styled.div`
  padding-top: 5px;
  line-height: 14px;
  font-size: 14px;
  font-weight: 100;
  cursor: pointer;
  display: flex;
  align-items: center;
  color: #014353;
`;

export const Form = styled.form`
  display: flex;
  height: 85%;
  margin: 1rem 0 0;
  flex-direction: column;
  justify-content: space-between;
`;

export const InputGroup = styled.div`
  margin: 1.5rem 0 0;
  display: flex;
  flex-direction: column;
  height: 100%;
  flex: 1;
`;

export const Row = styled.div`
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  width: 100%;
`;

export const Input = styled.input`
  padding: 4px 8px;
  outline: 0;
  border: 0px;
  background: #eaf7fc;
  border-radius: 5px;
  padding: 1.5rem 1.5rem;
  margin-top: 0.25rem;

  ::placeholder {
    color: rgba(1, 67, 83, 0.56);
  }

  ${props =>
    props.id === 'price' &&
    `
      display: inline-block;
      width: 50%;
    `};
`;

export const Select = styled.select`
  margin: 0.4rem 0 0.2rem 0;
  outline: 0;
  border: 0px;
  background: #eaf7fc;
  border-radius: 15px;
  padding: 1.5rem 1.5rem;
  margin-top: 0.25rem;
  color: rgba(1, 67, 83, 0.56);
`;

export const Label = styled.label`
  line-height: 1.4rem;
  font-size: 1.2rem;
  font-weight: 400;
  cursor: default;
`;

export const Sublabel = styled.label`
  line-height: 1.4rem;
  font-size: 1.1rem;
  font-weight: 400;
  opacity: 0.65;
  cursor: default;
  padding: 5px 0 0 5px;
`;

export const SublabelGreen = styled(Sublabel)`
  font-weight: 800;
`;

export const LeftBtn = styled(BaseBtn)`
  width: 45%;
  height: 40px;
  font-size: 1.5rem;
  border-radius: 5px;
  border: 1px solid ${p => p.theme.colors.primary};
  background-color: ${p => p.theme.colors.morMain};
  color: black;

  @media (min-width: 1040px) {
    margin-left: 0;
  }
`;

export const RightBtn = styled(BaseBtn)`
  width: 45%;
  height: 40px;
  font-size: 1.5rem;
  border-radius: 5px;
  background-color: ${p => p.theme.colors.morMain};
  color: black;
  font-weight: 600;

  @media (min-width: 1040px) {
    margin-left: 0;
  }
`;

export const ErrorLabel = styled(Sublabel)`
  padding: 5px 0 0 5px;
  color: red;
`;

export const ApplyBtn = styled(RightBtn)`
  width: 15%;
  height: 15%;
  font-size: 1.2rem;
  border-radius: 15px;
  margin-left: 10px;
  background-color: ${p => p.theme.colors.primary};
  color: ${p => p.theme.colors.light};

  @media (min-width: 1040px) {
    margin-left: 10px;
  }
`;

export const ProfitLabel = styled.div`
  cursor: pointer;
  text-align: center;
  margin-top: 1.5rem;
  padding: 1rem 0;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 0.5rem;
  border: 1px solid rgba(0, 0, 0, 0.125);
`;

export const ProfitMessageLabel = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  line-height: 1.4rem;
  font-size: 1.1rem;
  font-weight: bold;
  opacity: 0.65;
  cursor: default;
  padding: 0 1rem;
  border-bottom: ${p => (p.show ? '1px solid rgba(0,0,0,.125)' : '')};
  padding-bottom: ${p => (p.show ? '1rem' : '')};
`;

export const SearchContainer = styled.div`
    input {
      background-color: transparent;
      border: 1px solid ${p => p.theme.colors.morMain};
      color: white;
    }
    input::placeholder {
      color: white;
    }
    .form-control:focus {
      color: white;
      background-color: transparent;
      border-color: ${p => p.theme.colors.morMain};
    }
    .input-group-text {
      background-color: transparent;
      border: 1px solid ${p => p.theme.colors.morMain};
      color: white;
    }
`