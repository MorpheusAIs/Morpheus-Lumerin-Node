import { useRef, useState } from 'react';

import withModelsState from '../../store/hocs/withModelsState';
import styled from 'styled-components';

import Card from 'react-bootstrap/Card';
import { abbreviateAddress } from '../../utils';
import { IconDownload, IconCopy, IconCoin, IconTag, IconHash } from '@tabler/icons-react';
import Form from 'react-bootstrap/esm/Form';


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
    font-size: 1.25rem;
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
    font-size: 0.9rem;
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
    font-size: 0.8rem;
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

function ModelCard({ onSelect, model, openSelectDownloadFolder, downloadModelFromIpfs, toasts }) {
  const handleFolderSelect = async (e) => {
    e.stopPropagation();
    try {
      const result = await openSelectDownloadFolder();
      const { canceled, filePaths } = result;
      if (canceled) {
        return;
      }
      const folderPath = filePaths[0];
      const response = await downloadModelFromIpfs(model.IpfsCID, folderPath);
      if (response) {
        toasts.toast("success", "Model downloaded successfully");
      } else {
        toasts.toast("error", "Failed to download model");
      }
    } catch (error) {
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

  return (
    <CustomCard style={{ width: '36rem' }} onClick={() => onSelect(model.Id)}>
      <Card.Body>
        <Card.Title
          as={'div'}
          style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
        >
          {model.Name || "Unnamed Model"}
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
                {abbreviateAddress(model.Id, 6)}
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
                {abbreviateAddress(model.IpfsCID, 6)}
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

          <div className="model-info-item">
            <span className="info-label">
              <IconCoin size={16} strokeWidth={2} />
              Fee:
            </span>
            <span className="info-value">{formatMorValue(model.Fee)}</span>
          </div>
          
          <div className="model-info-item">
            <span className="info-label">
              <IconCoin size={16} strokeWidth={2} />
              Stake:
            </span>
            <span className="info-value">{formatMorValue(model.Stake)}</span>
          </div>

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
  downloadModelFromIpfs,
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
            {ModelCard({ onSelect, model: x, openSelectDownloadFolder, downloadModelFromIpfs, toasts })}
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
