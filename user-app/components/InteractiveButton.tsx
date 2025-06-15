import React, { useState } from 'react';

export default function InteractiveButton() {
  const [clicked, setClicked] = useState(false);
  return (
    <button
      className="mt-4 px-6 py-2 text-lg rounded-lg font-semibold bg-cyan-400/80 text-gray-900 shadow-lg hover:bg-cyan-300/90 active:scale-95 transition-all duration-150 border border-cyan-300 backdrop-blur-md focus:outline-none focus:ring-2 focus:ring-cyan-300 focus:ring-offset-2"
      onClick={() => setClicked(c => !c)}
    >
      {clicked ? 'You clicked me!' : 'Click me!'}
    </button>
  );
} 