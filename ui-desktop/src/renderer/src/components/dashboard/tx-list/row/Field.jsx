//@ts-check
import React from 'react';
import styled from 'styled-components';

const FieldWithTitle = styled.div`
  flex: 1 1 0;
`;

const FieldTitle = styled.div`
  font-size: 1.3rem;
  font-weight: 500;
  color: #0e7a4d;
`;

const FieldValue = styled.div`
  text-wrap: nowrap;
`;

export const Field = ({ title, children }) => (
  <FieldWithTitle>
    <FieldTitle>{title}</FieldTitle>
    <FieldValue>{children}</FieldValue>
  </FieldWithTitle>
);
