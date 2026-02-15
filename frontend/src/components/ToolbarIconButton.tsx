import type { ComponentType, SVGProps } from 'react';

interface ToolbarIconButtonProps {
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  label: string;
  onClick: () => void;
  isActive?: boolean;
}

export default function ToolbarIconButton({ icon: Icon, label, onClick, isActive = false }: ToolbarIconButtonProps) {
  return (
    <button
      type="button"
      aria-label={label}
      className={`w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] transition-colors ${
        isActive ? 'text-primary-500' : 'text-[var(--color-text-faint)]'
      }`}
      onClick={onClick}
    >
      <Icon className="w-4 h-4" />
    </button>
  );
}
