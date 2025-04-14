import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';
import { Sp } from '../common';
import withSettingsState from '../../store/hocs/withSettingsState';
import { StyledBtn, Subtitle, StyledParagraph, Input } from '../tools/common';
import { useContext, useEffect, useState } from 'react';
import Tabs from 'react-bootstrap/esm/Tabs';
import Tab from 'react-bootstrap/esm/Tab';
import { Client } from 'src/renderer/src/client';
import { StartupItemComponent } from '@renderer/components/StartupItem';
import withServicesState from '@renderer/store/hocs/withServicesState';
import { LoadingState } from 'src/main/orchestrator/orchestrator.types';
import { ToastsContext } from '@renderer/components/toasts';

type CommonProps = {
  client: Client;
  getConfig: () => Promise<{
    DerivedConfig: { EthNodeURLs: string[] };
  }>;
  logout: () => Promise<void>;
  updateEthNodeUrl: (url: string) => Promise<void>;
  updateFailoverSetting: (setting: boolean) => Promise<void>;
};

const Common = (props: CommonProps) => {
  const [ethNodeUrl, setEthUrl] = useState<string>('');
  const [useFailover, setUseFailover] = useState<boolean>(false);

  useEffect(() => {
    (async () => {
      const cfg = await props.getConfig();
      const customUrl = cfg?.DerivedConfig?.EthNodeURLs[0] || '';
      setEthUrl(customUrl);
      const failoverSettings = await props.client.getFailoverSetting();
      setUseFailover(Boolean(failoverSettings.isEnabled));
    })();
  }, []);

  return (
    <>
      <Sp mt={3}>
        <Subtitle>Reset</Subtitle>
        <StyledParagraph>Set up your wallet from scratch.</StyledParagraph>
        <StyledBtn onClick={() => props.logout()}>Reset</StyledBtn>
      </Sp>
      <Sp mt={3}>
        <Subtitle>Set Custom ETH Node</Subtitle>
        <Input
          id="eth-node-url"
          placeholder={'{wss|https}://{url}'}
          style={{ width: '500px', marginBottom: '1.5em', marginTop: '0' }}
          value={ethNodeUrl}
          onChange={(e) => setEthUrl(e.value)}
        />
        <StyledBtn onClick={() => props.updateEthNodeUrl(ethNodeUrl)}>
          Set
        </StyledBtn>
      </Sp>
      <Sp mt={3}>
        <Subtitle>Failover Mechanism</Subtitle>
        <StyledParagraph>
          A failover policy is applied when a provider is unable to service an
          open session. This policy ensures continuity by automatically
          rerouting or reassigning sessions to an alternate provider, minimizing
          service disruptions and maintaining a seamless user experience
        </StyledParagraph>
        <Sp mt={2} mb={2}>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'left',
            }}
          >
            <input
              type="checkbox"
              checked={useFailover}
              onChange={(e) => {
                setUseFailover(Boolean(e.target.checked));
              }}
              style={{ marginRight: '5px' }}
            />
            <div>Use Default Policy (set by proxy-router)</div>
          </div>
        </Sp>
        <StyledBtn onClick={() => props.updateFailoverSetting(useFailover)}>
          Apply
        </StyledBtn>
      </Sp>
    </>
  );
};

type SettingsProps = CommonProps & {
  services: LoadingState;
};

const Settings = (props: SettingsProps) => {
  const toast = useContext(ToastsContext);

  return (
    <View data-testid="agents-container">
      <LayoutHeader title="Settings" />

      <Tabs defaultActiveKey="common" className="mb-3">
        <Tab eventKey="common" title="Common">
          <Common
            client={props.client}
            getConfig={props.getConfig}
            logout={props.logout}
            updateEthNodeUrl={props.updateEthNodeUrl}
            updateFailoverSetting={props.updateFailoverSetting}
          />
        </Tab>
        <Tab eventKey="services" title="Services">
          {props.services.startup.map((item) => (
            <StartupItemComponent
              key={item.id}
              item={item}
              alwaysShowPingRestart={true}
              onRestart={() =>
                props.client.restartService({ service: item.id })
              }
              onPing={async () => {
                const res = await props.client.pingService({
                  service: item.id,
                });
                if (res === true) {
                  toast.toast('success', 'Ping successful');
                } else {
                  toast.toast('error', 'Ping failed');
                }
              }}
            />
          ))}
        </Tab>
      </Tabs>
    </View>
  );
};

export default withServicesState(withSettingsState(Settings));
