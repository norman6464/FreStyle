import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ExerciseLanguageSelectPage from '../ui/ExerciseLanguageSelectPage';
import ExerciseRepository from '@/repositories/ExerciseRepository';

vi.mock('@/repositories/ExerciseRepository', () => ({
  default: {
    listLanguageSummary: vi.fn(),
  },
}));

const mockSummary = vi.mocked(ExerciseRepository.listLanguageSummary);

function renderPage() {
  return render(
    <MemoryRouter>
      <ExerciseLanguageSelectPage />
    </MemoryRouter>,
  );
}

describe('ExerciseLanguageSelectPage (FRESTYLE-152)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('言語カードを表示し、その言語の問題一覧へリンクする', async () => {
    mockSummary.mockResolvedValue([{ language: 'go', total: 10, solved: 3 }]);
    renderPage();

    await waitFor(() => expect(screen.getByRole('heading', { name: 'Go' })).toBeInTheDocument());
    const link = screen.getByRole('link', { name: /Go の問題一覧へ/ });
    expect(link).toHaveAttribute('href', '/code-editor/lang/go');
    expect(screen.getByText('3/10 問完了')).toBeInTheDocument();
    expect(screen.getByText('10 問')).toBeInTheDocument();
  });

  it('ボタン文言は進捗によらず「問題を見る」で統一する（遷移先が同じため）', async () => {
    mockSummary.mockResolvedValue([
      { language: 'php', total: 20, solved: 0 },
      { language: 'go', total: 10, solved: 4 },
      { language: 'docker', total: 5, solved: 5 },
    ]);
    renderPage();

    await waitFor(() => expect(screen.getAllByText('問題を見る')).toHaveLength(3));
    // 「続きからはじめる」は未解答の問題から再開すると読めるが、実際は一覧に戻るだけだった。
    expect(screen.queryByText('続きからはじめる')).not.toBeInTheDocument();
    expect(screen.queryByText('もう一度解く')).not.toBeInTheDocument();
    // 進捗の伝達は進捗バーと「すべて完了」バッジが担う。
    expect(screen.getByText('すべて完了')).toBeInTheDocument();
  });

  it('進捗バーが完了率を aria-valuenow で公開する', async () => {
    mockSummary.mockResolvedValue([{ language: 'go', total: 10, solved: 3 }]);
    renderPage();

    await waitFor(() => expect(screen.getByRole('progressbar')).toBeInTheDocument());
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '30');
  });

  it('EXERCISE_LANGUAGES の定義順に並べる（API の返却順に依存しない）', async () => {
    // API は言語名昇順で返すが、表示は学習の入口として見せたい定義順(php → go → ...)にする。
    mockSummary.mockResolvedValue([
      { language: 'go', total: 10, solved: 0 },
      { language: 'php', total: 20, solved: 0 },
    ]);
    renderPage();

    await waitFor(() => expect(screen.getAllByRole('heading', { level: 2 })).toHaveLength(2));
    const names = screen.getAllByRole('heading', { level: 2 }).map((h) => h.textContent);
    expect(names).toEqual(['PHP', 'Go']);
  });

  it('未定義の言語は key をそのまま表示する（教材が先に増えても壊れない）', async () => {
    mockSummary.mockResolvedValue([{ language: 'rust', total: 3, solved: 1 }]);
    renderPage();
    await waitFor(() => expect(screen.getByRole('heading', { name: 'rust' })).toBeInTheDocument());
  });

  it('0 件なら未公開メッセージを出す', async () => {
    mockSummary.mockResolvedValue([]);
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('公開されている問題がまだありません。')).toBeInTheDocument(),
    );
  });

  it('取得失敗時はエラーを表示する', async () => {
    mockSummary.mockRejectedValue(new Error('boom'));
    renderPage();
    await waitFor(() => expect(screen.getByText('演習の取得に失敗しました')).toBeInTheDocument());
  });
});
