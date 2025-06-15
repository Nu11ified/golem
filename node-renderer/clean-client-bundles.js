import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const distDir = path.join(__dirname, 'dist');
const manifestPath = path.join(distDir, 'client-manifest.json');

let keepFile = null;
try {
  const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
  keepFile = manifest.clientJs;
} catch (e) {
  // If manifest missing, keep nothing
}

const files = fs.readdirSync(distDir);
for (const file of files) {
  if (/^client\..*\.js(\.map)?$/.test(file)) {
    if (file !== keepFile && file !== keepFile + '.map') {
      fs.unlinkSync(path.join(distDir, file));
      console.log('Deleted', file);
    }
  }
} 