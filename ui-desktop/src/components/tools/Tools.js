import { withRouter, NavLink } from 'react-router-dom';
import withToolsState from '../../store/hocs/withToolsState';
import PropTypes from 'prop-types';
import styled from 'styled-components';
import React, { useState, useContext, useEffect } from 'react';

import ConfirmModal from './ConfirmModal';
import WalletStatus from './WalletStatus';
import { ConfirmationWizard, TextInput, Flex, BaseBtn, Sp } from '../common';
import Spinner from '../common/Spinner';
import { View } from '../common/View';
import { ToastsContext } from '../../components/toasts';
import ConfirmProxyConfigModal from './ConfirmProxyConfigModal';
import RevealSecretPhraseModal from './RevealSecretPhraseModal';
import { Message } from './ConfirmModal.styles';
import ExportPrivateKeyModal from './ExportPrivateKeyModal';
import { ContractsTab } from './featureTabs/ContractsTab';
import { ProxyConfigPanel } from './ProxyConfigPanel';

import { Tab, Tabs, TabList, TabPanel } from 'react-tabs';
import { StyledBtn, Subtitle, StyledParagraph, Input } from './common';

import 'react-tabs/style/react-tabs.css';
import './styles.css';

const Container = styled.div`
  margin-left: 2rem;
  height: 85vh;
  overflow-y: scroll;
`;

const TitleContainer = styled.div`
  display: flex;
  padding: 1.8rem 0;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  position: sticky;
  width: 100%;
  z-index: 2;
  right: 0;
  left: 0;
  top: 0;
`;

const Title = styled.label`
  font-size: 2.4rem;
  line-height: 3rem;
  white-space: nowrap;
  margin: 0;
  font-weight: 600;
  color: ${p => p.theme.colors.dark};
  margin-bottom: 4.8px;
  margin-right: 2.4rem;
  cursor: default;

  @media (min-width: 1140px) {
    margin-right: 0.8rem;
  }

  @media (min-width: 1200px) {
    margin-right: 1.6rem;
  }
`;

const Confirmation = styled.div`
  color: ${p => p.theme.colors.danger};
  background-color: ${p => p.theme.colors.medium};
  border-radius: 4px;
  padding: 0.8rem 1.6rem;
`;

const ValidationMsg = styled.div`
  font-size: 1.4rem;
  margin-left: 1.6rem;
  opacity: 0.75;
`;

const WalletInfo = styled.h4`
  color: ${p => p.theme.colors.dark};
`;

const getPoolAndAccount = url => {
  if (!url) return {};
  const addressParts = url.replace('stratum+tcp://', '').split(':');
  return {
    account: decodeURIComponent(addressParts[0]),
    pool: `${addressParts[1].slice(1)}:${addressParts[2]}`
  };
};

const Select = styled.select`
  width: 35%;
  height: 40px;
  outline: 0;
  border: 0px;
  background: #eaf7fc;
  border-radius: 15px;
  padding: 1.2rem 1.2rem;
  margin-top: 0.25rem;
`;

