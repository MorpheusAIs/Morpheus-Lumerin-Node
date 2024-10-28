import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import { Sp } from '../common';
import withSettingsState from '../../store/hocs/withSettingsState';
import { StyledBtn, Subtitle, StyledParagraph, Input } from '../tools/common';
import { useEffect, useState } from 'react';

const Settings = (props) => {
    const [ethNodeUrl, setEthUrl] = useState("");

    useEffect(() => {
       props.getConfig().then(cfg => {
        const customUrl = cfg?.DerivedConfig?.EthNodeURLs[0] || "";
        setEthUrl(customUrl);
       })
    },[])
    
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
                        style={{ width: '500px'}}
                        value={ethNodeUrl}
                        onChange={(e) => setEthUrl(e.value)} />
                </StyledParagraph>
                
                <StyledBtn onClick={() => props.updateEthNodeUrl(ethNodeUrl)}>
                    Set
                </StyledBtn>
            </Sp>
        </View>)
}

export default withSettingsState(Settings);