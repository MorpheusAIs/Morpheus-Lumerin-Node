import Card from 'react-bootstrap/Card';
import { abbreviateAddress } from '../../utils';
import {
  IconPinnedOff,
  IconCopy,
  IconFile,
  IconTag,
  IconHash,
} from '@tabler/icons-react';
import { ModelActionButton } from './ModelActionButton';
import {
  ModelCardEmptyState,
  ModelCardGrid,
  ModelCardSurface,
} from './ModelCardSurface';

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

function ModelCard({
  model,
  toasts,
  unpinFile,
}: {
  model: PinnedFile;
  toasts: any;
  unpinFile: any;
}) {
  const onUnpinFile = (e) => {
    e.stopPropagation();
    unpinFile(model.fileCIDHash);
    unpinFile(model.metadataCIDHash);
  };

  const copyToClipboard = async (text: string, label: string) => {
    try {
      // Use the Electron clipboard bridge; navigator.clipboard silently
      // fails in the renderer (focus/permissions)
      await window.copyToClipboard(text);
      toasts.toast('success', `${label} copied to clipboard`, {
        autoClose: 700,
      });
    } catch (e) {
      toasts.toast('error', `Failed to copy ${label} to clipboard`);
    }
  };

  const copyHash = () => copyToClipboard(model.metadataCIDHash, 'Hash');

  const copyCID = () => copyToClipboard(model.metadataCID, 'CID');

  const copyId = () => copyToClipboard(model.id, 'ID');

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

  return (
    <ModelCardSurface>
      <Card.Body>
        <Card.Title as={'div'} className="model-card-title">
          <span className="model-card-name">
            {model.fileName || 'Unnamed File'}
          </span>
          <ModelActionButton
            type="button"
            aria-label={`Unpin ${model.fileName || 'file'}`}
            title="Unpin model file"
            onClick={onUnpinFile}
          >
            <IconPinnedOff size={18} aria-hidden="true" />
          </ModelActionButton>
        </Card.Title>

        <div className="model-info-section">
          <div className="model-info-item">
            <span className="info-label">
              <IconHash size={16} strokeWidth={2} />
              CID:
            </span>
            <div className="info-value">
              <span className="hash-container monospace">
                {abbreviateAddress(model.metadataCID, 6)}
                <ModelActionButton
                  type="button"
                  aria-label={`Copy CID for ${model.fileName || 'file'}`}
                  title="Copy CID"
                  disabled={!model.metadataCID}
                  onClick={() => copyCID()}
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
                {abbreviateAddress(model.metadataCIDHash, 6)}
                <ModelActionButton
                  type="button"
                  aria-label={`Copy CID hash for ${model.fileName || 'file'}`}
                  title="Copy CID hash"
                  disabled={!model.metadataCIDHash}
                  onClick={() => copyHash()}
                >
                  <IconCopy size={16} aria-hidden="true" />
                </ModelActionButton>
              </span>
            </div>
          </div>

          {model.fileSize ? (
            <div className="model-info-item">
              <span className="info-label">
                <IconFile size={16} strokeWidth={2} />
                Size:
              </span>
              <span className="info-value">
                {formatFileSize(model.fileSize)}
              </span>
            </div>
          ) : null}

          {model.modelName ? (
            <div className="model-info-item">
              <span className="info-label">
                <IconHash size={16} strokeWidth={2} />
                Name:
              </span>
              <span className="info-value">{model.modelName}</span>
            </div>
          ) : null}

          {model.id && model.id.length > 2 ? (
            <div className="model-info-item">
              <span className="info-label">
                <IconHash size={16} strokeWidth={2} />
                ID:
              </span>
              <div className="info-value">
                <span className="hash-container monospace">
                  {abbreviateAddress(model.id, 6)}
                  <ModelActionButton
                    type="button"
                    aria-label={`Copy model ID for ${model.fileName || 'file'}`}
                    title="Copy model ID"
                    onClick={() => copyId()}
                  >
                    <IconCopy size={16} aria-hidden="true" />
                  </ModelActionButton>
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
    </ModelCardSurface>
  );
}

function PinnedFilesTable({ pinnedFiles, toasts, unpinFile }: any) {
  return (
    <ModelCardGrid>
      {pinnedFiles?.length ? (
        pinnedFiles.map((x) => (
          <div key={x.fileCIDHash}>
            {ModelCard({ model: x, toasts, unpinFile })}
          </div>
        ))
      ) : (
        <ModelCardEmptyState>No pinned files found</ModelCardEmptyState>
      )}
    </ModelCardGrid>
  );
}

export default PinnedFilesTable;
