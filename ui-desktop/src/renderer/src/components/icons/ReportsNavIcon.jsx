import React from 'react';
import styled from 'styled-components';

import BaseIcon from './BaseIcon';

const ReportsNavIcon = ({ isActive, size }, props) => {
  const Circle = styled.div`
    background-color: ${p => p.theme.colors.medium};
    border-radius: 50%;
  `;
  const fill = isActive ? '#11B4BF' : 'black';

  return (
    <Circle>
      <BaseIcon size={size} {...props}>
        <path
          fill={fill}
          transform="rotate(-180 12 12)"
          d="M7.67 18.384l3.973 3.984V6.72h.958v15.638l3.973-3.984.68.682-4.788 4.8a.473.473 0 0 1-.68 0L7 19.056l.67-.672zm-2.677-1.25A9.97 9.97 0 0 1 2 10C2 4.477 6.477 0 12 0s10 4.477 10 10a9.969 9.969 0 0 1-2.929 7.071l-.707-.707a9 9 0 1 0-12.649.078l-.722.692z"
        />
      </BaseIcon>
    </Circle>
  );
};

export default ReportsNavIcon;
