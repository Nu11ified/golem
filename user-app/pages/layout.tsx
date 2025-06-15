import React from 'react';
import siteMetadata from '../site-metadata.js';

export const metadata = siteMetadata;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ minHeight: '100vh' }}>
      <header style={{ padding: '1rem', background: '#61dafb', color: '#222', fontWeight: 'bold', fontSize: '1.5rem' }}>
        🚀 {siteMetadata.title}
      </header>
      <main style={{ padding: '2rem' }}>{children}</main>
    </div>
  );
} 