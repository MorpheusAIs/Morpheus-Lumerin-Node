import React from 'react';
import styled from 'styled-components';

const Container = styled.div`
  position: relative;
  line-height: 1.5;
  padding-left: 2em;
  margin-bottom: 0.8rem;
`;

const Icon = styled.svg`
  position: absolute;
  fill: ${p =>
    p.isActive ? p.theme.colors.primary : p.theme.colors.translucentDark};
  left: 0;
  top: 1px;
`;

const Label = styled.span`
  transition: opacity 0.3s;
  opacity: ${p => (p.isActive ? 1 : 0.5)};
`;

export default function ChecklistItem({ isActive, text }) {
  return (
    <Container>
      <Icon isActive={isActive} viewBox="0 0 24 24" height="24" width="24">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm3.88-11.71L10 14.17l-1.88-1.88a.996.996 0 1 0-1.41 1.41l2.59 2.59c.39.39 1.02.39 1.41 0L17.3 9.7a.996.996 0 0 0 0-1.41c-.39-.39-1.03-.39-1.42 0z" />
      </Icon>
      <Label isActive={isActive}>{text}</Label>
    </Container>
  );
}
