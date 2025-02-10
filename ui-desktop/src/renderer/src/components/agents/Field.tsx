import { ReactNode, forwardRef } from 'react';
import styled from 'styled-components';

const FieldWithTitle = styled.div`
  flex: 1 1 0;
`;

const FieldTitle = styled.div`
  font-size: 1.2rem;
  font-weight: 500;
  color: #0e7a4d;
`;

const FieldValue = styled.div`
  text-wrap: nowrap;
`;

type FieldProps = {
  title: string;
  children: ReactNode;
};

export const Field = forwardRef<HTMLDivElement, FieldProps>((props, ref) => (
  <FieldWithTitle>
    <FieldTitle>{props.title}</FieldTitle>
    <FieldValue ref={ref}>{props.children}</FieldValue>
  </FieldWithTitle>
));
