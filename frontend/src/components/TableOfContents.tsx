import type { TocHeading } from '../hooks/useTableOfContents';

interface TableOfContentsProps {
  headings: TocHeading[];
  onHeadingClick: (id: string) => void;
}

const INDENT: Record<number, string> = {
  1: 'pl-0',
  2: 'pl-4',
  3: 'pl-8',
};

export default function TableOfContents({ headings, onHeadingClick }: TableOfContentsProps) {
  if (headings.length === 0) return null;

  return (
    <div className="mb-4 p-3 bg-surface-2 rounded-lg border border-surface-3">
      <p className="text-xs font-semibold text-[var(--color-text-tertiary)] mb-2">目次</p>
      <nav>
        <ul className="space-y-0.5">
          {headings.map((heading) => (
            <li key={heading.id}>
              <button
                onClick={() => onHeadingClick(heading.id)}
                className={`${INDENT[heading.level] ?? 'pl-0'} text-left w-full text-sm text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] transition-colors duration-100 py-0.5 truncate`}
              >
                {heading.text}
              </button>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
}
