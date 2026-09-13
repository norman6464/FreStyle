import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import HeaderRecentPagesNav from '../HeaderRecentPagesNav';
import type { KbRecentPage } from '@/entities/kb';

const hoisted = vi.hoisted(() => ({
  fetchRecentPages: vi.fn(),
}));

vi.mock('@/entities/kb', async () => {
  const actual = await vi.importActual<typeof import('@/entities/kb')>('@/entities/kb');
  return {
    ...actual,
    KbRepository: {
      fetchRecentPages: hoisted.fetchRecentPages,
    },
  };
});

function recentPage(pageId: string, title: string, spaceName = 'エンジニアリング'): KbRecentPage {
  return {
    pageId,
    workspaceSlug: 'acme',
    title,
    spaceId: 'space-1',
    spaceName,
    viewedAt: '2026-09-13T00:00:00Z',
  };
}

function renderNav() {
  return render(
    <MemoryRouter>
      <HeaderRecentPagesNav className="nav-link" />
    </MemoryRouter>,
  );
}

describe('HeaderRecentPagesNav', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('開くと取得し、最近見たページの一覧を出す', async () => {
    hoisted.fetchRecentPages.mockResolvedValue([
      recentPage('p-1', '設計メモ'),
      recentPage('p-2', '議事録'),
    ]);
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /最近見たページ/ }));

    expect(await screen.findByRole('link', { name: /設計メモ/ })).toHaveAttribute('href', '/kb/p-1');
    expect(screen.getByRole('link', { name: /議事録/ })).toHaveAttribute('href', '/kb/p-2');
    expect(hoisted.fetchRecentPages).toHaveBeenCalledTimes(1);
  });

  it('外側をクリックすると閉じる', async () => {
    hoisted.fetchRecentPages.mockResolvedValue([recentPage('p-1', '設計メモ')]);
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /最近見たページ/ }));
    await screen.findByRole('link', { name: /設計メモ/ });

    fireEvent.mouseDown(document.body);

    expect(screen.queryByRole('link', { name: /設計メモ/ })).not.toBeInTheDocument();
  });

  it('0 件なら「まだ最近見たページはありません」を出す', async () => {
    hoisted.fetchRecentPages.mockResolvedValue([]);
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /最近見たページ/ }));

    expect(await screen.findByText('まだ最近見たページはありません')).toBeInTheDocument();
  });

  it('取得に失敗しても壊れず、0 件と同じ表示になる', async () => {
    hoisted.fetchRecentPages.mockRejectedValue(new Error('boom'));
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /最近見たページ/ }));

    expect(await screen.findByText('まだ最近見たページはありません')).toBeInTheDocument();
  });
});
