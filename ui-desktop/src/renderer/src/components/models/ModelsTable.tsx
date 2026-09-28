import { useEffect, useRef, useState } from 'react';

import styled from 'styled-components';

import Card from 'react-bootstrap/Card';
import { abbreviateAddress } from '../../utils';
import {
  IconDownload,
  IconCopy,
  IconCoin,
  IconTag,
  IconHash,
  IconX,
} from '@tabler/icons-react';
import ProgressBar from 'react-bootstrap/ProgressBar';
import { ModelActionButton } from './ModelActionButton';
import {
  ModelCardEmptyState,
  ModelCardGrid,
  ModelCardSurface,
} from './ModelCardSurface';

// Event payload for download progress events from the SSE stream
interface DownloadProgressEvent {
  status: 'downloading' | 'completed' | 'error';
  downloaded: number;
  total: number;
  percentage: number;
  error?: string;
  timeUpdated: number;
}

// Type for the progress callback function
type DownloadProgressCallback = (event: DownloadProgressEvent) => void;

// Registry cards open the model when clicked, which the pinned files list does
// not, so the affordance lives here rather than on the shared surface.
const SelectableCard = styled(ModelCardSurface)`
  cursor: pointer;
`;

const ResultsFooter = styled.div`
  align-items: center;
  color: var(--text-muted);
  display: flex;
  flex-basis: 100%;
  flex-direction: column;
  gap: 0.8rem;
  justify-content: center;
  padding: 1.6rem 0 2.4rem;
`;

const LoadMoreButton = styled.button`
  background: var(--surface-hover);
  border: 1px solid var(--border-strong);
  border-radius: 8px;
  color: var(--accent);
  cursor: pointer;
  font: inherit;
  min-height: 4rem;
  padding: 0.8rem 1.6rem;

  &:hover:not(:disabled) {
    background: var(--surface-raised);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

// Overlays the card while a model file is coming down from IPFS.
const DownloadProgressContainer = styled.div`
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  /* Opaque rather than translucent black: the card underneath still shows its
     own copy, and a see-through overlay made both sets of text unreadable. */
  background: var(--surface-base);
  border: 1px solid var(--border-strong);
  padding: 1rem;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  z-index: 10;

  .progress-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    h4 {
      margin: 0;
      font-size: 1.5rem;
      color: var(--accent);
    }
  }

  .progress-info {
    display: flex;
    justify-content: space-between;
    font-size: 0.85rem;
    color: var(--text-muted);
    margin-top: 0.5rem;
  }

  .progress {
    --bs-progress-bg: var(--surface-hover);
    --bs-progress-bar-bg: var(--accent);
    height: 8px;
    border-radius: 4px;
  }
