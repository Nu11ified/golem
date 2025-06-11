import React from 'react';
import { createRoot } from 'react-dom/client';
import { pages, layouts } from './importMap.generated';

const pagePath = (window as any).__SSR_PAGE__;
const props = (window as any).__SSR_PROPS__ || {};
const layoutPath = (window as any).__SSR_LAYOUT__;

async function hydrate() {
  const pageLoader = pages[pagePath];
  if (!pageLoader) throw new Error(`Page not found: ${pagePath}`);
  const pageMod = await pageLoader();
  const Page = pageMod.default;

  let element: React.ReactElement;
  if (layoutPath && layouts[layoutPath]) {
    const layoutMod = await layouts[layoutPath]!();
    const Layout = layoutMod.default;
    element = <Layout {...props}><Page {...props} /></Layout>;
  } else {
    element = <Page {...props} />;
  }

  createRoot(document.getElementById('root')!).render(
    <React.StrictMode>{element}</React.StrictMode>
  );
}

hydrate(); 