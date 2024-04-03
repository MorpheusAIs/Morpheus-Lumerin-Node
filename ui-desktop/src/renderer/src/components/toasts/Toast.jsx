import { TransitionMotion, spring } from 'react-motion';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import React from 'react';

import CloseIcon from '../icons/CloseIcon';

const Separator = styled.div`
  padding: 8px 0;
`;

const Container = styled.div`
  z-index: 20;
  border-top-style: solid;
  border-top-width: 4px;
  border-top-color: ${({ type, theme }) =>
    type === 'success'
      ? theme.colors.success
      : type === 'error'
      ? theme.colors.danger
      : theme.colors.primary};
  background-color: #454545;
  border-radius: 4px;
  padding: 8px 16px 12px;
  position: relative;
  box-shadow: 0 4px 12px -1px rgba(0, 0, 0, 0.35),
    0 1px 3px 2px rgba(0, 0, 0, 0.1);
`;

const DismissBtn = styled.button`
  position: absolute;
  top: 0;
  right: 0;
  background: transparent;
  outline: none;
  cursor: pointer;
  color: white;
  border: none;
  padding: 16px;
  line-height: 1;
  font-size: 16px;
  margin: 0;
  opacity: 0.6;
  z-index: 1;

  &:hover {
    opacity: 1;
  }
`;

const Scroller = styled.div`
  overflow-x: hidden;
  overflow-y: auto;
  max-height: 410px;
`;

const Message = styled.div`
  color: white;
  font-size: 13px;
  line-height: 1.25;
  padding: 12px 32px 12px 0;
`;

const ShowMoreBtn = styled.button`
  background: none;
  outline: none;
  cursor: pointer;
  color: white;
  border: none;
  line-height: 1;
  font-size: 13px;
  margin: 0;
  padding: 12px 0;
  opacity: 0.9;
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 2px;

  &:hover {
    opacity: 1;
  }
`;

export default class Toast extends React.Component {
  static propTypes = {
    messagesPerToast: PropTypes.number.isRequired,
    onShowMore: PropTypes.func.isRequired,
    onDismiss: PropTypes.func.isRequired,
    messages: PropTypes.arrayOf(PropTypes.string).isRequired,
    type: PropTypes.oneOf(['info', 'success', 'error']).isRequired
  };

  state = { showMore: false };

  handleDismiss = () => this.props.onDismiss(this.props.type);

  handleShowMore = () => {
    this.setState({ showMore: true });
    // cancel autoClose
    this.props.onShowMore(this.props.type);
  };

  willEnter = () => ({ maxHeight: 0, opacity: 0 });

  render() {
    const { type, messages } = this.props;
    const { showMore } = this.state;

    const shownMessages = showMore
      ? messages
      : messages.slice(0, this.props.messagesPerToast);

    const hiddenMessages = showMore
      ? []
      : messages.slice(this.props.messagesPerToast);

    return (
      <Separator>
        <Container type={type}>
          <DismissBtn type="button" onClick={this.handleDismiss}>
            <CloseIcon />
          </DismissBtn>

          <TransitionMotion
            willEnter={this.willEnter}
            styles={(hiddenMessages.length > 0
              ? [...shownMessages, 'more']
              : shownMessages
            ).map(msg => ({
              key: msg,
              data: msg,
              style: {
                opacity: spring(1, { stiffness: 60, damping: 5 })
              }
            }))}
          >
            {interpolatedStyles => (
              <Scroller>
                {interpolatedStyles.map(({ key, style, data }) =>
                  key === 'more' ? (
                    <div style={style} key="more">
                      <ShowMoreBtn
                        data-type={type}
                        onClick={this.handleShowMore}
                        type="button"
                      >
                        View {hiddenMessages.length} more{' '}
                        {hiddenMessages.length > 1 ? 'messages' : 'message'}
                      </ShowMoreBtn>
                    </div>
                  ) : (
                    <div style={style} key={key}>
                      <Message>{data}</Message>
                    </div>
                  )
                )}
              </Scroller>
            )}
          </TransitionMotion>
        </Container>
      </Separator>
    );
  }
}
