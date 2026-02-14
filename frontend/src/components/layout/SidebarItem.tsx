import { ComponentType, SVGProps } from 'react';
import { Link } from 'react-router-dom';

interface SidebarItemProps {
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  label: string;
  to: string;
  active: boolean;
  onClick?: () => void;
  badge?: number;
}

export default function SidebarItem({ icon: Icon, label, to, active, onClick, badge }: SidebarItemProps) {
  return (
    <Link
      to={to}
      onClick={onClick}
      className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors duration-150 ${
        active
          ? 'bg-surface-3 text-primary-300'
          : 'text-[#A0A0A0] hover:bg-surface-2 hover:text-[#F0F0F0]'
      }`}
    >
      <Icon className="w-5 h-5 flex-shrink-0" />
      <span className="flex-1">{label}</span>
      {badge !== undefined && badge > 0 && (
        <span
          data-testid="sidebar-badge"
          className="bg-primary-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full"
        >
          {badge > 99 ? '99+' : badge}
        </span>
      )}
    </Link>
  );
}
