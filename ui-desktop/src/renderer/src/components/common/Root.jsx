import { withClient } from '../../store/hocs/clientContext'
import selectors from '../../store/selectors'
import { connect } from 'react-redux'
import React from 'react'
import { ToastsContext } from '../../components/toasts'

class Root extends React.Component {
  // static propTypes = {
  //   OnboardingComponent: PropTypes.func.isRequired,
  //   LoadingComponent: PropTypes.func.isRequired,
  //   RouterComponent: PropTypes.func.isRequired,
  //   isSessionActive: PropTypes.bool.isRequired,
  //   LoginComponent: PropTypes.func.isRequired,
  //   hasEnoughData: PropTypes.bool.isRequired,
  //   dispatch: PropTypes.func.isRequired,
  //   client: PropTypes.shape({
  //     onOnboardingCompleted: PropTypes.func.isRequired,
  //     onLoginSubmit: PropTypes.func.isRequired,
  //     onInit: PropTypes.func.isRequired
  //   }).isRequired
  // };

  static contextType = ToastsContext

  state = {
    onboardingComplete: null
  }

  componentDidMount() {
    this.props.client
      .onInit()
      .then(({ onboardingComplete, persistedState, config }) => {
        this.props.dispatch({
          type: 'initial-state-received',
          payload: { ...persistedState, config }
        })
        this.setState({ onboardingComplete })
      })
      .then(() => {
        if (this.props.isAuthBypassed) {
          // TODO: replace dummy password
          this.props.client
            .onLoginSubmit({ password: 'password' })
            .then(() => this.props.dispatch({ type: 'session-started' }))
            .catch((e) => {
              this.context.toast('error', 'Bypass auth failed')
            })
        }
      })
      .then(() => this.props.client.getDefaultCurrencySetting())
      .then((defaultCurr) => {
        this.props.dispatch({
          type: 'set-seller-currency',
          payload: defaultCurr || this.props.sellerDefaultCurrency || 'BTC'
        })
      })
      .then(() => this.props.client.getMarketplaceFee({}))
      .then((fee) =>
        this.props.dispatch({
          type: 'set-marketplace-fee',
          payload: fee
        })
      )
      // eslint-disable-next-line no-console
      .catch((e) => {
        console.error('root component error', e.message)
        this.context.toast(
          'error',
          'Failed to startup wallet. Please wait a few minutes and try again'
        )
      })
  }

  onOnboardingCompleted = ({ password, mnemonic, proxyRouterConfig }) => {
    return (
      this.props.client
        .onOnboardingCompleted({ password, mnemonic, proxyRouterConfig })
        .then(() => {
          this.setState({ onboardingComplete: true })
          this.props.dispatch({ type: 'session-started' })
        })
        // eslint-disable-next-line no-console
        .catch((e) => {
          this.context.toast(
            'error',
            'Failed to finish onboarding. Please wait a few minutes and try again'
          )
        })
    )
  }

  onLoginSubmit = ({ password }) =>
    this.props.client
      .onLoginSubmit({ password })
      .then(() => this.props.dispatch({ type: 'session-started' }))

  render() {
    console.log('Root component render')
    const {
      OnboardingComponent,
      LoadingComponent,
      RouterComponent,
      isSessionActive,
      LoginComponent,
      hasEnoughData
    } = this.props

    const { onboardingComplete } = this.state

    if (onboardingComplete === null) return null

    // eslint-disable-next-line no-negated-condition
    return !onboardingComplete ? (
      <OnboardingComponent onOnboardingCompleted={this.onOnboardingCompleted} />
    ) : // eslint-disable-next-line no-negated-condition
    !isSessionActive ? (
      <LoginComponent onLoginSubmit={this.onLoginSubmit} />
    ) : hasEnoughData ? (
      <RouterComponent />
    ) : (
      <LoadingComponent />
    )
  }
}

const mapStateToProps = (state) => ({
  isSessionActive: selectors.isSessionActive(state),
  hasEnoughData: selectors.hasEnoughData(state),
  isAuthBypassed: selectors.getIsAuthBypassed(state),
  sellerDefaultCurrency: selectors.getSellerDefaultCurrency(state)
})

export default connect(mapStateToProps)(withClient(Root))
