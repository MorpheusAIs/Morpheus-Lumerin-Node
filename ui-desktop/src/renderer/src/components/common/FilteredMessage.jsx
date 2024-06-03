import React from 'react';
import smartRounder from 'smart-round';
import { connect } from 'react-redux';

import { withClient } from '../../store/hocs/clientContext';
// import { getChainConfig, getCoinSymbol } from '../store/selectors';
import selectors from '../../store/selectors';

const FilteredMessage = props => {
  // static propTypes = {
  //   withDefault: PropTypes.func,
  //   coinSymbol: PropTypes.string.isRequired,
  //   children: PropTypes.node.isRequired,
  //   config: PropTypes.shape({
  //     mainTokenAddress: PropTypes.string.isRequired
  //   }).isRequired,
  //   client: PropTypes.shape({
  //     fromWei: PropTypes.func.isRequired
  //   }).isRequired
  // }

  // static defaultProps = {
  //   withDefault: t => t
  // }

  const messageParser = str => {
    const replacements = [
      {
        search: props.config.webFacingAddress,
        replaceWith: 'WEB FACING CONTRACT'
      },
      {
        search: props.config.mainTokenAddress,
        replaceWith: 'LMR TOKEN CONTRACT'
      },
      { search: /(.*gas too low.*)/gi, replaceWith: () => 'Gas too low.' },
      {
        search: /[\s\S]*Transaction has been reverted by the EVM[\s\S]*/gi,
        replaceWith: () => 'Transaction failed'
      },
      {
        search: /[\s\S]*CONNECTION TIMEOUT[\s\S]*/gi,
        replaceWith: () => 'Connection timeout'
      },
      {
        search: /[\s\S]*Couldn't connect to node on WS[\s\S]*/gi,
        replaceWith: () => `Couldn't connect to blockchain node`
      },
      {
        search: /(.*insufficient funds for gas \* price \+ value.*)/gim,
        replaceWith: () => "You don't have enough funds for this transaction."
      },
      {
        search: /(.*Insufficient\sfunds.*Required\s)(\d+)(\sand\sgot:\s)(\d+)(.*)/gim,
        // eslint-disable-next-line max-params
        replaceWith: (match, p1, p2, p3, p4, p5) => {
          const rounder = smartRounder(6, 0, 18);
          return [
            p1,
            rounder(props.client.fromWei(p2), true),
            ` ${props.coinSymbol}`,
            p3,
            rounder(props.client.fromWei(p4), true),
            ` ${props.coinSymbol}`,
            p5
          ].join('');
        }
      }
    ];

    return replacements.reduce(
      (output, { search, replaceWith }) => output.replace(search, replaceWith),
      str
    );
  };

  const filteredMessage = messageParser(props.children);

  return filteredMessage === props.children
    ? props.withDefault(props.children)
    : filteredMessage;
};

const mapStateToProps = state => ({
  coinSymbol: selectors.getCoinSymbol(state),
  config: selectors.getChainConfig(state)
});

export default connect(mapStateToProps)(withClient(FilteredMessage));
