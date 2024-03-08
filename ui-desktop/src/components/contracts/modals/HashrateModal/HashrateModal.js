//@ts-check
import React, { useState } from 'react';
import Highcharts from 'highcharts';
import HighchartsReact from 'highcharts-react-official';
import {
  Modal,
  Body,
  TitleWrapper,
  Title,
  CloseModal
} from '../CreateContractModal.styles';
import { withClient } from '../../../../store/hocs/clientContext';
import { renderChart } from './chartRenderer';
import { useInterval } from '../../../hooks/useInterval';
import { roundTime } from './utils';

const UpdateIntervalMs = 10 * 1000; // how often data will be checked for updates
const TimeResolution = 5 * 60 * 1000; // how granular will be the chart data
const MaxDuration = 24 * 60 * 60 * 1000; // how far back in time will be the chart data

function HashrateModal({ isActive, close, contractId, client }) {
  const [chart, setChart] = useState([]);

  const handleClose = e => close(e);
  const handlePropagation = e => e.stopPropagation();

  useInterval(
    async () => {
      if (!contractId) return;

      const now = new Date();
      const fromDate = new Date(now.getTime() - MaxDuration);

      const storedHashrate = await client.getContractHashrate({
        contractId,
        fromDate
      });
      const chartData = mapInputDataToDataPoints(storedHashrate, fromDate, now);

      //@ts-ignore
      setChart(renderChart(chartData));
    },
    UpdateIntervalMs,
    true,
    [contractId]
  );

  if (!isActive) {
    return <></>;
  }

  return (
    <Modal onClick={handleClose}>
      <Body
        style={{ width: '100%', maxWidth: '80%' }}
        onClick={handlePropagation}
      >
        {CloseModal(handleClose)}
        <TitleWrapper style={{ height: 'auto' }}>
          <Title>Dashboard</Title>
        </TitleWrapper>
        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <div>Recent Hashrate (last 24 hours)</div>
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center'
            }}
          ></div>
        </div>
        <HighchartsReact highcharts={Highcharts} options={chart} />
      </Body>
    </Modal>
  );
}

/**
 * Maps input hashrate data to the chart data points filling the gaps with zero hashrate
 * @param {{hashrate: number, timestamp: number}[]} storedHashrate
 * @param {Date} fromDate
 * @param {Date} toDate
 * @returns {[number, number][]}
 */
function mapInputDataToDataPoints(storedHashrate, fromDate, toDate) {
  /** @type {[number, number][]} */
  const chartData = [];
  let sourceIndex = 0;
  const fromDateRounded = roundTime(fromDate, TimeResolution).getTime();
  const toDateRounded = roundTime(toDate, TimeResolution).getTime();

  for (let t = fromDateRounded; t < toDateRounded; t += TimeResolution) {
    if (sourceIndex < storedHashrate.length) {
      const source = storedHashrate[sourceIndex];
      const sourceRounded = roundTime(
        new Date(source.timestamp),
        TimeResolution
      );

      if (sourceRounded.getTime() === t) {
        chartData.push([t, source.hashrate]);
        sourceIndex++;
        continue;
      }
    }

    chartData.push([t, 0]);
  }

  return chartData;
}

export default withClient(HashrateModal);
