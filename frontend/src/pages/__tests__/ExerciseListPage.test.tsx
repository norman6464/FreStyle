import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ExerciseListPage from '../ExerciseListPage';
import ExerciseRepository from '../../repositories/ExerciseRepository';
import { ExercisePage } from '../../types';

vi.mock('../../repositories/ExerciseRepository', () => ({
  default: {
    listExercises: vi.fn(),
  },
}));

const mockListExercises = vi.mocked(ExerciseRepository.listExercises);

const emptyPage: ExercisePage = { items: [], hasNext: false, offset: 0, limit: 20 };

function renderPage() {
  return render(
    <MemoryRouter>
      <ExerciseListPage />
    </MemoryRouter>,
  );
}

describe('ExerciseListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ヘッダーに「コード学習」が表示される', async () => {
    mockListExercises.mockResolvedValue(emptyPage);
    renderPage();
    await waitFor(() => expect(screen.getByRole('heading', { name: 'コード学習' })).toBeInTheDocument());
  });

  it('ロード中は読み込み中メッセージを出す', () => {
    mockListExercises.mockReturnValue(new Promise(() => {}));
    renderPage();
    expect(screen.getByText(/読み込み中/)).toBeInTheDocument();
  });

  it('空配列なら「該当する問題がありません」を表示する', async () => {
    mockListExercises.mockResolvedValue(emptyPage);
    renderPage();
    await waitFor(() => expect(screen.getByText('該当する問題がありません。')).toBeInTheDocument());
  });

  it('問題カードが status バッジと統計を表示する', async () => {
    mockListExercises.mockResolvedValue({
      items: [{
        id: 1, slug: 'php-1', language: 'php', orderIndex: 1, category: '基礎',
        title: 'Hello World', difficulty: 1, mode: 'execute', isPublished: true,
        status: 'solved' as const,
        stats: { totalSubmissions: 12, solvedUsers: 10 },
      }],
      hasNext: false,
      offset: 0,
      limit: 20,
    });
    renderPage();
    await waitFor(() => expect(screen.getByText('Hello World')).toBeInTheDocument());
    expect(screen.getByText('解いた')).toBeInTheDocument();
    expect(screen.getByText(/提出 12/)).toBeInTheDocument();
    expect(screen.getByText(/正答ユーザ 10/)).toBeInTheDocument();
  });

  it('未着手のカードには状態バッジを表示しない', async () => {
    mockListExercises.mockResolvedValue({
      items: [{
        id: 2, slug: 'php-2', language: 'php', orderIndex: 2, category: '基礎',
        title: '変数', difficulty: 1, mode: 'execute', isPublished: true,
        status: '' as const,
        stats: { totalSubmissions: 0, solvedUsers: 0 },
      }],
      hasNext: false,
      offset: 0,
      limit: 20,
    });
    renderPage();
    await waitFor(() => expect(screen.getByText('変数')).toBeInTheDocument());
    // 未着手はデフォルト状態なのでバッジを出さない(視覚ノイズ削減 / FRESTYLE-64)。
    expect(screen.queryByText('未着手')).not.toBeInTheDocument();
  });

  it('言語チップが常時表示され、クリックで language 付き再取得になる (FRESTYLE-101)', async () => {
    mockListExercises.mockResolvedValue(emptyPage);
    renderPage();
    // 初期言語は localStorage 復元(既定 php)なので PHP チップがアクティブ。
    await waitFor(() => expect(mockListExercises).toHaveBeenCalledWith('php', 0, 20));
    const group = screen.getByRole('group', { name: '言語で絞り込み' });
    expect(group).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'PHP' })).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(screen.getByRole('button', { name: 'Go' }));
    await waitFor(() => expect(mockListExercises).toHaveBeenCalledWith('go', 0, 20));
    expect(screen.getByRole('button', { name: 'Go' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'PHP' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('「すべて」チップで全言語(クエリなし)の再取得になる', async () => {
    mockListExercises.mockResolvedValue(emptyPage);
    renderPage();
    await waitFor(() => expect(mockListExercises).toHaveBeenCalled());

    fireEvent.click(screen.getByRole('button', { name: 'すべて' }));
    await waitFor(() => expect(mockListExercises).toHaveBeenCalledWith(undefined, 0, 20));
    expect(screen.getByRole('button', { name: 'すべて' })).toHaveAttribute('aria-pressed', 'true');
  });

  it('アクティブな言語チップの再クリックで「すべて」に戻る(コース一覧と同じトグル操作)', async () => {
    mockListExercises.mockResolvedValue(emptyPage);
    renderPage();
    await waitFor(() => expect(mockListExercises).toHaveBeenCalled());

    // Docker を選んでから再クリック → すべて('')へ。
    fireEvent.click(screen.getByRole('button', { name: 'Docker' }));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'Docker' })).toHaveAttribute('aria-pressed', 'true'),
    );
    fireEvent.click(screen.getByRole('button', { name: 'Docker' }));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'すべて' })).toHaveAttribute('aria-pressed', 'true'),
    );
    expect(screen.getByRole('button', { name: 'Docker' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('カードクリックで /code-editor/:slug へのリンクを描画する', async () => {
    mockListExercises.mockResolvedValue({
      items: [{
        id: 3, slug: 'php-3', language: 'php', orderIndex: 3, category: '基礎',
        title: '配列', difficulty: 2, mode: 'execute', isPublished: true,
        status: 'in_progress' as const,
        stats: { totalSubmissions: 5, solvedUsers: 2 },
      }],
      hasNext: false,
      offset: 0,
      limit: 20,
    });
    renderPage();
    await waitFor(() => expect(screen.getByText('配列')).toBeInTheDocument());
    const link = screen.getByRole('link', { name: /配列/ });
    expect(link).toHaveAttribute('href', '/code-editor/php-3');
  });
});
