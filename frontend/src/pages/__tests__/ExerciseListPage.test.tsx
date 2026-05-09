import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ExerciseListPage from '../ExerciseListPage';
import ExerciseRepository from '../../repositories/ExerciseRepository';

vi.mock('../../repositories/ExerciseRepository', () => ({
  default: {
    listExercises: vi.fn(),
  },
}));

const mockListExercises = vi.mocked(ExerciseRepository.listExercises);

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
    mockListExercises.mockResolvedValue([]);
    renderPage();
    await waitFor(() => expect(screen.getByRole('heading', { name: 'コード学習' })).toBeInTheDocument());
  });

  it('ロード中は読み込み中メッセージを出す', () => {
    mockListExercises.mockReturnValue(new Promise(() => {}));
    renderPage();
    expect(screen.getByText(/読み込み中/)).toBeInTheDocument();
  });

  it('空配列なら「該当する問題がありません」を表示する', async () => {
    mockListExercises.mockResolvedValue([]);
    renderPage();
    await waitFor(() => expect(screen.getByText('該当する問題がありません。')).toBeInTheDocument());
  });

  it('問題カードが status バッジと統計を表示する', async () => {
    mockListExercises.mockResolvedValue([
      {
        id: 1, slug: 'php-1', language: 'php', orderIndex: 1, category: '基礎',
        title: 'Hello World', description: '出力する', starterCode: '', hintText: '',
        expectedOutput: 'Hello', difficulty: 1, isPublished: true,
        createdAt: '', updatedAt: '',
        status: 'solved' as const,
        stats: { totalSubmissions: 12, solvedUsers: 10 },
      },
    ]);
    renderPage();
    await waitFor(() => expect(screen.getByText('Hello World')).toBeInTheDocument());
    expect(screen.getByText('解いた')).toBeInTheDocument();
    expect(screen.getByText(/提出 12/)).toBeInTheDocument();
    expect(screen.getByText(/正答ユーザ 10/)).toBeInTheDocument();
  });

  it('未着手のカードには未着手バッジが付く', async () => {
    mockListExercises.mockResolvedValue([
      {
        id: 2, slug: 'php-2', language: 'php', orderIndex: 2, category: '基礎',
        title: '変数', description: '変数を使う', starterCode: '', hintText: '',
        expectedOutput: '1', difficulty: 1, isPublished: true,
        createdAt: '', updatedAt: '',
        status: '' as const,
        stats: { totalSubmissions: 0, solvedUsers: 0 },
      },
    ]);
    renderPage();
    await waitFor(() => expect(screen.getByText('変数')).toBeInTheDocument());
    expect(screen.getByText('未着手')).toBeInTheDocument();
  });

  it('カードクリックで /code-editor/:slug へのリンクを描画する', async () => {
    mockListExercises.mockResolvedValue([
      {
        id: 3, slug: 'php-3', language: 'php', orderIndex: 3, category: '基礎',
        title: '配列', description: '', starterCode: '', hintText: '',
        expectedOutput: '', difficulty: 2, isPublished: true,
        createdAt: '', updatedAt: '',
        status: 'in_progress' as const,
        stats: { totalSubmissions: 5, solvedUsers: 2 },
      },
    ]);
    renderPage();
    await waitFor(() => expect(screen.getByText('配列')).toBeInTheDocument());
    const link = screen.getByRole('link', { name: /配列/ });
    expect(link).toHaveAttribute('href', '/code-editor/php-3');
  });
});
