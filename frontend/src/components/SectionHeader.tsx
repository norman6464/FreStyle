import { ComponentType, SVGProps } from 'react';

interface SectionHeaderProps {
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  title: string;
}

export default function SectionHeader({ icon: Icon, title }: SectionHeaderProps) {
  return (
    <div className="flex items-center gap-2 mb-3">
      <Icon className="w-4 h-4 text-primary-500" />
      <h3 className="text-sm font-bold text-[var(--color-text-primary)]">{title}</h3>
    </div>
  );
}
