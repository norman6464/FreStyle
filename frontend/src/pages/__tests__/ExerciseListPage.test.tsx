import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
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

// 対象言語は URL(`/code-editor/lang/:language`)が正なので、ルータ経由で描画する(FRESTYLE-152)。
function renderPage(language = 'php') {
  return render(
    <MemoryRouter initialEntries={[`/code-editor/lang/${language}`]}>
      <Routes>
        <Route path="/code-editor/lang/:language" element={<ExerciseListPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('ExerciseListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('URL の言語で一覧を取得し、その言語名を見出しに出す', async () => {
    mockListExercises.mockResolvedValue(emptyPage);
    renderPage('go');
    await waitFor(() => expect(mockListExercises).toHaveBeenCalledWith('go', 0, 20));
    expect(screen.getByRole('heading', { name: 'Go' })).toBeInTheDocument();
  });

  it('定義済みの表示名(Bash / Linux)に解決される', async () => {
    mockListExercises.mockResolvedValue(emptyPage);
    renderPage('bash');
    await waitFor(() => expect(mockListExercises).toHaveBeenCalledWith('bash', 0, 20));
    expect(screen.getByRole('heading', { name: 'Bash / Linux' })).toBeInTheDocument();
  });

  it('言語選択へ戻る導線がある', async () => {
    mockListExercises.mockResolvedValue(emptyPage);
    renderPage();
    await waitFor(() => expect(mockListExercises).toHaveBeenCalled());
    expect(screen.getByRole('link', { name: /言語を選びなおす/ })).toHaveAttribute(
      'href',
      '/code-editor',
    );
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
