import { ChangeEvent } from 'react';

interface TextareaFieldProps {
  label: string;
  name: string;
  value: string;
  onChange: (e: ChangeEvent<HTMLTextAreaElement>) => void;
  placeholder?: string;
  rows?: number;
  maxLength?: number;
  error?: string;
}

export default function TextareaField({ label, name, value, onChange, placeholder, rows = 3, maxLength, error }: TextareaFieldProps) {
  return (
    <div>
      <label htmlFor={name} className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
        {label}
      </label>
      <textarea
        id={name}
        name={name}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        rows={rows}
        maxLength={maxLength}
        aria-invalid={!!error}
        aria-describedby={error ? `${name}-error` : undefined}
        className={`w-full border rounded-lg px-3 py-2 text-sm focus:ring-1 transition-colors resize-none ${
          error
            ? 'border-rose-500 focus:border-rose-500 focus:ring-rose-500'
            : 'border-surface-3 focus:border-primary-500 focus:ring-primary-500'
        }`}
      />
      {error && (
        <p id={`${name}-error`} role="alert" className="text-xs text-rose-400 mt-1">
          {error}
        </p>
      )}
      {maxLength && (
        <p className={`text-xs text-right mt-1 ${
          value.length >= maxLength
            ? 'text-rose-400'
            : value.length >= maxLength * 0.9
              ? 'text-amber-400'
              : 'text-[var(--color-text-muted)]'
        }`}>
          {value.length} / {maxLength}
        </p>
      )}
    </div>
  );
}
