import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import NoteEditor from '../NoteEditor';

const defaultProps = {
  title: 'テストノート',
  content: 'テスト内容',
  onTitleChange: vi.fn(),
  onContentChange: vi.fn(),
};

describe('NoteEditor', () => {
  it('タイトル入力欄を表示する', () => {
    render(<NoteEditor {...defaultProps} />);
    const input = screen.getByDisplayValue('テストノート');
    expect(input).toBeInTheDocument();
  });

  it('内容入力欄を表示する', () => {
    render(<NoteEditor {...defaultProps} />);
    const textarea = screen.getByDisplayValue('テスト内容');
    expect(textarea).toBeInTheDocument();
  });

  it('タイトル変更でonTitleChangeが呼ばれる', () => {
    render(<NoteEditor {...defaultProps} />);
    const input = screen.getByDisplayValue('テストノート');
    fireEvent.change(input, { target: { value: '更新タイトル' } });
    expect(defaultProps.onTitleChange).toHaveBeenCalledWith('更新タイトル');
  });

  it('内容変更でonContentChangeが呼ばれる', () => {
    render(<NoteEditor {...defaultProps} />);
    const textarea = screen.getByDisplayValue('テスト内容');
    fireEvent.change(textarea, { target: { value: '更新内容' } });
    expect(defaultProps.onContentChange).toHaveBeenCalledWith('更新内容');
  });

  it('タイトルのプレースホルダーが表示される', () => {
    render(<NoteEditor {...defaultProps} title="" />);
    expect(screen.getByPlaceholderText('無題')).toBeInTheDocument();
  });

  it('内容のプレースホルダーが表示される', () => {
    render(<NoteEditor {...defaultProps} content="" />);
    expect(screen.getByPlaceholderText('ここに入力...')).toBeInTheDocument();
  });

  it('文字数カウントが表示される', () => {
    render(<NoteEditor {...defaultProps} content="テスト内容" />);
    expect(screen.getByText('5文字')).toBeInTheDocument();
  });

  it('読了時間が表示される', () => {
    render(<NoteEditor {...defaultProps} content="テスト内容" />);
    expect(screen.getByText(/約\d+分/)).toBeInTheDocument();
  });

  it('空の内容で0文字と表示される', () => {
    render(<NoteEditor {...defaultProps} content="" />);
    expect(screen.getByText('0文字')).toBeInTheDocument();
  });

  it('空の内容で読了時間が表示されない', () => {
    render(<NoteEditor {...defaultProps} content="" />);
    expect(screen.queryByText(/約\d+分/)).not.toBeInTheDocument();
  });

  it('長い内容で正しい文字数を表示する', () => {
    const longContent = 'あ'.repeat(1000);
    render(<NoteEditor {...defaultProps} content={longContent} />);
    expect(screen.getByText('1000文字')).toBeInTheDocument();
  });

  it('スペースのみの内容で0文字を表示する', () => {
    render(<NoteEditor {...defaultProps} content="   " />);
    expect(screen.getByText('0文字')).toBeInTheDocument();
  });

  it('編集タブとプレビュータブが表示される', () => {
    render(<NoteEditor {...defaultProps} />);
    expect(screen.getByRole('tab', { name: '編集' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'プレビュー' })).toBeInTheDocument();
  });

  it('初期状態で編集タブがアクティブ', () => {
    render(<NoteEditor {...defaultProps} />);
    expect(screen.getByRole('tab', { name: '編集' })).toHaveAttribute('aria-selected', 'true');
    expect(screen.getByRole('tab', { name: 'プレビュー' })).toHaveAttribute('aria-selected', 'false');
  });

  it('プレビュータブクリックでマークダウンが表示される', () => {
    render(<NoteEditor {...defaultProps} content={"- りんご\n- みかん"} />);
    fireEvent.click(screen.getByRole('tab', { name: 'プレビュー' }));
    const items = screen.getAllByRole('listitem');
    expect(items).toHaveLength(2);
  });

  it('プレビュータブクリック後に編集タブがアクティブでなくなる', () => {
    render(<NoteEditor {...defaultProps} />);
    fireEvent.click(screen.getByRole('tab', { name: 'プレビュー' }));
    expect(screen.getByRole('tab', { name: '編集' })).toHaveAttribute('aria-selected', 'false');
    expect(screen.getByRole('tab', { name: 'プレビュー' })).toHaveAttribute('aria-selected', 'true');
  });

  it('編集タブクリックでテキストエリアに戻る', () => {
    render(<NoteEditor {...defaultProps} content={"- りんご"} />);
    fireEvent.click(screen.getByRole('tab', { name: 'プレビュー' }));
    fireEvent.click(screen.getByRole('tab', { name: '編集' }));
    expect(screen.getByLabelText('ノートの内容')).toBeInTheDocument();
  });

  it('アクティブタブのtabIndexが0、非アクティブが-1', () => {
    render(<NoteEditor {...defaultProps} />);
    expect(screen.getByRole('tab', { name: '編集' })).toHaveAttribute('tabIndex', '0');
    expect(screen.getByRole('tab', { name: 'プレビュー' })).toHaveAttribute('tabIndex', '-1');
  });

  it('tabpanelが存在しaria-labelledbyで紐付けされる', () => {
    render(<NoteEditor {...defaultProps} />);
    const panel = screen.getByRole('tabpanel');
    expect(panel).toBeInTheDocument();
    expect(panel).toHaveAttribute('aria-labelledby');
  });

  it('ArrowRightキーでプレビュータブに移動する', () => {
    render(<NoteEditor {...defaultProps} />);
    const editTab = screen.getByRole('tab', { name: '編集' });
    fireEvent.keyDown(editTab, { key: 'ArrowRight' });
    expect(screen.getByRole('tab', { name: 'プレビュー' })).toHaveAttribute('aria-selected', 'true');
  });

  it('ArrowLeftキーで編集タブに戻る', () => {
    render(<NoteEditor {...defaultProps} />);
    fireEvent.click(screen.getByRole('tab', { name: 'プレビュー' }));
    const previewTab = screen.getByRole('tab', { name: 'プレビュー' });
    fireEvent.keyDown(previewTab, { key: 'ArrowLeft' });
    expect(screen.getByRole('tab', { name: '編集' })).toHaveAttribute('aria-selected', 'true');
  });
});
