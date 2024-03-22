import selectors from '../selectors';
import { connect } from 'react-redux';

const mapStateToProps = state => ({
  syncStatus: selectors.getContractsSyncStatus(state)
});

export default Component => connect(mapStateToProps)(Component);
