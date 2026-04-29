import { ChangeEvent, useState } from 'react';
import { XMarkIcon, EyeIcon, EyeSlashIcon } from '@heroicons/react/24/solid';
import FormFieldError from './FormFieldError';
import { getFieldBorderClass } from '../utils/fieldStyles';

interface InputFieldProps {
  label: string;
  name: string;
  type?: string;
  value: string;
  /**
   * onChange は `(e: ChangeEvent<HTMLInputElement>) => void` を厳密に要求する。
   *
   * 旧コードは `ChangeEvent | { target: {name, value} }` の union を許していたが、
   * 関数パラメータの contravariance により呼び出し側の narrow handler が代入不能になる
   * TS2322 を 9 ページで起こしていた。handleClear 側で synthetic event を組み立てて
   * 型を満たすことで解消する。
   */
  onChange: (e: ChangeEvent<HTMLInputElement>) => void;
  placeholder?: string;
  error?: string;
  disabled?: boolean;
  maxLength?: number;
}

export default function InputField({
  label,
  name,
  type = 'text',
  value,
  onChange,
  placeholder,
  error,
  disabled,
  maxLength,
}: InputFieldProps) {
  const [inputValue, setInputValue] = useState(value || '');
  const [showPassword, setShowPassword] = useState(false);
  const isPasswordField = type === 'password';

  const handleClear = () => {
    setInputValue('');
    // 呼び出し側が e.target.value / e.target.name のみ参照する前提で
    // ChangeEvent<HTMLInputElement> 互換の最小オブジェクトを synthesize する。
    // 完全な ChangeEvent ではないが構造的に target.{name,value} を保証する。
    const syntheticTarget = Object.assign(document.createElement('input'), { name, value: '' });
    const syntheticEvent = {
      target: syntheticTarget,
      currentTarget: syntheticTarget,
    } as unknown as ChangeEvent<HTMLInputElement>;
    onChange(syntheticEvent);
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
          type={isPasswordField && showPassword ? 'text' : type}
          value={inputValue}
          onChange={(e) => {
            setInputValue(e.target.value);
            onChange(e);
          }}
          placeholder={placeholder}
          disabled={disabled}
          maxLength={maxLength}
          aria-invalid={!!error}
          aria-describedby={error ? `${name}-error` : undefined}
          className={`w-full border rounded-lg px-4 py-2.5 pr-10 focus:outline-none focus:ring-1 transition-colors duration-150 disabled:opacity-50 disabled:cursor-not-allowed ${getFieldBorderClass(!!error)}`}
        />
        {isPasswordField && !disabled ? (
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            aria-label={showPassword ? 'パスワードを非表示' : 'パスワードを表示'}
            className="absolute right-3 top-1/2 transform -translate-y-1/2 text-[var(--color-text-faint)] hover:text-[var(--color-text-tertiary)] transition-colors"
          >
            {showPassword ? <EyeSlashIcon className="w-5 h-5" /> : <EyeIcon className="w-5 h-5" />}
          </button>
        ) : inputValue && !disabled ? (
          <button
            type="button"
            onClick={handleClear}
            aria-label="入力をクリア"
            className="absolute right-3 top-1/2 transform -translate-y-1/2 text-[var(--color-text-faint)] hover:text-[var(--color-text-tertiary)] transition-colors"
          >
            <XMarkIcon className="w-5 h-5" />
          </button>
        ) : null}
      </div>
      <FormFieldError name={name} error={error} />
    </div>
  );
}
