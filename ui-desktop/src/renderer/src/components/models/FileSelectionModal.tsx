//import Modal from '../../contracts/modals/Modal';
import Modal from '../contracts/modals/Modal';
import styled from 'styled-components';
import {
    TitleWrapper,
    Title,
    RightBtn
} from '../contracts/modals/CreateContractModal.styles';
import { useState } from 'react';
import Form from 'react-bootstrap/Form';
import { Sp } from '../common'
import { IconFile, IconUpload, IconX, IconHash, IconTag } from '@tabler/icons-react';

const bodyProps = {
    height: '750px',
    width: '70%',
    maxWidth: '100%',
    overflow: 'hidden',
    onClick: e => e.stopPropagation()
}

const RowContainer = styled.div`
  padding: 1rem;
  border: 1px solid rgba(33, 220, 143, 0.2);
  background: rgba(0, 0, 0, 0.2);
  margin-bottom: 1rem;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  transition: all 0.2s;
  
  &:hover {
    border-color: rgba(33, 220, 143, 0.4);
    background: rgba(0, 0, 0, 0.25);
    transform: translateY(-2px);
  }
  
  .file-info {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #fff;
    font-size: 0.9rem;
    
    svg {
      color: #21dc8f;
    }
  }
  
  .file-path {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8rem;
    opacity: 0.7;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100%;
  }
  
  .file-size {
    background: rgba(33, 220, 143, 0.15);
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 0.8rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
`

const HelperText = styled.div`
    color: #8a8a8a;
    font-size: 0.875rem;
    margin-top: 0.25rem;
`

const StyledForm = styled(Form)`
  .form-control {
    background-color: rgba(0, 0, 0, 0.2);
    border: 1px solid rgba(33, 220, 143, 0.2);
    color: white;
    transition: all 0.2s;
    
    &:focus {
      background-color: rgba(0, 0, 0, 0.3);
      border-color: rgba(33, 220, 143, 0.5);
      box-shadow: 0 0 0 0.25rem rgba(33, 220, 143, 0.15);
      color: white;
    }
    
    &::placeholder {
      color: rgba(255, 255, 255, 0.5);
    }
  }
  
  .form-label {
    color: #21dc8f;
    font-weight: 500;
    margin-bottom: 0.5rem;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .form-control-feedback {
    margin-top: 0.25rem;
  }
`

const StyledButton = styled(RightBtn)`
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 0.5rem 1.25rem;
  background: linear-gradient(135deg, #21dc8f 0%, #1baf71 100%);
  border-radius: 8px;
  transition: all 0.2s;
  border: none;
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
  }
  
  &:active {
    transform: translateY(0);
  }
`

const EmptyFilesMessage = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  border: 2px dashed rgba(33, 220, 143, 0.3);
  border-radius: 8px;
  margin: 1rem 0;
  color: rgba(255, 255, 255, 0.6);
  
  svg {
    font-size: 2rem;
    margin-bottom: 1rem;
    color: #21dc8f;
    opacity: 0.5;
  }
