import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import { Sp } from '../common';
import withSettingsState from '../../store/hocs/withSettingsState';
import { StyledBtn, Subtitle, StyledParagraph, Input } from '../tools/common';
import { useEffect, useState } from 'react';

const Settings = (props) => {
    const [ethNodeUrl, setEthUrl] = useState<string>("");
    const [useFailover, setUseFailover] = useState<boolean>(false);

    useEffect(() => {
        (async () => {
            const cfg = await props.getConfig();
            const customUrl = cfg?.DerivedConfig?.EthNodeURLs[0] || "";
            setEthUrl(customUrl);
            const failoverSettings = await props.client.getFailoverSetting();
            setUseFailover(Boolean(failoverSettings.isEnabled));
        })()
    }, [])

    return (
        <View data-testid="agents-container">
            <LayoutHeader title="Settings" />
            <Sp mt={2}>
                <Subtitle>Reset</Subtitle>
                <StyledParagraph>
                    Set up your wallet from scratch.
                </StyledParagraph>
                <StyledBtn onClick={() => props.logout()}>
                    Reset
                </StyledBtn>
            </Sp>
            <Sp mt={2}>
                <Subtitle>Set Custom ETH Node</Subtitle>
                <StyledParagraph>
                    <Input
                        placeholder={"{wss|https}://{url}"}
                        style={{ width: '500px' }}
                        value={ethNodeUrl}
                        onChange={(e) => setEthUrl(e.value)} />
                </StyledParagraph>

                <StyledBtn onClick={() => props.updateEthNodeUrl(ethNodeUrl)}>
                    Set
                </StyledBtn>
            </Sp>
            <Sp mt={2}>
                <Subtitle>Failover Mechanism</Subtitle>
                <StyledParagraph>
                    A failover policy is applied when a provider is unable to service an open session. This policy ensures continuity by automatically rerouting or reassigning sessions to an alternate provider, minimizing service disruptions and maintaining a seamless user experience
                </StyledParagraph>
                <Sp mt={2} mb={2}>
                    <div style={{ display: "flex", alignItems: 'center', justifyContent: "left" }}>
                        <input 
                            type="checkbox"
                            checked={useFailover}
                            onChange={(e) => {
                                setUseFailover(Boolean(e.target.checked))
                            } }
                            style={{ marginRight: '5px' }}

                        />
                        <div>Use Default Policy (set by proxy-router)</div>
                    </div>
                </Sp>
                <StyledBtn onClick={() => props.updateFailoverSetting(useFailover)}>
                    Apply
                </StyledBtn>
            </Sp>
        </View>)
}

export default withSettingsState(Settings);