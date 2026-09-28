import { useEffect, useRef, useState } from 'react';
import styled from 'styled-components';
import Form from 'react-bootstrap/Form';
import { IconFile, IconUpload, IconHash, IconTag } from '@tabler/icons-react';
import Modal from '../contracts/modals/Modal';
import {
  TitleWrapper,
  Title,
  RightBtn,
} from '../contracts/modals/CreateContractModal.styles';

const bodyProps = { width: '640px', maxWidth: '100%' };

const StyledForm = styled(Form)`
  display: flex;
  flex-direction: column;
  gap: 1.6rem;

  .form-control {
    min-height: 4rem;
    padding: 0.8rem 1.2rem;
    background: var(--surface-base, #071711);
    border: 1px solid var(--border-strong, rgba(170, 216, 193, 0.3));
    border-radius: 8px;
    color: var(--text-primary, #edf7f0);
    font-size: 1.4rem;
    line-height: 1.5;
  }

  .form-control:focus {
    border-color: var(--accent, #19d695);
    background: var(--surface-base, #071711);
    color: var(--text-primary, #edf7f0);
  }

  .form-control::placeholder {
    color: var(--text-muted, #9ab4a7);
  }

  .form-label {
    display: flex;
    align-items: center;
    gap: 0.8rem;
    margin-bottom: 0.6rem;
    color: var(--text-primary, #edf7f0);
    font-size: 1.4rem;
    font-weight: 600;
  }

  .invalid-feedback {
    color: #ffb7a8;
    font-size: 1.2rem;
  }
`;

const HelperText = styled.p`
  color: var(--text-muted, #9ab4a7);
  font-size: 1.2rem;
  line-height: 1.5;
  margin: 0.6rem 0 0;
`;

