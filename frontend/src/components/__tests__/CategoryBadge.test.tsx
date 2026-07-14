import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import CategoryBadge from '../CategoryBadge';
import { assignCategoryColors, categoryColorClass } from '../../utils/categoryColor';

describe('CategoryBadge', () => {
  it('カテゴリ名を表示する', () => {
    render(<CategoryBadge category="基礎" />);
    expect(screen.getByText('基礎')).toBeInTheDocument();
  });

  it('同じカテゴリ名は常に同じ色クラスになる（決定的）', () => {
    expect(categoryColorClass('基礎')).toBe(categoryColorClass('基礎'));
    expect(categoryColorClass('コンテナ操作')).toBe(categoryColorClass('コンテナ操作'));
  });

  it('濃色塗り + 白文字のパレットから色を選ぶ', () => {
    render(<CategoryBadge category="制御構文" />);
    const badge = screen.getByText('制御構文');
    expect(badge.className).toMatch(/bg-(rose|orange|amber|lime|emerald|teal|sky|blue|violet|fuchsia)-600/);
    expect(badge.className).toContain('text-white');
  });

  it('colorClass の明示指定が優先される', () => {
    render(<CategoryBadge category="基礎" colorClass="bg-violet-600 text-white" />);
    expect(screen.getByText('基礎').className).toContain('bg-violet-600');
  });
});

describe('assignCategoryColors', () => {
  it('隣り合うカテゴリが同色にならない（ハッシュ衝突時は次の色へずらす）', () => {
    // 本番 bash 演習で実際に名前ハッシュが衝突する並び。
    const seq = ['基本操作', '入出力', 'ファイル操作', 'テキスト処理', 'プロセス'];
    const colors = assignCategoryColors(seq);
    for (let i = 1; i < seq.length; i++) {
      expect(colors.get(seq[i])).not.toBe(colors.get(seq[i - 1]));
    }
  });

  it('衝突しないカテゴリは名前ハッシュの色のまま（フィルタが変わっても安定）', () => {
    const colors = assignCategoryColors(['基礎', '関数']);
    expect(colors.get('基礎')).toBe(categoryColorClass('基礎'));
    expect(colors.get('関数')).toBe(categoryColorClass('関数'));
  });

  it('本番に実在する代表的なカテゴリ群で色が分散する（全部同色にならない）', () => {
    const categories = ['基礎', '制御構文', '配列', '関数', 'コンテナ操作', '並行処理', '型', 'ブランチ'];
    const colors = assignCategoryColors(categories);
    expect(new Set(colors.values()).size).toBeGreaterThan(2);
  });
});
