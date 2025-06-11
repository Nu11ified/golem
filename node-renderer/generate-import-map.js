const fs = require('fs');
const path = require('path');

const pagesDir = path.join(__dirname, '../user-app/pages');
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

function toMap(arr, type) {
  return arr.map(rel => {
    return `  '${'pages/' + rel}': () => import('./pages/${rel}'),`;
  }).join('\n');
}

const content = `// AUTO-GENERATED FILE. DO NOT EDIT.
export const pages = {
${toMap(pages, 'page')}
};

export const layouts = {
${toMap(layouts, 'layout')}
};
`;

fs.writeFileSync(output, content);
console.log('Generated importMap.generated.js'); 