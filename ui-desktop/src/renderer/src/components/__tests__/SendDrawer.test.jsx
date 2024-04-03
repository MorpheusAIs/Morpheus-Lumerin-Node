/* eslint-disable require-path-exists/exists */
/* eslint-disable import/no-unresolved */
import testUtils from '../../testUtils';
import { Simulate } from 'react-testing-library';
import SendDrawer from '../dashboard/SendDrawer';
import config from '../../config';
import React from 'react';
import 'react-testing-library/extend-expect';

const closeHandler = jest.fn();

const getElement = defaultTab => (
  <SendDrawer onRequestClose={closeHandler} defaultTab={defaultTab} isOpen />
);

describe('<SendDrawer/>', () => {
  it.skip('displays SEND ETH form when clicking the tab', () => {
    const { queryByTestId, getByTestId } = testUtils.reduxRender(
      getElement(),
      getInitialState()
    );
    expect(queryByTestId('sendEth-form')).toBeNull();
    Simulate.click(testUtils.withDataset(getByTestId('eth-tab'), 'tab'));
    expect(queryByTestId('sendEth-form')).not.toBeNull();
  });

  it.skip('displays SEND LMR form when clicking the tab', () => {
    const { queryByTestId, getByTestId } = testUtils.reduxRender(
      getElement('eth'),
      getInitialState()
    );
    expect(queryByTestId('sendLmr-form')).toBeNull();
    Simulate.click(testUtils.withDataset(getByTestId('lmr-tab'), 'tab'));
    expect(queryByTestId('sendLmr-form')).not.toBeNull();
  });

  describe('SEND LMR tab is disabled and displays tooltip', () => {
    it('if user HAS NO LMR', () => {
      const { queryByTestId, getByTestId } = testUtils.reduxRender(
        getElement('eth'),
        getInitialState({ lmrBalance: '0' })
      );
      Simulate.click(testUtils.withDataset(getByTestId('lmr-tab'), 'tab'));
      expect(queryByTestId('sendLmr-form')).not.toBeInTheDOM();
      expect(getByTestId('lmr-tab').getAttribute('data-rh')).not.toBeNull();
      expect(getByTestId('lmr-tab').getAttribute('data-rh')).toBe(
        'You need some LMR to send'
      );
    });

    it('if user IS OFFLINE', () => {
      const { queryByTestId, getByTestId, store } = testUtils.reduxRender(
        getElement('eth'),
        getInitialState()
      );
      expect(getByTestId('lmr-tab').getAttribute('data-rh')).toBeNull();
      store.dispatch(goOffline());
      expect(getByTestId('lmr-tab').getAttribute('data-rh')).not.toBeNull();
      Simulate.click(testUtils.withDataset(getByTestId('lmr-tab'), 'tab'));
      expect(queryByTestId('sendLmr-form')).not.toBeInTheDOM();
      expect(getByTestId('lmr-tab').getAttribute('data-rh')).toBe(
        "Can't send while offline"
      );
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
  isInitialAuction = false,
  lmrBalance = '5000000000000000000000'
} = {}) {
  return testUtils.getInitialState({
    auction: { status: { currentAuction: isInitialAuction ? 0 : 1 } },
    rates: { ETH: { token: 'ETH', price: 250 } },
    wallets: {
      active: 'foo',
      allIds: ['foo'],
      byId: {
        foo: {
          addresses: {
            '0x15dd2028C976beaA6668E286b496A518F457b5Cf': {
              token: {
                [config.LMR_TOKEN_ADDR]: { balance: lmrBalance }
              },
              balance: '5000000000000000000000'
            }
          }
        }
      }
    }
  });
}
