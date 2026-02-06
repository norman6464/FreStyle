import React from 'react';

export default function AuthLayout({ children }) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4 py-12">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
        <div className="bg-primary-500 h-28 flex items-center justify-center">
          <h1 className="text-3xl font-bold text-white">FreStyle</h1>
        </div>
        <div className="p-8">{children}</div>
      </div>
    </div>
  );
}
