import styled from 'styled-components';

import { BaseBtn } from '../common';

const SecondaryBtn = styled(BaseBtn)`
  font-size: 1.2rem;
  color: ${p => p.theme.colors.dark}
  :hover {
    opacity: 0.75;
  }
`;

export default SecondaryBtn;
