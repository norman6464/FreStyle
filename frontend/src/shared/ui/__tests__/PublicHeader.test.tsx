import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import PublicHeader from '../PublicHeader';

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <PublicHeader />
    </MemoryRouter>
  );
}

describe('PublicHeader', () => {
  it('企業の利用申請への導線がある', () => {
    renderAt('/login');
    const apply = screen.getByRole('link', { name: /企業の利用申請/ });
    expect(apply).toHaveAttribute('href', '/company-application');
  });

  it('招待制のためログイン/新規登録の CTA ボタンは出さない', () => {
    renderAt('/login');
    expect(screen.queryByRole('link', { name: '新規登録' })).not.toBeInTheDocument();
    // ロゴリンクの aria-label は「FreStyle ホーム」なので「ログイン」名のリンクは存在しない。
    expect(screen.queryByRole('link', { name: 'ログイン' })).not.toBeInTheDocument();
  });
});
