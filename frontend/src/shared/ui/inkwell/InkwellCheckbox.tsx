import { useId } from 'react';
import { useRipple } from './useRipple';

export interface InkwellCheckboxProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type' | 'size'> {
  label?: string;
}

/** チェックで塗り＋レ点、周囲に押下波紋とホバーの淡い円。 */
export default function InkwellCheckbox({
  label,
  className = '',
  id,
  checked,
  disabled,
  ...props
}: InkwellCheckboxProps) {
  const autoId = useId();
  const inputId = id ?? autoId;
  const { addRipple, rippleOverlay } = useRipple('rgba(25,118,210,0.3)');

  return (
    <label
      htmlFor={inputId}
      className={`inline-flex cursor-pointer items-center font-roboto text-inkwell-text-primary ${
        disabled ? 'cursor-not-allowed opacity-60' : ''
      } ${className}`}
    >
      <span
        onPointerDown={(e) => {
          if (!disabled) addRipple(e);
        }}
        className="relative flex h-[38px] w-[38px] items-center justify-center rounded-full transition-colors hover:bg-inkwell-primary/[0.04]"
      >
        <input
          id={inputId}
          type="checkbox"
          checked={checked}
          disabled={disabled}
          className="peer sr-only"
          {...props}
        />
        {/* 枠（未チェック時のみ表示） */}
        <span className="pointer-events-none absolute h-[18px] w-[18px] rounded-[2px] border-2 border-inkwell-text-secondary transition-opacity peer-checked:opacity-0 peer-disabled:border-inkwell-text-disabled" />
        {/* 塗り + レ点（チェック時に表示） */}
        <span className="pointer-events-none absolute flex h-[18px] w-[18px] items-center justify-center rounded-[2px] bg-inkwell-primary opacity-0 transition-opacity peer-checked:opacity-100 peer-disabled:bg-inkwell-text-disabled">
          <svg viewBox="0 0 24 24" className="h-[18px] w-[18px] text-white" aria-hidden="true">
            <path fill="currentColor" d="M9 16.17 4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
          </svg>
        </span>
        {rippleOverlay}
      </span>
      {label && <span className="pl-1 pr-2 text-base">{label}</span>}
    </label>
  );
}
