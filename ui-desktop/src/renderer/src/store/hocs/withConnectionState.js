import { connect } from 'react-redux';

import selectors from '../selectors';

const mapStateToProps = state => ({
  // default to null until initial state is received
  isConnected: state.chain ? selectors.getChainConnectionStatus(state) : null,

  // default to null until initial state is received
  chainName: state.chain ? selectors.getChainDisplayName(state) : null
});

export default connect(mapStateToProps);
