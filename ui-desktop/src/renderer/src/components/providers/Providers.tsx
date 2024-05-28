import { useEffect } from 'react'
import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import ProvidersList from './ProvidersList'
import withProvidersState  from '../../store/hocs/withProvidersState';
import { BtnAccent } from '../dashboard/BalanceBlock.styles';

export const Providers = (props) => {

    useEffect(() => {

    }, [])

    return (    
    <View data-testid="models-container">
        <LayoutHeader title="Providers">
            <BtnAccent style={{ padding: '1.5rem'}}>Add provider</BtnAccent>
        </LayoutHeader>
        <ProvidersList />
    </View>)
}

export default withProvidersState(Providers)