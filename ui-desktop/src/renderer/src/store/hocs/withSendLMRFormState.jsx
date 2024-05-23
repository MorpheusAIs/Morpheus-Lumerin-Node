import * as validators from '../validators';
import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';
import * as utils from '../utils';
import debounce from 'lodash/debounce';
import React, { useState } from 'react';

const withSendLMRFormState = Component => {
  const WrappedComponent = ({
    lmrDefaultGasLimit,
    mainTokenAddress,
    chainGasPrice,
    availableLMR,
    walletId,
    client,
    from
  }) => {
    const props = {
      lmrDefaultGasLimit,
      mainTokenAddress,
      chainGasPrice,
      availableLMR,
      walletId,
      client,
      from
    };

    const displayName = `withSendLMRFormState(${Component.displayName ||
      Component.name})`;

    const [gasEstimateError, setGasEstimateError] = useState(false);
    const [useCustomGas, setUseCustomGas] = useState(false);
    const [gasPrice, setGasPrice] = useState(
      client.fromWei(this.props.chainGasPrice, 'gwei')
    );
    const [gasLimit, setGasLimit] = useState(lmrDefaultGasLimit);
    const [errors, setErrors] = useState({});

    const [inputs, setInputs] = useState({ toAddress: null, lmrAmount: null });

    const state = {
      gasEstimateError,
      useCustomGas,
      gasPrice,
      gasLimit,
      errors,
      ...inputs
    };

    const resetForm = () => {
      setGasEstimateError(false);
      setUseCustomGas(false);
      setGasPrice(client.fromWei(this.props.chainGasPrice, 'gwei'));
      setGasLimit(lmrDefaultGasLimit);
      setErrors({});
      setInputs({ toAddress: null, lmrAmount: null });
    };

    const onInputChange = ({ id, value }) => {
      setGasEstimateError(id === 'gasLimit' ? false : gasEstimateError);
      setErrors({ errors, [id]: null }), setInputs({ ...inputs, [id]: value });

      // Estimate gas limit again if parameters changed
      if (['toAddress', 'lmrAmount'].includes(id)) getGasEstimate();
    };

    const getGasEstimate = debounce(() => {
      if (
        !client.isAddress(inputs.toAddress) ||
        !utils.isWeiable(client, inputs.lmrAmount)
      ) {
        return;
      }

      client
        .getTokenGasLimit({
          value: client.toWei(utils.sanitize(inputs.lmrAmount)),
          token: mainTokenAddress,
          chain,
          from,
          to: inputs.toAddress
        })
        .then(({ gasLimit }) => {
          setGasEstimateError(false);
          setGasLimit(gasLimit.toString());
        })
        .catch(() => setGasEstimateError(true));
    }, 500);

    const onSubmit = () =>
      client.sendLmr({
        gasPrice: client.toWei(gasPrice, 'gwei'),
        walletId,
        value: client.toWei(utils.sanitize(inputs.lmrAmount)),
        chain,
        from,
        gas: gasLimit,
        to: inputs.toAddress
      });

    const validate = () => {
      const max = client.fromWei(availableLMR);
      const errors = {
        ...validators.validateToAddress(client, inputs.toAddress),
        ...validators.validateLmrAmount(client, inputs.lmrAmount, max),
        ...validators.validateGasPrice(client, gasPrice),
        ...validators.validateGasLimit(client, gasLimit)
      };
      const hasErrors = Object.keys(errors).length > 0;

      if (hasErrors) setErrors(errors);

      return !hasErrors;
    };

    const onMaxClick = () => {
      const lmrAmount = client.fromWei(availableLMR);

      onInputChange({ id: 'lmrAmount', value: lmrAmount });
    };

    const amountFieldsProps = utils.getAmountFieldsProps({
      lmrAmount: inputs.lmrAmount
    });

    return (
      <Component
        onInputChange={onInputChange}
        onMaxClick={onMaxClick}
        resetForm={resetForm}
        onSubmit={onSubmit}
        lmrPlaceholder={amountFieldsProps.lmrPlaceholder}
        lmrAmount={amountFieldsProps.lmrAmount}
        validate={validate}
        {...props}
        {...state}
      />
    );
  };

  const mapStateToProps = state => ({
    lmrDefaultGasLimit: selectors.getChainConfig(state).lmrDefaultGasLimit,
    mainTokenAddress: selectors.getChainConfig(state).mainTokenAddress,
    chainGasPrice: selectors.getChainGasPrice(state),
    availableLMR: selectors.getLmrBalanceWei(state),
    from: selectors.getWalletAddress(state),
    symbol: selectors.getCoinSymbol(state)
  });

  return connect(mapStateToProps)(withClient(WrappedComponent));
};

export default withSendLMRFormState;
