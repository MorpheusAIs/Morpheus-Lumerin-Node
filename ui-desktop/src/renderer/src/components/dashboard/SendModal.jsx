import withSendLMRFormState from '../../store/hocs/withSendLMRFormState';

import PropTypes from 'prop-types';
import React from 'react';
import Modal, { HeaderButton } from '../common/Modal';

import { Drawer, Tabs } from '../common';
import SendLMRForm from './SendLMRForm';

class SendModal extends React.Component {
  // static propTypes = {
  //   sendLmrDisabledReason: PropTypes.string,
  //   sendLmrDisabled: PropTypes.bool.isRequired,
  //   coinSymbol: PropTypes.string.isRequired,
  //   onRefreshRequest: PropTypes.func.isRequired,
  //   onRequestClose: PropTypes.func.isRequired,
  //   refreshStatus: PropTypes.oneOf(['init', 'pending', 'success', 'failure'])
  //     .isRequired,
  //   isOpen: PropTypes.bool.isRequired,
  //   hash: PropTypes.string
  // };

  componentDidUpdate(prevProps) {
    if (
      this.props.refreshStatus !== prevProps.refreshStatus &&
      this.props.refreshStatus === 'failure'
    ) {
      this.context.toast('error', 'Could not refresh');
    }
  }

  render() {
    if (!this.props.hash) return null;
    const tabs = (
      <Tabs
        onClick={this.onTabChange}
        active={this.props.activeTab}
        items={[
          {
            id: 'lmr',
            label: this.props.symbol,
            'data-rh': this.props.sendLmrDisabledReason,
            disabled: this.props.sendLmrDisabled
          },
          { id: 'coin', label: this.props.coinSymbol }
        ]}
      />
    );

    return (
      <Modal
        shouldReturnFocusAfterClose={false}
        onRequestClose={this.props.onRequestClose}
        // headerChildren={
        //   <HeaderButton
        //     disabled={this.props.refreshStatus === 'pending'}
        //     onClick={this.props.onRefreshRequest}
        //   >
        //     {this.props.refreshStatus === 'pending' ? 'Syncing...' : 'Refresh'}
        //   </HeaderButton>
        // }
        isOpen={this.props.isOpen}
        title="Send"
      >
        {this.state.activeTab === 'receive' && <SendLMRForm tabs={tabs} />}
      </Modal>
    );
  }
}

export default withSendLMRFormState(SendModal);
