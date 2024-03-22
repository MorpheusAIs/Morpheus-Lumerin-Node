import React, { useEffect, useState } from 'react';
import withCreateContractModalState from '../../../../store/hocs/withCreateContractModalState';
import { withRouter } from 'react-router-dom';
import { useForm } from 'react-hook-form';

import { PurchaseFormModalPage } from './PurchaseFormModalPage';
import { PurchasePreviewModalPage } from './PurchasePreviewModalPage';
import { toRfc2396 } from '../../../../utils';
import { PurchaseSuccessPage } from './PurchaseSuccessPage';
import Modal from '../Modal';

function PurchaseContractModal(props) {
  const {
    isActive,
    handlePurchase,
    close,
    contract,
    explorerUrl,
    portCheckErrorLink,
    lmrRate,
    history,
    pool,
    showSuccess,
    symbol,
    marketplaceFee,
    isProxyPortPublic
  } = props;

  const [isPreview, setIsPreview] = useState(false);
  const [isPurchasing, setIsPurchasing] = useState(false);

  const {
    register,
    handleSubmit,
    formState,
    setValue,
    getValues,
    reset,
    trigger
  } = useForm({ mode: 'onChange' });

  useEffect(() => {
    setValue('address', `${props.ip}:${props.buyerPort}`);
    trigger('address');

    setValue('worker', contract?.id);
    trigger('worker');
  }, [contract]);

  useEffect(() => {
    setValue('address', `${props.ip}:${props.buyerPort}`);
    trigger('address');
  }, [isActive]);

  const handleClose = e => {
    reset();
    setIsPreview(false);
    close(e);
  };

  const wrapHandlePurchase = async () => {
    setIsPurchasing(true);
    const inputs = getValues();
    await handlePurchase(
      inputs,
      contract,
      toRfc2396(inputs.address, inputs.worker)
    );
    setIsPurchasing(false);
  };

  const onEditPool = () => {
    history.push('/tools');
  };

  if (!isActive) {
    return <></>;
  }

  const pagesProps = {
    explorerUrl,
    onEditPool,
    inputs: getValues(),
    pool,
    contract,
    rate: lmrRate,
    marketplaceFee
  };

  return (
    <Modal onClose={handleClose}>
      {showSuccess ? (
        <PurchaseSuccessPage
          close={handleClose}
          contractId={contract.id}
          price={contract.price}
          symbol={symbol}
        />
      ) : isPreview ? (
        <PurchasePreviewModalPage
          {...pagesProps}
          isPurchasing={isPurchasing}
          onBackToForm={() => setIsPreview(false)}
          onPurchase={wrapHandlePurchase}
          symbol={symbol}
          isProxyPortPublic={isProxyPortPublic}
          portCheckErrorLink={portCheckErrorLink}
        />
      ) : (
        <PurchaseFormModalPage
          {...pagesProps}
          close={close}
          register={register}
          handleSubmit={handleSubmit}
          formState={formState}
          onFinished={() => setIsPreview(true)}
          symbol={symbol}
        />
      )}
    </Modal>
  );
}

export default withRouter(withCreateContractModalState(PurchaseContractModal));
