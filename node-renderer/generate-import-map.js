import fs from 'fs';
import path from 'path';
import { fileURLToPath, pathToFileURL } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const pagesDir = path.join(__dirname, '../user-app/pages');
const distPagesDir = path.join(__dirname, '../user-app/dist/pages');
const output = path.join(__dirname, 'importMap.generated.js');

function walk(dir, filelist = []) {
  fs.readdirSync(dir).forEach(file => {
    const filepath = path.join(dir, file);
    if (fs.statSync(filepath).isDirectory()) {
      walk(filepath, filelist);
    } else if (file.endsWith('.tsx') && !file.endsWith('.d.tsx') && !file.endsWith('.test.tsx')) {
      filelist.push(filepath);
    }
  });
  return filelist;
}

const files = walk(pagesDir);
const pages = [];
const layouts = [];

for (const file of files) {
  const rel = path.relative(pagesDir, file).replace(/\\/g, '/');
  if (rel.endsWith('layout.tsx')) {
    layouts.push(rel);
  } else {
    pages.push(rel);
  }
}

function toDynamicMap(arr, type) {
  return arr.map(rel => {
    const jsRel = rel.replace(/\.tsx?$/, '.js');
    const absPath = path.join(distPagesDir, jsRel);
    const fileUrl = pathToFileURL(absPath).href;
    return `  'pages/${rel}': () => import('${fileUrl}'),`;
  }).join('\n');
}

const pageMap = toDynamicMap(pages, 'page');
const layoutMap = toDynamicMap(layouts, 'layout');

const content = `// AUTO-GENERATED FILE. DO NOT EDIT.

export const pages = {
${pageMap}
};

export const layouts = {
${layoutMap}
};
`;

fs.writeFileSync(output, content);
console.log('Generated importMap.generated.js (dynamic imports, file URLs)'); 