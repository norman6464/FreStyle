import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LanguageBadge from '../LanguageBadge';

describe('LanguageBadge', () => {
  it('docker は青系(sky)の色クラスになる', () => {
    render(<LanguageBadge language="docker" />);
    const badge = screen.getByText('Docker');
    expect(badge.className).toContain('bg-sky-500/25');
    expect(badge.className).toContain('text-sky-700');
  });

  it('言語ごとに異なる色になる（go=cyan / php=indigo / git=orange / bash=slate / javascript=yellow / typescript=blue）', () => {
    const cases: Array<[string, string]> = [
      ['go', 'bg-cyan-500/25'],
      ['php', 'bg-indigo-500/25'],
      ['git', 'bg-orange-500/25'],
      ['bash', 'bg-slate-500/25'],
      ['javascript', 'bg-yellow-500/25'],
      ['typescript', 'bg-blue-500/25'],
    ];
    for (const [lang, cls] of cases) {
      const { unmount } = render(<LanguageBadge language={lang} />);
      const label = lang.charAt(0).toUpperCase() + lang.slice(1);
      expect(screen.getByText(label).className).toContain(cls);
      unmount();
    }
  });

  it('大文字小文字を無視して色を引き、表記は先頭のみ大文字に整形される', () => {
    render(<LanguageBadge language="DOCKER" />);
    expect(screen.getByText('Docker').className).toContain('bg-sky-500/25');
  });

  it('未知の言語は無彩色（surface-3）にフォールバックする', () => {
    render(<LanguageBadge language="rust" />);
    const badge = screen.getByText('Rust');
    expect(badge.className).toContain('bg-surface-3');
    expect(badge.className).not.toContain('bg-sky-500/25');
  });

  it('mono 指定で等幅フォントになる', () => {
    render(<LanguageBadge language="go" mono />);
    expect(screen.getByText('Go').className).toContain('font-mono');
  });
});
