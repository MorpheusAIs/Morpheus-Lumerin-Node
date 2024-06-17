import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import { Sp } from '../common';
import withSettingsState from '../../store/hocs/withSettingsState';
import { StyledBtn, Subtitle, StyledParagraph } from '../tools/common';

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
        </View>)
}

export default withSettingsState(Settings);