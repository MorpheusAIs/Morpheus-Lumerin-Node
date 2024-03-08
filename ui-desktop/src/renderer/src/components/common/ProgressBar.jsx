import React from 'react';
import styled from 'styled-components';
import theme from '../../ui/theme';

const ProgressBarWrapper = styled.div`
  width: 100%;
  min-width: 30px;
  display: flex;
  border-radius: 5px;
  overflow: hidden;
`;

const Completed = styled.div`
  background-color: ${theme.colors.active};
  height: 30px;
  width: ${props => props.width};
  display: flex;
  align-items: center;
`;

const Remaining = styled.div`
  background-color: ${theme.colors.cancelled};
  height: 30px;
  width: ${props => props.width};
  display: flex;
  align-items: center;
`;

const Value = styled.span`
  color: #ffffff;
  font-weight: bold;
  padding: 0 8px;
  display: inline-block;
  vertical-align: middle;
`;

const ProgressBar = ({ completed, remaining }) => {
  completed = Number(completed);
  remaining = Number(remaining);

  const isZero = completed + remaining === 0;
  const progress = isZero
    ? 100
    : Math.floor((completed / (completed + remaining)) * 100);

  const completedWidth = `${progress}%`;
  const remainingWidth = remaining === 0 ? '0%' : `${100 - progress}%`;

  return (
    <ProgressBarWrapper>
      <Completed width={completedWidth} data-rh={`${completed} Completed`}>
        {/* <Value>{completed}</Value> */}
      </Completed>
      <Remaining width={remainingWidth} data-rh={`${remaining} Cancelled`}>
        {/* <Value>{remaining}</Value> */}
      </Remaining>
    </ProgressBarWrapper>
  );
};

export default ProgressBar;
