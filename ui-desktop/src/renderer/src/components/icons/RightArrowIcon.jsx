import React from 'react';
import styled from 'styled-components';

import BaseIcon from './BaseIcon';

const RightArrowIcon = ({ size, fill }, props) => {
  return (
    <BaseIcon size={size} width="28" height="28" viewBox="0 0 28 28" {...props}>
      <circle
        r="13"
        transform="matrix(1 0 0 -1 14 14)"
        fill="white"
        stroke={fill}
        strokeWidth="2"
      />
      <path
        d="M18.7568 14.5494L15.3642 17.7769C15.0516 18.0744 14.5305 18.0744 14.2179 17.7769C13.9053 17.4795 13.9053 16.9838 14.2179 16.6864L16.2211 14.7697H8.81053C8.35895 14.7697 8 14.4282 8 13.9986C8 13.569 8.35895 13.2275 8.81053 13.2275H16.2211L14.2179 11.3219C13.9053 11.0244 13.9053 10.5287 14.2179 10.2313C14.38 10.0771 14.5884 10 14.7968 10C15.0053 10 15.2137 10.0771 15.3758 10.2313L18.7568 13.4478C18.9074 13.5911 19 13.7893 19 13.9986C19 14.2079 18.9189 14.4062 18.7568 14.5494V14.5494Z"
        fill={fill}
      />
    </BaseIcon>
  );
};

export default RightArrowIcon;
