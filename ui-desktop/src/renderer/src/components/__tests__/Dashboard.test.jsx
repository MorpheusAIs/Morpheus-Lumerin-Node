/* eslint-disable require-path-exists/exists */
/* eslint-disable import/no-unresolved */
import testUtils from '../../testUtils';
import { Simulate } from 'react-testing-library';
import Dashboard from '../dashboard/Dashboard';
import config from '../../config';
import React from 'react';
import 'react-testing-library/extend-expect';

const ACTIVE_ADDRESS = '0x15dd2028C976beaA6668E286b496A518F457b5Cf';

describe('<Dashboard/>', () => {
  it('Displays active address in header', () => {
    const { getByTestId } = testUtils.reduxRender(
      <Dashboard />,
      getInitialState()
    );
    expect(getByTestId('address')).toBeInTheDOM();
    expect(getByTestId('address')).toHaveTextContent(ACTIVE_ADDRESS);
  });

  it('Displays active address ETH BALANCE', () => {
    const { getByTestId } = testUtils.reduxRender(
      <Dashboard />,
      getInitialState()
    );
    expect(getByTestId('eth-balance')).toBeInTheDOM();
    expect(getByTestId('eth-balance')).toHaveTextContent('0.05');
  });

  it('Displays active address ETH BALANCE in USD', () => {
    const { getByTestId } = testUtils.reduxRender(
      <Dashboard />,
      getInitialState()
    );
    expect(getByTestId('eth-balance-usd')).toHaveTextContent('$5.00 (USD)');
  });

  it.skip('Displays an abbreviation if USD balance is very small', () => {
    const { getByTestId } = testUtils.reduxRender(
      <Dashboard />,
      getInitialState({ ethBalance: '1', lmrBalance: '1' })
    );
    expect(getByTestId('eth-balance-usd')).toHaveTextContent('< $0.01 (USD)');
  });

  it('Displays active address LMR BALANCE', () => {
    const { getByTestId } = testUtils.reduxRender(
      <Dashboard />,
      getInitialState()
    );
    expect(getByTestId('lmr-balance')).toHaveTextContent('0.025');
  });

  it('Displays the RECEIVE DRAWER when Receive button is clicked', () => {
    const { queryByTestId, getByTestId } = testUtils.reduxRender(
      <Dashboard />,
      getInitialState()
    );
    const btn = testUtils.withDataset(getByTestId('receive-btn'), 'modal');
    expect(queryByTestId('receive-drawer')).not.toBeInTheDOM();
    Simulate.click(btn);
    expect(queryByTestId('receive-drawer')).toBeInTheDOM();
  });

  it('Displays the SEND DRAWER when Send button is clicked', () => {
    const { queryByTestId, getByTestId } = testUtils.reduxRender(
      <Dashboard />,
      getInitialState()
    );
    const btn = testUtils.withDataset(getByTestId('send-btn'), 'modal');
    expect(queryByTestId('send-drawer')).not.toBeInTheDOM();
    Simulate.click(btn);
    expect(queryByTestId('send-drawer')).toBeInTheDOM();
  });

  describe('if the user has NO FUNDS', () => {
    it('SEND button is disabled', () => {
      const { queryByTestId, getByTestId } = testUtils.reduxRender(
        <Dashboard />,
        getInitialState({ ethBalance: '0', lmrBalance: '0' })
      );
      const btn = testUtils.withDataset(getByTestId('send-btn'), 'modal');
      expect(queryByTestId('send-drawer')).not.toBeInTheDOM();
      Simulate.click(btn);
      expect(queryByTestId('send-drawer')).not.toBeInTheDOM();
    });

    it('SEND button shows a tooltip when hovered', () => {
      const { getByTestId } = testUtils.reduxRender(
        <Dashboard />,
        getInitialState({ ethBalance: '0', lmrBalance: '0' })
      );
      expect(getByTestId('send-btn').getAttribute('data-rh')).toBe(
        'You need some funds to send'
      );
    });
  });

  describe('if connectivity is lost', () => {
    it('SEND button is disabled', () => {
      const { queryByTestId, getByTestId, store } = testUtils.reduxRender(
        <Dashboard />,
        getInitialState()
      );
      store.dispatch(goOffline());
      const btn = testUtils.withDataset(getByTestId('send-btn'), 'modal');
      expect(queryByTestId('send-drawer')).not.toBeInTheDOM();
      Simulate.click(btn);
      expect(queryByTestId('send-drawer')).not.toBeInTheDOM();
    });

    it('SEND button shows a tooltip when hovered', () => {
      const { getByTestId, store } = testUtils.reduxRender(
        <Dashboard />,
        getInitialState()
      );
      expect(getByTestId('send-btn').getAttribute('data-rh')).toBeNull();
      store.dispatch(goOffline());
      expect(getByTestId('send-btn').getAttribute('data-rh')).toBe(
        "Can't send while offline"
      );
    });
  });

  describe('If there are transactions to display', () => {
    it('displays the transactions list', () => {
      const { queryByTestId } = testUtils.reduxRender(
        <Dashboard />,
        getInitialState({
          transactions: [testUtils.getDummyTransaction()]
        })
      );
      expect(queryByTestId('notx-msg')).not.toBeInTheDOM();
      expect(queryByTestId('tx-list')).toBeInTheDOM();
    });
  });

  describe('If there are NO transactions to display', () => {
    it('displays a message saying "No transactions"', () => {
      const { getByTestId } = testUtils.reduxRender(
        <Dashboard />,
        getInitialState()
      );
      expect(getByTestId('no-tx-placeholder')).toBeInTheDOM();
    });
  });
});

function goOffline() {
  return {
    type: 'connectivity-state-changed',
    payload: { ok: false }
  };
}

function getInitialState({
  symbol = 'LMR',
  lmrBalance = '25000000000000000',
  ethBalance = '50000000000000000',
  transactions = []
} = {}) {
  return testUtils.getInitialState({
    rates: { ETH: { token: 'ETH', price: 100 } },
    wallet: {
      address: ACTIVE_ADDRESS,
      balance: lmrBalance,
      transactions,
      token: {
        contract: config.LMR_TOKEN_ADDR,
        balance: lmrBalance,
        transactions,
        symbol
      }
    }
  });
}
