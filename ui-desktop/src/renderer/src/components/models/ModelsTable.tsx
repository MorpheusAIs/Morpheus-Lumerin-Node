import { useRef, useState } from 'react';

import styled from 'styled-components';

import Card from 'react-bootstrap/Card';
import { abbreviateAddress } from '../../utils';
import { IconDownload, IconCopy, IconCoin, IconTag, IconHash, IconX } from '@tabler/icons-react';
import ProgressBar from 'react-bootstrap/ProgressBar';
import path from 'path';


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


const CustomCard = styled(Card)`
  background: linear-gradient(145deg, #244a47 0%, #1d3c39 100%) !important;
  color: #21dc8f !important;
  border: 1px solid rgba(33, 220, 143, 0.2) !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: all 0.2s ease-in-out;
  border-radius: 12px !important;
  overflow: hidden;
  cursor: pointer !important;
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.25);
    border-color: rgba(33, 220, 143, 0.4) !important;
  }

  p {
    color: white !important;
  }

  .card-title {
    margin-bottom: 5px;
    font-weight: 600;
    font-size: 1.3rem;
    letter-spacing: 0.02em;
    color: #21dc8f;
  }

  .card-subtitle {
    font-size: 0.85rem;
    color: rgba(255, 255, 255, 0.7) !important;
    margin-bottom: 16px;
  }

  .card-body {
    padding: 1.5rem;
  }

  .model-info-section {
    display: flex;
    flex-direction: column;
    gap: 3px;
    // margin-bottom: 5px;
    padding-top: 8px;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
  }

  .model-info-item {
    display: flex;
    align-items: center;
    font-size: 1.1rem;
    padding: 4px 0;
  }

  .info-label {
    font-weight: 600;
    min-width: 90px;
    color: rgba(255, 255, 255, 0.9);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .info-value {
    color: white;
    display: flex;
    align-items: center;
  }

  .icon-button {
    cursor: pointer;
    padding: 8px;
    border-radius: 50%;
    transition: all 0.2s;
    background: rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.8);
    
    &:hover {
      background: rgba(33, 220, 143, 0.15);
      color: #21dc8f;
      transform: translateY(-2px);
    }
  }

  .copy-button {
    background: rgba(33, 220, 143, 0.15);
    color: white;
    padding: 4px 8px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    cursor: pointer;
    margin-left: 10px;
    transition: all 0.2s;
    
    &:hover {
      background: rgba(33, 220, 143, 0.3);
    }
    
    svg {
      margin-right: 4px;
    }
  }

  .tag-container {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }

  .tag-item {
    background: rgba(33, 220, 143, 0.15);
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 1rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 22px;
    line-height: 1;
    transition: all 0.2s;
    border: 1px solid rgba(33, 220, 143, 0.1);
    
    &:hover {
      background: rgba(33, 220, 143, 0.25);
      transform: translateY(-2px);
    }
  }
  
  .monospace {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.85rem;
    letter-spacing: -0.03em;
  }
  
  .hash-container {
    background: rgba(0, 0, 0, 0.2);
    border-radius: 6px;
    padding: 6px 10px;
    display: flex;
    align-items: center;
    font-size: 1.1rem;
  }
`;

const Container = styled.div`
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  gap: 28px;
  max-height: 75vh;
  padding: 8px 4px;
  overflow-y: auto;
  
  &::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }
  
  &::-webkit-scrollbar-track {
    background: rgba(0, 0, 0, 0.1);
    border-radius: 4px;
  }
  
  &::-webkit-scrollbar-thumb {
    background: rgba(33, 220, 143, 0.3);
    border-radius: 4px;
  }
  
  &::-webkit-scrollbar-thumb:hover {
    background: rgba(33, 220, 143, 0.5);
  }
`;

// New styled component for progress bar container
const DownloadProgressContainer = styled.div`
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.85);
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
      color: #21dc8f;
    }
    
    .cancel-button {
      cursor: pointer;
      background: rgba(255, 0, 0, 0.2);
      border-radius: 50%;
      width: 28px;
      height: 28px;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: all 0.2s;
      
      &:hover {
        background: rgba(255, 0, 0, 0.3);
        transform: scale(1.05);
      }
    }
  }
  
  .progress-info {
    display: flex;
    justify-content: space-between;
    font-size: 0.85rem;
    color: rgba(255, 255, 255, 0.7);
    margin-top: 0.5rem;
  }
  
  .progress-bar {
    height: 8px;
    border-radius: 4px;
    background-color:rgb(137, 138, 137);
  }
`;

