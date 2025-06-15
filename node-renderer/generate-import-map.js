import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

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

function toStaticImports(arr, varName) {
  // Import from dist/pages and use .js extension
  return arr.map((rel, i) => {
    const jsRel = rel.replace(/\.tsx?$/, '.js');
    return `import ${varName}${i} from '../user-app/dist/pages/${jsRel}';`;
  }).join('\n');
}

function toStaticMap(arr, varName, type) {
  return arr.map((rel, i) => `  'pages/${rel}': () => Promise.resolve({ default: ${varName}${i} }),`).join('\n');
}

const pageImports = toStaticImports(pages, 'Page');
const layoutImports = toStaticImports(layouts, 'Layout');
const pageMap = toStaticMap(pages, 'Page', 'page');
const layoutMap = toStaticMap(layouts, 'Layout', 'layout');

const content = `// AUTO-GENERATED FILE. DO NOT EDIT.
${pageImports}
${layoutImports}

export const pages = {
${pageMap}
};

export const layouts = {
${layoutMap}
};
`;

fs.writeFileSync(output, content);
console.log('Generated importMap.generated.js (static imports)'); 