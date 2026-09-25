import styled from 'styled-components';
import { AltLayout, AltLayoutNarrow, Btn, Sp } from '../common';
import SecondaryBtn from './SecondaryBtn';

const Description = styled.p`
  color: var(--text-muted);
  font-size: 1.4rem;
  line-height: 1.6;
  text-align: center;
`;

export default function WalletChoiceStep({ onWalletModeSelected }) {
  return (
    <AltLayout title="Welcome to Morpheus" data-testid="wallet-choice">
      <AltLayoutNarrow>
        <Description>
          Create a new wallet or import one you already own. Next, set an app
          password to protect access on this device.
        </Description>
        <Sp mt={4}>
          <Btn block onClick={() => onWalletModeSelected('create')}>
            Create a new wallet
          </Btn>
        </Sp>
        <Sp mt={2}>
          <SecondaryBtn block onClick={() => onWalletModeSelected('import')}>
            Import an existing wallet
          </SecondaryBtn>
        </Sp>
      </AltLayoutNarrow>
    </AltLayout>
  );
}
