import * as testUtils from '../../testUtils';
import ReceiveDrawer from '../dashboard/ReceiveDrawer';
import { Simulate } from 'react-testing-library';
import React from 'react';
import 'react-testing-library/extend-expect';

const initialState = testUtils.getInitialState();

const element = <ReceiveDrawer onRequestClose={jest.fn()} isOpen />;

describe('<ReceiveDrawer/>', () => {
  it('matches its snapshot', () => {
    const { container } = testUtils.reduxRender(element, initialState);
    expect(container.firstChild).toMatchSnapshot();
  });

  it('displays the active address', () => {
    const { getByTestId } = testUtils.reduxRender(element, initialState);
    const address = Object.keys(initialState.wallet.address);
    expect(getByTestId('address')).toHaveTextContent(address);
  });

  it.skip('displays a message when the Copy button is clicked and the address copied', async () => {
    const { getByTestId, queryByTestId } = testUtils.reduxRender(
      element,
      initialState
    );
    expect(queryByTestId('btn-label')).toHaveTextContent('Copy');
    Simulate.click(getByTestId('copy-btn'));
    expect(queryByTestId('btn-label')).toHaveTextContent(
      'Copied to clipboard!'
    );
  });
});
