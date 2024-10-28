import styled from 'styled-components';

// import 'react-tabs/style/react-tabs.css';
import './styles.css';
import { Btn, BaseBtn, TextInput } from '../common';

export const Sublabel = styled.label`
  line-height: 1.4rem;
  font-size: 1.1rem;
  font-weight: 400;
  opacity: 0.65;
  cursor: default;
  padding: 5px 0 0 5px;
`;

export const ErrorLabel = styled(Sublabel)`
  padding: 5px 0 0 5px;
  color: red;
`;

export const StyledBtn = styled(Btn)`
  width: 40%;
  height: 40px;
  font-size: 1.5rem;
  border-radius: 5px;
  padding: 0 0.6rem;
  background-color: ${p => p.theme.colors.morMain};
  color: black;

  @media (min-width: 1040px) {
    width: 35%;
    height: 40px;
    margin-left: 0;
    margin-top: 1.6rem;
  }
`;

export const Subtitle = styled.h3`
  color: ${p => p.theme.colors.dark};
`;

export const StyledParagraph = styled.p`
  color: ${p => p.theme.colors.dark};

  span {
    font-weight: bold;
  }
`;

export const Input = styled(TextInput)`
  outline: 0;
  border: 0px;
  background: #eaf7fc;
  border-radius: 15px;
  padding: 1.2rem 1.2rem;
  margin-top: 0.25rem;
`;
