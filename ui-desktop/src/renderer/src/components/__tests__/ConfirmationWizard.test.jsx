import { Simulate, flushPromises } from 'react-testing-library'
import ConfirmationWizard from '../common/ConfirmationWizard'
import * as testUtils from '../../testUtils'
import React from 'react'

const INVALID_PASSWORD = 'wrong!'
const VALID_PASSWORD = 'foo'

const onWizardSubmit = jest.fn(pass =>
  pass === VALID_PASSWORD ? Promise.resolve() : Promise.reject(new Error())
)

const failValidation = jest.fn(() => false)
const passValidation = jest.fn(() => true)

const renderForm = goToReview => (
  <button onClick={goToReview} type="button" data-testid="review-btn" />
)

const renderConfirmation = () => <div data-testid="confirmation" />

const getElement = validate => (
  <ConfirmationWizard
    renderConfirmation={renderConfirmation}
    onWizardSubmit={onWizardSubmit}
    renderForm={renderForm}
    validate={validate}
  />
)

describe('<ConfirmationWizard/>', () => {
  it('initially displays the provided form', () => {
    const { queryByTestId } = testUtils.reduxRender(getElement())
    expect(queryByTestId('review-btn')).not.toBeNull()
  })

  it('displays the confirmation view after submitting the form', () => {
    const { queryByTestId, getByTestId } = testUtils.reduxRender(getElement())
    expect(queryByTestId('confirmation')).toBeNull()
    goToConfirmationView(getByTestId)
    expect(queryByTestId('confirmation')).not.toBeNull()
  })

  it('returns to form if confirmation is rejected', () => {
    const { queryByTestId, getByTestId } = testUtils.reduxRender(getElement())
    goToConfirmationView(getByTestId)
    expect(queryByTestId('review-btn')).toBeNull()
    Simulate.click(getByTestId('cancel-btn'))
    expect(queryByTestId('review-btn')).not.toBeNull()
  })

  describe('When submitting the confirmation form', () => {
    it.skip('displays an error if password is not provided', () => {
      const { getByTestId } = testUtils.reduxRender(getElement())
      goToConfirmationView(getByTestId)
      testUtils.testValidation(getByTestId, 'confirm-form', {
        formData: { 'pass-field': '' },
        errors: { 'pass-field': 'Password is required' }
      })
    })

    it('calls onWizardSubmit with password as an argument', () => {
      const { getByTestId } = testUtils.reduxRender(getElement())
      goToConfirmationView(getByTestId)
      confirm(getByTestId, VALID_PASSWORD)
      expect(onWizardSubmit).toHaveBeenCalledWith(VALID_PASSWORD)
    })

    it('displays a WAITING message while the call is on-flight', () => {
      const { getByTestId, queryByTestId } = testUtils.reduxRender(getElement())
      goToConfirmationView(getByTestId)
      expect(queryByTestId('waiting')).toBeNull()
      confirm(getByTestId, VALID_PASSWORD)
      expect(queryByTestId('waiting')).not.toBeNull()
    })

    it('displays a SUCCESS message if the call resolved', async () => {
      const { getByTestId, queryByTestId } = testUtils.reduxRender(getElement())
      goToConfirmationView(getByTestId)
      confirm(getByTestId, VALID_PASSWORD)
      expect(queryByTestId('success')).toBeNull()
      await flushPromises()
      expect(queryByTestId('success')).not.toBeNull()
    })

    it('displays a FAILURE message if the call was rejected', async () => {
      const { getByTestId, queryByTestId } = testUtils.reduxRender(getElement())
      goToConfirmationView(getByTestId)
      confirm(getByTestId, INVALID_PASSWORD)
      expect(queryByTestId('failure')).toBeNull()
      await flushPromises()
      expect(queryByTestId('failure')).not.toBeNull()
    })

    it('returns to form if call failed and user clicks TRY AGAIN', async () => {
      const { getByTestId, queryByTestId } = testUtils.reduxRender(getElement())
      goToConfirmationView(getByTestId)
      confirm(getByTestId, INVALID_PASSWORD)
      await flushPromises()
      expect(queryByTestId('review-btn')).toBeNull()
      Simulate.click(getByTestId('try-again-btn'))
      expect(queryByTestId('review-btn')).not.toBeNull()
    })
  })

  describe('If a custom validation function is provided', () => {
    it('will be called when the form is submitted', () => {
      const { getByTestId } = testUtils.reduxRender(getElement(passValidation))
      expect(passValidation).toHaveBeenCalledTimes(0)
      Simulate.click(getByTestId('review-btn'))
      expect(passValidation).toHaveBeenCalledTimes(1)
    })

    it('displays the confirmation view if validate() returns TRUE', () => {
      const { queryByTestId, getByTestId } = testUtils.reduxRender(
        getElement(passValidation)
      )
      expect(queryByTestId('confirmation')).toBeNull()
      goToConfirmationView(getByTestId)
      expect(queryByTestId('confirmation')).not.toBeNull()
    })

    it('remains in form view if validate() returns FALSE', () => {
      const { queryByTestId, getByTestId } = testUtils.reduxRender(
        getElement(failValidation)
      )
      expect(queryByTestId('review-btn')).not.toBeNull()
      expect(queryByTestId('confirmation')).toBeNull()
      goToConfirmationView(getByTestId)
      expect(queryByTestId('review-btn')).not.toBeNull()
      expect(queryByTestId('confirmation')).toBeNull()
    })
  })
})

function goToConfirmationView(getByTestId) {
  Simulate.click(getByTestId('review-btn'))
}

function confirm(getByTestId, pass) {
  const passwordField = getByTestId('pass-field')
  passwordField.value = pass
  Simulate.change(passwordField)
  Simulate.submit(getByTestId('confirm-form'))
}
