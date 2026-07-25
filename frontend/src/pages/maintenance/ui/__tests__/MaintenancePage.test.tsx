import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import MaintenancePage from '../MaintenancePage';

describe('MaintenancePage', () => {
  it('確定済みのミニマル構成で描画される（見出し + 案内文のみ）', () => {
    render(<MaintenancePage />);
    expect(
      screen.getByRole('heading', { name: 'ただいまメンテナンス中です' }),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/自動的に再接続を試みていますので、しばらくお待ちください。/),
    ).toBeInTheDocument();
    // ユーザ要望で削除済みの要素を復活させない（仕様ガード）
    expect(screen.queryByRole('button', { name: /再試行/ })).not.toBeInTheDocument();
    expect(screen.queryByText(/定期メンテナンス時間帯/)).not.toBeInTheDocument();
  });

  it('メンテナンス文言のコンテナに data-nosnippet が付き検索スニペットに使われない', () => {
    render(<MaintenancePage />);
    const heading = screen.getByRole('heading', { name: 'ただいまメンテナンス中です' });
    expect(heading.closest('[data-nosnippet]')).not.toBeNull();
  });
});
