import TermsAndConditions from '../common/TermsAndConditions';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import React, { useState } from 'react';
import { TextInput, AltLayout, AltLayoutNarrow, Btn, Sp } from '../common';
import { abbreviateAddress } from '../../utils';

import Message from './Message';

const DisclaimerWarning = styled.div`
  text-align: left;
  color: ${(p) => p.theme.colors.dark};
  font-size: 16px;
  margin-top: 16px;
  text-align: justify;
`;

const DisclaimerMessage = styled.div`
  width: 100%;
  height: 130px;
  border-radius: 2px;
  background-color: rgba(0, 0, 0, 0.5);
  color: ${(p) => p.theme.colors.dark};
  overflow: auto;
  font-size: 12px;
  padding: 10px 16px 0 16px;
  margin: 16px 0;
`;

const P = styled.p`
  color: ${(p) => p.theme.colors.dark};
`;

const Subtext = styled.span`
  color: ${(p) => p.theme.colors.dark};
`;

const Mnemonic = styled.div`
  font-size: 1.8rem;
  font-weight: 600;
  line-height: 2;
  text-align: center;
  color: ${(p) => p.theme.colors.morMain};
  word-spacing: 1.6rem;
`;

const Select = styled.select`
  outline: 0;
  border: 1px solid grey;
  padding: 1.2rem 2.4rem;
  letter-spacing: 1.4px;
  line-height: 1.2rem;
  font-size: 1.2rem;
  background: transparent;
  border-radius: 5px;
  font-weight: bold;
  font: inherit;
  color: white;
`;

export const ImportFlow = (props) => {
  const onCheckboxToggle = (e) => {
    props.onInputChange({ id: e.target.id, value: e.target.checked });
  };

  const [mode, setMode] = useState('mnemonic');
  const [isSelectingAddress, setIsSelectingAddress] = useState(false);
  const [addresses, setAddresses] = useState([]);
  const [derivationIndex, setDerivationIndex] = useState(0);

  const handleSetMnemonic = async (e) => {
    setIsSelectingAddress(true);
    e.stopPropagation();
    const addresses = await props.onSuggestAddress();
    setAddresses(addresses);
  };

  return (
    <AltLayout title="Access to wallet" data-testid="onboarding-container">
      <AltLayoutNarrow>
        {isSelectingAddress ? (
          <>
            <Sp mt={2} mb={2}>
              <Mnemonic data-testid="mnemonic-label">
                {props.userMnemonic}
              </Mnemonic>
            </Sp>

            <AltLayoutNarrow>
              <Message>
                Select one of 10 accounts derivied from mnemonic
              </Message>
            </AltLayoutNarrow>
            <Sp mt={3}>
              <Select
                style={{ width: '100%' }}
                id={'derivationPath'}
                error={props.errors.derivationPath}
                value={derivationIndex || 0}
                onChange={(e) => setDerivationIndex(e.target.value)}
              >
                {addresses.length &&
                  addresses.map((a, i) => {
                    return (
                      <option key={a} value={i}>
                        {abbreviateAddress(a, 10)}
                      </option>
                    );
                  })}
              </Select>
            </Sp>
          </>
        ) : (
          <>
            <DisclaimerWarning>
              Import your wallet using a private key or mnemonic
            </DisclaimerWarning>
            <Sp mt={2} mb={2}>
              <Select onChange={(e) => setMode(e.target.value)}>
                <option key={'mnemonic'} value={'mnemonic'} selected={true}>
                  Mnemonic
                </option>
                <option key={'key'} value={'key'}>
                  Private Key
                </option>
              </Select>
            </Sp>

            {mode == 'mnemonic' ? (
              <>
                <AltLayoutNarrow>
                  <Message>
                    Enter a valid 12 word mnemonic to import a previously
                    created wallet and select address
                  </Message>
                </AltLayoutNarrow>
                <Sp mt={3}>
                  <TextInput
                    data-testid="mnemonic-field"
                    autoFocus
                    onChange={props.onInputChange}
                    onPaste={(e) => {
                      e.preventDefault();
                      const value = e.clipboardData.getData('Text').trim();
                      props.onInputChange({ value, id: 'userMnemonic' });
                    }}
                    label="Import Mnemonic"
                    error={props.errors.userMnemonic}
                    value={props.userMnemonic || ''}
                    rows={2}
                    id={'userMnemonic'}
                  />
                </Sp>
              </>
            ) : (
              <>
                <AltLayoutNarrow>
                  <Message>Enter private key to import wallet</Message>
                </AltLayoutNarrow>
                <Sp mt={3}>
                  <TextInput
                    data-testid="pKey-field"
                    autoFocus
                    onChange={props.onInputChange}
                    onPaste={(e) => {
                      e.preventDefault();
                      const value = e.clipboardData.getData('Text').trim();
                      props.onInputChange({ value, id: 'userPrivateKey' });
                    }}
                    label="Import Private Key"
                    error={props.errors.userPrivateKey}
                    value={props.userPrivateKey || ''}
                    rows={2}
                    id={'userPrivateKey'}
                  />
                </Sp>
              </>
            )}
          </>
        )}

        {/* Select address - generate addresses - use */}
        {mode == 'mnemonic' ? (
          <Sp mt={6}>
            <Btn
              data-testid="accept-terms-btn"
              autoFocus
              onClick={(e) =>
                !isSelectingAddress
                  ? handleSetMnemonic(e)
                  : props.onMnemonicSet(e, derivationIndex)
              }
              block
            >
              {!isSelectingAddress ? 'Select Address' : 'Confirm'}
            </Btn>
          </Sp>
        ) : (
          <Sp mt={6}>
            <Btn
              data-testid="accept-terms-btn"
              autoFocus
              onClick={(e) => props.onPrivateKeyAccepted(e)}
              block
            >
              Confirm
            </Btn>
          </Sp>
        )}
      </AltLayoutNarrow>
    </AltLayout>
  );
};

// RecoverFromPrivateKey.propTypes = {
//   onTermsLinkClick: PropTypes.func.isRequired,
//   onTermsAccepted: PropTypes.func.isRequired,
//   licenseCheckbox: PropTypes.bool.isRequired,
//   termsCheckbox: PropTypes.bool.isRequired,
//   onInputChange: PropTypes.func.isRequired
// };
