import React from 'react';
import InteractiveButton from '../components/InteractiveButton';

export default function HomePage({ frameworkName }: { frameworkName: string }) {
  return (
    <div style={{ fontFamily: 'sans-serif', padding: '2rem', border: '2px solid #61DAFB', borderRadius: '8px', maxWidth: '600px', margin: '2rem auto', textAlign: 'center' }}>
      <h1>Welcome to the Future!</h1>
      <p>This React component was server-side rendered by a <strong>{frameworkName}</strong> stack.</p>
      <InteractiveButton />
    </div>
  );
} 