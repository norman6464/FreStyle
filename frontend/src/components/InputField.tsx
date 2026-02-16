import { ChangeEvent, useState } from 'react';
import { XMarkIcon } from '@heroicons/react/24/solid';
import FormFieldError from './FormFieldError';
import { getFieldBorderClass } from '../utils/fieldStyles';

interface InputFieldProps {
  label: string;
  name: string;
  type?: string;
  value: string;
  onChange: (e: ChangeEvent<HTMLInputElement> | { target: { name: string; value: string } }) => void;
  placeholder?: string;
  error?: string;
  disabled?: boolean;
}

export default function InputField({
  label,
  name,
  type = 'text',
  value,
  onChange,
  error,
  disabled,
}: InputFieldProps) {
  const [inputValue, setInputValue] = useState(value || '');

  const handleClear = () => {
    setInputValue('');
    onChange({ target: { name, value: '' } });
  };

  return (
    <div className="mb-6">
      <label
        className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2"
        htmlFor={name}
      >
        {label}
      </label>
      <div className="relative">
        <input
          id={name}
          name={name}
          type={type}
          value={inputValue}
          onChange={(e) => {
            setInputValue(e.target.value);
            onChange(e);
          }}
          disabled={disabled}
          aria-invalid={!!error}
          aria-describedby={error ? `${name}-error` : undefined}
          className={`w-full border rounded-lg px-4 py-2.5 pr-10 focus:outline-none focus:ring-1 transition-colors duration-150 disabled:opacity-50 disabled:cursor-not-allowed ${getFieldBorderClass(!!error)}`}
        />
        {inputValue && !disabled && (
          <button
            type="button"
            onClick={handleClear}
            className="absolute right-3 top-1/2 transform -translate-y-1/2 text-[var(--color-text-faint)] hover:text-[var(--color-text-tertiary)] transition-colors"
          >
            <XMarkIcon className="w-5 h-5" />
          </button>
        )}
      </div>
      <FormFieldError name={name} error={error} />
    </div>
  );
}
