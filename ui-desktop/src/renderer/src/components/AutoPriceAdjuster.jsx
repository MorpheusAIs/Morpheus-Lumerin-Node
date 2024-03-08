import React, { useContext } from 'react';

import withContractsState from '../store/hocs/withContractsState';
import { ToastsContext } from './toasts';
import { lmrDecimals } from '../utils/coinValue';
import { formatBtcPerTh, calculateSuggestedPrice } from './contracts/utils';

import useInterval from 'use-interval';

function AutoPriceAdjuster({
  contracts,
  address,
  client,
  networkDifficulty,
  lmrCoinPrice,
  btcCoinPrice,
  autoAdjustPriceInterval,
  autoAdjustContractPriceTimeout
}) {
  const context = useContext(ToastsContext);

  const adjustContractPrices = async (
    allContracts,
    lmrRate,
    btcRate,
    sellerAddress,
    networkDifficulty
  ) => {
    const profitSettings = await client.getProfitSettings();
    const autoAdjustSettings = await client.getAutoAdjustPriceData();

    const contractsWithEnabledAutoAdjust = allContracts.filter(
      c =>
        c.seller === sellerAddress &&
        !c.isDead &&
        autoAdjustSettings[c.id?.toLowerCase()]?.enabled
    );

    const reward = formatBtcPerTh(networkDifficulty);
    const deviation = +profitSettings.deviation;

    const result = contractsWithEnabledAutoAdjust.reduce((curr, contract) => {
      const contractProfitTarget =
        +contract.futureTerms?.profitTarget || +contract.profitTarget;
      const profitTarget =
        contractProfitTarget !== 0
          ? contractProfitTarget
          : profitSettings?.adaptExisting
          ? +profitSettings?.target
          : 0;

      if (profitTarget === 0) {
        return curr;
      }

      const speed = (contract.futureTerns?.speed || contract.speed) / 10 ** 12;
      const length = (contract.futureTerns?.length || contract.length) / 3600;
      const price =
        (contract.futureTerms?.price || contract.price) / lmrDecimals;

      const profitTargetPercent = profitTarget / 100;
      const deviationPercent = deviation / 100;

      const left =
        1 +
        (+profitTarget > 0
          ? deviationPercent - profitTargetPercent
          : profitTargetPercent - deviationPercent) /
          100;
      const right = 1 + deviationPercent + profitTargetPercent;

      const estimatedLeft = calculateSuggestedPrice(
        length,
        speed,
        btcRate,
        lmrRate,
        reward,
        left
      );
      const estimatedRight = calculateSuggestedPrice(
        length,
        speed,
        btcRate,
        lmrRate,
        reward,
        right
      );
      const targetEstimate = calculateSuggestedPrice(
        length,
        speed,
        btcRate,
        lmrRate,
        reward,
        1 + profitTargetPercent
      );

      const isPriceWithinRange =
        estimatedLeft <= price && estimatedRight >= price;

      if (!isPriceWithinRange) {
        curr.push({
          ...contract,
          newPrice: targetEstimate
        });
      }
      return curr;
    }, []);

    for (const contract of result) {
      const autoAdjustContractSettings =
        autoAdjustSettings[contract.id.toLowerCase()];
      const lastUpdatedAt = autoAdjustContractSettings?.lastUpdatedAt;

      if (
        !lastUpdatedAt ||
        lastUpdatedAt + autoAdjustContractPriceTimeout < Date.now()
      ) {
        await client
          .editContract({
            id: contract.id,
            price: (contract.newPrice * lmrDecimals).toString(),
            speed: contract.speed,
            duration: contract.length,
            sellerAddress: contract.seller,
            profit: contract.profitTarget
          })
          .then(() => {
            client.setAutoAdjustPriceData({
              [contract.id.toLowerCase()]: {
                ...autoAdjustContractSettings,
                lastUpdatedAt: Date.now()
              }
            });
          });
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    }
  };

  useInterval(() => {
    if (contracts.length) {
      adjustContractPrices(
        contracts,
        lmrCoinPrice,
        btcCoinPrice,
        address,
        networkDifficulty
      ).catch(e => {
        context.toast('error', `Failed to auto adjust prices: ${e.message}`);
      });
    }
  }, autoAdjustPriceInterval);

  return <></>;
}

export default withContractsState(AutoPriceAdjuster);
