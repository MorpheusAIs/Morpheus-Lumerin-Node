import styled from 'styled-components';

export const Container = styled.div`
  margin-top: 2px;
  padding: 0 15px;
`;

/**
 * @component
 * @param {Object} param
 * @param {string} param.color css color
 * @param {string} param.width css width
 */
export const BarElem = styled.div`
  height: 2px;
  margin-top: -2px;

  &:before {
    margin-bottom: 2px;
    content: '';
    display: block;
    background-color: ${({ color }) => color};
    width: ${({ width }) => width};
    height: 2px;
    transition: 0.5s;
  }
`;

/**
 * @component
 * @param {Object} param
 * @param {string} param.color css color
 */
export const Message = styled.div`
  line-height: 1.6rem;
  height: 1.6rem;
  font-size: 1.3rem;
  font-weight: 600;
  letter-spacing: 0.4px;
  color: ${({ color }) => color};
  text-align: right;
  margin-bottom: -18px;
`;
