import React, { useEffect, useState } from "react";

function JsonBlock({ data }: { data: any }) {
  const [open, setOpen] = useState(false);
  if (data == null) return <span style={{ color: '#aaa' }}>No response</span>;
  return (
    <div style={{ background: '#181c24', borderRadius: 6, padding: 12, marginTop: 8, fontSize: 14, fontFamily: 'monospace', color: '#e6e6e6', overflowX: 'auto' }}>
      <button
        onClick={() => setOpen(o => !o)}
        style={{ background: 'none', border: 'none', color: '#61dafb', cursor: 'pointer', fontWeight: 600, marginBottom: 4 }}
      >
        {open ? '▼ Hide JSON' : '▶ Show JSON'}
      </button>
      {open && <pre style={{ margin: 0 }}>{JSON.stringify(data, null, 2)}</pre>}
    </div>
  );
}

function StatusIcon({ status }: { status: 'success' | 'error' | 'loading' }) {
  if (status === 'loading') return <span style={{ color: '#61dafb', marginRight: 6 }}>⏳</span>;
  if (status === 'success') return <span style={{ color: '#4caf50', marginRight: 6 }}>✔️</span>;
  if (status === 'error') return <span style={{ color: '#f44336', marginRight: 6 }}>❌</span>;
  return null;
}

export default function ServerFunctionDemo() {
  const [goResult, setGoResult] = useState<any>(null);
  const [tsResult, setTsResult] = useState<any>(null);
  const [goStatus, setGoStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [tsStatus, setTsStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setGoStatus('loading');
    setTsStatus('loading');
    setError(null);
    fetch("/api/go/hello", { method: "POST" })
      .then(r => r.json())
      .then(data => {
        setGoResult(data);
        setGoStatus(data.error ? 'error' : 'success');
      })
      .catch(e => {
        setGoResult({ error: e.message });
        setGoStatus('error');
      });
    fetch("/api/ts/hello", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ sample: "data" })
    })
      .then(r => r.json())
      .then(data => {
        setTsResult(data);
        setTsStatus(data.error ? 'error' : 'success');
      })
      .catch(e => {
        setTsResult({ error: e.message });
        setTsStatus('error');
      });
  }, []);

  return (
    <div style={{
      background: "#23272f",
      color: "#fff",
      padding: 32,
      borderRadius: 12,
      maxWidth: 700,
      margin: "32px auto 24px auto",
      boxShadow: "0 4px 24px #0004"
    }}>
      <h2 style={{ textAlign: 'center', marginBottom: 24, letterSpacing: 1 }}>
        <span style={{ color: '#61dafb', marginRight: 8 }}>🧩</span>
        Server Function Demo
      </h2>
      <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap', justifyContent: 'center' }}>
        <div style={{ flex: 1, minWidth: 280, background: '#252a33', borderRadius: 8, padding: 20, boxShadow: '0 2px 8px #0002' }}>
          <div style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
            <StatusIcon status={goStatus} />
            <span style={{ fontWeight: 600, fontSize: 17 }}>Go Function <span style={{ color: '#90caf9' }}>/api/go/hello</span></span>
          </div>
          <JsonBlock data={goResult} />
        </div>
        <div style={{ flex: 1, minWidth: 280, background: '#252a33', borderRadius: 8, padding: 20, boxShadow: '0 2px 8px #0002' }}>
          <div style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
            <StatusIcon status={tsStatus} />
            <span style={{ fontWeight: 600, fontSize: 17 }}>TypeScript Function <span style={{ color: '#ffd54f' }}>/api/ts/hello</span></span>
          </div>
          <JsonBlock data={tsResult} />
        </div>
      </div>
      {error && <p style={{ color: "#f55", marginTop: 24, textAlign: 'center' }}>Error: {error}</p>}
      <div style={{ textAlign: 'center', marginTop: 32, fontSize: 13, color: '#aaa' }}>
        <span>🔄 Hot reload supported for both Go and TypeScript server functions.</span>
      </div>
    </div>
  );
} 