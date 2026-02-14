interface SecondaryPanelProps {
  title: string;
  headerContent?: React.ReactNode;
  children: React.ReactNode;
}

export default function SecondaryPanel({ title, headerContent, children }: SecondaryPanelProps) {
  return (
    <div className="w-72 bg-surface-1 border-r border-surface-3 flex flex-col h-full flex-shrink-0">
      <div className="px-4 py-3 border-b border-surface-3">
        <h2 className="text-sm font-semibold text-[#F0F0F0]">{title}</h2>
        {headerContent && <div className="mt-2">{headerContent}</div>}
      </div>
      <div className="flex-1 overflow-y-auto">{children}</div>
    </div>
  );
}
