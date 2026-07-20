export function getFieldBorderClass(hasError: boolean): string {
  return hasError
    ? 'border-rose-500 focus:border-rose-500 focus:ring-rose-500'
    : 'border-surface-3 focus:border-brand-400 focus:ring-brand-400';
}
