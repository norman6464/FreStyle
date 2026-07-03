import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import CoursesListPage from '../CoursesListPage';
import { ToastProvider } from '../../components/ToastProvider';
import authReducer from '../../store/authSlice';
import CourseRepository from '../../repositories/CourseRepository';
import { COURSE_CATEGORIES, findCourseCategory } from '../../constants/courseCategories';
import type { Course } from '../../types';

vi.mock('../../repositories/CourseRepository', () => ({
  default: {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

const mockList = vi.mocked(CourseRepository.list);

function makeCourse(overrides: Partial<Course> = {}): Course {
  return {
    id: 1,
    companyId: 10,
    createdByUserId: 1,
    title: 'PostgreSQL 徹底入門',
    description: 'DB の基礎',
    category: 'database',
    sortOrder: 100,
    isPublished: true,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
    ...overrides,
  };
}

function renderPage(role = 'trainee') {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { role } as never,
    },
  });
  return render(
    <Provider store={store}>
      <ToastProvider>
        <MemoryRouter>
          <CoursesListPage />
        </MemoryRouter>
      </ToastProvider>
    </Provider>,
  );
}

describe('CoursesListPage カテゴリ色分け (FRESTYLE-67)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('カテゴリ付きコースはカテゴリ名バッジを表示する', async () => {
    mockList.mockResolvedValue([makeCourse({ category: 'database' })]);
    renderPage();
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.getByText('データベース')).toBeInTheDocument();
  });

  it('未分類コースはカテゴリバッジを表示しない', async () => {
    mockList.mockResolvedValue([makeCourse({ category: '' })]);
    renderPage();
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    for (const c of COURSE_CATEGORIES) {
      expect(screen.queryByText(c.label)).not.toBeInTheDocument();
    }
  });

  it('管理者の作成フォームにカテゴリ選択（未分類 + 全カテゴリ）が表示される', async () => {
    mockList.mockResolvedValue([]);
    renderPage('company_admin');
    await waitFor(() => expect(screen.getByRole('button', { name: /新しいコース/ })).toBeInTheDocument());
    fireEvent.click(screen.getAllByRole('button', { name: /新しいコース/ })[0]);
    const select = screen.getByRole('combobox');
    expect(select).toBeInTheDocument();
    expect(screen.getByRole('option', { name: '未分類' })).toBeInTheDocument();
    for (const c of COURSE_CATEGORIES) {
      expect(screen.getByRole('option', { name: c.label })).toBeInTheDocument();
    }
  });
});

describe('findCourseCategory', () => {
  it('定義済み key は定義を返す', () => {
    expect(findCourseCategory('database')?.label).toBe('データベース');
    expect(findCourseCategory('dev-basics')?.label).toBe('開発基礎');
  });

  it('空文字・未知の key は undefined（無色 = 従来表示）', () => {
    expect(findCourseCategory('')).toBeUndefined();
    expect(findCourseCategory(undefined)).toBeUndefined();
    expect(findCourseCategory('unknown')).toBeUndefined();
  });

  it('カテゴリ key に重複がない', () => {
    const keys = COURSE_CATEGORIES.map((c) => c.key);
    expect(new Set(keys).size).toBe(keys.length);
  });
});
