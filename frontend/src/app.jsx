import { useState } from 'preact/hooks';
import { EmailInbox } from './components/EmailInbox';
import { RequestForm } from './components/RequestForm';
import './App.css';

export function App() {
  const [view, setView] = useState('inbox');
  const [selectedRequestId, setSelectedRequestId] = useState(null);

  return (
    <div className="app">
      <header className="app-header">
        <h1>PxBox</h1>
        <nav>
          <button onClick={() => setView('inbox')}>Inbox</button>
        </nav>
      </header>
      <main className="app-main">
        {view === 'inbox' && (
          <EmailInbox
            onSelectRequest={(id) => {
              setSelectedRequestId(id);
              setView('form');
            }}
          />
        )}
        {view === 'form' && selectedRequestId && (
          <RequestForm
            requestId={selectedRequestId}
            onBack={() => {
              setView('inbox');
              setSelectedRequestId(null);
            }}
          />
        )}
      </main>
    </div>
  );
}
