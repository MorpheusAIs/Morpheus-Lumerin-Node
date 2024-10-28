import * as validators from '../validators';
import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import * as utils from '../utils';
import PropTypes from 'prop-types';
import debounce from 'lodash/debounce';
import React from 'react';
import { lmrDecimals, ethDecimals } from '../../utils/coinValue';

const withTransactionModalState = WrappedComponent => {
  class Container extends React.Component {
    // static propTypes = {
    //   chainGasPrice: PropTypes.string.isRequired,
    //   availableCoin: PropTypes.string.isRequired,
    //   coinSymbol: PropTypes.string.isRequired,
    //   coinPrice: PropTypes.number.isRequired,
    //   walletId: PropTypes.oneOfType([PropTypes.string, PropTypes.number])
    //     .isRequired,
    //   client: PropTypes.shape({
    //     copyToClipboard: PropTypes.func.isRequired,
    //     isAddress: PropTypes.func.isRequired,
    //     sendCoin: PropTypes.func.isRequired,
    //     fromWei: PropTypes.func.isRequired,
    //     toWei: PropTypes.func.isRequired
    //   }).isRequired,
    //   // sendLmrFeatureStatus: PropTypes.oneOf(['no-funds', 'offline', 'ok'])
    //   // .isRequired,
    //   from: PropTypes.string.isRequired
    // }

    static displayName = `withTransactionModalState(${WrappedComponent.displayName ||
      WrappedComponent.name})`;

    rangeSelectOptions = [
      {
        label: this.props.symbol,
        value: 'LMR'
      },
      {
        label: this.props.symbolEth,
        value: 'ETH'
      }
    ];

    initialState = {
      copyBtnLabel: 'Copy to clipboard',
      gasEstimateError: false,
      useCustomGas: false,
      coinAmount: 0,
      usdAmount: 0,
      toAddress: '',
      estimatedFee: null,
      selectedCurrency: this.rangeSelectOptions[0],
      errors: {
        coinAmount: '',
        toAddress: '',
        gasLimit: '',
        gasPrice: ''
      }
    };

    state = this.initialState;

    resetForm = () => this.setState(this.initialState);

    setSelectedCurrency = e => {
      this.setState({ ...this.state, selectedCurrency: e });
      this.onInputChange({
        id: 'coinAmount',
        value: this.state.coinAmount,
        selectedCurrency: e
      });
    };

    onInputChange = ({ id, value, selectedCurrency }) => {
      const { client, lmrCoinPrice, ethCoinPrice } = this.props;
      const coinPrice =
        (selectedCurrency || this.state.selectedCurrency)?.value === 'LMR'
          ? lmrCoinPrice
          : ethCoinPrice;
      this.setState(state => {
        return {
          ...state,
          ...utils.syncAmounts({ state, coinPrice, id, value, client }),
          gasEstimateError: id === 'gasLimit' ? false : state.gasEstimateError,
          errors: { ...state.errors, [id]: null },
          [id]: utils.sanitizeInput(value)
        };
      });

      // Estimate gas limit again if parameters changed
      if (['coinAmount', 'toAddress'].includes(id)) {
        this.getGasEstimate();
      }
    };

    onSubmit = type => {
      const payload = {
        gasPrice: this.props.client.toWei(this.state.gasPrice, 'gwei'),
        walletId: this.props.walletId,
        value: utils.sanitize(this.state.coinAmount),
        chain: this.props.chain,
        from: this.props.from,
        gas: this.state.gasLimit,
        to: this.state.toAddress
      };
      return type === 'ETH'
        ? this.props.client.sendEth(payload)
        : this.props.client.sendLmr(payload);
    };

    validate = () => {
      const { coinAmount, toAddress, gasPrice, gasLimit } = this.state;
      const { client, lmrBalanceWei, ethBalanceWei } = this.props;
      const balance =
        this.state.selectedCurrency.value === 'LMR'
          ? lmrBalanceWei
          : ethBalanceWei;

      const errors = {
        ...validators.validateToAddress(client, toAddress),
        ...validators.validateCoinAmount(client, coinAmount, balance),
        ...validators.validateGasPrice(client, gasPrice),
        ...validators.validateGasLimit(client, gasLimit)
      };
      const hasErrors = Object.keys(errors).length > 0;
      if (hasErrors) this.setState({ errors });
      return hasErrors ? errors : false;
    };

    onMaxClick = () => {
      const coinAmount = this.props.client.fromWei(this.props.availableCoin);
      this.onInputChange({ id: 'coinAmount', value: coinAmount });
    };

    copyToClipboard = () => {
      this.props.client
        .copyToClipboard(this.props.address)
        .then(() => this.setState({ copyBtnLabel: 'Copied to clipboard!' }))
        .catch(err => this.setState({ copyBtnLabel: err.message }));
    };

    render() {
      const amountFieldsProps = utils.getAmountFieldsProps({
        coinAmount: this.state.coinAmount,
        usdAmount: this.state.usdAmount
      });
      const { sendLmrFeatureStatus, symbol } = this.props;

      const sendLmrDisabledReason =
        sendLmrFeatureStatus === 'no-funds'
          ? `You need some ${symbol} to send`
          : sendLmrFeatureStatus === 'offline'
          ? "Can't send while offline"
          : null;

      return (
        <WrappedComponent
          copyToClipboard={this.copyToClipboard}
          sendLmrDisabledReason={sendLmrDisabledReason}
          sendLmrDisabled={sendLmrFeatureStatus !== 'ok'}
          onInputChange={this.onInputChange}
          onMaxClick={this.onMaxClick}
          resetForm={this.resetForm}
          onSubmit={this.onSubmit}
          setSelectedCurrency={this.setSelectedCurrency}
          {...this.props}
          {...this.state}
          coinPlaceholder={amountFieldsProps.coinPlaceholder}
          usdPlaceholder={amountFieldsProps.usdPlaceholder}
          coinAmount={amountFieldsProps.coinAmount}
          usdAmount={amountFieldsProps.usdAmount}
          validate={this.validate}
        />
      );
    }
  }

  const mapStateToProps = state => ({
    address: selectors.getWalletAddress(state),
    explorerUrl: selectors.getContractExplorerUrl(state, {
      hash: selectors.getWalletAddress(state)
    }),
    // availableCoin: selectors.getCoinBalanceWei(state),
    coinSymbol: selectors.getCoinSymbol(state),
    lmrBalanceUSD: selectors.getWalletLmrBalanceUSD(state),
    lmrBalanceWei: selectors.getWalletLmrBalance(state),
    ethBalanceUSD: selectors.getWalletEthBalanceUSD(state),
    ethBalanceWei: selectors.getWalletEthBalance(state),
    lmrCoinPrice: selectors.getRate(state),
    ethCoinPrice: selectors.getRateEth(state),
    from: selectors.getWalletAddress(state),
    symbol: selectors.getCoinSymbol(state),
    symbolEth: selectors.getSymbolEth(state)
  });

  return connect(mapStateToProps)(withClient(Container));
};

export default withTransactionModalState;
