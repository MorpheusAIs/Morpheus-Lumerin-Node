import React from 'react';
import styled from 'styled-components';

import BaseIcon from './BaseIcon';

const AcceptedIcon = props => (
  <BaseIcon width="19" height="19" viewBox="0 0 19 19" {...props}>
    <path
      fillRule="evenodd"
      clipRule="evenodd"
      d="M17 9.5C17 13.6421 13.6421 17 9.5 17C5.35786 17 2 13.6421 2 9.5C2 5.35786 5.35786 2 9.5 2C13.6421 2 17 5.35786 17 9.5ZM19 9.5C19 14.7467 14.7467 19 9.5 19C4.25329 19 0 14.7467 0 9.5C0 4.25329 4.25329 0 9.5 0C14.7467 0 19 4.25329 19 9.5ZM13.3757 6.98278C13.6424 6.49912 13.4664 5.89089 12.9828 5.62426C12.4991 5.35763 11.8909 5.53356 11.6243 6.01722L8.76046 11.212L7.27875 9.37267C6.93229 8.94258 6.30276 8.87478 5.87267 9.22125C5.44258 9.56771 5.37478 10.1972 5.72125 10.6273L7.67043 13.047C8.33951 13.8776 9.63726 13.7642 10.1522 12.8302L13.3757 6.98278Z"
      fill="#11B4BF"
    />
  </BaseIcon>
);

export default AcceptedIcon;
