import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import HeaderSpacesNav from '../HeaderSpacesNav';
import { setCurrentKbSpace } from '@/entities/kb';
import type { KbMySpace } from '@/entities/kb';

const hoisted = vi.hoisted(() => ({
  fetchMySpaces: vi.fn(),
  createSpace: vi.fn(),
  showToast: vi.fn(),
}));

vi.mock('@/shared/lib/hooks/useToast', () => ({
  useToast: () => ({ showToast: hoisted.showToast, toasts: [], removeToast: vi.fn() }),
}));

vi.mock('@/entities/kb', async () => {
  const actual = await vi.importActual<typeof import('@/entities/kb')>('@/entities/kb');
  return {
    ...actual,
    KbRepository: {
      fetchMySpaces: hoisted.fetchMySpaces,
      createSpace: hoisted.createSpace,
    },
  };
});

function mySpace(id: string, name = id): KbMySpace {
  return { id, name, role: 'editor' };
}

const NAV_CLASS = 'nav-link';

function renderNav() {
  return render(
    <MemoryRouter>
      <HeaderSpacesNav className={NAV_CLASS} />
    </MemoryRouter>,
  );
}

describe('HeaderSpacesNav', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setCurrentKbSpace(null);
  });

  it('今いるスペースが分からないときは素のリンク（/kb/spaces）にする', () => {
    renderNav();
    const link = screen.getByRole('link', { name: 'スペース' });
    expect(link).toHaveAttribute('href', '/kb/spaces');
    expect(screen.queryByRole('button', { name: 'スペース' })).not.toBeInTheDocument();
  });

  it('今いるスペースが分かっていればドロップダウンのボタンにする', async () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    hoisted.fetchMySpaces.mockResolvedValue([mySpace('space-1', '開発部'), mySpace('space-2', '営業部')]);
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /スペース/ }));

    expect(await screen.findByRole('link', { name: '営業部' })).toBeInTheDocument();
    expect(hoisted.fetchMySpaces).toHaveBeenCalledWith('acme');
  });

  it('外側をクリックすると閉じる', async () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    hoisted.fetchMySpaces.mockResolvedValue([mySpace('space-1', '開発部')]);
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /スペース/ }));
    await screen.findByRole('link', { name: '開発部' });

    fireEvent.mouseDown(document.body);

    expect(screen.queryByRole('link', { name: '開発部' })).not.toBeInTheDocument();
  });

  it('今いるスペースを一覧の中で強調する', async () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    hoisted.fetchMySpaces.mockResolvedValue([mySpace('space-1', '開発部'), mySpace('space-2', '営業部')]);
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /スペース/ }));

    const current = await screen.findByRole('link', { name: '開発部' });
    expect(current.className).toContain('font-semibold');
    const other = screen.getByRole('link', { name: '営業部' });
    expect(other.className).not.toContain('font-semibold');
  });

  it('「スペースを作成」から作れる', async () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    hoisted.fetchMySpaces.mockResolvedValue([mySpace('space-1', '開発部')]);
    hoisted.createSpace.mockResolvedValue({ id: 'space-2', name: '営業部' });
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /スペース/ }));
    fireEvent.click(await screen.findByRole('button', { name: 'スペースを作成' }));
    fireEvent.change(screen.getByLabelText('スペースの名前'), { target: { value: '営業部' } });
    fireEvent.click(screen.getByRole('button', { name: 'スペースを作る' }));

    await waitFor(() => expect(hoisted.createSpace).toHaveBeenCalledWith('acme', { name: '営業部' }));
  });

  it('「プライベートスペースを作成」から visibility=private で作る', async () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    hoisted.fetchMySpaces.mockResolvedValue([mySpace('space-1', '開発部')]);
    hoisted.createSpace.mockResolvedValue({ id: 'space-3', name: '川野の下書き' });
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /スペース/ }));
    fireEvent.click(await screen.findByRole('button', { name: 'プライベートスペースを作成' }));
    fireEvent.change(screen.getByLabelText('プライベートスペースの名前'), {
      target: { value: '川野の下書き' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'プライベートスペースを作る' }));

    await waitFor(() =>
      expect(hoisted.createSpace).toHaveBeenCalledWith('acme', {
        name: '川野の下書き',
        visibility: 'private',
      }),
    );
  });

  it('作成に失敗したら知らせを出し、入力は消さない', async () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    hoisted.fetchMySpaces.mockResolvedValue([mySpace('space-1', '開発部')]);
    hoisted.createSpace.mockRejectedValue(new Error('boom'));
    renderNav();

    fireEvent.click(screen.getByRole('button', { name: /スペース/ }));
    fireEvent.click(await screen.findByRole('button', { name: 'スペースを作成' }));
    fireEvent.change(screen.getByLabelText('スペースの名前'), { target: { value: '営業部' } });
    fireEvent.click(screen.getByRole('button', { name: 'スペースを作る' }));

    await waitFor(() => expect(hoisted.showToast).toHaveBeenCalledWith('error', 'スペースを作成できませんでした'));
    expect(screen.getByLabelText('スペースの名前')).toHaveValue('営業部');
  });
});