const FileSummary = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  padding: 1.2rem 0;
  font-size: 1.4rem;
  border-block: 1px solid var(--border-subtle, rgba(170, 216, 193, 0.13));

  .file-info {
    display: flex;
    align-items: center;
    gap: 0.8rem;
    flex-wrap: wrap;
  }

  strong {
    overflow-wrap: anywhere;
  }

  svg {
    color: var(--accent, #19d695);
    flex: 0 0 auto;
  }

  .file-detail {
    color: var(--text-muted, #9ab4a7);
    font-size: 1.2rem;
    overflow-wrap: anywhere;
  }

  .file-path {
    font-family: var(--font-mono, monospace);
  }
`;

const SubmitRow = styled.div`
  display: flex;
  justify-content: flex-end;
  padding-top: 0.4rem;
`;

const StyledButton = styled(RightBtn)`
  width: auto;
  min-height: 4.4rem;
  height: auto;
  padding: 1rem 1.6rem;
  gap: 0.8rem;
  background: var(--accent, #19d695);
  border-radius: 10px;
  border: none;
  font-size: 1.4rem;
  line-height: 1.4;

  &:hover:not(:disabled) {
    background: #48e4ae;
  }
`;

const ErrorMessage = styled.p`
  color: #ffb7a8;
  font-size: 1.4rem;
  line-height: 1.5;
  margin: 0;
`;

type ModelFile = File & { path?: string };

const formatFileSize = (bytes: number) => {
  if (bytes < 1024) return `${bytes} bytes`;
  if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(2)} KB`;
  if (bytes < 1024 ** 3) return `${(bytes / 1024 ** 2).toFixed(2)} MB`;
  return `${(bytes / 1024 ** 3).toFixed(2)} GB`;
};

function modelIdError(id: string): string {
  if (!id) return '';
  if (!id.startsWith('0x')) return 'Model ID must start with "0x".';
  if (id.length !== 66)
    return 'Model ID must contain 64 hex characters after 0x.';
  if (!/^0x[0-9a-fA-F]{64}$/.test(id)) {
    return 'Model ID must contain only hex characters (0–9, a–f).';
  }
  return '';
}

const FileSelectionModal = ({
  isActive,
  handleClose,
  addFileToIpfs,
  pinFile,
  toasts,
}) => {
  const [file, setFile] = useState<ModelFile | null>(null);
  const [modelName, setModelName] = useState('');
  const [modelID, setModelID] = useState('');
  const [tags, setTags] = useState('');
  const [isPending, setIsPending] = useState(false);
  const [error, setError] = useState('');
  const pendingRef = useRef(false);
  const mountedRef = useRef(true);
  const idError = modelIdError(modelID.trim());

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  // Keep hooks stable when the caller toggles the dialog's visibility.
  if (!isActive) return null;

  const onPinModel = async (event: React.FormEvent) => {
    event.preventDefault();
    if (pendingRef.current) return;
    if (!file?.path) {
      setError('Choose one model file from your computer before pinning.');
      return;
    }
    if (idError) return;

    pendingRef.current = true;
    setIsPending(true);
    setError('');
    try {
      const response = await addFileToIpfs(
        file.path,
        modelID.trim(),
        modelName.trim(),
        tags.trim()
          ? tags
              .split(',')
              .map((tag) => tag.trim())
              .filter(Boolean)
          : undefined,
      );
      if (!response?.metadataCIDHash || !response?.fileCIDHash) {
        throw new Error('The node did not return the file identifiers.');
      }
      const results = await Promise.all([
        pinFile(response.metadataCIDHash),
        pinFile(response.fileCIDHash),
      ]);
      if (!results.every((result) => result?.result)) {
        throw new Error('The node did not confirm both pins.');
      }
      toasts.toast('success', 'Model pinned successfully');
      if (mountedRef.current) handleClose();
    } catch {
      const message =
        'Could not pin this model. Check your IPFS connection and try again.';
      if (mountedRef.current)
        setError(`${message} Your selection has been kept.`);
      else toasts.toast('error', message);
    } finally {
      pendingRef.current = false;
      if (mountedRef.current) setIsPending(false);
    }
  };

  return (
    <Modal
      ariaLabel="Pin a model file"
      onClose={handleClose}
      bodyProps={bodyProps}
    >
      <TitleWrapper>
        <Title as="h2">Pin a model file</Title>
      </TitleWrapper>
      <StyledForm
        onSubmit={onPinModel}
        aria-label="Pin model file"
        aria-busy={isPending}
      >
        <Form.Group controlId="modelName">
          <Form.Label>
            <IconFile size={18} aria-hidden="true" />
            Model name (optional)
          </Form.Label>
          <Form.Control
            type="text"
            value={modelName}
            disabled={isPending}
            onChange={(event) => setModelName(event.target.value)}
            placeholder="Enter model name"
          />
        </Form.Group>
        <Form.Group controlId="modelID">
          <Form.Label>
            <IconHash size={18} aria-hidden="true" />
            Model ID (optional)
          </Form.Label>
          <Form.Control
            type="text"
            value={modelID}
            disabled={isPending}
            onChange={(event) => setModelID(event.target.value)}
            placeholder="0x followed by 64 hex characters"
            isInvalid={!!idError}
            aria-invalid={!!idError}
            aria-describedby={idError ? 'model-id-error' : 'model-id-help'}
          />
          {idError && (
            <Form.Control.Feedback
              type="invalid"
              id="model-id-error"
              role="alert"
            >
              {idError}
            </Form.Control.Feedback>
          )}
          <HelperText id="model-id-help">
            Leave blank, or enter a 32-byte model hash starting with 0x.
          </HelperText>
        </Form.Group>
        <Form.Group controlId="tags">
          <Form.Label>
            <IconTag size={18} aria-hidden="true" />
            Tags (optional)
          </Form.Label>
          <Form.Control
            type="text"
            value={tags}
            disabled={isPending}
            onChange={(event) => setTags(event.target.value)}
            placeholder="Separate tags with commas"
          />
        </Form.Group>
        <Form.Group controlId="modelFile">
          <Form.Label>
            <IconUpload size={18} aria-hidden="true" />
            Model file
          </Form.Label>
          <Form.Control
            type="file"
            disabled={isPending}
            aria-describedby="model-file-help"
            onChange={(event) => {
              const selected = (event.currentTarget as HTMLInputElement).files;
              if (selected && selected.length > 1) {
                setError('Choose one model file at a time.');
                setFile(null);
                return;
              }
              setFile(selected?.[0] ?? null);
              setError('');
            }}
          />
          <HelperText id="model-file-help">
            Choose one file per pin, such as a .gguf model file.
          </HelperText>
        </Form.Group>
        {file && (
          <FileSummary>
            <div className="file-info">
              <IconFile size={18} aria-hidden="true" />
              <strong>{file.name}</strong>
              <span className="file-detail">{formatFileSize(file.size)}</span>
            </div>
            {file.path && (
              <div className="file-detail file-path">{file.path}</div>
            )}
          </FileSummary>
        )}
        {error && <ErrorMessage role="alert">{error}</ErrorMessage>}
        <SubmitRow>
          <StyledButton submit disabled={!file || !!idError || isPending}>
            <IconUpload size={18} aria-hidden="true" />
            {isPending
              ? 'Pinning file…'
              : error
                ? 'Try pinning again'
                : 'Pin model file'}
          </StyledButton>
        </SubmitRow>
      </StyledForm>
    </Modal>
  );
};

export default FileSelectionModal;
