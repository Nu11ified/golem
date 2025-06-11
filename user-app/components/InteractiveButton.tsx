import React, { useState } from 'react';

export default function InteractiveButton() {
  const [clicked, setClicked] = useState(false);
  return (
    <button
      style={{ marginTop: '2rem', padding: '0.5rem 1.5rem', fontSize: '1.1rem', borderRadius: '6px', border: 'none', background: '#61DAFB', color: '#222', cursor: 'pointer' }}
      onClick={() => setClicked(c => !c)}
    >
      {clicked ? 'You clicked me!' : 'Click me!'}
    </button>
  );
} 