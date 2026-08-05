import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import React from 'react';
import { toBaseUnits, isValidAddress } from '../utils/amount';

// Send/receive state for the wallet transaction modal.
//
// This used to be Lumerin-era code: it offered an "LMR" currency, dispatched to
// `client.sendLmr` (a handler that does not exist anywhere in the app), and
// computed gas for a token this product doesn't use. The MOR path simply had no
// implementation — the proxy-router has exposed POST /blockchain/send/mor the
// whole time, but nothing in the UI ever called it. This is now a straight
// MOR/ETH send against that endpoint.

const withTransactionModalState = (WrappedComponent) => {
  class Container extends React.Component {
    static displayName = `withTransactionModalState(${
      WrappedComponent.displayName || WrappedComponent.name
    })`;

    currencyOptions = [
      { label: this.props.symbol || 'MOR', value: 'MOR' },
      { label: this.props.symbolEth || 'ETH', value: 'ETH' },
    ];

    initialState = {
      copyBtnLabel: 'Copy to clipboard',
      coinAmount: '',
      toAddress: '',
      txHash: null,
      selectedCurrency: this.currencyOptions[0],
      errors: { coinAmount: '', toAddress: '' },
    };

    state = this.initialState;

    resetForm = () => this.setState(this.initialState);

    setSelectedCurrency = (option) =>
      this.setState({ selectedCurrency: option, errors: {} });

    onInputChange = ({ id, value }) =>
      this.setState((state) => ({
        ...state,
        [id]: value,
        errors: { ...state.errors, [id]: null },
      }));

    /** Balance of the currently selected asset, as a decimal number. */
    getAvailableBalance = () =>
      this.state.selectedCurrency.value === 'ETH'
        ? Number(this.props.eth?.value ?? 0)
        : Number(this.props.mor?.value ?? 0);

    validate = () => {
      const { coinAmount, toAddress } = this.state;
      const errors = {};

      if (!isValidAddress(toAddress)) {
        errors.toAddress = 'Enter a valid 0x… address';
      }

      try {
        toBaseUnits(coinAmount);
        if (Number(coinAmount) > this.getAvailableBalance()) {
          errors.coinAmount = `Amount exceeds your ${this.state.selectedCurrency.label} balance`;
        }
      } catch (e) {
        errors.coinAmount = e.message;
      }

      // Sending the entire ETH balance leaves nothing for gas, so the
      // transaction is guaranteed to fail on-chain. Warn before the user pays
      // to find that out.
      if (
        !errors.coinAmount &&
        this.state.selectedCurrency.value === 'ETH' &&
        Number(coinAmount) >= this.getAvailableBalance()
      ) {
        errors.coinAmount = 'Leave some ETH to cover the network fee';
      }

      const hasErrors = Object.keys(errors).length > 0;
      if (hasErrors) {
        this.setState({ errors });
      }
      return hasErrors ? errors : false;
    };

    onSubmit = async () => {
      const amount = toBaseUnits(this.state.coinAmount);
      const to = this.state.toAddress.trim();

      const txHash =
        this.state.selectedCurrency.value === 'ETH'
          ? await this.props.client.sendEth({ to, amount })
          : await this.props.client.sendMor({ to, amount });

      this.setState({ txHash });
      return txHash;
    };

    onMaxClick = () => {
      const balance = this.getAvailableBalance();
      // Never offer a true "max" on ETH — see the gas note in validate().
      const value =
        this.state.selectedCurrency.value === 'ETH'
          ? String(Math.max(balance - 0.0005, 0))
          : String(balance);
      this.onInputChange({ id: 'coinAmount', value });
    };

    copyToClipboard = () => {
      this.props.client
        .copyToClipboard(this.props.address)
        .then(() => this.setState({ copyBtnLabel: 'Copied to clipboard!' }))
        .catch((err) => this.setState({ copyBtnLabel: err.message }));
    };

    render() {
      return (
        <WrappedComponent
          copyToClipboard={this.copyToClipboard}
          onInputChange={this.onInputChange}
          onMaxClick={this.onMaxClick}
          resetForm={this.resetForm}
          onSubmit={this.onSubmit}
          setSelectedCurrency={this.setSelectedCurrency}
          currencyOptions={this.currencyOptions}
          availableBalance={this.getAvailableBalance()}
          {...this.props}
          {...this.state}
          validate={this.validate}
        />
      );
    }
  }

  const mapStateToProps = (state) => ({
    address: selectors.getWalletAddress(state),
    explorerUrl: selectors.getContractExplorerUrl(state, {
      hash: selectors.getWalletAddress(state),
    }),
    from: selectors.getWalletAddress(state),
    symbol: selectors.getCoinSymbol(state),
    symbolEth: selectors.getSymbolEth(state),
    txUrlResolver: selectors.getTransactionExplorerUrlResolver(state),
  });

  return connect(mapStateToProps)(withClient(Container));
};

export default withTransactionModalState;
