import { useRef, useState } from 'react';

import withModelsState from '../../store/hocs/withModelsState';
import styled from 'styled-components';

import Card from 'react-bootstrap/Card';
import { abbreviateAddress } from '../../utils';
import { IconPinnedOff, IconCopy, IconFile, IconCalendar, IconTag, IconHash } from '@tabler/icons-react';
import Form from 'react-bootstrap/esm/Form';
import client from '../../client';


const CustomCard = styled(Card)`
  background: linear-gradient(145deg, #244a47 0%, #1d3c39 100%) !important;
  color: #21dc8f !important;
  border: 1px solid rgba(33, 220, 143, 0.2) !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: all 0.2s ease-in-out;
  border-radius: 12px !important;
  overflow: hidden;
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.25);
    border-color: rgba(33, 220, 143, 0.4) !important;
  }

  p {
    color: white !important;
  }

  .card-title {
    font-weight: 600;
    font-size: 1.3rem;
    letter-spacing: 0.02em;
    text-overflow: ellipsis;
    color: #21dc8f;
  }

  .card-subtitle {
    font-size: 0.85rem;
    color: rgba(255, 255, 255, 0.7) !important;
  }

  .card-body {
    padding: 1.5rem;
  }

  .model-info-section {
    display: flex;
    flex-direction: column;
    gap: 3px;
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
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    display: inline-block;
  }

  .icon-button {
    cursor: pointer;
    padding: 8px;
    border-radius: 50%;
    transition: all 0.2s;
    background: rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.8);
    
    &:hover {
      background: rgba(255, 0, 0, 0.15);
      color: #ff6b6b;
      transform: rotate(8deg);
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

interface PinnedFile {
  fileCID: string;
  fileCIDHash: string;
  metadataCID: string;
  metadataCIDHash: string;
  fileName: string;
  fileSize: number;
  tags: string[] | null;
  modelName: string;
  id: string;
}

function ModelCard({ model, toasts, unpinFile }: { model: PinnedFile, toasts: any, unpinFile: any }) {
  const onUnpinFile = (e) => {
    e.stopPropagation();
    unpinFile(model.fileCIDHash);
    unpinFile(model.metadataCIDHash);
    toasts.toast("success", "File unpinned successfully", { autoClose: 2000 });
  };

  const copyHash = () => {
    navigator.clipboard.writeText(model.metadataCIDHash);
    toasts.toast("success", "Hash copied to clipboard", {
      autoClose: 700
    });
  };

  const copyCID = () => {
    navigator.clipboard.writeText(model.metadataCID);
    toasts.toast("success", "CID copied to clipboard", {
      autoClose: 700
    });
  };

  const copyId = () => {
    navigator.clipboard.writeText(model.id);
    toasts.toast("success", "ID copied to clipboard", {
      autoClose: 700
    });
  };

  const formatFileSize = (bytes: number) => {
    if (!bytes) return '';

    const KB = bytes / 1024;
    const MB = KB / 1024;
    const GB = MB / 1024;

    if (GB >= 1) {
      return `${GB.toFixed(2)} GB`;
    } else if (MB >= 1) {
      return `${MB.toFixed(2)} MB`;
    } else {
      return `${KB.toFixed(2)} KB`;
    }
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return '';

    try {
      const date = new Date(dateString);
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch (e) {
      return dateString;
    }
  };

  return (
    <CustomCard style={{ width: '36rem' }}>
      <Card.Body>
        <Card.Title
          as={'div'}
          style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
        >
          <span style={{ textOverflow: 'ellipsis', overflow: 'hidden', maxWidth: '90%' }}>
            {model.fileName || "Unnamed File"}
          </span>
          <IconPinnedOff
            className="icon-button"
            style={{ width: '2.5rem', height: '2.5rem' }}
            onClick={onUnpinFile}
          />
        </Card.Title>

        <div className="model-info-section">
        <div className="model-info-item">
            <span className="info-label">
              <IconHash size={16} strokeWidth={2} />
              CID:</span>
              <div className="info-value">
              <span className="hash-container monospace">
                {abbreviateAddress(model.metadataCID, 6)}
                <IconCopy
                  style={{ width: '1rem', height: '1rem', marginLeft: '8px', cursor: 'pointer', opacity: 0.8 }}
                  onClick={() => copyCID()}
                />
              </span>
            </div>
          </div>

          <div className="model-info-item">
            <span className="info-label">
              <IconHash size={16} strokeWidth={2} />
              CID Hash:</span>
            <div className="info-value">
              <span className="hash-container monospace">
                {abbreviateAddress(model.metadataCIDHash, 6)}
                <IconCopy
                  style={{ width: '1rem', height: '1rem', marginLeft: '8px', cursor: 'pointer', opacity: 0.8 }}
                  onClick={() => copyHash()}
                />
              </span>
            </div>
          </div>
          
          {model.fileSize ? (
            <div className="model-info-item">
              <span className="info-label">
                <IconFile size={16} strokeWidth={2} />
                Size:
              </span>
              <span className="info-value">{formatFileSize(model.fileSize)}</span>
            </div>
          ) : null}

          {model.modelName ? (
            <div className="model-info-item">
              <span className="info-label">
                <IconHash size={16} strokeWidth={2} />
                Name:</span>
              <span className="info-value">{model.modelName}</span>
            </div>
          ) : null}

          {model.id && model.id.length > 2 ? (
            <div className="model-info-item">
              <span className="info-label">
                <IconHash size={16} strokeWidth={2} />
                ID:</span>
              <div className="info-value">
              <span className="hash-container monospace">
                {abbreviateAddress(model.id, 6)}
                <IconCopy
                  style={{ width: '1rem', height: '1rem', marginLeft: '8px', cursor: 'pointer', opacity: 0.8 }}
                  onClick={() => copyId()}
                  />
                </span>
              </div>
            </div>
          ) : null}

          {model.tags && model.tags.length > 0 ? (
            <div className="model-info-item">
              <span className="info-label">
                <IconTag size={16} strokeWidth={2} />
                Tags:
              </span>
              <div className="info-value">
                <div className="tag-container">
                  {model.tags.map((tag, index) => (
                    <span key={index} className="tag-item">
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          ) : null}
        </div>
      </Card.Body>
    </CustomCard>
  );
}


function PinnedFilesTable({
  pinnedFiles,
  toasts,
  unpinFile
}: any) {
  return (
    <Container>
      {pinnedFiles?.length ?
        pinnedFiles.map(x => (
          <div key={x.fileCIDHash}>
            {ModelCard({ model: x, toasts, unpinFile })}
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
          No pinned files found
        </div>
      }
    </Container>
  );
}

export default PinnedFilesTable;
