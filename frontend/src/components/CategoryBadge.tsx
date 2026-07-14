import { categoryColorClass } from '../utils/categoryColor';

interface CategoryBadgeProps {
  /** 演習のカテゴリ名（例: '基礎' / 'コンテナ操作'）。 */
  category: string;
  /**
   * 色クラスの明示指定。一覧のように複数カテゴリを並べる画面では
   * utils/categoryColor の assignCategoryColors で隣接同色を避けた色を渡す。
   * 未指定なら名前ハッシュの色にフォールバック。
   */
  colorClass?: string;
  className?: string;
}

/** 演習のカテゴリを、カテゴリごとに安定した色のバッジで表示する。 */
export default function CategoryBadge({ category, colorClass, className = '' }: CategoryBadgeProps) {
  return (
    <span
      className={`inline-flex items-center rounded-md px-2.5 py-1 text-xs font-semibold ${colorClass ?? categoryColorClass(category)} ${className}`}
    >
      {category}
    </span>
  );
}
