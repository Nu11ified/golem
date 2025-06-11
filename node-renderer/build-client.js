const esbuild = require('esbuild');
const path = require('path');

const entry = 'hydrate.tsx';
const outfile = path.join(__dirname, 'dist', 'client.js');

const isWatch = process.argv.includes('--watch') || process.env.NODE_ENV !== 'production';

const buildOptions = {
  absWorkingDir: __dirname,
  entryPoints: [entry],
  bundle: true,
  outfile,
  platform: 'browser',
  format: 'iife',
  globalName: 'HomePageBundle',
  jsx: 'automatic',
  sourcemap: true,
  define: { 'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development') },
  loader: { '.ts': 'ts', '.tsx': 'tsx' }
};

async function run() {
  if (isWatch) {
    const ctx = await esbuild.context(buildOptions);
    await ctx.watch();
    console.log('Watching for changes...');
  } else {
    await esbuild.build(buildOptions);
    console.log('Client bundle built.');
  }
}

run().catch((e) => {
  console.error(e);
  process.exit(1);
}); 