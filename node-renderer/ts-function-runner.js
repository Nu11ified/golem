#!/usr/bin/env node

const path = require('path');
const fs = require('fs');
const { pathToFileURL } = require('url');

process.on('uncaughtException', err => {
  process.stderr.write(JSON.stringify({ error: 'Uncaught Exception', message: err.message, stack: err.stack }));
  process.exit(1);
});
process.on('unhandledRejection', err => {
  process.stderr.write(JSON.stringify({ error: 'Unhandled Rejection', message: err && err.message, stack: err && err.stack }));
  process.exit(1);
});

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
    process.stderr.write(JSON.stringify({ error: 'Invalid JSON input', stack: e.stack }));
    process.exit(1);
  }

  const { functionName, body, query, headers } = req;

  const isProd = process.env.NODE_ENV === 'production';
  const extension = isProd ? '.js' : '.ts';
  const basePath = isProd
    ? path.resolve(__dirname, '../user-app/dist/ts')
    : path.resolve(__dirname, '../user-app/server/ts');
  
  const funcPath = path.join(basePath, functionName + extension);

  if (!fs.existsSync(funcPath)) {
    process.stderr.write(JSON.stringify({ error: `Function file not found: ${funcPath}` }));
    process.exit(1);
  }

  try {
    let mod;
    if (isProd) {
      // Use pathToFileURL for valid file:// URL
      const fileUrl = pathToFileURL(funcPath);
      mod = await import(fileUrl.href);
    } else {
      require('ts-node').register();
      mod = await import(funcPath);
    }

    if (typeof mod.default !== 'function') {
      throw new Error('No default export function found');
    }
    const result = await mod.default({ body, query, headers });
    process.stdout.write(JSON.stringify(typeof result === 'object' && result !== null ? result : { result }));
  } catch (err) {
    process.stderr.write(JSON.stringify({ error: err.message, stack: err.stack }));
    process.exit(1);
  }
}

main(); 