import { reduxRender, testValidation } from '../../testUtils'
import { Simulate } from 'react-testing-library'
import Onboarding from '../onboarding/Onboarding'
import React from 'react'

const mockCallback = jest.fn(() => Promise.resolve())
const element = <Onboarding onOnboardingCompleted={mockCallback} />
const VALID_PASSWORD = 'asdasdasdasdasdasd'
const VALID_MNEMONIC =
  'discover prepare cause retire pitch web curious own hollow total initial simple'

describe('<Onboarding/>', () => {
  it('initially displays the terms and conditions form', () => {
    const { queryByTestId } = reduxRender(element)
    expect(queryByTestId('accept-terms-btn')).not.toBeNull()
  })

  it.skip('displays the password form after accepting terms', () => {
    const { getByTestId, queryByTestId } = reduxRender(element)
    expect(queryByTestId('pass-form')).toBeNull()
    acceptTerms(getByTestId)
    expect(queryByTestId('pass-form')).not.toBeNull()
  })

  describe.skip('When submitting the password form', () => {
    it('displays an error if password is not provided', () => {
      const { getByTestId } = reduxRender(element)
      acceptTerms(getByTestId)
      testValidation(getByTestId, 'pass-form', {
        formData: { 'pass-field': '' },
        errors: { 'pass-field': 'Password is required' }
      })
    })

    it('displays an error if password is weak', () => {
      const { getByTestId } = reduxRender(element)
      acceptTerms(getByTestId)
      testValidation(getByTestId, 'pass-form', {
        formData: { 'pass-field': '1' },
        errors: { 'pass-field': 'Password is not strong enough' }
      })
    })

    it('displays an error if password is not repeated', () => {
      const { getByTestId } = reduxRender(element)
      acceptTerms(getByTestId)
      testValidation(getByTestId, 'pass-form', {
        formData: { 'pass-field': VALID_PASSWORD },
        errors: { 'pass-again-field': 'Repeat the password' }
      })
    })

    it("displays an error if passwords don't match", () => {
      const { getByTestId } = reduxRender(element)
      acceptTerms(getByTestId)
      testValidation(getByTestId, 'pass-form', {
        formData: { 'pass-field': VALID_PASSWORD, 'pass-again-field': '1' },
        errors: { 'pass-again-field': "Passwords don't match" }
      })
    })

    it('displays the mnemonic if there are no errors', () => {
      const { getByTestId, queryByTestId } = reduxRender(element)
      expect(queryByTestId('mnemonic-form')).toBeNull()
      acceptTerms(getByTestId)
      createPassword(getByTestId)
      expect(queryByTestId('mnemonic-form')).not.toBeNull()
    })
  })

  describe('If creating a new wallet', () => {
    it.skip('displays verify mnemonic form after accepting given mnemonic', () => {
      const { getByTestId, queryByTestId } = reduxRender(element)
      acceptTerms(getByTestId)
      createPassword(getByTestId)
      expect(queryByTestId('mnemonic-field')).toBeNull()
      acceptMnemonic(getByTestId)
      expect(queryByTestId('mnemonic-field')).not.toBeNull()
    })

    it.skip('returns to mnemonic view when clicking cancel button', () => {
      const { getByTestId, queryByTestId } = reduxRender(element)
      acceptTerms(getByTestId)
      createPassword(getByTestId)
      acceptMnemonic(getByTestId)
      expect(queryByTestId('mnemonic-label')).toBeNull()
      Simulate.click(getByTestId('goback-btn'))
      expect(queryByTestId('mnemonic-label')).not.toBeNull()
    })

    describe.skip('when submitting the verify mnemonic form', () => {
      it("displays an error if mnemonics don't match", () => {
        const { getByTestId } = reduxRender(element)
        acceptTerms(getByTestId)
        createPassword(getByTestId)
        acceptMnemonic(getByTestId)
        testValidation(getByTestId, 'mnemonic-form', {
          formData: { 'mnemonic-field': 'foo' },
          errors: {
            'mnemonic-field':
              'The text provided does not match your recovery passphrase.'
          }
        })
      })

      it('calls onOnboardingCompleted with pass and mnemonic as arguments', () => {
        const { getByTestId } = reduxRender(element)
        acceptTerms(getByTestId)
        createPassword(getByTestId)
        const mnemonic = getByTestId('mnemonic-label').textContent
        acceptMnemonic(getByTestId)

        const mnemonicField = getByTestId('mnemonic-field')
        mnemonicField.value = mnemonic
        Simulate.change(mnemonicField)

        Simulate.submit(getByTestId('mnemonic-form'))

        expect(mockCallback).toHaveBeenCalledWith({
          password: VALID_PASSWORD,
          mnemonic
        })
      })
    })
  })

  describe.skip('If recovering a new wallet from mnemonic', () => {
    it('displays recovery mnemonic form after clicking recover button', () => {
      const { getByTestId, queryByTestId } = reduxRender(element)
      acceptTerms(getByTestId)
      createPassword(getByTestId)
      expect(queryByTestId('mnemonic-field')).toBeNull()
      recoverFromMnemonic(getByTestId)
      expect(queryByTestId('mnemonic-field')).not.toBeNull()
    })

    it('returns to mnemonic view when clicking cancel button', () => {
      const { getByTestId, queryByTestId } = reduxRender(element)
      acceptTerms(getByTestId)
      createPassword(getByTestId)
      recoverFromMnemonic(getByTestId)
      expect(queryByTestId('mnemonic-label')).toBeNull()
      Simulate.click(getByTestId('cancel-btn'))
      expect(queryByTestId('mnemonic-label')).not.toBeNull()
    })

    describe.skip('when submitting the recovery form', () => {
      it('displays an error if mnemonic is not provided', () => {
        const { getByTestId } = reduxRender(element)
        acceptTerms(getByTestId)
        createPassword(getByTestId)
        recoverFromMnemonic(getByTestId)
        testValidation(getByTestId, 'mnemonic-form', {
          formData: { 'mnemonic-field': '' },
          errors: { 'mnemonic-field': 'The phrase is required' }
        })
      })

      it('displays an error if mnemonic is invalid', () => {
        const { getByTestId } = reduxRender(element)
        acceptTerms(getByTestId)
        createPassword(getByTestId)
        recoverFromMnemonic(getByTestId)
        testValidation(getByTestId, 'mnemonic-form', {
          formData: { 'mnemonic-field': 'foo' },
          errors: {
            'mnemonic-field':
              "These words don't look like a valid recovery phrase"
          }
        })
      })

      it('calls onOnboardingCompleted with pass and mnemonic as arguments', () => {
        const { getByTestId } = reduxRender(element)
        acceptTerms(getByTestId)
        createPassword(getByTestId)
        recoverFromMnemonic(getByTestId)

        const mnemonicField = getByTestId('mnemonic-field')
        mnemonicField.value = VALID_MNEMONIC
        Simulate.change(mnemonicField)

        Simulate.submit(getByTestId('mnemonic-form'))

        expect(mockCallback).toHaveBeenCalledWith({
          password: VALID_PASSWORD,
          mnemonic: VALID_MNEMONIC
        })
      })
    })
  })
})

function acceptTerms(getByTestId) {
  Simulate.change(getByTestId('accept-terms-chb'))
  Simulate.change(getByTestId('accept-license-chb'))
  Simulate.click(getByTestId('accept-terms-btn'))
}

function createPassword(getByTestId) {
  const passField = getByTestId('pass-field')
  passField.value = VALID_PASSWORD
  Simulate.change(passField)
  const passAgainField = getByTestId('pass-again-field')
  passAgainField.value = VALID_PASSWORD
  Simulate.change(passAgainField)
  Simulate.submit(getByTestId('pass-form'))
}

function acceptMnemonic(getByTestId) {
  Simulate.click(getByTestId('copied-mnemonic-btn'))
}

function recoverFromMnemonic(getByTestId) {
  Simulate.click(getByTestId('recover-btn'))
}
