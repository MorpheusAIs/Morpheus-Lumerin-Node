import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import { Sp } from '../common';
import withSettingsState from '../../store/hocs/withSettingsState';
import { StyledBtn, Subtitle, StyledParagraph, Input } from '../tools/common';

const Settings = (props) => {
    return (
        <View data-testid="agents-container">
            <LayoutHeader title="Settings" />
            <Sp mt={1}>
                <Subtitle>Reset</Subtitle>
                <StyledParagraph>
                    Set up your wallet from scratch.
                </StyledParagraph>
                <StyledBtn onClick={() => props.logout()}>
                    Reset
                </StyledBtn>
            </Sp>
            <Sp mt={1}>
                <Subtitle>Set Custom ETH Node</Subtitle>
                <StyledParagraph>
                    <Input />
                </StyledParagraph>
                
                <StyledBtn onClick={() => props.logout()}>
                    Set
                </StyledBtn>
            </Sp>
        </View>)
}

export default withSettingsState(Settings);