import React from 'react';
import InteractiveButton from '../components/InteractiveButton.js';
import ServerFunctionDemo from "../components/ServerFunctionDemo.js";
import siteMetadata from '../site-metadata.js';

const frameworks = [
  { name: 'Go/React', time: 37, color: 'bg-orange-500', logo: '⚡️' },
  { name: 'Next.js', time: 120, color: 'bg-gray-900', logo: '▲' },
  { name: 'Astro', time: 95, color: 'bg-purple-700', logo: '🪐' },
  { name: 'SvelteKit', time: 110, color: 'bg-red-500', logo: '🧡' },
];

function FrameworkComparisonCard() {
  const max = Math.max(...frameworks.map(f => f.time));
  return (
    <div className="w-full max-w-2xl mx-auto mb-8 p-6 rounded-2xl shadow-xl bg-white/10 backdrop-blur-md border border-orange-400 animate-fade-in">
      <h3 className="text-xl font-bold text-orange-400 mb-4 text-center tracking-wide">SSR Speed Comparison</h3>
      <div className="space-y-4">
        {frameworks.map(f => (
          <div key={f.name} className="flex items-center gap-4 group">
            <span className={`text-2xl ${f.color} rounded-lg w-10 h-10 flex items-center justify-center shadow-md`}>{f.logo}</span>
            <span className="font-semibold text-white w-28">{f.name}</span>
            <div className="flex-1 h-3 bg-gray-800 rounded-full overflow-hidden">
              <div
                className={`h-full ${f.color} transition-all duration-700 group-hover:scale-x-105 origin-left`}
                style={{ width: `${(f.time / max) * 100}%` }}
              />
            </div>
            <span className="ml-4 font-mono text-sm text-gray-200">{f.time} ms</span>
          </div>
        ))}
      </div>
      <div className="text-xs text-gray-400 text-center mt-3">Lower is better. Go/React is blazing fast!</div>
    </div>
  );
}

export default function HomePage({ frameworkName, ssrTimeMs }: { frameworkName: string, ssrTimeMs?: number }) {
  return (
    <div
      className="flex flex-col items-center justify-center min-h-screen bg-gradient-to-br from-gray-950 via-gray-900 to-black"
      style={{ fontFamily: `var(--site-font, ${siteMetadata.fontFamily})` }}
    >
      <style>{`:root { --site-font: ${siteMetadata.fontFamily}; }`}</style>
      <link rel="stylesheet" href={siteMetadata.fontUrl} />
      <div className="w-full max-w-2xl mx-auto mt-12 mb-8 p-8 rounded-2xl shadow-2xl bg-white/10 backdrop-blur-md border border-orange-400 animate-fade-in">
        <h1 className="text-5xl font-extrabold text-orange-400 mb-2 text-center drop-shadow-lg tracking-tight">Ship faster with Go/React</h1>
        <p className="text-lg text-gray-200 mb-4 text-center">
          The database developers trust, on a serverless platform designed to help you build reliable and scalable applications faster.
        </p>
        {typeof ssrTimeMs === 'number' && (
          <div className="flex flex-col items-center justify-center mb-4">
            <span className="text-xs uppercase tracking-widest text-gray-400">SSR Render Time</span>
            <span className="text-2xl font-mono font-bold text-orange-300 mt-1 animate-pulse">{ssrTimeMs} ms</span>
          </div>
        )}
        <div className="flex flex-col items-center gap-4 mt-6">
          <InteractiveButton />
        </div>
      </div>
      <FrameworkComparisonCard />
      <div className="w-full max-w-3xl mx-auto mt-8">
        <ServerFunctionDemo />
      </div>
    </div>
  );
} 