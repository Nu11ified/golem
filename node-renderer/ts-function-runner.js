#!/usr/bin/env node

const path = require('path');
const fs = require('fs');

async function main() {
  let input = '';
  process.stdin.setEncoding('utf8');
  for await (const chunk of process.stdin) {
    input += chunk;
  }
  let req;
  try {
    req = JSON.parse(input);
  } catch (e) {
    console.error(JSON.stringify({ error: 'Invalid JSON input', stack: e.stack }));
    process.exit(1);
  }

  const { functionName, body, query, headers } = req;
  const tsPath = path.resolve(__dirname, '../user-app/server/ts', functionName + '.ts');
  if (!fs.existsSync(tsPath)) {
    console.error(JSON.stringify({ error: `Function file not found: ${functionName}.ts` }));
    process.exit(1);
  }

  try {
    // Use dynamic import with ts-node support
    require('ts-node').register();
    const mod = await import(tsPath);
    if (typeof mod.default !== 'function') {
      throw new Error('No default export function found');
    }
    const result = await mod.default({ body, query, headers });
    console.log(JSON.stringify({ result }));
  } catch (err) {
    console.error(JSON.stringify({ error: err.message, stack: err.stack }));
    process.exit(1);
  }
}

main(); 