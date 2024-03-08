import { layout as element } from '../Router'
import * as testUtils from '../../testUtils'
import { Simulate } from 'react-testing-library'

describe('<Router/>', () => {
  test('navigates to WALLET screen when clicking WALLET sidebar button', () => {
    clickAndExpect('wallets-nav-btn', 'dashboard-container', '/auction')
  })

  test('navigates to AUCTION screen when clicking AUCTION sidebar button', () => {
    clickAndExpect('auction-nav-btn', 'auction-container')
  })

  test('navigates to CONVERTER screen when clicking CONVERTER sidebar button', () => {
    clickAndExpect('converter-nav-btn', 'converter-container')
  })

  test('navigates to TOOLS screen when clicking TOOLS sidebar button', () => {
    clickAndExpect('tools-nav-btn', 'tools-container')
  })

  test('redirects to WALLET route for / route', () => {
    const { queryByTestId } = testUtils.routerRender(
      element,
      testUtils.getInitialState(),
      '/'
    )
    expect(queryByTestId('dashboard-container')).not.toBeNull()
  })
})

function clickAndExpect(linkTestId, pageTestId, initialRoute) {
  const { getByTestId, queryByTestId } = testUtils.routerRender(
    element,
    testUtils.getInitialState(),
    initialRoute
  )
  expect(queryByTestId(pageTestId)).toBeNull()
  Simulate.click(getByTestId(linkTestId), { button: 0 })
  expect(queryByTestId(pageTestId)).not.toBeNull()
}
