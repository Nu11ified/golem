const fs = require('fs');
const path = require('path');
const siteMetadata = require('../site-metadata');

console.log('siteMetadata:', siteMetadata);
if (!siteMetadata) throw new Error('siteMetadata is undefined!');

const outPath = path.resolve(__dirname, '../../node-renderer/metadata.json');
fs.writeFileSync(outPath, JSON.stringify(siteMetadata, null, 2));
console.log('Synced site-metadata.js to node-renderer/metadata.json'); 