`

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

const FileSelectionModal = ({ isActive, handleClose, addFileToIpfs, pinFile, toasts }) => {

    if (!isActive) {
        return <></>;
    }

    const [files, setFiles] = useState<any>([]);
    const [modelName, setModelName] = useState<string>('');
    const [modelID, setModelID] = useState<string>('');
    const [modelIDError, setModelIDError] = useState<string>('');
    const [tags, setTags] = useState<string>('');

    // Validate modelID - should be a hash of 32 bytes starting with 0x
    const validateModelID = (id: string): boolean => {
        if (!id) return true; // Empty is valid (optional field)
        
        // Check if starts with 0x
        if (!id.startsWith('0x')) {
            setModelIDError('Model ID must start with "0x"');
            return false;
        }
        
        // Remove 0x prefix for length check (0x + 64 characters for 32 bytes)
        const hexPart = id.substring(2);
        if (hexPart.length !== 64) {
            setModelIDError('Model ID must be 32 bytes (64 hex characters after 0x)');
            return false;
        }
        
        // Check if it's a valid hex string
        if (!/^[0-9a-fA-F]+$/.test(hexPart)) {
            setModelIDError('Model ID must contain only hex characters (0-9, a-f, A-F)');
            return false;
        }
        
        setModelIDError('');
        return true;
    };

    const handleModelIDChange = (e) => {
        const value = e.target.value;
        setModelID(value);
        validateModelID(value);
    };

    const onPinModel = async () => {
        // Validate modelID before proceeding
        if (!validateModelID(modelID)) {
            return; // Stop if validation fails
        }

        try {
            const response = await addFileToIpfs(files[0].path, modelID, modelName, tags ? tags.split(',').map(tag => tag.trim()) : undefined);
            console.log("🚀 ~ onPinModel ~ response:", response)
            if (response) {
                await Promise.all([
                    pinFile(response.metadataCIDHash),
                    pinFile(response.fileCIDHash)
                ]).then((res) => {
                    console.log("🚀 ~ ]).then ~ res:", res)
                    if (res.every(r => r.result)) {
                        handleClose();
                        toasts.toast("success", "Model pinned successfully");
                    } else {
                        handleClose();
                        toasts.toast("error", "Failed to pin model");
                    }
                }).catch(() => {
                    handleClose();
                    toasts.toast("error", "Failed to pin model");
                });
            } else {
                handleClose();
                toasts.toast("error", "Failed to pin model");
            }
        } catch (error) {
            handleClose();
            toasts.toast("error", "Failed to pin model");
            console.error("Error", error);
        }
    }

    return (
        <Modal
            onClose={() => {
                handleClose();
            }}
            bodyProps={bodyProps}
        >
            <TitleWrapper>
                <Title>Select File for IPFS</Title>
            </TitleWrapper>

            <StyledForm>
                <Sp mt={2}>
                    <Form.Group controlId="modelName" className="mb-3">
                        <Form.Label>
                            <IconFile size={16} strokeWidth={2} />
                            Model Name (optional)
                        </Form.Label>
                        <Form.Control 
                            type="text" 
                            value={modelName}
                            onChange={(e) => setModelName(e.target.value)}
                            placeholder="Enter model name"
                        />
                    </Form.Group>
                </Sp>
                
                <Sp mt={2}>
                    <Form.Group controlId="modelID" className="mb-3">
                        <Form.Label>
                            <IconHash size={16} strokeWidth={2} />
                            Model ID (optional)
                        </Form.Label>
                        <Form.Control 
                            type="text" 
                            value={modelID}
                            onChange={handleModelIDChange}
                            placeholder="Enter model ID (0x followed by 64 hex characters)"
                            isInvalid={!!modelIDError}
                        />
                        {modelIDError && (
                            <Form.Control.Feedback type="invalid">
                                {modelIDError}
                            </Form.Control.Feedback>
                        )}
                        <HelperText>
                            Must be a 32-byte hash starting with 0x (e.g., 0x1234...abcd)
                        </HelperText>
                    </Form.Group>
                </Sp>
                
                <Sp mt={2}>
                    <Form.Group controlId="tags" className="mb-3">
                        <Form.Label>
                            <IconTag size={16} strokeWidth={2} />
                            Tags (optional, comma-separated)
                        </Form.Label>
                        <Form.Control 
                            type="text" 
                            value={tags}
                            onChange={(e) => setTags(e.target.value)}
                            placeholder="tag1, tag2, tag3"
                        />
                    </Form.Group>
                </Sp>
                
                <Sp mt={2}>
                    <Form.Group controlId="formFile" className="mb-3">
                        <Form.Label>
                            <IconUpload size={16} strokeWidth={2} />
                            Select files required to run model (including .gguf)
                        </Form.Label>
                        <Form.Control type="file" multiple onChange={(e => {
                            setFiles(Object.values((e.currentTarget as any).files))
                        })} />
                    </Form.Group>
                </Sp>
            </StyledForm>

            {!files.length ? (
                <EmptyFilesMessage>
                    <IconUpload size={36} strokeWidth={1.5} />
                    <div>No files selected yet</div>
                    <div style={{ fontSize: '0.9rem', marginTop: '0.5rem' }}>
                        Please select files to pin to IPFS
                    </div>
                </EmptyFilesMessage>
            ) : (
                files.map((f, index) => (
                    <RowContainer key={index}>
                        <div className="file-info">
                            <IconFile size={18} strokeWidth={2} />
                            <strong>{f.name}</strong>
                            <div className="file-size">{formatFileSize(f.size)}</div>
                        </div>
                        <div className="file-path">{f.path}</div>
                    </RowContainer>
                ))
            )}

            <Sp mt={3} style={{ display: 'flex', justifyContent: 'center' }}>
                <StyledButton onClick={onPinModel} disabled={!files.length}>
                    <IconUpload size={16} strokeWidth={2} />
                    Pin Model Files
                </StyledButton>
            </Sp>
        </Modal>
    );
}

export default FileSelectionModal;