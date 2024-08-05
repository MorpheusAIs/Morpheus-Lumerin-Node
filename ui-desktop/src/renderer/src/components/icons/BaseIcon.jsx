import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

const Container = styled.svg`
  display: ${p => p.display || 'block'};
  width: ${p => (p.size ? p.size : '2.4rem')};
  fill: ${p => (p.color ? p.color : 'transparent')};
`;

export default function BaseIcon({ children, ...other }) {
  // static propTypes = {
  //   children: PropTypes.node.isRequired,
  //   color: PropTypes.string,
  //   size: PropTypes.string
  // }

  return (
    <Container viewBox="0 0 24 24" {...other}>
      {children}
    </Container>
  );
}
