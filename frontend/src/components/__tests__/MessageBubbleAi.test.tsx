import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import MessageBubbleAi from '../MessageBubbleAi';

describe('MessageBubbleAi', () => {
  it('テキストメッセージが表示される', () => {
    render(<MessageBubbleAi isSender={false} content="AIの応答です" id={1} />);

    expect(screen.getByText('AIの応答です')).toBeInTheDocument();
  });

  it('送信者メッセージが表示される', () => {
    render(<MessageBubbleAi isSender={true} content="ユーザーの質問" id={1} />);

    expect(screen.getByText('ユーザーの質問')).toBeInTheDocument();
  });

  it('削除済みメッセージが表示される', () => {
    render(<MessageBubbleAi isSender={false} content="テスト" id={1} isDeleted={true} />);

    expect(screen.getByText('メッセージを削除しました')).toBeInTheDocument();
  });

  it('画像タイプでimgが表示される', () => {
    render(<MessageBubbleAi isSender={false} type="image" content="https://example.com/image.png" id={1} />);

    expect(screen.getByAltText('画像')).toBeInTheDocument();
  });

  it('送信者と受信者で異なるスタイルが適用される', () => {
    const { container: senderContainer } = render(
      <MessageBubbleAi isSender={true} content="送信" id={1} />
    );
    const { container: receiverContainer } = render(
      <MessageBubbleAi isSender={false} content="受信" id={2} />
    );
    const senderBubble = senderContainer.firstElementChild as HTMLElement;
    const receiverBubble = receiverContainer.firstElementChild as HTMLElement;
    expect(senderBubble.className).not.toBe(receiverBubble.className);
  });

  it('長いメッセージも表示される', () => {
    const longMessage = 'テスト'.repeat(100);
    render(<MessageBubbleAi isSender={false} content={longMessage} id={1} />);
    expect(screen.getByText(longMessage)).toBeInTheDocument();
  });
});
