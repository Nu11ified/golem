import React, { useState } from 'react';

export default function HomePage({ frameworkName }: { frameworkName: string }) {
  const [clicked, setClicked] = useState(false);

  return (
    <div style={{ fontFamily: 'sans-serif', padding: '2rem', border: '2px solid #61DAFB', borderRadius: '8px', maxWidth: '600px', margin: '2rem auto', textAlign: 'center' }}>
      <h1>Welcome to the Future!</h1>
      <p>This React component was server-side rendered by a <strong>{frameworkName}</strong> stack.</p>
      <button
        style={{ marginTop: '2rem', padding: '0.5rem 1.5rem', fontSize: '1.1rem', borderRadius: '6px', border: 'none', background: '#61DAFB', color: '#222', cursor: 'pointer' }}
        onClick={() => setClicked(c => !c)}
      >
        {clicked ? 'You clicked me!' : 'Click me!'}
      </button>
    </div>
  );
} 