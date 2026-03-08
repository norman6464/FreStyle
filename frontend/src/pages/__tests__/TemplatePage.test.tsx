import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import TemplatePage from '../TemplatePage';
import { useTemplates } from '../../hooks/useTemplates';

vi.mock('../../hooks/useTemplates');
const mockedUseTemplates = vi.mocked(useTemplates);

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

function renderPage() {
  return render(<MemoryRouter><TemplatePage /></MemoryRouter>);
}

describe('TemplatePage', () => {
  const mockTemplates = [
    { id: 1, title: '会議の進行', description: 'ファシリテーション', category: 'meeting', openingMessage: 'msg', difficulty: 'beginner' },
    { id: 2, title: 'プレゼン質疑', description: '質問対応', category: 'presentation', openingMessage: 'msg2', difficulty: 'intermediate' },
  ];
  const categories = [
    { key: '', label: 'すべて' },
    { key: 'meeting', label: '会議' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockedUseTemplates.mockReturnValue({
      templates: mockTemplates,
      category: '',
      categories,
      changeCategory: vi.fn(),
      loading: false,
      error: null,
    });
  });

  it('タイトルを表示する', () => {
    renderPage();
    expect(screen.getByText('会話テンプレート')).toBeInTheDocument();
  });

  it('テンプレートカードを表示する', () => {
    renderPage();
    expect(screen.getByText('会議の進行')).toBeInTheDocument();
    expect(screen.getByText('プレゼン質疑')).toBeInTheDocument();
  });

  it('ローディング中はローディング表示する', () => {
    mockedUseTemplates.mockReturnValue({ templates: [], category: '', categories, changeCategory: vi.fn(), loading: true, error: null });
    renderPage();
    expect(screen.getByText('テンプレートを読み込み中...')).toBeInTheDocument();
  });

  it('練習開始ボタンクリックでAIチャットに遷移する', () => {
    renderPage();
    fireEvent.click(screen.getAllByText('このテンプレートで練習する')[0]);
    expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai', {
      state: { templateTitle: '会議の進行', templateMessage: 'msg' },
    });
  });

  it('エラー時にエラーメッセージを表示する', () => {
    mockedUseTemplates.mockReturnValue({ templates: [], category: '', categories, changeCategory: vi.fn(), loading: false, error: 'テンプレートの取得に失敗しました' });
    renderPage();
    expect(screen.getByText('テンプレートの取得に失敗しました')).toBeInTheDocument();
  });
});
