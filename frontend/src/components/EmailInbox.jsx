import { useState, useEffect, useMemo } from 'preact/hooks';
import { useBrokerWS } from '../hooks/useBrokerWS';
import { getSetting, setSetting } from '../utils/indexeddb';
import './EmailInbox.css';

function getTimeAgo(date) {
  const now = new Date();
  const diffMs = now - date;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);
  
  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

function getStatusLabel(status) {
  const labels = {
    'PENDING': 'Pending',
    'CLAIMED': 'In Progress',
    'ANSWERED': 'Answered',
    'CANCELLED': 'Cancelled',
    'ATTENTION': 'Needs Attention',
  };
  return labels[status] || status;
}

export function EmailInbox({ onSelectRequest }) {
  const [inquiries, setInquiries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedFolder, setSelectedFolder] = useState('all'); // 'all' or entity ID
  const [hideAnswered, setHideAnswered] = useState(false);
  const [settingsLoaded, setSettingsLoaded] = useState(false);
  const [selectedItems, setSelectedItems] = useState(new Set());
  const [showDeleteAllConfirm, setShowDeleteAllConfirm] = useState(false);
  
  // Load hideAnswered setting from IndexedDB on mount
  useEffect(() => {
    getSetting('hideAnswered').then(value => {
      if (value !== null && value !== undefined) {
        setHideAnswered(value === true);
      }
      setSettingsLoaded(true);
    });
  }, []);
  
  // Save hideAnswered setting to IndexedDB when it changes (only after initial load)
  useEffect(() => {
    if (settingsLoaded) {
      setSetting('hideAnswered', hideAnswered);
    }
  }, [hideAnswered, settingsLoaded]);
  
  const getEntityId = () => {
    if (typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search);
      const urlEntityId = params.get('entityId');
      if (urlEntityId) {
        localStorage.setItem('pxbox_entity_id', urlEntityId);
        return urlEntityId;
      }
      const stored = localStorage.getItem('pxbox_entity_id');
      if (stored) return stored;
    }
    return null;
  };
  
  const [entityId] = useState(getEntityId());

  useEffect(() => {
    const url = entityId ? `/v1/inquiries?entityId=${entityId}` : '/v1/inquiries';
    fetch(url)
      .then(res => res.json())
      .then(data => {
        setInquiries(data.items || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch inquiries:', err);
        setLoading(false);
      });
  }, [entityId]);

  // Subscribe to WebSocket events
  const ws = useBrokerWS();
  useEffect(() => {
    if (ws && entityId) {
      ws.subscribe(`entity:${entityId}`);
      ws.on(`entity:${entityId}`, (event) => {
        if (event.data?.event === 'request.created' || event.data?.event === 'request.answered') {
          const url = entityId ? `/v1/inquiries?entityId=${entityId}` : '/v1/inquiries';
          fetch(url)
            .then(res => res.json())
            .then(data => setInquiries(data.items || []));
        }
      });
    }
  }, [ws, entityId]);

  // Group inquiries by entity (createdBy) for folders
  const folders = useMemo(() => {
    const folderMap = new Map();
    folderMap.set('all', { name: 'All Inquiries', count: inquiries.length, items: inquiries });
    
    inquiries.forEach(inq => {
      const folderId = inq.createdBy || 'unknown';
      if (!folderMap.has(folderId)) {
        const displayName = folderId === 'unknown' ? 'Unknown' : (folderId.substring(0, 8) + '...');
        folderMap.set(folderId, { name: displayName, count: 0, items: [] });
      }
      const folder = folderMap.get(folderId);
      folder.count++;
      folder.items.push(inq);
    });
    
    return Array.from(folderMap.entries()).map(([id, data]) => ({ id, ...data }));
  }, [inquiries]);

  // Filter inquiries based on selected folder and hide answered
  const filteredInquiries = useMemo(() => {
    let filtered = selectedFolder === 'all' 
      ? inquiries 
      : inquiries.filter(inq => inq.createdBy === selectedFolder);
    
    if (hideAnswered) {
      filtered = filtered.filter(inq => inq.status !== 'ANSWERED');
    }
    
    return filtered;
  }, [inquiries, selectedFolder, hideAnswered]);

  const handleDelete = async (id) => {
    try {
      const response = await fetch(`/v1/inquiries/${id}`, { method: 'DELETE' });
      if (response.ok) {
        setInquiries(prev => prev.filter(inq => inq.id !== id));
        setSelectedItems(prev => {
          const next = new Set(prev);
          next.delete(id);
          return next;
        });
      }
    } catch (err) {
      console.error('Failed to delete:', err);
      alert('Failed to delete inquiry');
    }
  };

  const handleDeleteAll = async () => {
    try {
      const promises = inquiries.map(inq => 
        fetch(`/v1/inquiries/${inq.id}`, { method: 'DELETE' })
      );
      await Promise.all(promises);
      setInquiries([]);
      setSelectedItems(new Set());
      setShowDeleteAllConfirm(false);
    } catch (err) {
      console.error('Failed to delete all:', err);
      alert('Failed to delete all inquiries');
    }
  };

  const handleDeleteSelected = async () => {
    try {
      const promises = Array.from(selectedItems).map(id =>
        fetch(`/v1/inquiries/${id}`, { method: 'DELETE' })
      );
      await Promise.all(promises);
      setInquiries(prev => prev.filter(inq => !selectedItems.has(inq.id)));
      setSelectedItems(new Set());
    } catch (err) {
      console.error('Failed to delete selected:', err);
      alert('Failed to delete selected inquiries');
    }
  };

  const toggleSelect = (id) => {
    setSelectedItems(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const selectAll = () => {
    if (selectedItems.size === filteredInquiries.length) {
      setSelectedItems(new Set());
    } else {
      setSelectedItems(new Set(filteredInquiries.map(inq => inq.id)));
    }
  };

  if (loading) {
    return <div className="email-inbox-loading">Loading inquiries...</div>;
  }

  return (
    <div className="email-inbox">
      <div className="email-inbox-sidebar">
        <div className="email-inbox-sidebar-header">
          <h2>Folders</h2>
        </div>
        <div className="email-inbox-folders">
          {folders.map(folder => (
            <div
              key={folder.id}
              className={`email-inbox-folder ${selectedFolder === folder.id ? 'active' : ''}`}
              onClick={() => setSelectedFolder(folder.id)}
            >
              <span className="folder-name">{folder.name}</span>
              <span className="folder-count">{folder.count}</span>
            </div>
          ))}
        </div>
        <div className="email-inbox-actions">
          <label className="email-inbox-toggle">
            <input
              type="checkbox"
              checked={hideAnswered}
              onChange={(e) => setHideAnswered(e.target.checked)}
            />
            <span>Hide Answered</span>
          </label>
          <button
            className="email-inbox-delete-all"
            onClick={() => setShowDeleteAllConfirm(true)}
          >
            Delete All
          </button>
        </div>
      </div>

      <div className="email-inbox-main">
        <div className="email-inbox-toolbar">
          <div className="email-inbox-toolbar-left">
            <input
              type="checkbox"
              checked={selectedItems.size === filteredInquiries.length && filteredInquiries.length > 0}
              onChange={selectAll}
              className="email-inbox-checkbox"
            />
            {selectedItems.size > 0 && (
              <button
                className="email-inbox-delete-selected"
                onClick={handleDeleteSelected}
              >
                Delete Selected ({selectedItems.size})
              </button>
            )}
          </div>
          <div className="email-inbox-toolbar-right">
            <span className="email-inbox-count">
              {filteredInquiries.length} {filteredInquiries.length === 1 ? 'inquiry' : 'inquiries'}
            </span>
          </div>
        </div>

        {filteredInquiries.length === 0 ? (
          <div className="email-inbox-empty">
            <div className="email-inbox-empty-icon">📭</div>
            <div className="email-inbox-empty-text">No inquiries</div>
          </div>
        ) : (
          <div className="email-inbox-list">
            {filteredInquiries.map(inquiry => {
              const deadline = inquiry.deadlineAt ? new Date(inquiry.deadlineAt) : null;
              const isDueSoon = deadline && deadline <= new Date(Date.now() + 24 * 60 * 60 * 1000) && deadline > new Date();
              const isOverdue = deadline && deadline <= new Date();
              const createdAt = new Date(inquiry.createdAt);
              const timeAgo = getTimeAgo(createdAt);
              const isSelected = selectedItems.has(inquiry.id);
              const isRead = inquiry.readAt != null;
              
              return (
                <div
                  key={inquiry.id}
                  className={`email-inbox-item email-inbox-item-${inquiry.status.toLowerCase()} ${isSelected ? 'selected' : ''} ${isRead ? 'read' : ''}`}
                  onClick={(e) => {
                    if (e.target.type === 'checkbox' || e.target.closest('.email-inbox-item-actions')) {
                      return;
                    }
                    onSelectRequest(inquiry.id);
                  }}
                >
                  <div className="email-inbox-item-checkbox">
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => toggleSelect(inquiry.id)}
                      onClick={(e) => e.stopPropagation()}
                    />
                  </div>
                  <div className="email-inbox-item-content">
                    <div className="email-inbox-item-header">
                    <div className="email-inbox-item-from">
                      <span className={`status-indicator status-${inquiry.status.toLowerCase()}`}></span>
                      <span className="from-name">{inquiry.createdBy ? inquiry.createdBy.substring(0, 8) + '...' : 'Unknown'}</span>
                      {inquiry.createdBy && (
                        <span className="from-id">({inquiry.createdBy.substring(0, 8)}...)</span>
                      )}
                    </div>
                      <div className="email-inbox-item-time">{timeAgo}</div>
                    </div>
                    <div className="email-inbox-item-body">
                      <span className="email-inbox-item-status">{getStatusLabel(inquiry.status)}</span>
                      <span className="email-inbox-item-id">ID: {inquiry.id.substring(0, 12)}...</span>
                    </div>
                    <div className="email-inbox-item-footer">
                      {isOverdue && (
                        <span className="email-inbox-badge overdue">⚠️ Overdue</span>
                      )}
                      {isDueSoon && !isOverdue && (
                        <span className="email-inbox-badge due-soon">⏰ Due Soon</span>
                      )}
                      {deadline && !isDueSoon && !isOverdue && (
                        <span className="email-inbox-deadline">
                          Due {deadline.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="email-inbox-item-actions">
                    <button
                      className="email-inbox-action-btn"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDelete(inquiry.id);
                      }}
                      title="Delete"
                    >
                      🗑️
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {showDeleteAllConfirm && (
        <div className="email-inbox-modal-overlay" onClick={() => setShowDeleteAllConfirm(false)}>
          <div className="email-inbox-modal" onClick={(e) => e.stopPropagation()}>
            <h3>Delete All Inquiries?</h3>
            <p>This will permanently delete all {inquiries.length} inquiries. This action cannot be undone.</p>
            <div className="email-inbox-modal-actions">
              <button
                className="email-inbox-modal-btn email-inbox-modal-btn-danger"
                onClick={handleDeleteAll}
              >
                Delete All
              </button>
              <button
                className="email-inbox-modal-btn"
                onClick={() => setShowDeleteAllConfirm(false)}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

