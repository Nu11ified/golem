import React, { useEffect, useState } from "react";

function JsonBlock({ data }: { data: any }) {
  const [open, setOpen] = useState(false);
  if (data == null) return <span className="text-gray-400">No response</span>;
  return (
    <div className="bg-gray-900/80 rounded-lg p-3 mt-2 text-sm font-mono text-gray-200 overflow-x-auto shadow-inner border border-gray-700">
      <button
        onClick={() => setOpen(o => !o)}
        className="bg-transparent border-none text-cyan-300 cursor-pointer font-semibold mb-2 hover:underline focus:outline-none"
      >
        {open ? '▼ Hide JSON' : '▶ Show JSON'}
      </button>
      {open && <pre className="m-0 whitespace-pre-wrap">{JSON.stringify(data, null, 2)}</pre>}
    </div>
  );
}

function StatusIcon({ status }: { status: 'success' | 'error' | 'loading' }) {
  if (status === 'loading') return <span className="text-cyan-300 mr-2">⏳</span>;
  if (status === 'success') return <span className="text-green-400 mr-2">✔️</span>;
  if (status === 'error') return <span className="text-red-400 mr-2">❌</span>;
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
    <div className="bg-gray-900/80 text-white p-8 rounded-2xl max-w-2xl mx-auto shadow-2xl border border-cyan-700">
      <h2 className="text-2xl font-bold text-center mb-6 tracking-wide text-cyan-200 flex items-center justify-center gap-2">
        <span className="text-cyan-300">🧩</span>
        Server Function Demo
      </h2>
      <div className="flex flex-col md:flex-row gap-6 md:gap-8 justify-center">
        <div className="flex-1 min-w-[260px] bg-gray-800/80 rounded-xl p-5 shadow-md border border-gray-700">
          <div className="flex items-center mb-2">
            <StatusIcon status={goStatus} />
            <span className="font-semibold text-base">Go Function <span className="text-blue-200">/api/go/hello</span></span>
          </div>
          <JsonBlock data={goResult} />
        </div>
        <div className="flex-1 min-w-[260px] bg-gray-800/80 rounded-xl p-5 shadow-md border border-gray-700">
          <div className="flex items-center mb-2">
            <StatusIcon status={tsStatus} />
            <span className="font-semibold text-base">TypeScript Function <span className="text-yellow-200">/api/ts/hello</span></span>
          </div>
          <JsonBlock data={tsResult} />
        </div>
      </div>
      {error && <p className="text-red-400 mt-6 text-center">Error: {error}</p>}
      <div className="text-center mt-8 text-xs text-gray-400">
        <span>🔄 Hot reload supported for both Go and TypeScript server functions.</span>
      </div>
    </div>
  );
} 