import { ChangeEvent } from 'react';
import FormFieldError from './FormFieldError';
import { getFieldBorderClass } from '../utils/fieldStyles';

interface SelectOption {
  value: string;
  label: string;
}

interface SelectFieldProps {
  label: string;
  name: string;
  value: string;
  onChange: (e: ChangeEvent<HTMLSelectElement>) => void;
  options: SelectOption[];
  error?: string;
}

export default function SelectField({ label, name, value, onChange, options, error }: SelectFieldProps) {
  return (
    <div>
      <label htmlFor={name} className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
        {label}
      </label>
      <select
        id={name}
        name={name}
        value={value}
        onChange={onChange}
        aria-invalid={!!error}
        aria-describedby={error ? `${name}-error` : undefined}
        className={`w-full border rounded-lg px-3 py-2 text-sm focus:ring-1 transition-colors ${getFieldBorderClass(!!error)}`}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
      <FormFieldError name={name} error={error} />
    </div>
  );
}
