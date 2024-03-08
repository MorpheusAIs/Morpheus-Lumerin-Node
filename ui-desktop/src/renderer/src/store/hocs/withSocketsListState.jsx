import selectors from '../selectors';
import { connect } from 'react-redux';

const mapStateToProps = state => ({
  hasConnections: selectors.hasConnections(state),
  connections: selectors.getConnections(state),
  isLocalProxyRouter: selectors.getIsLocalProxyRouter(state)
});

export default Component => connect(mapStateToProps)(Component);
