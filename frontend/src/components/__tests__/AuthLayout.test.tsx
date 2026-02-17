import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import AuthLayout from '../AuthLayout';

describe('AuthLayout', () => {
  it('タイトルが表示される', () => {
    render(<AuthLayout title="テストタイトル"><div>テスト</div></AuthLayout>);

    expect(screen.getByText('テストタイトル')).toBeInTheDocument();
  });

  it('子要素が表示される', () => {
    render(<AuthLayout><p>子コンテンツ</p></AuthLayout>);

    expect(screen.getByText('子コンテンツ')).toBeInTheDocument();
  });

  it('複数の子要素が表示される', () => {
    render(
      <AuthLayout>
        <p>要素1</p>
        <p>要素2</p>
      </AuthLayout>
    );

    expect(screen.getByText('要素1')).toBeInTheDocument();
    expect(screen.getByText('要素2')).toBeInTheDocument();
  });

  it('中央寄せレイアウトが適用される', () => {
    const { container } = render(<AuthLayout><div>テスト</div></AuthLayout>);
    const wrapper = container.firstElementChild as HTMLElement;
    expect(wrapper.className).toContain('flex');
    expect(wrapper.className).toContain('items-center');
    expect(wrapper.className).toContain('justify-center');
  });

  it('カードに角丸が適用される', () => {
    const { container } = render(<AuthLayout><div>テスト</div></AuthLayout>);
    const card = container.querySelector('.rounded-xl');
    expect(card).toBeTruthy();
  });

  it('フッターが表示される', () => {
    render(
      <MemoryRouter>
        <AuthLayout footer={<p>フッター内容</p>}><div>テスト</div></AuthLayout>
      </MemoryRouter>
    );

    expect(screen.getByText('フッター内容')).toBeInTheDocument();
  });

  it('アイコンが表示される', () => {
    const { container } = render(<AuthLayout><div>テスト</div></AuthLayout>);
    const svg = container.querySelector('svg');
    expect(svg).toBeTruthy();
  });
});
