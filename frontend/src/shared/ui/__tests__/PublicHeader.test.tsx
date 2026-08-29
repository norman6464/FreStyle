import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import PublicHeader from '../PublicHeader';

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <PublicHeader />
    </MemoryRouter>
  );
}

describe('PublicHeader', () => {
  it('アカウント作成への導線がある', () => {
    renderAt('/login');
    const signup = screen.getByRole('link', { name: /アカウントを作成/ });
    expect(signup).toHaveAttribute('href', '/signup');
  });
});
