import PropTypes from 'prop-types'
import React from 'react'

import BaseIcon from './BaseIcon'

const LogoIcon = ({ size, negative, ...other }) => (
  <BaseIcon size={size || '3.2rem'} viewBox="0 0 32 32" {...other}>
    <g fill="none" fillRule="evenodd" opacity=".3">
      <mask id="b" fill="#fff">
        <rect id="a" width="32" height="32" rx="12" />
      </mask>
      <use fill={negative ? '#FFF' : '#323232'} xlinkHref="#a" />
      <path
        fill={negative ? '#323232' : '#FFF'}
        d="M15.913 11.888A4.942 4.942 0 0 0 12.03 10C9.304 10 7 12.268 7 14.954v5.953C7 21.47 7.543 22 8.119 22c.588 0 1.067-.49 1.067-1.093v-5.953c.027-1.554 1.276-2.769 2.844-2.769a2.739 2.739 0 0 1 2.767 2.766v5.956c.014.604.508 1.093 1.093 1.093.614 0 1.093-.49 1.093-1.093v-5.953c.027-1.554 1.276-2.769 2.843-2.769a2.739 2.739 0 0 1 2.768 2.769v5.953c0 .592.5 1.093 1.092 1.093.582 0 1.093-.51 1.093-1.093v-5.953A4.959 4.959 0 0 0 19.826 10c-1.562 0-2.983.747-3.913 1.888z"
        mask="url(#b)"
      />
    </g>
  </BaseIcon>
)

LogoIcon.propTypes = {
  negative: PropTypes.bool,
  size: PropTypes.string
}

export default LogoIcon
