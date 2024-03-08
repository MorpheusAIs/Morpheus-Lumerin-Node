import React, { useEffect, useContext } from 'react';
import styled from 'styled-components';

import withScanIndicatorState from '../../../store/hocs/withScanIndicatorState';
import { ToastsContext } from '../../toasts';
import Spinner from '../../common/Spinner';

const Container = styled.div`
  display: flex;
  align-items: center;
  border-radius: 12px;
  background-color: ${p => p.theme.colors.lightShade};
  padding: 0.4rem 1rem 0.4rem 0.4rem;
  margin-top: 3px;
  cursor: ${({ isDisabled }) => (isDisabled ? 'auto' : 'pointer')};

  &:hover {
    background-color: ${({ theme, isDisabled }) =>
      theme.colors[isDisabled ? 'lightShade' : 'darkShade']};
  }
`;

const Label = styled.div`
  font-size: 1.3rem;
  line-height: 1.4rem;
  font-weight: 600;
  letter-spacing: 0.5px;
  margin-left: 7px;
`;

const IndicatorLed = styled.div`
  width: 10px;
  height: 10px;
  background-color: ${({ isOnline, syncStatus, theme }) =>
    isOnline
      ? syncStatus === 'failed'
        ? theme.colors.danger
        : '#45d48d'
      : 'rgba(119, 132, 125, 0.68)'};
  border: 1px solid white;
  border-radius: 10px;
  margin: 3px;
`;

function ScanIndicator(props) {
  // static propTypes = {
  //   onLabelClick: PropTypes.func.isRequired,
  //   syncStatus: PropTypes.oneOf(['up-to-date', 'syncing', 'failed']).isRequired,
  //   isOnline: PropTypes.bool.isRequired,
  //   tooltip: PropTypes.string,
  //   label: PropTypes.string.isRequired
  // };

  const context = useContext(ToastsContext);

  useEffect(() => {
    if (props.syncStatus === 'failed') {
      context.toast('error', 'Could not refresh');
    }
  }, []);

  return (
    <Container
      isDisabled={props.syncStatus === 'syncing' || !props.isOnline}
      onClick={props.onLabelClick}
      data-rh={props.tooltip}
    >
      {props.isOnline && props.syncStatus === 'syncing' ? (
        <Spinner />
      ) : (
        <IndicatorLed syncStatus={props.syncStatus} isOnline={props.isOnline} />
      )}
      <Label>{props.label}</Label>
    </Container>
  );
}

export default withScanIndicatorState(ScanIndicator);
