import { useCallback } from 'react';

interface SkipLinkProps {
  targetId: string;
  label?: string;
}

export default function SkipLink({ targetId, label = 'メインコンテンツへスキップ' }: SkipLinkProps) {
  const handleClick = useCallback((e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault();
    const target = document.getElementById(targetId);
    if (target) {
      target.focus();
    }
  }, [targetId]);

  return (
    <a
      href={`#${targetId}`}
      onClick={handleClick}
      className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-50 focus:px-4 focus:py-2 focus:bg-primary-500 focus:text-white focus:rounded-lg focus:font-medium focus:shadow-lg z-50"
    >
      {label}
    </a>
  );
}
