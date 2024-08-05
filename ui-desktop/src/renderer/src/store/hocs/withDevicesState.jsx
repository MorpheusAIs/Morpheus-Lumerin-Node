//@ts-check
import { connect } from 'react-redux';
import selectors from '../selectors';
import { withClient } from './clientContext';
import { EVENT_DEVICES_RESET } from '../events/devices';

const mapStateToProps = state => ({
  devices: selectors.getDevicesList(state),
  address: selectors.getWalletAddress(state),
});

const mapDispatchToProps = dispatch => ({
  resetDevices: () => dispatch({ type: EVENT_DEVICES_RESET })
});

export default Component =>
  withClient(connect(mapStateToProps, mapDispatchToProps)(Component));
