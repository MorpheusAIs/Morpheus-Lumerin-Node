import { useQuery } from '@tanstack/react-query'
import { LayoutHeader } from '../common/LayoutHeader'
import { View } from '../common/View'
import ProvidersList from './ProvidersList'

import { BtnAccent } from '../dashboard/BalanceBlock.styles';

import withProvidersState from "../../store/hocs/withProvidersState";
import { queryKeys } from '../../store/queries';
import QueryError from '../common/QueryError'
const Providers = ({ fetchData, providerId }) => {

    // Cached, stale-while-revalidate: revisiting the Providers tab renders the
    // last result instantly and revalidates in the background instead of
    // re-running the expensive per-session balance fetch on every mount.
    const { data, error, refetch } = useQuery({
        queryKey: queryKeys.providerData(providerId),
        queryFn: () => fetchData(providerId),
        enabled: !!providerId,
    });

    return (
    <View data-testid="providers-container">
        <QueryError error={error} what="provider sessions" onRetry={() => refetch()} />
        <LayoutHeader title="Providers">
            <BtnAccent style={{ padding: '1.5rem'}} disabled>Add provider</BtnAccent>
        </LayoutHeader>
        <ProvidersList data={data} providerId={providerId} />
    </View>)
}

export default withProvidersState(Providers);