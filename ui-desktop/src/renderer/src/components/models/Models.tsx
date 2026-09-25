import { lazy, Suspense, useState } from 'react';
import {
  useInfiniteQuery,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import withModelsState from '../../store/hocs/withModelsState';

import { LayoutHeader } from '../common/LayoutHeader';
import { View } from '../common/View';
import { BtnAccent } from '../dashboard/BalanceBlock.styles';
import ModelsTable from './ModelsTable';
import Tab from 'react-bootstrap/Tab';
import Tabs from 'react-bootstrap/Tabs';
import styled from 'styled-components';
import { queryKeys } from '../../store/queries';
import { modelPagesQueryOptions } from '../../store/modelQueries';
import { normalizeModelList } from '../../store/utils/modelMetadata';
import QueryError from '../common/QueryError';

const FileSelectionModal = lazy(() => import('./FileSelectionModal'));
const PinnedFilesTable = lazy(() => import('./PinnedFilesTable'));

const Container = styled.div`
    overflow-y: auto;
    
    .nav-link {
        color: ${(p) => p.theme.colors.morMain}
    }

    .nav-link.active {
        color: ${(p) => p.theme.colors.morMain}
        border-color: ${(p) => p.theme.colors.morMain}
        background-color: rgba(0,0,0,0.4);
    }
`;

const IpfsStatus = styled.div`
  color: ${(p) => p.theme.colors.morMain};
  font-size: 1.2rem;
`;

export const Models = ({
  setSelectedModel,
  getIpfsVersion,
  getModelsPage,
  openSelectDownloadFolder,
  addFileToIpfs,
  getPinnedFiles,
  pinFile,
  unpinFile,
  toasts,
  client,
  config,
}: any) => {
  const [openChangeModal, setOpenChangeModal] = useState(false);
  const [activeTab, setActiveTab] = useState('registry');
  const queryClient = useQueryClient();

  // Cached, stale-while-revalidate data so revisiting the Models tab renders
  // instantly and refreshes in the background instead of refetching on mount.
  const ipfsVersionQuery = useQuery({
    queryKey: queryKeys.ipfsVersion,
    queryFn: getIpfsVersion,
  });
  const modelsQuery = useInfiniteQuery({
    ...modelPagesQueryOptions(getModelsPage),
  });
  const pinnedFilesQuery = useQuery({
    queryKey: queryKeys.pinnedFiles,
    queryFn: getPinnedFiles,
    // Enumerating the local IPFS repository can be expensive. Nothing on
    // the registry tab needs it, so wait until the user opens Pinned Models.
    enabled: activeTab === 'pinned',
  });

  const ipfsVersion = (ipfsVersionQuery.data as any)?.version ?? null;
  const isIpfsConnected = !!ipfsVersion;
  const models = Array.from(
    new Map(
      normalizeModelList((modelsQuery.data?.pages ?? []).flat()).map(
        (model: any) => [model.Id, model],
      ),
    ).values(),
  );
  const pinnedFiles = pinnedFilesQuery.data ?? [];

  const reload = () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.ipfsVersion });
    queryClient.invalidateQueries({ queryKey: queryKeys.allModels });
    queryClient.invalidateQueries({ queryKey: queryKeys.modelPages });
    queryClient.invalidateQueries({ queryKey: queryKeys.pinnedFiles });
  };

  const handleUnpinFile = async (hash) => {
    try {
      const response = await unpinFile(hash);
      if (response) {
        toasts.toast('success', 'File unpinned successfully');
        queryClient.setQueryData(queryKeys.pinnedFiles, (old: any[] = []) =>
          old.filter(
            (file: any) =>
              file.metadataCIDHash !== hash && file.fileCIDHash !== hash,
          ),
        );
      } else {
        toasts.toast('error', 'Failed to unpin file');
      }
    } catch (error) {
      toasts.toast('error', 'Failed to unpin file');
      console.error('Error', error);
    }
  };

  const onPinModel = async (hash) => {
    const response = await pinFile(hash);
    reload();
    return response;
  };

  return (
    <View data-testid="models-container">
      {isIpfsConnected ? (
        <IpfsStatus>
          <span>IPFS Connected. Version: {ipfsVersion}</span>
        </IpfsStatus>
      ) : (
        <IpfsStatus>
          <span>IPFS is not connected</span>
        </IpfsStatus>
      )}
      <QueryError
        error={modelsQuery.error}
        what="models"
        onRetry={() => modelsQuery.refetch()}
      />
      <LayoutHeader title="Models">
        <BtnAccent
          style={{ padding: '1.5rem' }}
          onClick={() => setOpenChangeModal(true)}
        >
          Pin Model
        </BtnAccent>
      </LayoutHeader>
      <Container>
        <Tabs
          activeKey={activeTab}
          onSelect={(key) => setActiveTab(key ?? 'registry')}
          id="tab-models"
          className="mb-3"
        >
          <Tab eventKey="registry" title="Registry">
            <ModelsTable
              setSelectedModel={setSelectedModel}
              models={models}
              isLoading={modelsQuery.isPending}
              hasMore={modelsQuery.hasNextPage}
              isFetchingMore={modelsQuery.isFetchingNextPage}
              onLoadMore={() => modelsQuery.fetchNextPage()}
              openSelectDownloadFolder={openSelectDownloadFolder}
              toasts={toasts}
              client={client}
              config={config}
            />
          </Tab>
          <Tab eventKey="pinned" title="Pinned Models">
            {activeTab === 'pinned' && (
              <Suspense
                fallback={<IpfsStatus>Loading pinned models…</IpfsStatus>}
              >
                <PinnedFilesTable
                  pinnedFiles={pinnedFiles}
                  unpinFile={handleUnpinFile}
                  toasts={toasts}
                />
              </Suspense>
            )}
          </Tab>
        </Tabs>
      </Container>
      {openChangeModal && (
        <Suspense fallback={null}>
          <FileSelectionModal
            isActive
            addFileToIpfs={addFileToIpfs}
            pinFile={onPinModel}
            toasts={toasts}
            handleClose={() => setOpenChangeModal(false)}
          />
        </Suspense>
      )}
    </View>
  );
};

export default withModelsState(Models);
