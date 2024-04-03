import PropTypes from 'prop-types';
import styled from 'styled-components';
import theme from '../../../../ui/theme';
import React from 'react';

import LeftArrowIcon from '../../../icons/LeftArrowIcon';
import RightArrowIcon from '../../../icons/RightArrowIcon';
import { ContractIcon } from '../../../icons/ContractIcon';

export const TxIcon = ({ txType, size = '3.6rem' }) => {
  if (txType === 'received') {
    return <LeftArrowIcon fill={theme.colors.primaryLight} />;
  }

  if (txType === 'sent') {
    return <RightArrowIcon fill={theme.colors.tertiary} />;
  }

  return (
    <>
      <ContractIcon fill={theme.colors.primaryLight} />
    </>
  );
};
