import styled from 'styled-components';

export const Label = styled.label`
  line-height: 1.5;
  font-size: 1.3rem;
  font-weight: 600;
  letter-spacing: 0;
  color: ${(p) => (p.hasErrors ? p.theme.colors.danger : p.theme.colors.dark)};
`;

export const Input = styled.input`
  border: 1px solid
    ${(p) =>
      p.hasErrors ? p.theme.colors.danger : 'var(--border-strong, #456a57)'};
  display: block;
  border-radius: 10px;
  padding: 1rem 1.2rem;
  min-height: 4.4rem;
  background-color: var(--surface-base, #071711);
  margin-top: 0.8rem;
  width: 100%;
  line-height: 1.5;
  color: ${(p) => (p.disabled ? p.theme.colors.copy : 'white')};
  font-size: 1.4rem;
  font-weight: 400;
  transition: border-color 150ms ease-out;
  resize: vertical;

  &:focus {
    border-color: ${(p) =>
      p.hasErrors ? p.theme.colors.danger : 'var(--accent, #19d695)'};
  }
`;

export const TextArea = Input.withComponent('textarea');

export const ErrorMsg = styled.div`
  color: ${(p) => p.theme.colors.danger};
  line-height: 1.6rem;
  font-size: 1.3rem;
  font-weight: 600;
  text-align: left;
  margin-top: 0.4rem;
  width: 100%;
  display: inline-block;
`;
