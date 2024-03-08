import * as testUtils from '../../testUtils'
import { Simulate } from 'react-testing-library'
import TxList from '../dashboard/tx-list/TxList'
import React from 'react'
import 'react-testing-library/extend-expect'

const ACTIVE_ADDRESS = '0x15dd2028C976beaA6668E286b496A518F457b5Cf'

const element = <TxList syncStatus="up-to-date" hasTransactions />

describe('<TxList/>', () => {
  it('displays the receipt when clicking a transaction', () => {
    const { queryByTestId, getByTestId } = testUtils.reduxRender(
      element,
      getInitialState()
    )
    expect(getByTestId('tx-row')).toBeInTheDOM()
    expect(queryByTestId('receipt-modal')).not.toBeInTheDOM()
    const txRow = testUtils.withDataset(getByTestId('tx-row'), 'hash')
    Simulate.click(txRow)
    expect(queryByTestId('receipt-modal')).toBeInTheDOM()
  })
})

function getInitialState() {
  return testUtils.getInitialState({
    wallets: {
      byId: {
        foo: {
          addresses: {
            [ACTIVE_ADDRESS]: {
              transactions: [testUtils.getDummyTransaction()]
            }
          }
        }
      }
    }
  })
}
