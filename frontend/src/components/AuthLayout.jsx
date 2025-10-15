import React from 'react';

export default function AuthLayout({ children }) {
  return (
    <div className="min-h-screen flex item-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md bg-white p-8 rounded shadow">
        {children}
      </div>
    </div>
  );
}
