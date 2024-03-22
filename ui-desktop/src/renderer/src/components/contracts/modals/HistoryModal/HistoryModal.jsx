import React, { useEffect, useState } from 'react';
import { List as RVList, AutoSizer } from 'react-virtualized';
import {
  Modal,
  Body,
  TitleWrapper,
  Title,
  Subtitle,
  CloseModal
} from '../CreateContractModal.styles';
import HistoryRow from './HistoryRow';
import { withClient } from '../../../../store/hocs/clientContext';
import { lmrDecimals } from '../../../../utils/coinValue';

function HistroyModal(props) {
  const { isActive, close, historyContracts, client } = props;

  const handleClose = e => {
    close(e);
  };
  const handlePropagation = e => e.stopPropagation();

  const history = historyContracts
    .map(hc => hc.history)
    .flat()
    .map(h => {
      return {
        id: h.id,
        isSuccess: h[0],
        price: +h[3],
        startedAt: +h[1],
        finishedAt: +h[2],
        speed: +h[4],
        duration: +h[5],
        actualDuration: +h[2] - +h[1]
      };
    })
    .map(h => {
      return {
        ...h,
        payed: h.isSuccess ? h.price : (h.actualDuration / h.duration) * h.price
      };
    })
    .sort((a, b) => b.finishedAt - a.finishedAt);

  if (!isActive) {
    return <></>;
  }

  const rowRenderer = historyContracts => ({ key, index, style }) => (
    <HistoryRow
      key={historyContracts[index].id}
      contract={historyContracts[index]}
    />
  );

  return (
    <Modal onClick={handleClose}>
      <Body height={'400px'} onClick={handlePropagation}>
        {CloseModal(handleClose)}
        <TitleWrapper>
          <Title>Purchase history</Title>
        </TitleWrapper>
        <AutoSizer width={400}>
          {({ width, height }) => (
            <RVList
              rowRenderer={rowRenderer(history)}
              rowHeight={50}
              rowCount={history.length}
              height={height || 500} // defaults for tests
              width={width || 500} // defaults for tests
            />
          )}
        </AutoSizer>
      </Body>
    </Modal>
  );
}

export default withClient(HistroyModal);
