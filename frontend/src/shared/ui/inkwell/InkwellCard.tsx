import { ReactNode } from 'react';

export type InkwellElevation = 0 | 1 | 2 | 3 | 4 | 8;

export interface InkwellCardProps extends React.HTMLAttributes<HTMLDivElement> {
  elevation?: InkwellElevation;
  children: ReactNode;
}

const ELEVATION: Record<InkwellElevation, string> = {
  0: 'border border-inkwell-divider',
  1: 'shadow-inkwell-1',
  2: 'shadow-inkwell-2',
  3: 'shadow-inkwell-3',
  4: 'shadow-inkwell-4',
  8: 'shadow-inkwell-8',
};

/** 標高シャドウで浮遊感を出す面。elevation=0 は影の代わりに枠線。 */
export default function InkwellCard({ elevation = 1, children, className = '', ...props }: InkwellCardProps) {
  return (
    <div
      className={`overflow-hidden rounded bg-white font-roboto text-inkwell-text-primary ${ELEVATION[elevation]} ${className}`}
      {...props}
    >
      {children}
    </div>
  );
}

/** カード本文の内側余白。 */
export function InkwellCardContent({ children, className = '', ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={`p-4 ${className}`} {...props}>
      {children}
    </div>
  );
}

/** カード下部のアクション行（ボタンを右寄せ等）。 */
export function InkwellCardActions({ children, className = '', ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={`flex items-center gap-2 p-2 ${className}`} {...props}>
      {children}
    </div>
  );
}