function ModelCard({ onSelect, model, openSelectDownloadFolder, toasts, client, config }) {
  const [isDownloading, setIsDownloading] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const [downloadedSize, setDownloadedSize] = useState('0 KB');
  const [totalSize, setTotalSize] = useState('0 KB');
  const [latestUploadTime, setLatestUploadTime] = useState(0);
  const cancelDownloadRef = useRef<(() => void) | null>(null);

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
      if (error.includes("invalid CID")) {
        toasts.toast("error", "Invalid CID specified in the model.");
      } else if (error.includes("failed to find")) {
        toasts.toast("error", "Model is not found in IPFS.");
      } else {
        toasts.toast("error", "Failed to download model");
      }
    } else {
      toasts.toast("error", "Failed to download model");
    }
    setIsDownloading(false);
  }

  const handleFolderSelect = async (e) => {
    e.stopPropagation();
    try {
      const result = await openSelectDownloadFolder();
      const { canceled, filePaths } = result;
      if (canceled) {
        return;
      }

      const folderPath = filePaths[0];

      // Start download with progress tracking
      setIsDownloading(true);
      setDownloadProgress(0);
      setDownloadedSize('0 KB');
      setTotalSize('0 KB');
      setLatestUploadTime(Date.now());

      const filePath = path.join(folderPath, model.IpfsCID || model.metadataCIDHash);
      // Use streaming download
      cancelDownloadRef.current = streamIpfsFileDownload({
        cid: model.IpfsCID || model.metadataCIDHash,
        destinationPath: filePath,
        onProgress: (progressEvent) => {
          const { downloaded, total, percentage, timeUpdated } = progressEvent;
          setDownloadProgress(percentage);
          setDownloadedSize(formatBytes(downloaded));
          setTotalSize(formatBytes(total));
          setLatestUploadTime(timeUpdated);
        },
        onComplete: () => {
          setIsDownloading(false);
          toasts.toast("success", "Model downloaded successfully");
          cancelDownloadRef.current = null;
        },
        onError: (error) => {
          setIsDownloading(false);
          toasts.toast("error", `Failed to download model: ${error}`);
          cancelDownloadRef.current = null;
        }
      });
    } catch (error) {
      handleDownloadError(error);
    }
  };

  const streamIpfsFileDownload = ({
    cid,
    destinationPath,
    onProgress,
    onComplete,
    onError
  }: {
    cid: string,
    destinationPath: string,
    onProgress: DownloadProgressCallback,
    onComplete: DownloadProgressCallback,
    onError: (error: string) => void
  }): () => void => {
    // Create AbortController for cancellation
    const controller = new AbortController();
    const { signal } = controller;

    // Start the download
    (async () => {
      try {
        const authHeaders = await client.getAuthHeaders();
        const destEncoded = encodeURIComponent(destinationPath);
        const url = `${config.chain.localProxyRouterUrl}/ipfs/download/stream/${cid}?dest=${destEncoded}`;

        // Use fetch API with streaming enabled
        const response = await fetch(url, {
          method: 'GET',
          headers: authHeaders,
          signal: signal,
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        // Set up a reader for the response body stream
        const reader = response.body?.getReader();
        if (!reader) {
          throw new Error('Failed to get response reader');
        }

        // Initial progress state
        let downloaded = 0;
        let lastProgressUpdate = Date.now();
        const progressUpdateInterval = 100; // Update progress at most every 100ms
        const textDecoder = new TextDecoder();

        // Process the stream
        while (true) {
          const { done, value } = await reader.read();

          if (done) {
            // Download completed successfully
            onComplete({
              status: 'completed',
              downloaded,
              total: downloaded,
              percentage: 100,
              timeUpdated: Date.now()
            });
            break;
          }
          const decodedString = textDecoder.decode(value, { stream: true });
          const objects = decodedString.split('data: ').filter(Boolean).map(s => {
            try {
              return JSON.parse(s);
            } catch (e) {
              return null;
            }
          }).filter(Boolean);

          if (objects.length === 0) {
            continue;
          }

          const latestProgress = objects[objects.length - 1];

          if (latestProgress.error) {
            handleDownloadError(latestProgress.error);
            break;
          }

          const now = Date.now();
          if (now - lastProgressUpdate > progressUpdateInterval) {
            lastProgressUpdate = now;

            onProgress({
              status: 'downloading',
              downloaded: latestProgress.downloaded,
              total: latestProgress.total,
              percentage: latestProgress.percentage,
              timeUpdated: lastProgressUpdate
            });
          }
        }
      } catch (error: unknown) {
        if (error instanceof Error && error.name === 'AbortError') {
          return;
        } else {
          const errorMessage = error instanceof Error ? error.message : String(error);
          onError(`Failed to download: ${errorMessage || 'Unknown error'}`);
        }
      }
    })();

    // Return cancel function
    return () => controller.abort();
  }

  const cancelDownload = (e) => {
    e.stopPropagation();
    if (cancelDownloadRef.current) {
      cancelDownloadRef.current();
      cancelDownloadRef.current = null;
      setIsDownloading(false);
      toasts.toast("info", "Download canceled");
    }
  };

  const copyId = () => {
    navigator.clipboard.writeText(model.Id);
    toasts.toast("success", "ID copied to clipboard", {
      autoClose: 700
    });
  };

  const copyCIDHash = () => {
    navigator.clipboard.writeText(model.IpfsCID);
    toasts.toast("success", "CID Hash copied to clipboard", {
      autoClose: 700
    });
  };

  // Format MOR values to prevent scientific notation and limit decimals
  const formatMorValue = (value) => {
    if (!value) return '0 MOR';

    // Convert to MOR by dividing by 10^18
    const morValue = value / (10 ** 18);

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
      second: '2-digit'
    });
  };

  return (
    <CustomCard style={{ width: '36rem', position: 'relative' }} onClick={() => onSelect(model.Id)}>
      {isDownloading && (
        <DownloadProgressContainer>
          <div className="progress-header">
            <h4>Downloading Model</h4>
            <div className="cancel-button" onClick={cancelDownload}>
              <IconX size={16} />
            </div>
          </div>

          <ProgressBar
            variant="success"
            now={downloadProgress}
            className="progress-bar"
          />

          <div className="progress-info">
            <span>{downloadedSize} / {totalSize}</span>
            <span>{downloadProgress.toFixed(1)}%</span>
          </div>
          <div className="progress-info">
            <span>Last updated at: {formatDate(new Date(latestUploadTime))}</span>  
          </div>
        </DownloadProgressContainer>
      )}

      <Card.Body>
        <Card.Title
          as={'div'}
          style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
        >
          <span style={{ textOverflow: 'ellipsis', overflow: 'hidden', maxWidth: '90%' }}>
            {model.Name || "Unnamed Model"}
          </span>
          <IconDownload
            className="icon-button"
            style={{ width: '2.5rem', height: '2.5rem' }}
            onClick={handleFolderSelect}
          />
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
                <IconCopy
                  style={{ width: '1rem', height: '1rem', marginLeft: '8px', cursor: 'pointer', opacity: 0.8 }}
                  onClick={(e) => {
                    e.stopPropagation();
                    copyId();
                  }}
                />
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
                <IconCopy
                  style={{ width: '1rem', height: '1rem', marginLeft: '8px', cursor: 'pointer', opacity: 0.8 }}
                  onClick={(e) => {
                    e.stopPropagation();
                    copyCIDHash();
                  }}
                />
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
    </CustomCard>
  );
}


function ModelsTable({
  setSelectedModel,
  models,
  client,
  openSelectDownloadFolder,
  config,
  toasts,
}: any) {
  const onSelect = (id) => {
    setSelectedModel(models.find((x) => x.Id == id));
  };

  return (
    <Container>
      {models.length ?
        models.map(x => (
          <div key={x.Id}>
            <ModelCard
              onSelect={onSelect}
              model={x}
              openSelectDownloadFolder={openSelectDownloadFolder}
              toasts={toasts}
              client={client}
              config={config}
            />
          </div>
        )) :
        <div style={{
          width: '100%',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          padding: '40px 0',
          color: 'rgba(255, 255, 255, 0.6)',
          fontSize: '1.1rem',
          fontStyle: 'italic'
        }}>
          No models found
        </div>
      }
    </Container>
  );
}

export default ModelsTable;
