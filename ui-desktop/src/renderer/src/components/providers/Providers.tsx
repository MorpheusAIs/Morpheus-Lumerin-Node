import { useEffect } from 'react'
import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import ProvidersList from './ProvidersList'
import { withModelsState } from '../../store/hocs/withModelsState';

export const Providers = (props) => {

    useEffect(() => {

    }, [])

    return (    
    <View data-testid="models-container">
        <LayoutHeader title="Providers">
            Add provider
        </LayoutHeader>
        <ProvidersList />
    </View>)
}

export default withModelsState(Providers)