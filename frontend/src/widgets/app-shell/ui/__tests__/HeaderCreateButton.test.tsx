import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import HeaderCreateButton from '../HeaderCreateButton';
import { setCurrentKbSpace } from '@/entities/kb';
import type { KbPage } from '@/entities/kb';

const hoisted = vi.hoisted(() => ({
  createPage: vi.fn(),
  showToast: vi.fn(),
  emitKbTreeEvent: vi.fn(),
}));

vi.mock('@/shared/lib/hooks/useToast', () => ({
  useToast: () => ({ showToast: hoisted.showToast, toasts: [], removeToast: vi.fn() }),
}));

vi.mock('@/entities/kb', async () => {
  const actual = await vi.importActual<typeof import('@/entities/kb')>('@/entities/kb');
  return {
    ...actual,
    KbRepository: { createPage: hoisted.createPage },
    emitKbTreeEvent: hoisted.emitKbTreeEvent,
  };
});

function page(id: string, title: string): KbPage {
  return {
    id,
    spaceId: 'space-1',
    title,
    createdByUserId: 1,
    createdAt: '2026-09-13T00:00:00Z',
    updatedAt: '2026-09-13T00:00:00Z',
  };
}

function renderButton() {
  return render(
    <MemoryRouter>
      <HeaderCreateButton className="create-btn" />
    </MemoryRouter>,
  );
}

describe('HeaderCreateButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setCurrentKbSpace(null);
  });

  it('今いるスペースが分からなければ何も出さない（ナレッジ以外の画面）', () => {
    const { container } = renderButton();
    expect(container).toBeEmptyDOMElement();
  });

  it('今いるスペースが分かっていれば「作成」ボタンを出す', () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    renderButton();
    expect(screen.getByRole('button', { name: /作成/ })).toBeInTheDocument();
  });

  it('押すと今いるスペース直下にページを作り、木へ通知してから開く', async () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    hoisted.createPage.mockResolvedValue(page('p-1', '無題'));
    renderButton();

    fireEvent.click(screen.getByRole('button', { name: /作成/ }));

    await waitFor(() =>
      expect(hoisted.createPage).toHaveBeenCalledWith('acme', 'space-1', { title: '無題' }),
    );
    expect(hoisted.emitKbTreeEvent).toHaveBeenCalledWith({ type: 'page-created', page: page('p-1', '無題') });
  });

  it('失敗したら知らせを出す', async () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    hoisted.createPage.mockRejectedValue(new Error('boom'));
    renderButton();

    fireEvent.click(screen.getByRole('button', { name: /作成/ }));

    await waitFor(() => expect(hoisted.showToast).toHaveBeenCalledWith('error', 'ページを作成できませんでした'));
  });
});
