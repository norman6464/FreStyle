import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import CodeEditorPage from '../CodeEditorPage';
import type { PhpExercise, CodeExecutionResult } from '../../types';

// Monaco Editor は jsdom で動かないためスタブ化
vi.mock('../../components/CodeEditor', () => ({
  default: ({ value, onChange }: { value: string; onChange: (v: string) => void }) => (
    <textarea
      data-testid="code-editor"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));

vi.mock('../../hooks/useTheme', () => ({
  useTheme: () => ({ theme: 'dark', toggleTheme: vi.fn() }),
}));

const mockRunCode = vi.fn();
const mockSelectExercise = vi.fn();
const mockSetShowHint = vi.fn();
const mockResetCode = vi.fn();

const baseExercise: PhpExercise = {
  id: 1,
  orderIndex: 1,
  category: '基礎',
  title: 'こんにちは世界',
  description: 'Hello World を表示しましょう',
  starterCode: '<?php echo "Hello";',
  hintText: 'echo を使います',
  expectedOutput: 'Hello',
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

function makeHook(overrides: Partial<ReturnType<typeof import('../../hooks/usePhpEditor').usePhpEditor>> = {}) {
  return {
    exercises: [baseExercise],
    categories: ['基礎'],
    selectedExercise: baseExercise,
    code: baseExercise.starterCode,
    setCode: vi.fn(),
    result: null as CodeExecutionResult | null,
    running: false,
    showHint: false,
    setShowHint: mockSetShowHint,
    loadingExercises: false,
    error: null as string | null,
    selectExercise: mockSelectExercise,
    runCode: mockRunCode,
    resetCode: mockResetCode,
    ...overrides,
  };
}

vi.mock('../../hooks/usePhpEditor', () => ({
  usePhpEditor: vi.fn(),
}));

import { usePhpEditor } from '../../hooks/usePhpEditor';

function renderPage() {
  return render(
    <MemoryRouter>
      <CodeEditorPage />
    </MemoryRouter>
  );
}

describe('CodeEditorPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(usePhpEditor).mockReturnValue(makeHook());
  });

  it('演習タイトルと説明が表示される', () => {
    renderPage();
    expect(screen.getAllByText('こんにちは世界').length).toBeGreaterThan(0);
    expect(screen.getByText('Hello World を表示しましょう')).toBeInTheDocument();
    expect(screen.getAllByText('基礎').length).toBeGreaterThan(0);
  });

  it('期待される出力が表示される', () => {
    renderPage();
    expect(screen.getByText('期待される出力')).toBeInTheDocument();
    expect(screen.getAllByText('Hello').length).toBeGreaterThan(0);
  });

  it('実行ボタンをクリックすると runCode が呼ばれる', () => {
    renderPage();
    fireEvent.click(screen.getByText('▶ 実行'));
    expect(mockRunCode).toHaveBeenCalledTimes(1);
  });

  it('ヒントボタンをクリックすると setShowHint が呼ばれる', () => {
    renderPage();
    fireEvent.click(screen.getByText('ヒントを見る'));
    expect(mockSetShowHint).toHaveBeenCalledWith(true);
  });

  it('showHint=true のときヒント文が表示される', () => {
    vi.mocked(usePhpEditor).mockReturnValue(makeHook({ showHint: true }));
    renderPage();
    expect(screen.getByText(/echo を使います/)).toBeInTheDocument();
  });

  it('running=true のとき「実行中...」と表示される', () => {
    vi.mocked(usePhpEditor).mockReturnValue(makeHook({ running: true }));
    renderPage();
    expect(screen.getByText('実行中...')).toBeInTheDocument();
  });

  it('result に stdout がある場合出力パネルに表示される', () => {
    vi.mocked(usePhpEditor).mockReturnValue(
      makeHook({ result: { stdout: 'Hello, World!', stderr: '', exitCode: 0 } })
    );
    renderPage();
    expect(screen.getAllByText(/Hello, World!/).length).toBeGreaterThan(0);
    expect(screen.getByText('✓ 成功')).toBeInTheDocument();
  });

  it('result にエラーがある場合エラーパネルに表示される', () => {
    vi.mocked(usePhpEditor).mockReturnValue(
      makeHook({ result: { stdout: '', stderr: 'Parse error', exitCode: 255 } })
    );
    renderPage();
    expect(screen.getByText('Parse error')).toBeInTheDocument();
    expect(screen.getByText(/✗ エラー/)).toBeInTheDocument();
  });

  it('ローディング中は「読み込み中...」と表示される', () => {
    vi.mocked(usePhpEditor).mockReturnValue(makeHook({ loadingExercises: true }));
    renderPage();
    expect(screen.getByText('読み込み中...')).toBeInTheDocument();
  });

  it('リセットボタンをクリックすると resetCode が呼ばれる', () => {
    renderPage();
    fireEvent.click(screen.getByText('リセット'));
    expect(mockResetCode).toHaveBeenCalledTimes(1);
  });
});
