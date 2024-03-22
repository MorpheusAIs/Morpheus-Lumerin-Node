import React from 'react';
import ReactHintFactory from 'react-hint';
import { Container, GlobalStyles, ContainerLocal } from './Tooltips.styles';

export const ReactHint = ReactHintFactory(React);

const onRenderContentDefault = (target, content) => (
  <Container
    data-testid="tooltip"
    negative={target.dataset.rhNegative}
    maxWidth={target.dataset.rhWidth}
    darker={target.dataset.rhDarker}
  >
    {content}
  </Container>
);

// Global tooltip component
export const GlobalTooltips = () => (
  <React.Fragment>
    <ReactHint events delay={100} onRenderContent={onRenderContentDefault} />
    <GlobalStyles />
  </React.Fragment>
);

// Local tooltip component
export const Tooltip = ({ content, show, ...props }) => (
  <ContainerLocal show={show} {...props}>
    {content}
  </ContainerLocal>
);
