import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

const Container = styled.div`
  min-height: 100%;
  background-color: ${p => p.theme.colors.light};
`;

const Header = styled.header`
  background-color: ${p => p.theme.colors.light};

  padding: 4.5rem 4.8rem;
  padding-bottom: 0;
`;

const Title = styled.h1`
  background-color: ${p => p.theme.colors.light};
  color: ${p => p.theme.colors.dark};
  margin: 0;
  line-height: 3rem;
  font-size: 2.4rem;
  cursor: default;
`;

const LightLayout = props => {
  const { children, title, ...other } = props;

  return (
    <Container {...other}>
      <Header>
        <Title>{title}</Title>
      </Header>
      {children}
    </Container>
  );
};

LightLayout.propTypes = {
  children: PropTypes.node.isRequired,
  title: PropTypes.string.isRequired
};

export default LightLayout;
