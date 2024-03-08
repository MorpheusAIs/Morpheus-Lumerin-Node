import PropTypes from 'prop-types';
import styled from 'styled-components';

// Every element can potentially be a flexbox item...
const Item = styled.div`
  flex-shrink: ${p => p.shrink || '0'};
  flex-basis: ${p => p.basis || 'auto'};
  flex-grow: ${p => p.grow || '0'};
  order: ${p => p.order || '0'};
`;

Item.propTypes = {
  shrink: PropTypes.string,
  basis: PropTypes.string,
  order: PropTypes.string,
  grow: PropTypes.string
};

// ...even flexbox containers, that's why we extend from Item...
const Row = styled(Item)`
  justify-content: ${p => p.justify || 'flex-start'};
  align-items: ${p => p.align || 'stretch'};
  flex-wrap: ${p => (p.rowwrap ? 'wrap' : 'nowrap')};
  display: flex;
`;

// ...and Columns are just Rows with a diferent flex-direction.
const Column = styled(Row)`
  flex-direction: column;
`;

const containerProptypes = {
  justify: PropTypes.oneOf([
    'space-between',
    'space-around',
    'space-evenly',
    'flex-start',
    'flex-end',
    'center'
  ]),
  rowwrap: PropTypes.bool,
  align: PropTypes.oneOf(['center', 'flex-start', 'flex-end', 'baseline'])
};

Column.propTypes = containerProptypes;
Row.propTypes = containerProptypes;

const Flex = { Column, Row, Item };
export default Flex;
