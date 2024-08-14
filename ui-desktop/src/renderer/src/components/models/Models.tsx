import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import ModelsTable from './ModelsTable';

export const Models = () => {
    return (    
    <View data-testid="models-container">
        <LayoutHeader title="Models" />
        <ModelsTable />
    </View>)
}