import { connect } from 'react-redux';

import { withClient } from './clientContext';
import selectors from '../selectors';

const mapStateToProps = (state, { client }) => ({
  address: selectors.getWalletAddress(state),
  copyToClipboard: client.copyToClipboard
});

export default Component => withClient(connect(mapStateToProps)(Component));
