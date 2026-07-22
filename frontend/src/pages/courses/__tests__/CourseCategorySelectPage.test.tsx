import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import CourseCategorySelectPage from '../ui/CourseCategorySelectPage';
import { ToastProvider } from '@/app/providers/ToastProvider';
import authReducer from '@/entities/user/model/authSlice';
import { CourseRepository } from '@/entities/course';
import type { CourseWithProgress } from '@/entities/course';

vi.mock('@/entities/course/api/courseRepository', () => ({
  default: { list: vi.fn(), create: vi.fn(), update: vi.fn(), remove: vi.fn() },
}));

const mockList = vi.mocked(CourseRepository.list);

function makeCourse(overrides: Partial<CourseWithProgress> = {}): CourseWithProgress {
  return {
    id: 1,
    companyId: 10,
    createdByUserId: 1,
    title: 'コース',
    description: '',
    category: 'database',
    language: '',
    sortOrder: 100,
    isPublished: true,
    materialCount: 0,
    completedCount: 0,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
    ...overrides,
  };
}

function renderPage(role = 'trainee') {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { role } as never },
  });
  return render(
    <Provider store={store}>
      <ToastProvider>
        <MemoryRouter>
          <CourseCategorySelectPage />
        </MemoryRouter>
      </ToastProvider>
    </Provider>,
  );
}

describe('CourseCategorySelectPage (FRESTYLE-177)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('コースがある領域だけカードを出し、コース数とリンクを表示する', async () => {
    mockList.mockResolvedValue([
      makeCourse({ id: 1, category: 'database' }),
      makeCourse({ id: 2, category: 'database' }),
      makeCourse({ id: 3, category: 'infra' }),
    ]);
    renderPage('trainee');

    const dbLink = await screen.findByRole('link', { name: /データベース のコース一覧へ/ });
    expect(dbLink).toHaveAttribute('href', '/courses/category/database');
    expect(dbLink).toHaveTextContent('2 コース');

    const infraLink = screen.getByRole('link', { name: /インフラ・クラウド のコース一覧へ/ });
    expect(infraLink).toHaveAttribute('href', '/courses/category/infra');

    // コースが無い領域（セキュリティ等）はカードを出さない。
    expect(screen.queryByRole('link', { name: /セキュリティ のコース一覧へ/ })).not.toBeInTheDocument();
  });

  it('未分類コースは uncategorized カードに集約する', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, category: '' })]);
    renderPage('trainee');
    const link = await screen.findByRole('link', { name: /未分類 のコース一覧へ/ });
    expect(link).toHaveAttribute('href', '/courses/category/uncategorized');
  });

  it('受講者にはその領域の学習進捗（章単位の集計）を出す', async () => {
    mockList.mockResolvedValue([
      makeCourse({ id: 1, category: 'database', materialCount: 4, completedCount: 1 }),
      makeCourse({ id: 2, category: 'database', materialCount: 4, completedCount: 2 }),
    ]);
    renderPage('trainee');
    await screen.findByRole('link', { name: /データベース のコース一覧へ/ });
    // 合計 3/8 章完了
    expect(screen.getByText('3/8 章完了')).toBeInTheDocument();
    expect(screen.getByRole('progressbar', { name: /データベース の進捗/ })).toHaveAttribute(
      'aria-valuenow',
      '38',
    );
  });

  it('管理ロールには進捗を出さない（進捗は受講者個人のもの）', async () => {
    mockList.mockResolvedValue([
      makeCourse({ id: 1, category: 'database', materialCount: 4, completedCount: 2 }),
    ]);
    renderPage('company_admin');
    await screen.findByRole('link', { name: /データベース のコース一覧へ/ });
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
  });

  it('管理ロールは「新しいコース」ボタンから作成フォームを開ける', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, category: 'database' })]);
    renderPage('company_admin');
    const btn = await screen.findByRole('button', { name: /新しいコース/ });
    fireEvent.click(btn);
    expect(screen.getByRole('combobox', { name: /カテゴリ（学習領域）/ })).toBeInTheDocument();
  });

  it('trainee には「新しいコース」ボタンを出さない', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, category: 'database' })]);
    renderPage('trainee');
    await screen.findByRole('link', { name: /データベース のコース一覧へ/ });
    expect(screen.queryByRole('button', { name: /新しいコース/ })).not.toBeInTheDocument();
  });

  it('コースが無いときは EmptyState を出す', async () => {
    mockList.mockResolvedValue([]);
    renderPage('trainee');
    await waitFor(() => expect(screen.getByText('コースがありません')).toBeInTheDocument());
  });

  it('API が null を返してもクラッシュせず EmptyState を出す (FRESTYLE-70)', async () => {
    mockList.mockResolvedValue(null as unknown as CourseWithProgress[]);
    renderPage('trainee');
    await waitFor(() => expect(screen.getByText('コースがありません')).toBeInTheDocument());
  });
});
