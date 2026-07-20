import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import BackLink from '../BackLink';

describe('BackLink', () => {
  it('演習一覧へのリンクを表示する', () => {
    render(<BackLink />, { wrapper: MemoryRouter });
    const link = screen.getByRole('link', { name: /問題一覧に戻る/ });
    expect(link).toHaveAttribute('href', '/code-editor');
  });
});
