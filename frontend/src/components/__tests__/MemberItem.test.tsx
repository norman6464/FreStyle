import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import MemberItem from '../MemberItem';
import ChatRepository from '../../repositories/ChatRepository';

vi.mock('../../repositories/ChatRepository');

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('MemberItem', () => {
  it('名前とメールが表示される', () => {
    render(
      <MemoryRouter>
        <MemberItem id={1} name="テストユーザー" email="test@example.com" />
      </MemoryRouter>
    );

    expect(screen.getByText('テストユーザー')).toBeInTheDocument();
    expect(screen.getByText('test@example.com')).toBeInTheDocument();
  });

  it('名前の頭文字がアバターに表示される', () => {
    render(
      <MemoryRouter>
        <MemberItem id={1} name="Test" email="test@example.com" />
      </MemoryRouter>
    );

    expect(screen.getByText('T')).toBeInTheDocument();
  });

  it('roomIdがある場合「チャット」ボタンが表示される', () => {
    render(
      <MemoryRouter>
        <MemberItem id={1} name="テスト" email="test@example.com" roomId={10} />
      </MemoryRouter>
    );

    expect(screen.getByText('チャット')).toBeInTheDocument();
  });

  it('roomIdがない場合「追加」ボタンが表示される', () => {
    render(
      <MemoryRouter>
        <MemberItem id={1} name="テスト" email="test@example.com" />
      </MemoryRouter>
    );

    expect(screen.getByText('追加')).toBeInTheDocument();
  });

  it('roomIdがある場合クリックで既存ルームに遷移する', async () => {
    render(
      <MemoryRouter>
        <MemberItem id={1} name="テスト" email="test@example.com" roomId={10} />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByText('テスト'));

    expect(mockNavigate).toHaveBeenCalledWith('/chat/users/10');
  });

  it('role="button"とtabIndex={0}を持つ', () => {
    render(
      <MemoryRouter>
        <MemberItem id={1} name="テスト" email="test@example.com" />
      </MemoryRouter>
    );

    const button = screen.getByRole('button', { name: /テスト/ });
    expect(button).toHaveAttribute('tabindex', '0');
  });

  it('Enterキーでクリックと同じ動作をする', async () => {
    render(
      <MemoryRouter>
        <MemberItem id={1} name="テスト" email="test@example.com" roomId={10} />
      </MemoryRouter>
    );

    const button = screen.getByRole('button', { name: /テスト/ });
    fireEvent.keyDown(button, { key: 'Enter' });

    expect(mockNavigate).toHaveBeenCalledWith('/chat/users/10');
  });

  it('Spaceキーでクリックと同じ動作をする', async () => {
    render(
      <MemoryRouter>
        <MemberItem id={1} name="テスト" email="test@example.com" roomId={10} />
      </MemoryRouter>
    );

    const button = screen.getByRole('button', { name: /テスト/ });
    fireEvent.keyDown(button, { key: ' ' });

    expect(mockNavigate).toHaveBeenCalledWith('/chat/users/10');
  });
});
