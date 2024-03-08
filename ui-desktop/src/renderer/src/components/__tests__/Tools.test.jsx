import * as testUtils from '../../testUtils'
import { Simulate } from 'react-testing-library'
import Tools from '../tools/Tools'
import React from 'react'
import 'react-testing-library/extend-expect'

const element = <Tools />

const VALID_MNEMONIC =
  'discover prepare cause retire pitch web curious own hollow total initial simple'

describe('<Tools/>', () => {
  describe('When submitting the form', () => {
    it('displays an error if MNEMONIC is not provided', () => {
      const { getByTestId } = testUtils.routerRender(element)
      testUtils.testValidation(getByTestId, 'recover-form', {
        formData: { 'mnemonic-field': '' },
        errors: { 'mnemonic-field': 'The phrase is required' }
      })
    })

    it('displays an error if MNEMONIC is invalid', () => {
      const { getByTestId } = testUtils.routerRender(element)
      testUtils.testValidation(getByTestId, 'recover-form', {
        formData: { 'mnemonic-field': 'foo' },
        errors: {
          'mnemonic-field':
            "These words don't look like a valid recovery phrase"
        }
      })
    })

    it('displays the confirmation view if there are no errors', () => {
      const { queryByTestId, getByTestId } = testUtils.routerRender(element)
      expect(queryByTestId('confirmation')).not.toBeInTheDOM()
      const mnemonicField = getByTestId('mnemonic-field')
      mnemonicField.value = VALID_MNEMONIC
      Simulate.change(mnemonicField)
      Simulate.submit(getByTestId('recover-form'))
      expect(queryByTestId('confirmation')).toBeInTheDOM()
    })
  })
})
