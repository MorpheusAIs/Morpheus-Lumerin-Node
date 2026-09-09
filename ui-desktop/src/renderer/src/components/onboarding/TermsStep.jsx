import styled from 'styled-components';
import { useState } from 'react';
import licenseText from '../../../../../LICENSE?raw';
import TermsAndConditions from '../common/TermsAndConditions';
import { AltLayout, AltLayoutNarrow, Btn, Sp } from '../common';
import SecondaryBtn from './SecondaryBtn';

const Description = styled.p`
  color: var(--text-muted);
  font-size: 1.4rem;
  line-height: 1.6;
`;
const LegalText = styled.section`
  max-height: clamp(14rem, calc(100vh - 50rem), 32rem);
  overflow: auto;
  background: var(--surface-sunken, #071a12);
  border: 1px solid var(--border-subtle);
  border-radius: 10px;
  padding: 1.6rem;
  margin: 2rem 0;
  color: var(--text-primary);
  font-size: 1.4rem;
  line-height: 1.6;
  overflow-wrap: anywhere;
`;
const Consent = styled.label`
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  min-height: 44px;
  padding-block: 0.8rem;
  color: var(--text-primary);
  font-size: 1.4rem;
  line-height: 1.5;
  cursor: pointer;
  input {
    flex-shrink: 0;
    margin-top: 0.35rem;
    accent-color: var(--accent);
  }
`;
const LicenseButton = styled.button`
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--accent);
  font: inherit;
  text-decoration: underline;
  text-underline-offset: 3px;
`;

export default function TermsStep(props) {
  const [showLicense, setShowLicense] = useState(false);
  const toggle = (event) =>
    props.onInputChange({ id: event.target.id, value: event.target.checked });
  return (
    <AltLayout title="Terms and conditions" data-testid="onboarding-container">
      <AltLayoutNarrow>
        <Description>
          Review the terms before setting up your wallet.
        </Description>
        <LegalText tabIndex={0} aria-label="Morpheus terms of use">
          <TermsAndConditions />
        </LegalText>
        <Consent htmlFor="termsCheckbox">
          <input
            id="termsCheckbox"
            data-testid="accept-terms-chb"
            type="checkbox"
            checked={props.termsCheckbox}
            onChange={toggle}
          />
          <span>I have read and accept these terms</span>
        </Consent>
        <Consent htmlFor="licenseCheckbox">
          <input
            id="licenseCheckbox"
            data-testid="accept-license-chb"
            type="checkbox"
            checked={props.licenseCheckbox}
            onChange={toggle}
          />
          <span>I have read and accept the software license</span>
        </Consent>
        <LicenseButton
          type="button"
          aria-expanded={showLicense}
          aria-controls="software-license-text"
          onClick={() => setShowLicense(!showLicense)}
        >
          {showLicense
            ? 'Hide the software license'
            : 'Read the software license'}
        </LicenseButton>
        {showLicense && (
          <LegalText
            id="software-license-text"
            tabIndex={0}
            aria-label="Morpheus software license"
          >
            <pre style={{ whiteSpace: 'pre-wrap', font: 'inherit', margin: 0 }}>
              {licenseText}
            </pre>
          </LegalText>
        )}
        <Sp mt={3}>
          <Btn
            data-testid="accept-terms-btn"
            disabled={!props.licenseCheckbox || !props.termsCheckbox}
            onClick={props.onTermsAccepted}
            block
          >
            Accept and continue
          </Btn>
        </Sp>
        <Sp mt={2}>
          <SecondaryBtn block onClick={props.onChooseWallet}>
            Back to wallet options
          </SecondaryBtn>
        </Sp>
      </AltLayoutNarrow>
    </AltLayout>
  );
}
