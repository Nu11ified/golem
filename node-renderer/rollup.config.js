import { fileURLToPath } from 'url';
import path from 'path';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import sucrase from '@rollup/plugin-sucrase';
import { terser } from 'rollup-plugin-terser';
import crypto from 'crypto';
import fs from 'fs';
import replace from '@rollup/plugin-replace';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Helper to generate a content hash for the output file
function contentHashPlugin() {
  return {
    name: 'content-hash',
    generateBundle(options, bundle) {
      for (const fileName of Object.keys(bundle)) {
        const asset = bundle[fileName];
        if (asset.type === 'chunk') {
          const hash = crypto.createHash('sha256').update(asset.code).digest('hex').slice(0, 8);
          const newName = fileName.replace(/client(\..*)?\.js$/, `client.${hash}.js`);
          asset.fileName = newName;
          // Write manifest for Go server
          fs.writeFileSync(path.join(__dirname, 'dist', 'client-manifest.json'), JSON.stringify({ clientJs: newName }, null, 2));
        }
      }
    }
  };
}

export default {
  input: 'hydrate.tsx',
  output: {
    dir: 'dist',
    format: 'iife',
    entryFileNames: 'client.js', // Will be renamed by contentHashPlugin
    sourcemap: true,
    name: 'HomePageBundle',
  },
  plugins: [
    replace({
      'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development'),
      'process.env': JSON.stringify({}),
      preventAssignment: true,
    }),
    resolve({ extensions: ['.js', '.ts', '.tsx'] }),
    commonjs(),
    sucrase({
      exclude: ['node_modules/**'],
      include: [
        'hydrate.tsx',
        '../user-app/pages/**/*.tsx',
        '../user-app/pages/**/*.ts',
        '../user-app/components/**/*.tsx',
        '../user-app/components/**/*.ts',
      ],
      transforms: ['typescript', 'jsx'],
    }),
    terser(),
    contentHashPlugin(),
  ],
}; 