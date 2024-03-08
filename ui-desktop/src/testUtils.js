/* eslint-disable require-path-exists/exists */
/* eslint-disable import/no-unresolved */
import { Provider as ClientProvider } from './store/hocs/clientContext';
import { Provider, createStore } from './store/store';
import { render, Simulate } from 'react-testing-library';
import { ThemeProvider } from 'styled-components';
import { MemoryRouter } from 'react-router';
import createClient from './client';
import { merge } from 'lodash';
import theme from './ui/theme';
import React from 'react';

import config from '../config';

/*
 * The same render method of 'react-testing-library' but wrapped with
 * Redux store Provider and 'styled-components' ThemeProvider.
 *
 * element      : the React element you want to wrap
 * initialState : optional state to initialize the Redux store
 *
 * Returns the same object as render() but with an additional 'store'
 * property useful for dispatching actions inside your tests.
 */
export function reduxRender(element, initialState) {
  const client = createClient(config, createStore, initialState);

  const renderResult = render(
    <ClientProvider value={client}>
      <Provider store={client.store}>
        <ThemeProvider theme={theme}>
          <React.Fragment>{element}</React.Fragment>
        </ThemeProvider>
      </Provider>
    </ClientProvider>
  );
  return { ...renderResult, store: client.store };
}

/*
 * The same as reduxRender() above but also wrapped with 'react-router'
 * in order to test navigation flows.
 *
 * Returns the same object as reduxRender()
 */
export function routerRender(element, initialState, initialRoute = '/') {
  return reduxRender(
    <MemoryRouter initialEntries={[initialRoute]}>{element}</MemoryRouter>,
    initialState
  );
}

/*
 * Declaratively test complex form validations
 *
 * getter   : function that receives a data-testid and returns an HTML element
 * formId   : data-testid of the form element
 * formData : object of type { [data-testid]: any } to populate the form
 * errors   : object of type { [data-testid]: String } with expected errors
 */
export function testValidation(getter, formId, { formData, errors }) {
  // Populate the specified form fields
  Object.keys(formData).forEach((fieldId, i) => {
    const field = getter(fieldId);
    field.value = formData[fieldId];
    Simulate.change(field);
  });

  // Submit the form
  Simulate.submit(getter(formId));

  // Check expected errors
  Object.keys(errors).forEach((fieldId, i) => {
    expect(getter(`${fieldId}-error`).textContent).toEqual(
      expect.stringContaining(errors[fieldId])
    );
  });
}

/*
 * Add dataset property (data-foo attribute values) to nodes
 *
 * This hack is required because jsdom has no support for dataset yet
 * which prevents us from testing components that rely on it.
 * See https://github.com/jsdom/jsdom/issues/961
 *
 * element   : DOM node to add the dataset to
 * dataAttrs : attribute names to add to the node
 *
 * Usage:
 * <button data-testid="btn" data-modal="buy" data-foo="bar">Click</button>
 *
 * const btn = withDataset(getByTestId('btn'), 'modal', 'foo')
 * // btn.dataset === { modal: 'buy', foo: 'bar'}
 *
 */
export function withDataset(element, ...dataAttrs) {
  element.dataset = dataAttrs.reduce((acc, attr) => {
    acc[attr] = element.getAttribute(`data-${attr}`);
    return acc;
  }, {});
  return element;
}

/**
 * Return UTC timestamps relative to the current time
 * Useful for setting up different auction scenarios.
 */
export const twoWeeksAgo = () => Date.now() / 1000 - 60 * 60 * 24 * 7 * 2;
export const oneHourAgo = () => Date.now() / 1000 - 60 * 60;
export const inOneHour = () => Date.now() / 1000 + 60 * 60;
export const inOneWeek = () => Date.now() / 1000 + 60 * 60 * 24 * 7;

/*
 * Returns a common initial state for Redux store that is useful for most tests
 * Accepts an optional object with overrides that will be deeply merged with
 * the base state object.
 */
export function getInitialState(overrides = {}) {
  const baseState = {
    config,
    connectivity: { isOnline: true },
    chain: { height: 1, gasPrice: config.DEFAULT_GAS_PRICE },
    session: { isLoggedIn: true },
    rates: { ETH: { token: 'ETH', price: 1 } },
    wallet: {
      syncStatus: 'up-to-date',
      isActive: true,
      address: '0x15dd2028C976beaA6668E286b496A518F457b5Cf',
      ethBalance: '5000000000000000000000',
      transactions: [],
      token: {
        contract: config.LMR_TOKEN_ADDR,
        lmrBalance: '5000000000000000000000',
        transactions: [],
        symbol: 'LMR'
      }
    }
  };
  return merge({}, baseState, overrides);
}

/*
 * A sample transaction object. Useful for populating state in tests.
 */
export function getDummyTransaction() {
  return {
    meta: {
      lumerin: {},
      token: {
        [config.LMR_TOKEN_ADDR]: {
          event: 'Transfer',
          processing: false,
          value:
            '0x000000000000000000000000000000000000000000000001e5b8fa8fe2ac0000'
        }
      }
    },
    receipt: {
      blockHash: '0x1234567890abcdef',
      blockNumber: 407574,
      cumulativeGasUsed: 38484,
      gasUsed: 38484,
      transactionHash: '0x1234567890abcdef',
      transactionIndex: 0
    },
    transaction: {
      transactionIndex: 0,
      blockNumber: 407574,
      blockHash: '0x1234567890abcdef',
      hash: '0x1234567890abcdef',
      value: '0',
      gasPrice: '5000000000',
      gas: 76968,
      from: '0x15dd2028C976beaA6668E286b496A518F457b5Cf',
      to: '0xde806D6efD432CDeE42573760682D99eDEdC1d89'
    }
  };
}
