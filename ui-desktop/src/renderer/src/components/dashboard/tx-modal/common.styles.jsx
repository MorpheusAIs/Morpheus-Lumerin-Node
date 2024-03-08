import styled from 'styled-components';
import { BaseBtn } from '../../common';

export const HeaderWrapper = styled.div`
  display: flex;
  position: relative;
  height: 10%;
  align-content: center;
  align-items: center;
`;

export const Header = styled.div`
  font-size: 1.6rem;
  font-weight: bold;
  color: ${p => p.theme.colors.dark};
  text-align: center;
  width: 100%;
`;

export const BackBtn = styled(BaseBtn)`
  position: absolute;
  color: ${p => p.theme.colors.dark};
  font-weight: bold;
  margin: 8px 0 0 5px;
`;

export const Footer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: left;
`;

export const FooterRow = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
`;

export const FooterBlock = styled.div`
  display: flex;
  flex-direction: column;
`;

export const FooterLabel = styled.label`
  color: ${p => p.theme.colors.dark};
  margin-top: 5px;
  font-size: 1.6rem;
  font-weight: bold;
`;

export const FooterSublabel = styled.label`
  color: ${p => p.theme.colors.primary};
  font-size: 1.4rem;
`;
