import styled from 'styled-components';

type BaseBtnProps = {
  submit?: boolean;
  block?: boolean;
};

type FieldBtnProps = BaseBtnProps & {
  float?: boolean;
};

export const BaseBtn = styled.button.attrs<BaseBtnProps>(({ submit }) => ({
  type: submit ? 'submit' : 'button',
}))<BaseBtnProps>`
  display: ${({ block }) => (block ? 'flex' : 'inline-flex')};
  align-items: center;
  justify-content: center;
  gap: 0.8rem;
  min-width: 0;
  max-width: 100%;
  width: ${({ block }) => (block ? '100%' : 'auto')};
  font: inherit;
  text-align: center;
  border: none;
  cursor: pointer;
  line-height: 1.4;
  overflow-wrap: anywhere;
  transition:
    background-color 150ms ease-out,
    border-color 150ms ease-out,
    color 150ms ease-out,
    box-shadow 150ms ease-out;
  background-color: transparent;
  padding: 0;
  color: ${(p) => p.theme.colors.light};
  outline: none;

  &[data-disabled='true'],
  &[disabled] {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

export const Btn = styled(BaseBtn)`
  min-height: 4.4rem;
  line-height: 1.4;
  font-size: 1.4rem;
  font-weight: 650;
  color: #032117;
  border-radius: 10px;
  background-color: var(--accent, ${(p) => p.theme.colors.morMain});
  padding: 1.1rem 1.6rem;

  &:not([disabled], [data-disabled]):hover,
  &:not([disabled], [data-disabled]):focus-visible {
    background-color: #48e4ae;
  }

  &:not([disabled], [data-disabled]):active {
    background-color: #12b77e;
  }
`;

export const FieldBtn = styled(BaseBtn)<FieldBtnProps>`
  float: ${(p) => (p.float ? 'right' : 'none')};
  line-height: 1.8rem;
  color: var(--text-muted, #9ab4a7);
  font-size: 1.4rem;
  font-weight: 600;
  letter-spacing: 0;
  margin-top: ${(p) => (p.float ? '0.4rem' : 0)};
  white-space: nowrap;

  &:hover {
    opacity: 1;
  }
`;
