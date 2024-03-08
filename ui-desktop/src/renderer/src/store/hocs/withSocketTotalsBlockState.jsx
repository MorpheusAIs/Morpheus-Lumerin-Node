import { withClient } from './clientContext';
import selectors from '../selectors';
import { connect } from 'react-redux';

const mapStateToProps = (state, { client }) => ({
  // coinBalanceUSD: selectors.getCoinBalanceUSD(state, client),
  // coinBalanceWei: selectors.getCoinBalanceWei(state),
  // lmrBalanceWei: selectors.getLmrBalanceWei(state),
  lmrBalance: selectors.getWalletLmrBalance(state)
});

export default Component => withClient(connect(mapStateToProps)(Component));
