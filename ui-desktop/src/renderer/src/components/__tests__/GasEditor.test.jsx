import testUtils from '../../testUtils'
import { Simulate } from 'react-testing-library'
import GasEditor from '../common/GasEditor'
import React from 'react'

const element = (
  <GasEditor
    onInputChange={jest.fn()}
    useCustomGas
    gasLimit="21000"
    gasPrice="1"
    errors={{}}
  />
)

describe('<GasEditor/>', () => {
  it('should match its snapshot', () => {
    const { container } = testUtils.reduxRender(element)
    expect(container).toMatchSnapshot()
  })
})

export function runValidateTests(el, initialState, formTestId) {
  it('displays an error if GAS LIMIT is not provided', async () => {
    const { getByTestId } = testUtils.reduxRender(el, initialState)
    Simulate.click(getByTestId('edit-gas-btn'))
    testUtils.testValidation(getByTestId, formTestId, {
      formData: { 'gas-limit-field': '' },
      errors: { 'gas-limit-field': 'Gas limit is required' }
    })
  })

  it.skip('displays an error if GAS LIMIT is an invalid value', async () => {
    const { getByTestId } = testUtils.reduxRender(element, initialState)
    Simulate.click(getByTestId('edit-gas-btn'))
    testUtils.testValidation(getByTestId, formTestId, {
      formData: { 'gas-limit-field': 'foo' },
      errors: { 'gas-limit-field': 'Invalid value' }
    })
  })

  it.skip('displays an error if GAS LIMIT is not an integer', async () => {
    const { getByTestId } = testUtils.reduxRender(element, initialState)
    Simulate.click(getByTestId('edit-gas-btn'))
    testUtils.testValidation(getByTestId, formTestId, {
      formData: { 'gas-limit-field': '1.1' },
      errors: { 'gas-limit-field': 'Gas limit must be an integer' }
    })
  })

  it.skip('displays an error if GAS LIMIT is negative', async () => {
    const { getByTestId } = testUtils.reduxRender(element, initialState)
    Simulate.click(getByTestId('edit-gas-btn'))
    testUtils.testValidation(getByTestId, formTestId, {
      formData: { 'gas-limit-field': '-1' },
      errors: { 'gas-limit-field': 'Gas limit must be greater than 0' }
    })
  })

  it.skip('displays an error if GAS PRICE is not provided', async () => {
    const { getByTestId } = testUtils.reduxRender(element, initialState)
    Simulate.click(getByTestId('edit-gas-btn'))
    testUtils.testValidation(getByTestId, formTestId, {
      formData: { 'gas-price-field': '' },
      errors: { 'gas-price-field': 'Gas price is required' }
    })
  })

  it.skip('displays an error if GAS PRICE is an invalid value', async () => {
    const { getByTestId } = testUtils.reduxRender(element, initialState)
    Simulate.click(getByTestId('edit-gas-btn'))
    testUtils.testValidation(getByTestId, formTestId, {
      formData: { 'gas-price-field': 'foo' },
      errors: { 'gas-price-field': 'Invalid value' }
    })
  })

  it.skip('displays an error if GAS PRICE is lower than 1', async () => {
    const { getByTestId } = testUtils.reduxRender(element, initialState)
    Simulate.click(getByTestId('edit-gas-btn'))
    testUtils.testValidation(getByTestId, formTestId, {
      formData: { 'gas-price-field': '0' },
      errors: { 'gas-price-field': 'Gas price can not be lower than 1' }
    })
  })
}
