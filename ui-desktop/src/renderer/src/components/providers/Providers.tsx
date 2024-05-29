import { useEffect, useState } from 'react'
import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import ProvidersList from './ProvidersList'
import { withRouter } from 'react-router-dom';
import { BtnAccent } from '../dashboard/BalanceBlock.styles';

import withProvidersState from "../../store/hocs/withProvidersState";
const Providers = ({ fetchData, providerId }) => {

    const [data, setData] = useState();

    useEffect(() => {
        (async () => {
            const data = await fetchData(providerId);
            console.log(data);
            setData(data);
        })()
    }, [])

    return (    
    <View data-testid="models-container">
        <LayoutHeader title="Providers">
            <BtnAccent style={{ padding: '1.5rem'}}>Add provider</BtnAccent>
        </LayoutHeader>
        <ProvidersList data={data} />
    </View>)
}

export default withRouter(withProvidersState(Providers));