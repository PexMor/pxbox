import { useState, useEffect, useMemo } from 'preact/hooks';
import { useBrokerWS } from '../hooks/useBrokerWS';
import './Inbox.css';

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

export function Inbox({ onSelectRequest }) {
  const [inquiries, setInquiries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [groupFilter, setGroupFilter] = useState('all'); // 'all', 'attention', 'dueSoon'
  // Get entity ID from localStorage or URL parameter
  const getEntityId = () => {
    if (typeof window !== 'undefined') {
      // Try URL parameter first (highest priority)
      const params = new URLSearchParams(window.location.search);
      const urlEntityId = params.get('entityId');
      if (urlEntityId) {
        localStorage.setItem('pxbox_entity_id', urlEntityId);
        return urlEntityId;
      }
      // Then check localStorage
      const stored = localStorage.getItem('pxbox_entity_id');
      if (stored) return stored;
    }
    return null; // No entity ID - will show all requests or prompt
  };
  
  const [entityId, setEntityId] = useState(getEntityId());

  useEffect(() => {
    // Fetch inquiries
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
          // Refresh inquiries
          const url = entityId ? `/v1/inquiries?entityId=${entityId}` : '/v1/inquiries';
          fetch(url)
            .then(res => res.json())
            .then(data => setInquiries(data.items || []));
        }
      });
    }
  }, [ws, entityId]);

  // Group inquiries
  const groupedInquiries = useMemo(() => {
    const now = new Date();
    const oneDayFromNow = new Date(now.getTime() + 24 * 60 * 60 * 1000);
    
    const groups = {
      attention: [],
      dueSoon: [],
      all: inquiries,
    };

    inquiries.forEach(inquiry => {
      // Needs attention: status is ATTENTION or has attentionAt in the past
      if (inquiry.status === 'ATTENTION' || 
          (inquiry.attentionAt && new Date(inquiry.attentionAt) <= now)) {
        groups.attention.push(inquiry);
      }
      
      // Due soon: has deadline within 24 hours
      if (inquiry.deadlineAt) {
        const deadline = new Date(inquiry.deadlineAt);
        if (deadline <= oneDayFromNow && deadline > now && inquiry.status === 'PENDING') {
          groups.dueSoon.push(inquiry);
        }
      }
    });

    return groups;
  }, [inquiries]);

  // Get filtered inquiries based on group filter
  const filteredInquiries = useMemo(() => {
    if (groupFilter === 'all') {
      return groupedInquiries.all;
    } else if (groupFilter === 'attention') {
      return groupedInquiries.attention;
    } else if (groupFilter === 'dueSoon') {
      return groupedInquiries.dueSoon;
    }
    return [];
  }, [groupFilter, groupedInquiries]);

  if (loading) {
    return <div className="inbox-loading">Loading inquiries...</div>;
  }

  return (
    <div className="inbox">
      <div className="inbox-header">
        <div className="inbox-header-top">
          <h2>Inbox</h2>
          {entityId ? (
            <div className="inbox-entity-info">
              <span>Entity: {entityId.substring(0, 8)}...</span>
              <button 
                onClick={() => {
                  localStorage.removeItem('pxbox_entity_id');
                  setEntityId(null);
                  window.history.replaceState({}, '', window.location.pathname);
                }}
                className="entity-clear-btn"
              >
                Clear
              </button>
            </div>
          ) : (
            <div className="inbox-entity-prompt">
              <span>No entity selected - showing all requests</span>
              <input
                type="text"
                placeholder="Enter entity ID"
                onKeyPress={(e) => {
                  if (e.key === 'Enter' && e.target.value.trim()) {
                    const newEntityId = e.target.value.trim();
                    setEntityId(newEntityId);
                    localStorage.setItem('pxbox_entity_id', newEntityId);
                    window.history.replaceState({}, '', `?entityId=${newEntityId}`);
                  }
                }}
                className="entity-input"
              />
            </div>
          )}
        </div>
        <div className="inbox-filters">
          <button
            className={groupFilter === 'all' ? 'active' : ''}
            onClick={() => setGroupFilter('all')}
          >
            All ({groupedInquiries.all.length})
          </button>
          <button
            className={groupFilter === 'attention' ? 'active' : ''}
            onClick={() => setGroupFilter('attention')}
          >
            Needs Attention ({groupedInquiries.attention.length})
          </button>
          <button
            className={groupFilter === 'dueSoon' ? 'active' : ''}
            onClick={() => setGroupFilter('dueSoon')}
          >
            Due Soon ({groupedInquiries.dueSoon.length})
          </button>
        </div>
      </div>
      {filteredInquiries.length === 0 ? (
        <div className="inbox-empty">
          <div className="inbox-empty-icon">📭</div>
          <div className="inbox-empty-text">No inquiries in this group</div>
          {!entityId && (
            <div className="inbox-empty-hint">Enter an entity ID above to filter inquiries</div>
          )}
        </div>
      ) : (
        <div className="inbox-list">
          {filteredInquiries.map(inquiry => {
            const deadline = inquiry.deadlineAt ? new Date(inquiry.deadlineAt) : null;
            const isDueSoon = deadline && deadline <= new Date(Date.now() + 24 * 60 * 60 * 1000) && deadline > new Date();
            const isOverdue = deadline && deadline <= new Date();
            const createdAt = new Date(inquiry.createdAt);
            const timeAgo = getTimeAgo(createdAt);
            
            return (
              <div
                key={inquiry.id}
                className={`inbox-item inbox-item-${inquiry.status.toLowerCase()}`}
                onClick={() => onSelectRequest(inquiry.id)}
              >
                <div className="inbox-item-main">
                  <div className="inbox-item-left">
                    <div className="inbox-item-status-badge">
                      <span className={`status-indicator status-${inquiry.status.toLowerCase()}`}></span>
                      <span className="status-text">{getStatusLabel(inquiry.status)}</span>
                    </div>
                    <div className="inbox-item-meta">
                      <span className="inbox-item-time">{timeAgo}</span>
                      {inquiry.id && (
                        <span className="inbox-item-id">ID: {inquiry.id.substring(0, 12)}...</span>
                      )}
                    </div>
                  </div>
                  <div className="inbox-item-right">
                    {isOverdue && (
                      <div className="inbox-item-badge overdue">
                        <span>⚠️</span>
                        <span>Overdue</span>
                      </div>
                    )}
                    {isDueSoon && !isOverdue && (
                      <div className="inbox-item-badge due-soon">
                        <span>⏰</span>
                        <span>Due Soon</span>
                      </div>
                    )}
                    {deadline && !isDueSoon && !isOverdue && (
                      <div className="inbox-item-deadline">
                        Due {deadline.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                      </div>
                    )}
                    {inquiry.status === 'ATTENTION' && (
                      <div className="inbox-item-badge attention">
                        <span>⚠️</span>
                        <span>Needs Attention</span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

