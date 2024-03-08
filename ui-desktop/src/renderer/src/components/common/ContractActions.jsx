import * as ReachUI from '@reach/menu-button';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

import { ErrorMsg, Label } from './TextInput.styles';
import SelectorCaret from '../icons/SelectorCaret';

const MenuButton = styled(ReachUI.MenuButton)`
  background-color: ${p => p.theme.colors.primary};
  color: ${p => p.theme.colors.light};
  font-size: 1.2rem;
  font-weight: 500;
  letter-spacing: 1px;
  padding: 0;
  border: none;
  border-radius: 12px;
  display: block;
  height: fit-content;
  text-align: left;
  width: 100%;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 0 0px
    ${p => (p.hasErrors ? p.theme.colors.danger : 'transparent')};

  &[disabled] {
    cursor: not-allowed;
  }

  &:focus {
    outline: none;
    box-shadow: 0 2px 0 0px ${p => p.theme.colors.primary};
    box-shadow: ${p =>
      p.noFocus && p.value.length > 0
        ? 'none'
        : `0 2px 0 0px ${p.theme.colors.primary}`};
  }
`;

const ValueContainer = styled.div`
  padding: 0.8rem 0 0.7rem 1.6rem;
  flex-grow: 1;
`;

const CaretContainer = styled.div`
  background-color: transparent;
  padding: 0.8rem 1rem;
  svg {
    fill: ${p => p.theme.colors.ligth};
  }

  [aria-expanded='true'] & {
    svg {
      fill: ${p => p.theme.colors.ligth};
    }
  }

  [disabled] & {
    opacity: 0.25;
  }
`;

const MenuList = styled(ReachUI.MenuList)`
  background-color: ${p => p.theme.colors.light};
  width: 100%;
`;

const MenuItem = styled(ReachUI.MenuItem)`
  color: ${p => p.theme.colors.primary};
  width: 100%;
  font-size: 1.3rem;
  font-weight: 600;
  letter-spacing: 0.5px;
  padding: 1.2rem 1.6rem;
  cursor: pointer;

  &[data-selected] {
    background-color: #eaf7fc;
    color: ${p => p.theme.colors.primary};
    outline: none;
  }

  &[data-disabled] {
    opacity: 0.25;
    cursor: not-allowed;
  }
`;

export default class ContractActions extends React.Component {
  static propTypes = {
    'data-testid': PropTypes.string,
    onChange: PropTypes.func.isRequired,
    options: PropTypes.arrayOf(
      PropTypes.shape({
        value: PropTypes.string.isRequired,
        label: PropTypes.string.isRequired
      })
    ).isRequired,
    error: PropTypes.oneOfType([
      PropTypes.arrayOf(PropTypes.string),
      PropTypes.string
    ]),
    label: PropTypes.string.isRequired,
    value: PropTypes.string,
    id: PropTypes.string.isRequired
  };

  onChange = e => {
    this.props.onChange({ id: this.props.id, value: e.target.value });
  };

  render() {
    const { onChange, options, error, label, value, id, ...other } = this.props;

    const hasErrors = error && error.length > 0;
    const activeItem = options.find(item => item.value === value);

    return (
      <div>
        <Label hasErrors={hasErrors} htmlFor={id}>
          {label}
        </Label>
        <ReachUI.Menu>
          <MenuButton {...other}>
            <ValueContainer>
              {activeItem ? activeItem.label : ''}{' '}
            </ValueContainer>
            <CaretContainer>
              <SelectorCaret style={{ width: '14px' }} />
            </CaretContainer>
          </MenuButton>
          <MenuList>
            {options
              .filter(i => !i.hidden)
              .map(item => (
                <MenuItem
                  onSelect={e => onChange({ id, value: item.value, ...e })}
                  key={item.value}
                  disabled={item.disabled}
                  data-rh={item.message}
                >
                  {item.label}
                </MenuItem>
              ))}
          </MenuList>
        </ReachUI.Menu>
        {hasErrors && (
          <ErrorMsg data-testid={`${this.props['data-testid']}-error`}>
            {typeof error === 'string' ? error : error.join('. ')}
          </ErrorMsg>
        )}
      </div>
    );
  }
}
