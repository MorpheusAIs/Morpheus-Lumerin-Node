import styled, { createGlobalStyle, css } from 'styled-components';
import withCurrencySelectorState from '../../store/hocs/withCurrencySelectorState';
import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import React from 'react';

import { DisplayValue, Flex } from '../common';
import CaretIcon from '../icons/CaretIcon';
import CoinIcon from '../icons/CoinIcon';
import { LumerinLightIcon } from '../icons/LumerinLightIcon';
import { EtherIcon } from '../icons/EtherIcon';

const relSize = ratio => `calc(100vw / ${ratio})`;

const wideOrHover = styles => ({ parent }) =>
  css`
    ${parent}:hover & {
      ${styles};
    }
    @media (min-width: 800px) {
      ${styles};
    }
  `;

const GlobalStyles = createGlobalStyle`
  [data-radix-popper-content-wrapper] {
    z-index: 4 !important;
    width: 20%;
  }
`;

const MenuButton = styled(DropdownMenu.Trigger)`
  background-color: ${({ theme }) => theme.colors.lightShade};
  font: inherit;
  color: ${({ theme }) => theme.colors.light};
  border-radius: 1.2rem;
  border: none;
  text-align: left;
  display: block;
  width: 100%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 0.4rem 1rem 0.8rem;
  transition: padding 0.3s;
  &:focus,
  &:hover {
    background-color: ${({ theme }) => theme.colors.darkShade};
    outline: none;
  }
  &[aria-expanded='true'] {
    visibility: hidden;
  }
  ${wideOrHover`
    justify-content: center;
    padding: 1.6rem 1.2rem;
  `};
`;

const MenuList = styled(DropdownMenu.Content)`
  box-shadow: 0 0 32px 0 rgba(0, 0, 0, 0.4);
  opactity: 0;
  display: block;
  white-space: nowrap;
  outline: none;
  overflow: hidden;
  border-radius: 1.2rem;
  transform: translateY(-56px);
`;

const MenuItem = styled(DropdownMenu.Item)`
  background-color: ${({ theme }) => theme.colors.lightShade};
  display: block;
  cursor: pointer;
  padding: 1.6rem 1.2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  &[data-highlighted] {
    background-color: ${({ theme }) => theme.colors.translucentPrimary};
    outline: none;
  }
  @media (min-width: 800px) {
    justify-content: center;
  }
`;

const Icon = styled(CoinIcon)`
  opacity: 0.5;
  transition: opacity 0.3s;
  height: 2.4rem;
  width: 2rem;
  transition: width 0.3s, opacity 0.3s;
  [role='menuitem'] &,
  ${MenuButton}:focus &,
  ${MenuButton}:hover & {
    opacity: 1;
  }
  ${wideOrHover`
    width: 2.4rem;
  `};
`;

const ItemBody = styled(Flex.Item)`
  overflow: hidden;
  margin: 0;
  opacity: 0;
  transition: opacity 0.3s, margin 0.3s;
  [role='menuitem'] & {
    opacity: 1;
    margin-left: 0.8rem;
    margin-right: 0.4rem;
  }
  ${wideOrHover`
    opacity: 1;
    margin-left: 0.8rem;
    margin-right: 0.4rem;
  `};
`;

const CurrencyName = styled.div`
  color: ${({ theme }) => theme.colors.light};
  font-size: ${relSize(48)};
  line-height: 1.4rem;
  letter-spacing: 1.6px;
  font-weight: 600;
  margin-top: -2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`;

const Balance = styled.div`
  color: ${({ theme }) => theme.colors.primary};
  font-size: 1.1rem;
  line-height: 1.4rem;
  letter-spacing: 1px;
  font-weight: 600;
  margin-bottom: -2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`;

const Caret = styled(CaretIcon)`
  transform: scaleY(${({ caret }) => (caret === 'up' ? -1 : 1)});
  opacity: ${({ caret }) => (caret === 'none' ? 0 : 0.5)};
  transition: opacity 0.3s;
  [role='menuitem'] &,
  ${MenuButton}:focus &,
  ${MenuButton}:hover & {
    opacity: ${({ caret }) => (caret === 'none' ? 0 : 1)};
  }
`;

function CurrencySelector({
  onCurrencyChange,
  activeCurrency,
  balances,
  parent,
  handleMouseEnter,
  handleMouseLeave
}) {
  return (
    <React.Fragment>
      <DropdownMenu.Root>
        <MenuButton>
          <Item
            displayName={activeCurrency}
            balance={balances[activeCurrency]}
            caret="down"
          />
        </MenuButton>
        <DropdownMenu.Portal>
          <MenuList
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
            sideOffset={0}
            align="start"
          >
            <MenuItem
              onSelect={() => onCurrencyChange(activeCurrency)}
              key={activeCurrency}
            >
              <Item
                displayName={activeCurrency}
                balance={balances[activeCurrency]}
                caret="up"
              />
            </MenuItem>
            {[
              {
                balance: balances['LMR'],
                name: 'LMR'
              },
              {
                balance: balances['ETH'],
                name: 'ETH'
              }
            ]
              .filter(({ name }) => name !== activeCurrency)
              .map(({ name, balance }) => (
                <MenuItem onSelect={() => onCurrencyChange(name)} key={name}>
                  <Item displayName={name} balance={balance} caret="none" />
                </MenuItem>
              ))}
          </MenuList>
        </DropdownMenu.Portal>
      </DropdownMenu.Root>
      <GlobalStyles />
    </React.Fragment>
  );
}

const Item = ({ displayName, balance, id, caret, parent }) => {
  const icons = {
    ETH: EtherIcon,
    LMR: LumerinLightIcon
  };

  return (
    <React.Fragment>
      <Flex.Item>
        {displayName === 'ETH' ? <EtherIcon /> : <LumerinLightIcon />}
      </Flex.Item>
      <ItemBody grow="1" shrink="1">
        <CurrencyName>{displayName}</CurrencyName>
        <Balance>
          <DisplayValue value={balance} post=" LMR" />
        </Balance>
      </ItemBody>
      <Flex.Item shrink="0">
        <Caret caret={caret} />
      </Flex.Item>
    </React.Fragment>
  );
};

export default withCurrencySelectorState(CurrencySelector);
