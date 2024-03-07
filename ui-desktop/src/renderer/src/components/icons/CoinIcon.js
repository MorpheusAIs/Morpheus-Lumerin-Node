import PropTypes from 'prop-types'
import React from 'react'

import BaseIcon from './BaseIcon'

const ETH = isTest => (
  <g fill="none" fillRule="evenodd">
    <circle cx="12" cy="12" r="12" fill="#FFF" fillRule="nonzero" />
    <g fillRule="nonzero">
      <path
        fill={isTest ? '#2dc62d' : '#7E61F8'}
        d="M12.288 11.6H6.75L12.288 3l5.539 8.6z"
      />
      <path
        fill={isTest ? '#27a327' : '#644DC6'}
        d="M12.288 14.096L6.75 11.6l5.538-2.767 5.539 2.767z"
      />
      <g>
        <path
          fill={isTest ? '#2dc62d' : '#7E61F8'}
          d="M12.288 15.32L6.75 12.4l5.538 8.6 5.539-8.6z"
        />
        <path
          fill={isTest ? '#27a327' : '#644DC6'}
          d="M12.288 15.32V21l5.539-8.6z"
        />
      </g>
    </g>
  </g>
)

const ETC = isTest => (
  <g fill="none" fillRule="evenodd">
    <circle cx="12" cy="12" r="12" fill="#FFF" fillRule="nonzero" />
    <path
      fill="gray"
      fillRule="nonzero"
      d="M6.916 12.138s.028 0 0 0c.028 0 .028 0 0 0 0 .028 0 .028 0 0z"
    />
    <g fillRule="nonzero">
      <path
        fill={isTest ? '#2dc62d' : '#7E61F8'}
        d="M12.288 8.68L6.75 11.6 12.288 3l5.539 8.6z"
      />
      <path
        fill={isTest ? '#27a327' : '#644DC6'}
        d="M12.288 8.68V3l5.539 8.6z"
      />
      <path
        fill={isTest ? '#2dc62d' : '#7E61F8'}
        d="M12.288 15.32L6.75 12.4l5.538 8.6 5.539-8.6z"
      />
      <path
        fill={isTest ? '#27a327' : '#644DC6'}
        d="M12.288 15.32V21l5.539-8.6z"
      />
      <g>
        <path
          fill={isTest ? '#2dc62d' : '#7E61F8'}
          d="M12.288 10.2l4.708 1.8-4.708 1.8L7.581 12z"
        />
        <path
          fill={isTest ? '#27a327' : '#644DC6'}
          d="M16.996 12l-4.708 1.8L7.581 12z"
        />
      </g>
    </g>
  </g>
)

const CoinIcon = ({ coin, ...other }) => (
  <BaseIcon {...other}>
    {coin === 'ethRopstenLocal' && ETH('local')}
    {coin === 'etcMordorLocal' && ETC('local')}
    {coin === 'ethMainnet' && ETH()}
    {coin === 'ethRopsten' && ETH('test')}
    {coin === 'etcMainnet' && ETC()}
    {coin === 'etcMordor' && ETC('test')}
  </BaseIcon>
)

CoinIcon.propTypes = {
  coin: PropTypes.oneOf([
    'ethRopstenLocal',
    'etcMordorLocal',
    'ethMainnet',
    'ethRopsten',
    'etcMainnet',
    'etcMordor'
  ]).isRequired
}

export default CoinIcon
