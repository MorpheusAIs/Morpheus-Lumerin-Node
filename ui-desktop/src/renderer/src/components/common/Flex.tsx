import styled from 'styled-components';

// Every element can potentially be a flexbox item...
const Item = styled.div<{
  shrink?: string;
  basis?: string;
  grow?: string;
  order?: string;
}>`
  flex-shrink: ${(p) => p.shrink || '0'};
  flex-basis: ${(p) => p.basis || 'auto'};
  flex-grow: ${(p) => p.grow || '0'};
  order: ${(p) => p.order || '0'};
`;

// ...even flexbox containers, that's why we extend from Item...
const Row = styled(Item)<{
  justify?: string;
  align?: string;
  rowwrap?: boolean;
  gap?: string;
}>`
  justify-content: ${(p) => p.justify || 'flex-start'};
  align-items: ${(p) => p.align || 'stretch'};
  flex-wrap: ${(p) => (p.rowwrap ? 'wrap' : 'nowrap')};
  gap: ${(p) => p.gap || '0'};
  display: flex;
`;

// ...and Columns are just Rows with a diferent flex-direction.
const Column = styled(Row)`
  flex-direction: column;
`;

const Flex = { Column, Row, Item };
export default Flex;
