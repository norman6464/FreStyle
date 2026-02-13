import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import MemberList from '../MemberList';

vi.mock('../../repositories/ChatRepository');

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('MemberList', () => {
  it('メンバー一覧が表示される', () => {
    const users = [
      { id: 1, name: 'ユーザー1', email: 'user1@example.com' },
      { id: 2, name: 'ユーザー2', email: 'user2@example.com', roomId: 5 },
    ];

    render(
      <MemoryRouter>
        <MemberList users={users} />
      </MemoryRouter>
    );

    expect(screen.getByText('ユーザー1')).toBeInTheDocument();
    expect(screen.getByText('ユーザー2')).toBeInTheDocument();
  });

  it('空のリストでは何も表示されない', () => {
    const { container } = render(
      <MemoryRouter>
        <MemberList users={[]} />
      </MemoryRouter>
    );

    expect(container.querySelector('.space-y-4')?.children).toHaveLength(0);
  });

  it('メールアドレスが表示される', () => {
    const users = [
      { id: 1, name: 'テスト', email: 'test@example.com' },
    ];

    render(
      <MemoryRouter>
        <MemberList users={users} />
      </MemoryRouter>
    );

    expect(screen.getByText('test@example.com')).toBeInTheDocument();
  });

  it('roomIdがあるメンバーには「チャット」が表示される', () => {
    const users = [
      { id: 1, name: 'テスト', email: 'test@example.com', roomId: 10 },
    ];

    render(
      <MemoryRouter>
        <MemberList users={users} />
      </MemoryRouter>
    );

    expect(screen.getByText('チャット')).toBeInTheDocument();
  });

  it('roomIdがないメンバーには「追加」が表示される', () => {
    const users = [
      { id: 1, name: 'テスト', email: 'test@example.com' },
    ];

    render(
      <MemoryRouter>
        <MemberList users={users} />
      </MemoryRouter>
    );

    expect(screen.getByText('追加')).toBeInTheDocument();
  });
});
