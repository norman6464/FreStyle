import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LanguageBadge from '../LanguageBadge';

describe('LanguageBadge', () => {
  it('docker は青系(sky)の色クラスになる', () => {
    render(<LanguageBadge language="docker" />);
    const badge = screen.getByText('docker');
    expect(badge.className).toContain('bg-sky-500/15');
    expect(badge.className).toContain('text-sky-700');
  });

  it('言語ごとに異なる色になる（go=cyan / php=indigo / git=orange / bash=slate / javascript=yellow / typescript=blue）', () => {
    const cases: Array<[string, string]> = [
      ['go', 'bg-cyan-500/15'],
      ['php', 'bg-indigo-500/15'],
      ['git', 'bg-orange-500/15'],
      ['bash', 'bg-slate-500/15'],
      ['javascript', 'bg-yellow-500/15'],
      ['typescript', 'bg-blue-500/15'],
    ];
    for (const [lang, cls] of cases) {
      const { unmount } = render(<LanguageBadge language={lang} />);
      expect(screen.getByText(lang).className).toContain(cls);
      unmount();
    }
  });

  it('大文字小文字を無視して色を引く', () => {
    render(<LanguageBadge language="Docker" />);
    expect(screen.getByText('Docker').className).toContain('bg-sky-500/15');
  });

  it('未知の言語は無彩色（surface-3）にフォールバックする', () => {
    render(<LanguageBadge language="rust" />);
    const badge = screen.getByText('rust');
    expect(badge.className).toContain('bg-surface-3');
    expect(badge.className).not.toContain('bg-sky-500/15');
  });

  it('mono 指定で等幅フォントになる', () => {
    render(<LanguageBadge language="go" mono />);
    expect(screen.getByText('go').className).toContain('font-mono');
  });
});
