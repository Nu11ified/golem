import { dirname, resolve } from 'path';
import { fileURLToPath } from 'url';
import fs from 'fs';

const __dirname = dirname(fileURLToPath(import.meta.url));

// Dynamically import the TS file using ts-node/esm loader
const { default: siteMetadata } = await import(resolve(__dirname, '../../user-app/site-metadata.ts'));
console.log('siteMetadata:', siteMetadata);
if (!siteMetadata) throw new Error('siteMetadata is undefined!');

const outPath = resolve(__dirname, '../../node-renderer/metadata.json');
fs.writeFileSync(outPath, JSON.stringify(siteMetadata, null, 2));
console.log('Synced site-metadata.ts to node-renderer/metadata.json'); 