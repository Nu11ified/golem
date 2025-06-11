import React from 'react';
import { createRoot } from 'react-dom/client';
import HomePage from './index';

// Props injected by the server for hydration
const props = (window as any).__SSR_PROPS__ || { frameworkName: 'Go/React' };

createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <HomePage {...props} />
  </React.StrictMode>
); 