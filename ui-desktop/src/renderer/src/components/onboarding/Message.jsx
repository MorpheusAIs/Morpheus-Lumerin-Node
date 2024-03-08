import styled from 'styled-components';

const Message = styled.div`
  font-size: 1.6rem;
  font-weight: 500;
  line-height: 1.5;
  color: ${p => p.theme.colors.dark};

  & span {
    font-size: 13px;
  }

  & a {
    text-decoration: underline;
    font-size: 13px;
    cursor: pointer;
    color: ${p => p.theme.colors.success};
  }
`;

export default Message;
