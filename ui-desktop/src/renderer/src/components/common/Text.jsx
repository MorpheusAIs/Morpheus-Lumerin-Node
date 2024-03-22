import styled from 'styled-components';

// sizes:  11, 13, 14, 16, 20, 24, 32, 48
// weights: normal, 600
// letter spacings: -1, 0.4, 0.5, 1.4, 1.6
// line-heights: 24, 25, 30
// shadows: 0 1px 1px ${p => p.theme.colors.darkShade}

const Text = styled.span`
  text-shadow: 0 1px 1px ${p => p.theme.colors.darkShade};
  line-height: 2.5rem;
  font-size: 2rem;
  font-weight: 600;
`;

export default Text;
