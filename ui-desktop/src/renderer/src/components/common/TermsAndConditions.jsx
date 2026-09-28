import React from 'react';
import styled from 'styled-components';
import Markdown from 'marked-react';
import terms from '../../termsAndConditions.md?raw';

const StyledTC = styled.div`
  text-align: left;

  h1 {
    font-size: 1.5em;
  }
  h2 {
    font-size: 1.25em;
  }
  h3 {
    font-size: 1.1em;
  }
`;

const TermsAndConditions = () => {
  return (
    <StyledTC>
      <Markdown>{terms}</Markdown>
    </StyledTC>
  );
};

export default TermsAndConditions;
