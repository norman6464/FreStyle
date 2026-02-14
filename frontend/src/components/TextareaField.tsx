import { ChangeEvent } from 'react';

interface TextareaFieldProps {
  label: string;
  name: string;
  value: string;
  onChange: (e: ChangeEvent<HTMLTextAreaElement>) => void;
  placeholder?: string;
  rows?: number;
  maxLength?: number;
}

export default function TextareaField({ label, name, value, onChange, placeholder, rows = 3, maxLength }: TextareaFieldProps) {
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
        className="w-full border border-surface-3 rounded-lg px-3 py-2 text-sm focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-colors resize-none"
      />
      {maxLength && (
        <p className="text-xs text-[var(--color-text-muted)] text-right mt-1">
          {value.length} / {maxLength}
        </p>
      )}
    </div>
  );
}
