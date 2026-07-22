import {
  CommandLineIcon,
  ServerStackIcon,
  CubeTransparentIcon,
  CircleStackIcon,
  CloudIcon,
  ShieldCheckIcon,
  ClipboardDocumentListIcon,
  PuzzlePieceIcon,
  Squares2X2Icon,
} from '@heroicons/react/24/outline';
import type { ComponentType, SVGProps } from 'react';

/**
 * CategoryIcon — 学習領域の代表アイコン（FRESTYLE-177）。
 *
 * 領域選択カードで使う。カテゴリは単一技術に対応しないため Devicon のロゴではなく
 * 概念を表す Heroicons を使い、色はカテゴリ色（accentClass）で塗る。
 * 未知 / 未分類（key = ''）は汎用のタイルアイコンにフォールバックする。
 */
const ICONS: Record<string, ComponentType<SVGProps<SVGSVGElement>>> = {
  'dev-basics': CommandLineIcon,
  backend: ServerStackIcon,
  architecture: CubeTransparentIcon,
  database: CircleStackIcon,
  infra: CloudIcon,
  security: ShieldCheckIcon,
  product: ClipboardDocumentListIcon,
  design: PuzzlePieceIcon,
};

export default function CategoryIcon({
  categoryKey,
  className = 'w-6 h-6',
}: {
  categoryKey: string;
  className?: string;
}) {
  const Icon = ICONS[categoryKey] ?? Squares2X2Icon;
  return <Icon className={className} aria-hidden="true" />;
}
