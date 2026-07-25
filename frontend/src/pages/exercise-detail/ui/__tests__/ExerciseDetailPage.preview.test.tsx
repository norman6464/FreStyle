import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import ExerciseDetailPage from '../ExerciseDetailPage';
import { ExerciseRepository } from '@/entities/exercise';
import type { MasterExercise } from '@/entities/exercise';

vi.mock('@/entities/exercise/api/exerciseRepository', () => ({
  default: {
    getDetail: vi.fn(),
    execute: vi.fn(),
    warmup: vi.fn(),
    submit: vi.fn(),
    listSubmissions: vi.fn(),
  },
}));

// MarkdownView は remark 一式を読み込むため、本文が渡ることだけ確認する
vi.mock('@/shared/ui/MarkdownView', () => ({
  default: ({ content }: { content: string }) => <div data-testid="markdown">{content}</div>,
}));

// Monaco は jsdom で動かないため、value/onChange だけ透過する textarea に差し替える
vi.mock('@/shared/ui/CodeEditor', () => ({
  default: ({ value, onChange }: { value: string; onChange: (v: string) => void }) => (
    <textarea aria-label="解答コードエディタ" value={value} onChange={(e) => onChange(e.target.value)} />
  ),
}));

const mocks = vi.mocked(ExerciseRepository);

const previewExercise: MasterExercise = {
  id: 51,
  slug: 'html-1',
  language: 'html',
  orderIndex: 1,
  category: '基礎',
  title: '見出しと段落',
  description: 'h1 見出しと p 段落を作ってみよう',
  starterCode: '<h1>タイトル</h1>',
  hintText: '',
  expectedOutput: '<h1>タイトル</h1><p>本文</p>',
  mode: 'preview',
  explanation: '',
  difficulty: 1,
  isPublished: true,
  createdAt: '',
  updatedAt: '',
};

const renderPage = () =>
  render(
    <MemoryRouter initialEntries={['/code-editor/html-1']}>
      <Routes>
        <Route path="/code-editor/:slug" element={<ExerciseDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );

describe('ExerciseDetailPage (mode=preview)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.getDetail.mockResolvedValue({ exercise: previewExercise, examples: [] });
    mocks.listSubmissions.mockResolvedValue([]);
    mocks.warmup.mockResolvedValue(undefined);
  });

  it('プレビュー iframe と見本 iframe を srcdoc + 空 sandbox 付きで表示する', async () => {
    renderPage();

    const preview = await screen.findByTitle('プレビュー');
    expect(preview).toHaveAttribute('srcdoc', '<h1>タイトル</h1>');
    // sandbox="" (スクリプト実行・同一オリジンとも不許可) はセキュリティ要件
    expect(preview).toHaveAttribute('sandbox', '');

    const sample = screen.getByTitle('見本');
    expect(sample).toHaveAttribute('srcdoc', '<h1>タイトル</h1><p>本文</p>');
    expect(sample).toHaveAttribute('sandbox', '');
  });

  it('エディタ入力でプレビューの srcdoc が更新される（見本は変わらない）', async () => {
    renderPage();
    const editor = await screen.findByLabelText('解答コードエディタ');

    fireEvent.change(editor, { target: { value: '<h1>変更後</h1>' } });

    // 反映は 300ms デバウンス後
    await waitFor(() =>
      expect(screen.getByTitle('プレビュー')).toHaveAttribute('srcdoc', '<h1>変更後</h1>'),
    );
    expect(screen.getByTitle('見本')).toHaveAttribute('srcdoc', '<h1>タイトル</h1><p>本文</p>');
  });

  it('「できた！」で submit API が現在のコードで呼ばれ、成功後はクリア済み表示になる', async () => {
    mocks.submit.mockResolvedValue({ submissionId: 9, isCorrect: true, results: [] });
    renderPage();
    const editor = await screen.findByLabelText('解答コードエディタ');

    fireEvent.change(editor, { target: { value: '<h1>完成</h1>' } });
    fireEvent.click(screen.getByRole('button', { name: /できた！/ }));

    await waitFor(() => expect(mocks.submit).toHaveBeenCalledWith('html-1', '<h1>完成</h1>'));
    expect(await screen.findByText('この演習はクリア済みです')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /できた！/ })).not.toBeInTheDocument();
  });

  it('実行ボタン・実行結果・テストケース表示を出さない', async () => {
    renderPage();
    await screen.findByTitle('プレビュー');

    expect(screen.queryByRole('button', { name: /コード実行/ })).not.toBeInTheDocument();
    expect(screen.queryByText(/複数のテストケースで採点します/)).not.toBeInTheDocument();
    expect(mocks.execute).not.toHaveBeenCalled();
  });
});
