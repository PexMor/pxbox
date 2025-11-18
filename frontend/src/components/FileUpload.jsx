import { useState } from 'preact/hooks';
import './FileUpload.css';

export function FileUpload({ requestId, filesPolicy, onFilesChange }) {
  const [files, setFiles] = useState([]);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState(null);

  const handleFileSelect = async (event) => {
    const selectedFiles = Array.from(event.target.files);
    setError(null);
    setUploading(true);

    try {
      const uploadedFiles = await Promise.all(
        selectedFiles.map(async (file) => {
          // Get presigned URL
          const signResponse = await fetch(
            `/v1/files/sign?name=${encodeURIComponent(file.name)}&contentType=${encodeURIComponent(file.type)}&requestId=${requestId}&size=${file.size}`
          );

          if (!signResponse.ok) {
            const err = await signResponse.json();
            throw new Error(err.message || 'Failed to get upload URL');
          }

          const { putUrl, getUrl } = await signResponse.json();

          // Upload file to presigned URL
          const uploadResponse = await fetch(putUrl, {
            method: 'PUT',
            body: file,
            headers: {
              'Content-Type': file.type,
            },
          });

          if (!uploadResponse.ok) {
            throw new Error('Failed to upload file');
          }

          // Calculate SHA-256 hash (simplified - in production use crypto API)
          const arrayBuffer = await file.arrayBuffer();
          const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
          const hashArray = Array.from(new Uint8Array(hashBuffer));
          const sha256 = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');

          return {
            name: file.name,
            url: getUrl,
            size: file.size,
            mimeType: file.type,
            sha256: sha256,
          };
        })
      );

      const newFiles = [...files, ...uploadedFiles];
      setFiles(newFiles);
      onFilesChange(newFiles);
    } catch (err) {
      setError(err.message || 'Failed to upload files');
    } finally {
      setUploading(false);
    }
  };

  const handleRemoveFile = (index) => {
    const newFiles = files.filter((_, i) => i !== index);
    setFiles(newFiles);
    onFilesChange(newFiles);
  };

  return (
    <div className="file-upload">
      <label className="file-upload-label">
        <input
          type="file"
          multiple
          onChange={handleFileSelect}
          disabled={uploading}
          className="file-upload-input"
        />
        <span className="file-upload-button">
          {uploading ? 'Uploading...' : 'Choose Files'}
        </span>
      </label>

      {filesPolicy && (
        <div className="file-upload-policy">
          {filesPolicy.maxFileMB && (
            <span>Max file size: {filesPolicy.maxFileMB} MB</span>
          )}
          {filesPolicy.mimeTypes && filesPolicy.mimeTypes.length > 0 && (
            <span>Allowed types: {filesPolicy.mimeTypes.join(', ')}</span>
          )}
        </div>
      )}

      {error && <div className="file-upload-error">{error}</div>}

      {files.length > 0 && (
        <div className="file-upload-list">
          {files.map((file, index) => (
            <div key={index} className="file-upload-item">
              <span className="file-upload-name">{file.name}</span>
              <span className="file-upload-size">
                {(file.size / 1024).toFixed(2)} KB
              </span>
              <button
                type="button"
                onClick={() => handleRemoveFile(index)}
                className="file-upload-remove"
              >
                ×
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

