import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import ExerciseDetailPage from '../ExerciseDetailPage';
import { ExerciseRepository } from '@/entities/exercise';
import type { ExerciseSubmission, MasterExercise } from '@/entities/exercise';

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

const solvedSubmission: ExerciseSubmission = {
  id: 1,
  userId: 1,
  exerciseKind: 'master',
  exerciseId: 51,
  submittedCode: '<h1>タイトル</h1><p>本文</p>',
  stdout: '',
  stderr: '',
  exitCode: 0,
  isCorrect: true,
  submittedAt: '2026-07-20T00:00:00Z',
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

  it('「できた！」の提出が失敗するとエラーを表示し、ボタンは押せるまま残る', async () => {
    mocks.submit.mockRejectedValue(new Error('server error'));
    renderPage();
    await screen.findByTitle('プレビュー');

    fireEvent.click(screen.getByRole('button', { name: /できた！/ }));

    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('提出に失敗しました');
    // 失敗時はクリア済みにならず、再提出できる
    expect(screen.queryByText('この演習はクリア済みです')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /できた！/ })).toBeEnabled();
  });

  it('hintText がある演習はヒントを開閉できる', async () => {
    mocks.getDetail.mockResolvedValue({
      exercise: { ...previewExercise, hintText: 'p タグを使ってみよう' },
      examples: [],
    });
    renderPage();
    await screen.findByTitle('プレビュー');

    // 初期状態は閉じている
    const toggle = screen.getByRole('button', { name: 'ヒントを見る' });
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(screen.queryByText('p タグを使ってみよう')).not.toBeInTheDocument();

    // 開く
    fireEvent.click(toggle);
    expect(screen.getByRole('button', { name: 'ヒントを隠す' })).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByText('p タグを使ってみよう')).toBeInTheDocument();

    // 閉じる
    fireEvent.click(screen.getByRole('button', { name: 'ヒントを隠す' }));
    expect(screen.getByRole('button', { name: 'ヒントを見る' })).toBeInTheDocument();
    expect(screen.queryByText('p タグを使ってみよう')).not.toBeInTheDocument();
  });

  it('hintText が空の演習はヒントボタンを表示しない', async () => {
    renderPage();
    await screen.findByTitle('プレビュー');

    expect(screen.queryByRole('button', { name: /ヒントを/ })).not.toBeInTheDocument();
  });

  it('提出履歴に正解があると初回からクリア済み表示になり「できた！」を出さない', async () => {
    mocks.listSubmissions.mockResolvedValue([solvedSubmission]);
    renderPage();

    expect(await screen.findByText('この演習はクリア済みです')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /できた！/ })).not.toBeInTheDocument();
    expect(mocks.submit).not.toHaveBeenCalled();
  });

  it('リセットで編集内容が starterCode に戻る', async () => {
    renderPage();
    const editor = await screen.findByLabelText('解答コードエディタ');

    fireEvent.change(editor, { target: { value: '<h1>編集した</h1>' } });
    expect(editor).toHaveValue('<h1>編集した</h1>');

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(editor).toHaveValue('<h1>タイトル</h1>');
  });

  it('starterCode が空の演習は「できた！」が無効化され、入力すると押せるようになる', async () => {
    mocks.getDetail.mockResolvedValue({
      exercise: { ...previewExercise, starterCode: '' },
      examples: [],
    });
    renderPage();
    const editor = await screen.findByLabelText('解答コードエディタ');

    expect(screen.getByRole('button', { name: /できた！/ })).toBeDisabled();

    fireEvent.change(editor, { target: { value: '<p>書いた</p>' } });
    expect(screen.getByRole('button', { name: /できた！/ })).toBeEnabled();
  });

  it('詳細の取得に失敗するとエラーメッセージを表示する', async () => {
    mocks.getDetail.mockRejectedValue(new Error('network error'));
    renderPage();

    expect(await screen.findByText('演習問題の取得に失敗しました')).toBeInTheDocument();
    expect(screen.queryByTitle('プレビュー')).not.toBeInTheDocument();
  });

  it('mode=qa の演習はプレビューではなくコマンド入力フォームに委譲する', async () => {
    mocks.getDetail.mockResolvedValue({
      exercise: { ...previewExercise, mode: 'qa' },
      examples: [],
    });
    renderPage();

    expect(await screen.findByLabelText('コマンドを入力')).toBeInTheDocument();
    expect(screen.queryByTitle('プレビュー')).not.toBeInTheDocument();
  });

  it('slug が無い場合は「演習問題が見つかりません」を表示する', async () => {
    render(
      <MemoryRouter initialEntries={['/code-editor']}>
        <Routes>
          <Route path="/code-editor" element={<ExerciseDetailPage />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByText('演習問題が見つかりません')).toBeInTheDocument();
    expect(mocks.getDetail).not.toHaveBeenCalled();
  });
});
