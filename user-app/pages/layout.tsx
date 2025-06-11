import React from 'react';

export const metadata = {
  title: 'My Custom Site',
  description: 'A modern Go/React meta-framework demo',
  favicon: '/favicon.ico',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ minHeight: '100vh', background: '#222', color: '#fff' }}>
      <header style={{ padding: '1rem', background: '#61dafb', color: '#222', fontWeight: 'bold', fontSize: '1.5rem' }}>
        🚀 My Custom Site
      </header>
      <main style={{ padding: '2rem' }}>{children}</main>
    </div>
  );
} 