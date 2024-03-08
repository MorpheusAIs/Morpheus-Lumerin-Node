import React from 'react';
import styled from 'styled-components';

import BaseIcon from './BaseIcon';

const LeftArrowIcon = ({ size, fill }, props) => {
  return (
    <BaseIcon size={size} width="28" height="28" viewBox="0 0 28 28" {...props}>
      <circle
        r="13"
        transform="matrix(-1 -8.74228e-08 -8.74228e-08 1 14 14)"
        fill="white"
        stroke={fill}
        strokeWidth="2"
      />
      <path
        d="M9.24316 13.4506L12.6358 10.2231C12.9484 9.92564 13.4695 9.92564 13.7821 10.2231C14.0947 10.5205 14.0947 11.0162 13.7821 11.3136L11.7789 13.2303L19.1895 13.2303C19.6411 13.2303 20 13.5718 20 14.0014C20 14.431 19.6411 14.7725 19.1895 14.7725L11.7789 14.7725L13.7821 16.6781C14.0947 16.9756 14.0947 17.4713 13.7821 17.7687C13.62 17.9229 13.4116 18 13.2032 18C12.9947 18 12.7863 17.9229 12.6242 17.7687L9.24316 14.5522C9.09263 14.4089 9 14.2107 9 14.0014C9 13.7921 9.08105 13.5938 9.24316 13.4506L9.24316 13.4506Z"
        fill={fill}
      />
    </BaseIcon>
  );
};

export default LeftArrowIcon;
