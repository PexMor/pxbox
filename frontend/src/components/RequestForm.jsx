import { useState, useEffect } from 'preact/hooks';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import { FileUpload } from './FileUpload';
import './RequestForm.css';

export function RequestForm({ requestId, onBack }) {
  const [request, setRequest] = useState(null);
  const [formData, setFormData] = useState({});
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const [hasDraft, setHasDraft] = useState(false);
  const [draftSaved, setDraftSaved] = useState(false);

  const draftKey = `draft_${requestId}`;

  // Load draft from localStorage
  const loadDraft = () => {
    try {
      const draft = localStorage.getItem(draftKey);
      if (draft) {
        const parsed = JSON.parse(draft);
        setFormData(parsed.formData || {});
        setFiles(parsed.files || []);
        setHasDraft(true);
        return true;
      }
    } catch (err) {
      console.error('Failed to load draft:', err);
    }
    return false;
  };

  // Save draft to localStorage
  const saveDraft = () => {
    try {
      const draft = {
        formData,
        files,
        savedAt: new Date().toISOString(),
      };
      localStorage.setItem(draftKey, JSON.stringify(draft));
      setHasDraft(true);
      setDraftSaved(true);
      setTimeout(() => setDraftSaved(false), 2000);
    } catch (err) {
      console.error('Failed to save draft:', err);
    }
  };

  // Clear draft
  const clearDraft = () => {
    localStorage.removeItem(draftKey);
    setHasDraft(false);
  };

  useEffect(() => {
    fetch(`/v1/requests/${requestId}`)
      .then(res => res.json())
      .then(data => {
        setRequest(data);
        // Try to load draft first, otherwise use prefill
        if (!loadDraft()) {
          setFormData(data.prefill || {});
        }
        setLoading(false);
      })
      .catch(err => {
        setError('Failed to load request');
        setLoading(false);
      });
  }, [requestId]);

  // Auto-save draft every 30 seconds
  useEffect(() => {
    if (!request) return;
    
    const interval = setInterval(() => {
      if (Object.keys(formData).length > 0 || files.length > 0) {
        saveDraft();
      }
    }, 30000); // Auto-save every 30 seconds

    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formData, files, request]);

  const getEntityId = () => {
    // Use the request's entityId as the responder (the entity the request was sent to)
    if (request && request.entityId) {
      return request.entityId;
    }
    
    // Fallback to localStorage or URL param
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('pxbox_entity_id');
      if (stored) return stored;
      const params = new URLSearchParams(window.location.search);
      const urlEntityId = params.get('entityId');
      if (urlEntityId) {
        localStorage.setItem('pxbox_entity_id', urlEntityId);
        return urlEntityId;
      }
    }
    
    // Last resort: return null and let backend handle it
    return null;
  };

  const handleSubmit = async ({ formData }) => {
    setSubmitting(true);
    setError(null);

    const entityId = getEntityId();
    if (!entityId) {
      setError('Unable to determine entity ID. Please ensure the request is loaded.');
      setSubmitting(false);
      return;
    }

    try {
      // Claim request first (ignore 409 if already claimed - we can still respond)
      const claimResponse = await fetch(`/v1/requests/${requestId}/claim`, {
        method: 'POST',
        headers: { 'X-Entity-ID': entityId },
      });
      
      // If claim fails with 409, it's already claimed - that's OK, we can still respond
      // Only throw if it's a different error
      if (!claimResponse.ok && claimResponse.status !== 409) {
        const err = await claimResponse.json();
        throw new Error(err.message || 'Failed to claim request');
      }

      // Submit response
      const responseBody = {
        payload: formData,
      };
      if (files.length > 0) {
        responseBody.files = files;
      }

      const response = await fetch(`/v1/requests/${requestId}/response`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Entity-ID': entityId,
        },
        body: JSON.stringify(responseBody),
      });

      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.message || 'Failed to submit response');
      }

      // Clear draft after successful submission
      clearDraft();
      alert('Response submitted successfully!');
      onBack();
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <div className="form-loading">Loading form...</div>;
  }

  if (error && !request) {
    return (
      <div className="form-error">
        <p>{error}</p>
        <button onClick={onBack}>Back to Inbox</button>
      </div>
    );
  }

  if (!request) {
    return <div className="form-error">Request not found</div>;
  }

  // Convert schema_payload to proper format for @rjsf
  const schema = request.schemaPayload || {};
  const uiSchema = request.uiHints || {};

  return (
    <div className="request-form">
      <div className="form-header">
        <button onClick={onBack} className="form-back">← Back</button>
        <h2>Request: {requestId}</h2>
      </div>

      {error && <div className="form-error-message">{error}</div>}

      {hasDraft && (
        <div className="form-draft-notice">
          <span>💾 Draft loaded</span>
          <button onClick={clearDraft} className="form-draft-clear">Clear Draft</button>
        </div>
      )}

      <Form
        schema={schema}
        uiSchema={uiSchema}
        formData={formData}
        validator={validator}
        onChange={({ formData }) => setFormData(formData)}
        onSubmit={handleSubmit}
      >
        {request.filesPolicy && (
          <div className="form-section">
            <h3>File Attachments</h3>
            <FileUpload
              requestId={requestId}
              filesPolicy={request.filesPolicy}
              onFilesChange={setFiles}
            />
          </div>
        )}
        <div className="form-actions">
          <button
            type="button"
            onClick={saveDraft}
            className="form-save-draft"
          >
            {draftSaved ? '✓ Draft Saved' : '💾 Save Draft'}
          </button>
          <button type="submit" disabled={submitting} className="form-submit">
            {submitting ? 'Submitting...' : 'Submit'}
          </button>
        </div>
      </Form>
    </div>
  );
}

