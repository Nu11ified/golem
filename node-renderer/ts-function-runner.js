#!/usr/bin/env node

const path = require('path');
const fs = require('fs');

process.on('uncaughtException', err => {
  console.error(JSON.stringify({ error: 'Uncaught Exception', message: err.message, stack: err.stack }));
  process.exit(1);
});
process.on('unhandledRejection', err => {
  console.error(JSON.stringify({ error: 'Unhandled Rejection', message: err && err.message, stack: err && err.stack }));
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
    console.error(JSON.stringify({ error: 'Invalid JSON input', stack: e.stack }));
    process.exit(1);
  }

  const { functionName, body, query, headers } = req;

  const isProd = process.env.NODE_ENV === 'production';
  const extension = isProd ? '.js' : '.ts';
  const basePath = isProd
    ? path.resolve(__dirname, '../user-app/dist/server/ts')
    : path.resolve(__dirname, '../user-app/server/ts');
  
  const funcPath = path.join(basePath, functionName + extension);

  if (!fs.existsSync(funcPath)) {
    console.error(JSON.stringify({ error: `Function file not found: ${funcPath}` }));
    process.exit(1);
  }

  try {
    let mod;
    if (isProd) {
      mod = require(funcPath);
    } else {
      // Use dynamic import with ts-node support for dev
      require('ts-node').register();
      mod = await import(funcPath);
    }

    if (typeof mod.default !== 'function') {
      throw new Error('No default export function found');
    }
    const result = await mod.default({ body, query, headers });
    // Ensure the result is an object that can be stringified
    const response = typeof result === 'object' && result !== null ? result : { result };
    console.log(JSON.stringify(response));
  } catch (err) {
    console.error(JSON.stringify({ error: err.message, stack: err.stack }));
    process.exit(1);
  }
}

main(); 