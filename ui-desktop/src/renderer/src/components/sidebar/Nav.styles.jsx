import styled, { css } from 'styled-components';
import { NavLink } from 'react-router-dom';

const itemStyle = css`
  display: flex;
  align-items: center;
  gap: 1.2rem;
  width: 100%;
  min-height: 4.4rem;
  padding: 1.1rem 1.2rem;
  border: 1px solid transparent;
  border-radius: 10px;
  background: transparent;
  color: var(--text-muted, #9ab4a7);
  font: inherit;
  font-size: 1.4rem;
  font-weight: 550;
  line-height: 1.4;
  text-decoration: none;
  text-align: left;
  cursor: pointer;

  svg {
    width: 2rem;
    height: 2rem;
    flex: 0 0 2rem;
  }
  span {
    white-space: nowrap;
  }
  &:hover {
    color: var(--text-primary, #edf7f0);
    background: var(--surface-hover, #18372a);
  }
  &.active {
    color: var(--accent, #19d695);
    background: #133b2a;
    border-color: rgba(25, 214, 149, 0.14);
  }
  @media (max-width: 799px) {
    [data-sidebar-expanded='false'] & span {
      display: none;
    }
  }
`;

export const NavItem = styled(NavLink)`
  ${itemStyle}
`;
export const NavAction = styled.button.attrs({ type: 'button' })`
  ${itemStyle}
`;
export const NavGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
`;
