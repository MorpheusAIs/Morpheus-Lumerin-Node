import styled from 'styled-components';

export const Label = styled.label`
  line-height: 1.6rem;
  font-size: 1.3rem;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: ${p => (p.hasErrors ? p.theme.colors.danger : p.theme.colors.dark)};
`;

export const Input = styled.input`
  border: 1px solid gray;
  display: block;
  border-radius: 2px;
  padding: 0.8rem 1.6rem;
  background-color: transparent;
  margin-top: 0.8rem;
  width: 100%;
  line-height: 2.5rem;
  color: ${p => (p.disabled ? p.theme.colors.copy : "white")};
  font-size: 1.3rem;
  font-weight: 600;
  letter-spacing: 0.5px;
  transition: box-shadow 300ms;
  resize: vertical;
  box-shadow: 0 2px 0 0px
    ${p => (p.hasErrors ? p.theme.colors.danger : 'transparent')};

  &:focus {
    outline: none;
    box-shadow: 0 2px 0 0px ${p => p.theme.colors.morMain};
    box-shadow: ${p =>
      p.noFocus && p.value.length > 0
        ? 'none'
        : `0 2px 0 0px ${p.theme.colors.morMains}`};
  }
`;

export const TextArea = Input.withComponent('textarea');

export const ErrorMsg = styled.div`
  color: ${p => p.theme.colors.danger};
  line-height: 1.6rem;
  font-size: 1.3rem;
  font-weight: 600;
  text-align: right;
  margin-top: 0.4rem;
  width: 100%;
  margin-bottom: -2rem;
  display: inline-block;
`;
