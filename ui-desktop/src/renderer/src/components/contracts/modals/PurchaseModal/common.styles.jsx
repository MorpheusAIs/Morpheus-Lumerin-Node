import styled from 'styled-components';
import { Label, InputGroup } from '../CreateContractModal.styles';

export const Divider = styled.div`
  margin-top: 5px
  width:100%;
  height: 0px;
  border: 0.5px solid rgba(0, 0, 0, 0.25);`;

export const HeaderFlex = styled.div`
  display: flex;
  justify-content: space-between;
`;

export const SmallTitle = styled(Label)`
  display: flex;
  align-items: center;
  font-size: 1rem !important;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.7);
`;

export const ContractInfoContainer = styled.div`
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
`;

export const UpperCaseTitle = styled(SmallTitle)`
  text-transform: uppercase;
`;

export const ActionsGroup = styled(InputGroup)`
  text-align: center;
  justify-content: space-between;
  height: 60px;
  margintop: 50px;
`;

export const UrlContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: start;
  margin-top: 50px;
`;

export const Values = styled.div`
  line-height: 1.4rem;
  font-size: 1.4rem;
  font-weight: 100;
  display: flex;
  align-items: center;
`;

export const EditBtn = styled.div`
  cursor: pointer;
  color: #014353;
  text-decoration: underline;
  font-size: 1rem;
  letter-spacing: 1px;
`;

export const PreviewCont = styled.div`
  display: flex;
  height: 85%;
  margin: 1rem 0 0;
  flex-direction: column;
  justify-content: space-between;
`;

export const PoolInfoContainer = styled.div`
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
  width: 100%;
`;
