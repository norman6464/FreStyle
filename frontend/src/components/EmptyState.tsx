interface EmptyStateProps {
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  description?: string;
  action?: { label: string; onClick: () => void };
}

export default function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center h-full text-center px-6">
      <div className="bg-surface-3 rounded-full p-4 mb-4">
        <Icon className="w-8 h-8 text-[#666666]" />
      </div>
      <h3 className="text-base font-semibold text-[#D0D0D0] mb-1">{title}</h3>
      {description && <p className="text-sm text-[#888888] max-w-xs">{description}</p>}
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 px-4 py-2 text-sm font-medium text-white bg-primary-500 rounded-lg hover:bg-primary-600 transition-colors"
        >
          {action.label}
        </button>
      )}
    </div>
  );
}