const Tools = props => {
  const {
    getProxyRouterSettings,
    saveProxyRouterSettings,
    restartProxyRouter,
    setDefaultCurrency,
    selectedCurrency,
    getCustomEnvs,
    setCustomEnvs,
    getProfitSettings,
    setProfitSettings,
    config,
    restartWallet,
    titanLightningDashboard
  } = props;

  const RenderForm = goToReview => {
    const defState = {
      activeModal: null,
      testSocket: '',
      toast: {
        Show: false,
        Message: null,
        Type: 'info'
      },
      selectedCurrency: selectedCurrency,
      customEnvs: {},
      profitSettings: {}
    };

    const [state, setState] = useState(defState);
    const [isRestarting, restartNode] = useState(false);
    const [proxyRouterSettings, setProxyRouterSettings] = useState({
      proxyRouterEditMode: false,
      isFetching: true
    });
    const [sellerPoolParts, setSellerPoolParts] = useState(null);
    const [isTitanLightning, setTitanLightning] = useState(false);

    const [httpNodeInput, setHttpNodeInput] = useState(
      state.customEnvs?.httpNode || config.chain.httpApiUrls[0]
    );
    const [wsNodeInput, setWsNodeInput] = useState(state.customEnvs?.wsNode);

    const context = useContext(ToastsContext);

    const logPath = (() => {
      if (window.navigator.userAgent.indexOf('Win') !== -1)
        return '%USERPROFILE%\\AppData\\Roaming\\morpheus-ui-desktop\\logs\\main.log';
      if (window.navigator.userAgent.indexOf('Mac') !== -1)
        return '~/Library/Logs/morpheus-ui-desktop/main.log';
      return '~/.config/morpheus-ui-desktop/logs/main.log';
    })();

    useEffect(() => {
      getProxyRouterSettings()
        .then(data => {
          setSellerPoolParts(getPoolAndAccount(data.sellerDefaultPool));

          setProxyRouterSettings({
            ...data,
            isFetching: false
          });
          if (data.isTitanLightning) {
            setTitanLightning(true);
          }
        })
        .catch(err => {
          context.toast('error', 'Failed to fetch proxy-router settings');
        });

      getCustomEnvs().then(envs => {
        setState({ ...state, customEnvs: envs });
        setHttpNodeInput(envs?.httpNode || config.chain.httpApiUrls[0]);
        setWsNodeInput(envs?.wsNode);
      });
      getProfitSettings().then(settings => {
        setState({ ...state, profitSettings: settings });
      });
    }, []);

    const onCloseModal = () => {
      setState({ ...state, activeModal: null });
    };

    const onActiveModalClick = modal => {
      setState({ ...state, activeModal: modal });
    };

    const resetCustomEnv = () => {
      setState({ ...state, customEnvs: {} });
      setHttpNodeInput(null);
      setWsNodeInput(null);
      setCustomEnvs({});
    };

    const setCustomEnvHandler = envs => {
      setCustomEnvs(envs);
      setState({ ...state, customEnvs: envs });
    };

    const proxyRouterEditClick = () => {
      setProxyRouterSettings({
        ...proxyRouterSettings,
        proxyRouterEditMode: true
      });
    };

    const saveProxyRouterConfig = () => {
      onCloseModal();
      setProxyRouterSettings({
        ...proxyRouterSettings,
        proxyRouterEditMode: false,
        isFetching: true
      });
      return saveProxyRouterSettings({
        sellerDefaultPool: proxyRouterSettings.sellerDefaultPool,
        buyerDefaultPool: proxyRouterSettings.sellerDefaultPool,
        isTitanLightning
      })
        .catch(() => {
          context.toast('error', 'Failed to save proxy-router settings');
        })
        .finally(() => {
          setProxyRouterSettings({
            ...proxyRouterSettings,
            isFetching: false,
            proxyRouterEditMode: false
          });
        });
    };

    const confirmProxyRouterRestart = () => {
      saveProxyRouterConfig().then(() => {
        restartProxyRouter({}).catch(err => {
          context.toast('error', 'Failed to restart proxy-router');
        });
      });
    };

    const toggleIsLightning = () => {
      setTitanLightning(!isTitanLightning);
      setSellerPoolParts({
        ...sellerPoolParts,
        pool: '',
        account: '',
        isTitanLightning: !isTitanLightning
      });
    };

    const onRestartClick = async () => {
      onCloseModal();
      restartNode(true);

      await restartProxyRouter({}).catch(() => {
        context.toast('error', 'Failed to restart proxy-router');
      });

      // for UX
      setTimeout(() => {
        restartNode(false);
      }, 6000);
    };

    const { onInputChange, mnemonic, errors } = props;
    const { testSocket } = state;

    return (
      <Container>
        <Sp mt={-4}>
          {/* <Subtitle>Recover a Wallet</Subtitle>
          <form data-testid="recover-form" onSubmit={goToReview}>
            <StyledParagraph>
              Enter a valid twelve-word recovery phrase to recover another
              wallet.
            </StyledParagraph>
            <StyledParagraph>
              This action will replace your current stored seed!
            </StyledParagraph>
            <TextInput
              data-testid="mnemonic-field"
              autoFocus
              onChange={onInputChange}
              label="Recovery phrase"
              error={errors.mnemonic}
              value={mnemonic || ''}
              rows={2}
              id="mnemonic"
            />
            <Sp mt={4}>
              <Flex.Row align="center">
                <StyledBtn disabled={!props.isRecoverEnabled} submit>
                  Recover
                </StyledBtn>
                {!props.isRecoverEnabled && (
                  <ValidationMsg>
                    A recovery phrase must have exactly 12 words
                  </ValidationMsg>
                )}
              </Flex.Row>
            </Sp>
          </form> */}
          <Tabs style={{ color: 'black' }}>
            <TabList>
              <Tab>Info</Tab>
              <Tab>Wallet</Tab>
              <Tab>Contracts</Tab>
              <Tab>Proxy Router</Tab>
              <Tab>Environment</Tab>
            </TabList>
            <TabPanel>
              <Sp mt={5}>
                <WalletInfo>Wallet Information</WalletInfo>
                <WalletStatus />
              </Sp>

              <Sp mt={5}>
                <Subtitle>Logs</Subtitle>
                <StyledParagraph>
                  You can find wallet logs in the file: <br />
                  <i>{logPath}</i>
                </StyledParagraph>
              </Sp>
            </TabPanel>
            <TabPanel>
              <Sp mt={5}>
                <Subtitle>Seller Default Currency</Subtitle>
                <StyledParagraph>
                  This will set default currency to display prices and balances
                  on Seller Hub.
                  <div style={{ marginTop: '1rem' }}>
                    <Select
                      onChange={e =>
                        setState({ ...state, selectedCurrency: e.target.value })
                      }
                    >
                      <option
                        selected={state.selectedCurrency === 'BTC'}
                        key={'BTC'}
                        value={'BTC'}
                      >
                        BTC
                      </option>
                      <option
                        selected={state.selectedCurrency === 'LMR'}
                        key={'LMR'}
                        value={'LMR'}
                      >
                        LMR
                      </option>
                    </Select>
                  </div>
                </StyledParagraph>
                <StyledBtn
                  disabled={state.selectedCurrency === selectedCurrency}
                  onClick={() => setDefaultCurrency(state.selectedCurrency)}
                >
                  Save
                </StyledBtn>
              </Sp>
              <Sp mt={5}>
                <Subtitle>Change Password</Subtitle>
                <StyledParagraph>
                  This will allow you to change the password you use to access
                  the wallet.
                </StyledParagraph>
                <NavLink data-testid="change-password-btn" to="/change-pass">
                  <StyledBtn>Change Password</StyledBtn>
                </NavLink>
              </Sp>
              <Sp mt={5}>
                <Subtitle>Rescan Transactions List</Subtitle>
                <StyledParagraph>
                  This will clear your local cache and rescan all your wallet
                  transactions.
                </StyledParagraph>
                <StyledBtn onClick={() => onActiveModalClick('confirm-rescan')}>
                  Rescan Transactions
                </StyledBtn>
                <ConfirmModal
                  onRequestClose={onCloseModal}
                  onConfirm={props.onRescanTransactions}
                  isOpen={state.activeModal === 'confirm-rescan'}
                />
              </Sp>
              <Sp mt={5}>
                <Subtitle>Sensitive Info</Subtitle>
                {!props.hasStoredSecretPhrase && (
                  <StyledParagraph>
                    To enable this feature you need to re-login to your wallet
                  </StyledParagraph>
                )}
                <div style={{ display: 'flex', flexDirection: 'column' }}>
                  <StyledBtn
                    disabled={!props.hasStoredSecretPhrase}
                    onClick={() => onActiveModalClick('reveal-secret-phrase')}
                  >
                    Reveal Secret Recovery Phrase
                  </StyledBtn>
                  <StyledBtn
                    style={{ marginTop: '1.6rem' }}
                    onClick={() => onActiveModalClick('export-private-key')}
                  >
                    Export private key
                  </StyledBtn>
                </div>
                <ExportPrivateKeyModal
                  onRequestClose={() => {
                    props.discardPrivateKey();
                    onCloseModal();
                  }}
                  onLater={onCloseModal}
                  onExportPrivateKey={props.onExportPrivateKey}
                  privateKey={props.privateKey}
                  copyToClipboard={props.copyToClipboard}
                  onRevealPhrase={props.onRevealPhrase}
                  isOpen={state.activeModal === 'export-private-key'}
                />
                <RevealSecretPhraseModal
                  onRequestClose={() => {
                    props.discardMnemonic();
                    onCloseModal();
                  }}
                  onLater={onCloseModal}
                  onShowMnemonic={props.onShowMnemonic}
                  mnemonic={props.mnemonic}
                  copyToClipboard={props.copyToClipboard}
                  onRevealPhrase={props.onRevealPhrase}
                  isOpen={state.activeModal === 'reveal-secret-phrase'}
                />
              </Sp>

              <Sp mt={5}>
                <Subtitle>Reset</Subtitle>
                <StyledParagraph>
                  Set up your wallet from scratch.
                </StyledParagraph>
                <StyledBtn onClick={() => onActiveModalClick('confirm-logout')}>
                  Reset
                </StyledBtn>

                <ConfirmProxyConfigModal
                  title={'Reset your wallet'}
                  message={
                    <>
                      <Message>
                        Make sure you have your recovery phrase before reseting
                        your wallet. If you don’t have your recovery phrase, we
                        suggest you transfer all funds out of your wallet before
                        you reset. Otherwise you will lock yourself out of your
                        wallet, and you won’t have access to the funds in this
                        wallet.
                      </Message>
                      <Message>Continue?</Message>
                    </>
                  }
                  onRequestClose={onCloseModal}
                  onConfirm={props.logout}
                  onLater={onCloseModal}
                  isOpen={state.activeModal === 'confirm-logout'}
                />
              </Sp>
            </TabPanel>
            <TabPanel>
              <ContractsTab
                settings={state.profitSettings}
                onCommit={settings => {
                  setState({ ...state, profitSettings: settings });
                  setProfitSettings(settings);
                  context.toast('success', 'Updated');
                }}
              />
            </TabPanel>
            <TabPanel>
              <ProxyConfigPanel
                {...props}
                onCloseModal={onCloseModal}
                onRestartClick={onRestartClick}
                state={state}
                saveProxyRouterConfig={saveProxyRouterConfig}
                saveProxyRouterSettings={saveProxyRouterSettings}
                confirmProxyRouterRestart={confirmProxyRouterRestart}
                toggleIsLightning={toggleIsLightning}
                setProxyRouterSettings={setProxyRouterSettings}
                proxyRouterSettings={proxyRouterSettings}
                setSellerPoolParts={setSellerPoolParts}
                sellerPoolParts={sellerPoolParts}
                proxyRouterEditClick={proxyRouterEditClick}
                onActiveModalClick={onActiveModalClick}
                isTitanLightning={isTitanLightning}
              />
            </TabPanel>
            <TabPanel>
              <Subtitle>HTTP ETH Node: </Subtitle>
              <StyledParagraph>
                <Input
                  placeholder={config.chain.httpApiUrls[0]}
                  onChange={e => setHttpNodeInput(e.value)}
                  value={httpNodeInput}
                />
              </StyledParagraph>

              <StyledParagraph>
                <Subtitle>Web Socket ETH Node: </Subtitle>
                <Input
                  placeholder={
                    state.customEnvs?.wsNode ||
                    'wss://arb-mainnet.g.alchemy.com/v2/API_KEY'
                  }
                  onChange={e => setWsNodeInput(e.value)}
                  value={wsNodeInput}
                />
                <StyledParagraph>
                  Warning: Modifying these settings with incorrect values may
                  impact the wallet's stability and disrupt blockchain
                  integration. If you experience unexpected issues, consider
                  resetting to the default values.
                </StyledParagraph>
              </StyledParagraph>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <StyledBtn
                  onClick={() => {
                    setCustomEnvHandler({
                      httpNode: httpNodeInput,
                      wsNode: wsNodeInput
                    });
                    onActiveModalClick('confirm-custom-env-change');
                  }}
                >
                  Save
                </StyledBtn>
                <StyledBtn
                  onClick={() => {
                    resetCustomEnv();
                    onActiveModalClick('confirm-custom-env-change');
                  }}
                >
                  Reset
                </StyledBtn>

                <ConfirmProxyConfigModal
                  title={'Wallet and Proxy restart'}
                  message={
                    <>
                      <Message>
                        You need to restart Wallet and Proxy Router to apply
                        changes.
                      </Message>
                    </>
                  }
                  onRequestClose={onCloseModal}
                  onConfirm={() => {
                    restartProxyRouter({})
                      .then(() => restartWallet())
                      .catch(err => console.log(err));
                  }}
                  onLater={onCloseModal}
                  isOpen={state.activeModal === 'confirm-custom-env-change'}
                />
              </div>
            </TabPanel>
          </Tabs>
          {/* <Sp mt={5}>
            <hr />
            <Subtitle>Run End-to-End Test</Subtitle>
            <StyledParagraph>
              Before running test, make sure that all your Lumerin node is up
              and running locally.
            </StyledParagraph>

            <TextInput
              data-testid="test-field"
              onChange={e => setState({ ...state, testSocket: e.targeValue })}
              label={process.env.PROXY_ROUTER_URL || ''}
              error={errors.mnemonic}
              value={testSocket || ''}
              rows={2}
              id="testSocket"
            />
            <br />
            <StyledBtn onClick={() => onActiveModalClick('confirm-test')}>
              Run Test
            </StyledBtn>
            <TestModal
              onRequestClose={onCloseModal}
              onConfirm={props.onRunTest}
              isOpen={state.activeModal === 'confirm-test'}
            />
          </Sp> */}
        </Sp>
      </Container>
    );
  };

  const onWizardSubmit = password =>
    props.onSubmit(password).then(() => props.history.push('/wallet'));

  const renderConfirmation = () => (
    <Confirmation data-testid="confirmation">
      <h3>Are you sure?</h3>
      <p>This operation will overwrite and restart the current wallet!</p>
    </Confirmation>
  );

  return (
    <View data-testid="tools-container">
      <TitleContainer>
        <Title>Tools</Title>
      </TitleContainer>
      <ConfirmationWizard
        renderConfirmation={renderConfirmation}
        confirmationTitle=""
        onWizardSubmit={onWizardSubmit}
        pendingTitle="Recovering..."
        successText="Wallet successfully recovered"
        RenderForm={RenderForm}
        validate={props.validate}
        noCancel
        styles={{
          confirmation: {
            padding: 0
          },
          btns: {
            background: 'none',
            marginTop: '3.2rem',
            maxWidth: '200px',
            padding: 0
          }
        }}
      />
    </View>
  );
};

Tools.propTypes = {
  onRescanTransactions: PropTypes.func.isRequired,
  onRunTest: PropTypes.func.isRequired,
  isRecoverEnabled: PropTypes.bool.isRequired,
  onInputChange: PropTypes.func.isRequired,
  validate: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  mnemonic: PropTypes.string,
  history: PropTypes.shape({
    push: PropTypes.func.isRequired
  }).isRequired,
  errors: PropTypes.shape({
    mnemonic: PropTypes.string
  }).isRequired
};

export default withToolsState(withRouter(Tools));