`;

function ModelCard({
  onSelect,
  model,
  openSelectDownloadFolder,
  toasts,
  client,
}) {
  const [isDownloading, setIsDownloading] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const [downloadedSize, setDownloadedSize] = useState('0 KB');
  const [totalSize, setTotalSize] = useState('0 KB');
  const [latestUploadTime, setLatestUploadTime] = useState(0);
  const cancelDownloadRef = useRef<(() => void) | null>(null);

  useEffect(
    () => () => {
      cancelDownloadRef.current?.();
      cancelDownloadRef.current = null;
    },
    [],
  );

  const formatBytes = (bytes, decimals = 2) => {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];

    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
  };

  const handleDownloadError = (error) => {
    if (typeof error === 'string') {
      if (error.includes('invalid CID')) {
        toasts.toast('error', 'Invalid CID specified in the model.');
      } else if (error.includes('failed to find')) {
        toasts.toast('error', 'Model is not found in IPFS.');
      } else {
        toasts.toast('error', 'Failed to download model');
      }
    } else {
      toasts.toast('error', 'Failed to download model');
    }
    setIsDownloading(false);
  };

  const handleFolderSelect = async (e) => {
    e.stopPropagation();
    try {
      const result = await openSelectDownloadFolder();
      const { canceled, folderToken } = result;
      if (canceled || !folderToken) {
        return;
      }

      // Start download with progress tracking
      setIsDownloading(true);
      setDownloadProgress(0);
      setDownloadedSize('0 KB');
      setTotalSize('0 KB');
      setLatestUploadTime(Date.now());

      // Use streaming download
      cancelDownloadRef.current = streamIpfsFileDownload({
        cid: model.IpfsCID || model.metadataCIDHash,
        folderToken,
        onProgress: (progressEvent) => {
          const { downloaded, total, percentage, timeUpdated } = progressEvent;
          setDownloadProgress(percentage);
          setDownloadedSize(formatBytes(downloaded));
          setTotalSize(formatBytes(total));
          setLatestUploadTime(timeUpdated);
        },
        onComplete: () => {
          setIsDownloading(false);
          toasts.toast('success', 'Model downloaded successfully');
          cancelDownloadRef.current = null;
        },
        onError: (error) => {
          setIsDownloading(false);
          toasts.toast('error', `Failed to download model: ${error}`);
          cancelDownloadRef.current = null;
        },
      });
    } catch (error) {
      handleDownloadError(error);
    }
  };

  const streamIpfsFileDownload = ({
    cid,
    folderToken,
    onProgress,
    onComplete,
    onError,
  }: {
    cid: string;
    folderToken: string;
    onProgress: DownloadProgressCallback;
    onComplete: DownloadProgressCallback;
    onError: (error: string) => void;
  }): (() => void) => {
    let cancelled = false;
    let settled = false;
    const requestId = window.crypto.randomUUID();
    const unsubscribe = client.onIpfsDownloadEvent({
      requestId,
      listener: (event) => {
        if (cancelled || settled) return;
        if (event.kind === 'error') {
          settled = true;
          unsubscribe();
          onError(event.message);
          return;
        }
        if (event.progress.status === 'completed') {
          settled = true;
          unsubscribe();
          onComplete(event.progress);
          return;
        }
        onProgress(event.progress);
      },
    });

    (async () => {
      try {
        onProgress({
          status: 'downloading',
          downloaded: 0,
          total: 0,
          percentage: 0,
          timeUpdated: Date.now(),
        });
        await client.startIpfsDownload({
          requestId,
          folderToken,
          cidHash: cid,
        });
      } catch (error: unknown) {
        if (!cancelled && !settled) {
          settled = true;
          unsubscribe();
          const errorMessage =
            error instanceof Error ? error.message : String(error);
          onError(`Failed to download: ${errorMessage || 'Unknown error'}`);
        }
      }
    })();

    return () => {
      if (cancelled || settled) return;
      cancelled = true;
      unsubscribe();
      client.cancelIpfsDownload({ requestId });
    };
  };

  const cancelDownload = (e) => {
    e.stopPropagation();
    if (cancelDownloadRef.current) {
      cancelDownloadRef.current();
      cancelDownloadRef.current = null;
      setIsDownloading(false);
      toasts.toast('info', 'Download canceled');
    }
  };

  const copyToClipboard = async (text: string, label: string) => {
    try {
      // Use the Electron clipboard bridge; navigator.clipboard silently
      // fails in the renderer (focus/permissions), see issue #793
      await window.copyToClipboard(text);
      toasts.toast('success', `${label} copied to clipboard`, {
        autoClose: 700,
      });
    } catch (e) {
      toasts.toast('error', `Failed to copy ${label} to clipboard`);
    }
  };

  const copyId = () => copyToClipboard(model.Id, 'ID');

  const copyCIDHash = () => copyToClipboard(model.IpfsCID, 'CID Hash');

  // Format MOR values to prevent scientific notation and limit decimals
  const formatMorValue = (value) => {
    if (!value) return '0 MOR';

    // Convert to MOR by dividing by 10^18
    const morValue = value / 10 ** 18;

    // For very small values, use a different format to avoid scientific notation
    if (morValue < 0.000001) {
      return morValue.toFixed(12).replace(/\.?0+$/, '') + ' MOR';
    } else if (morValue < 0.001) {
      return morValue.toFixed(8).replace(/\.?0+$/, '') + ' MOR';
    } else if (morValue < 1) {
      return morValue.toFixed(6).replace(/\.?0+$/, '') + ' MOR';
    } else {
      return morValue.toFixed(4).replace(/\.?0+$/, '') + ' MOR';
    }
  };

  const formatDate = (date) => {
    return date.toLocaleString(undefined, {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  return (
    <SelectableCard onClick={() => onSelect(model.Id)}>
      {isDownloading && (
        <DownloadProgressContainer>
          <div className="progress-header">
            <h4>Downloading Model</h4>
            <ModelActionButton
              type="button"
              aria-label={`Cancel download of ${model.Name || 'model'}`}
              title="Cancel download"
              onClick={cancelDownload}
            >
              <IconX size={18} aria-hidden="true" />
            </ModelActionButton>
          </div>

          {/* No variant and no className: react-bootstrap puts className on the
              track, where "progress-bar" collided with the class of the fill
              inside it, and bg-success would have overridden the app accent. */}
          <ProgressBar now={downloadProgress} />

          <div className="progress-info">
            <span>
              {downloadedSize} / {totalSize}
            </span>
            <span>{downloadProgress.toFixed(1)}%</span>
          </div>
          <div className="progress-info">
            <span>
              Last updated at: {formatDate(new Date(latestUploadTime))}
            </span>
          </div>
        </DownloadProgressContainer>
      )}

      <Card.Body>
        <Card.Title as={'div'} className="model-card-title">
          <span className="model-card-name">
            {model.Name || 'Unnamed Model'}
          </span>
          <ModelActionButton
            type="button"
            aria-label={`Download ${model.Name || 'model'}`}
            title="Download model file"
            disabled={isDownloading}
            onClick={handleFolderSelect}
          >
            <IconDownload size={18} aria-hidden="true" />
          </ModelActionButton>
        </Card.Title>

        <div className="model-info-section">
          <div className="model-info-item">
            <span className="info-label">
              <IconHash size={16} strokeWidth={2} />
              ID:
            </span>
            <div className="info-value">
              <span className="hash-container monospace">
                {abbreviateAddress(model?.Id || '', 6)}
                <ModelActionButton
                  type="button"
                  aria-label={`Copy model ID for ${model.Name || 'model'}`}
                  title="Copy model ID"
                  disabled={!model.Id}
                  onClick={(e) => {
                    e.stopPropagation();
                    copyId();
                  }}
                >
                  <IconCopy size={16} aria-hidden="true" />
                </ModelActionButton>
              </span>
            </div>
          </div>

          <div className="model-info-item">
            <span className="info-label">
              <IconHash size={16} strokeWidth={2} />
              CID Hash:
            </span>
            <div className="info-value">
              <span className="hash-container monospace">
                {abbreviateAddress(model?.IpfsCID, 6)}
                <ModelActionButton
                  type="button"
                  aria-label={`Copy CID hash for ${model.Name || 'model'}`}
                  title="Copy CID hash"
                  disabled={!model.IpfsCID}
                  onClick={(e) => {
                    e.stopPropagation();
                    copyCIDHash();
                  }}
                >
                  <IconCopy size={16} aria-hidden="true" />
                </ModelActionButton>
              </span>
            </div>
          </div>

          {model.Fee ? (
            <div className="model-info-item">
              <span className="info-label">
                <IconCoin size={16} strokeWidth={2} />
                Fee:
              </span>
              <span className="info-value">{formatMorValue(model.Fee)}</span>
            </div>
          ) : null}

          {model.Stake ? (
            <div className="model-info-item">
              <span className="info-label">
                <IconCoin size={16} strokeWidth={2} />
                Stake:
              </span>
              <span className="info-value">{formatMorValue(model.Stake)}</span>
            </div>
          ) : null}

          {model.Tags && model.Tags.length > 0 && (
            <div className="model-info-item">
              <span className="info-label">
                <IconTag size={16} strokeWidth={2} />
                Tags:
              </span>
              <div className="info-value">
                <div className="tag-container">
                  {model.Tags.map((tag, index) => (
                    <span key={index} className="tag-item">
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      </Card.Body>
    </SelectableCard>
  );
}

function ModelsTable({
  setSelectedModel,
  models,
  isLoading,
  hasMore,
  isFetchingMore,
  onLoadMore,
  client,
  openSelectDownloadFolder,
  toasts,
}: any) {
  const onSelect = (id) => {
    setSelectedModel(models.find((x) => x.Id == id));
  };

  return (
    <ModelCardGrid>
      {models.length ? (
        <>
          {models.map((x) => (
            <div key={x.Id}>
              <ModelCard
                onSelect={onSelect}
                model={x}
                openSelectDownloadFolder={openSelectDownloadFolder}
                toasts={toasts}
                client={client}
              />
            </div>
          ))}
          <ResultsFooter aria-live="polite">
            <span>{models.length} models loaded</span>
            {hasMore && (
              <LoadMoreButton
                type="button"
                disabled={isFetchingMore}
                onClick={onLoadMore}
              >
                {isFetchingMore ? 'Loading more…' : 'Show more models'}
              </LoadMoreButton>
            )}
          </ResultsFooter>
        </>
      ) : (
        <ModelCardEmptyState aria-live="polite">
          {isLoading ? 'Loading model registry…' : 'No models found'}
        </ModelCardEmptyState>
      )}
    </ModelCardGrid>
  );
}

export default ModelsTable;
