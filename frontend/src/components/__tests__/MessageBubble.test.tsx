import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import MessageBubble from '../MessageBubble';

describe('MessageBubble', () => {
  it('テキストメッセージが表示される（受信側）', () => {
    render(<MessageBubble isSender={false} content="こんにちは" id="m1" />);
    expect(screen.getByText('こんにちは')).toBeInTheDocument();
  });

  it('送信者メッセージが表示される', () => {
    render(<MessageBubble isSender={true} content="テスト送信" id="m2" />);
    expect(screen.getByText('テスト送信')).toBeInTheDocument();
  });

  it('受信側で送信者名が表示される', () => {
    render(<MessageBubble isSender={false} content="テスト" id="m3" senderName="田中太郎" />);
    expect(screen.getByText('田中太郎')).toBeInTheDocument();
  });

  it('送信側では送信者名が表示されない', () => {
    render(<MessageBubble isSender={true} content="テスト" id="m4" senderName="田中太郎" />);
    expect(screen.queryByText('田中太郎')).not.toBeInTheDocument();
  });

  it('削除済みメッセージは「削除しました」表示になる', () => {
    render(<MessageBubble isSender={true} content="テスト" id="m5" isDeleted={true} />);
    expect(screen.getByText('メッセージを削除しました')).toBeInTheDocument();
  });

  it('コピーボタンが表示され、クリックで onCopy が呼ばれる', () => {
    const mockCopy = vi.fn();
    render(<MessageBubble isSender={false} content="コピー対象" id="m6" onCopy={mockCopy} />);
    const btn = screen.getByTitle('コピー');
    fireEvent.click(btn);
    expect(mockCopy).toHaveBeenCalledWith('m6', 'コピー対象');
  });

  it('isCopied=true で「コピー済み」アイコンになる', () => {
    render(
      <MessageBubble
        isSender={false}
        content="テスト"
        id="m7"
        onCopy={vi.fn()}
        isCopied={true}
      />
    );
    expect(screen.getByTitle('コピー済み')).toBeInTheDocument();
  });

  it('onCopy 未指定の場合コピーボタンが表示されない', () => {
    render(<MessageBubble isSender={false} content="テスト" id="m8" />);
    expect(screen.queryByTitle('コピー')).not.toBeInTheDocument();
  });

  it('削除済みメッセージにはコピーボタンが表示されない', () => {
    render(
      <MessageBubble
        isSender={true}
        content="テスト"
        id="m9"
        isDeleted={true}
        onCopy={vi.fn()}
      />
    );
    expect(screen.queryByTitle('コピー')).not.toBeInTheDocument();
  });

  it('画像メッセージが表示される', () => {
    render(
      <MessageBubble isSender={false} content="https://example.com/img.jpg" id="m10" type="image" />
    );
    const img = screen.getByAltText('画像');
    expect(img).toBeInTheDocument();
  });

  it('時刻が正しく表示される', () => {
    render(
      <MessageBubble isSender={false} content="テスト" id="m11" createdAt="2026-02-14T10:30:00" />
    );
    expect(screen.getByText('10:30')).toBeInTheDocument();
  });

  it('受信メッセージに article ロールと aria-label が設定される', () => {
    render(<MessageBubble isSender={false} content="こんにちは" id="m12" senderName="田中太郎" />);
    const article = screen.getByRole('article');
    expect(article).toHaveAttribute('aria-label', '田中太郎のメッセージ');
  });

  it('送信者メッセージにも aria-label が設定される', () => {
    render(<MessageBubble isSender={true} content="テスト" id="m13" />);
    expect(screen.getByRole('article')).toHaveAttribute('aria-label', '自分のメッセージ');
  });

  it('送信側で onDelete を渡すと削除ボタンが描画される', () => {
    const mockDelete = vi.fn();
    render(<MessageBubble isSender={true} content="テスト" id="m14" onDelete={mockDelete} />);
    expect(screen.getByLabelText('メッセージを削除')).toBeInTheDocument();
  });

  it('受信メッセージで Markdown のコードブロックが <pre> として描画される', () => {
    const md = '```js\nconst a = 1;\n```';
    const { container } = render(
      <MessageBubble isSender={false} content={md} id="m15" />
    );
    expect(container.querySelector('pre')).not.toBeNull();
  });

  it('受信メッセージで Markdown の見出しが描画される', () => {
    render(<MessageBubble isSender={false} content="# タイトル" id="m16" />);
    expect(screen.getByRole('heading', { name: 'タイトル' })).toBeInTheDocument();
  });
});
