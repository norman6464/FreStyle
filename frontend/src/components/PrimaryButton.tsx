import { ReactNode } from 'react';
import Button from '@/shared/ui/Button';

interface PrimaryButtonProps {
  children: ReactNode;
  onClick?: () => void;
  type?: 'button' | 'submit' | 'reset';
  disabled?: boolean;
  loading?: boolean;
}

/** PrimaryButton は全幅・primaryバリアントの Button ショートハンド。 */
export default function PrimaryButton({
  children,
  onClick,
  type = 'button',
  disabled,
  loading,
}: PrimaryButtonProps) {
  return (
    <Button variant="primary" fullWidth type={type} onClick={onClick} disabled={disabled} loading={loading}>
      {children}
    </Button>
  );
}
