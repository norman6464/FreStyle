import { ChangeEvent, useState } from 'react';
import { XMarkIcon } from '@heroicons/react/24/solid';

interface InputFieldProps {
  label: string;
  name: string;
  type?: string;
  value: string;
  onChange: (e: ChangeEvent<HTMLInputElement> | { target: { name: string; value: string } }) => void;
  placeholder?: string;
}

export default function InputField({
  label,
  name,
  type = 'text',
  value,
  onChange,
}: InputFieldProps) {
  const [inputValue, setInputValue] = useState(value || '');

  const handleClear = () => {
    setInputValue('');
    onChange({ target: { name, value: '' } });
  };

  return (
    <div className="mb-6">
      <label
        className="block text-sm font-medium text-gray-700 mb-2"
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
          className="w-full border border-gray-300 rounded-lg px-4 py-2.5 pr-10 focus:outline-none focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-colors duration-150"
        />
        {inputValue && (
          <button
            type="button"
            onClick={handleClear}
            className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 transition-colors"
          >
            <XMarkIcon className="w-5 h-5" />
          </button>
        )}
      </div>
    </div>
  );
}
