import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import CourseDetailPage from '../CourseDetailPage';
import { ToastProvider } from '../../components/ToastProvider';
import authReducer from '../../store/authSlice';
import CourseRepository from '../../repositories/CourseRepository';
import TeachingMaterialRepository from '../../repositories/TeachingMaterialRepository';
import LessonProgressRepository from '../../repositories/LessonProgressRepository';
import DashboardRepository from '../../repositories/DashboardRepository';
import type { Course, TeachingMaterial, UserChapterView } from '../../types';

vi.mock('../../repositories/CourseRepository', () => ({
  default: {
    get: vi.fn(),
    listMaterials: vi.fn(),
    lastViewed: vi.fn(),
  },
}));

vi.mock('../../repositories/TeachingMaterialRepository', () => ({
  default: {
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

vi.mock('../../repositories/LessonProgressRepository', () => ({
  default: {
    list: vi.fn(),
    complete: vi.fn(),
    incomplete: vi.fn(),
  },
}));

vi.mock('../../repositories/DashboardRepository', () => ({
  default: {
    get: vi.fn(),
    recordChapterView: vi.fn(),
  },
}));

const mockGetCourse = vi.mocked(CourseRepository.get);
const mockListMaterials = vi.mocked(CourseRepository.listMaterials);
const mockLastViewed = vi.mocked(CourseRepository.lastViewed);
const mockGetMaterial = vi.mocked(TeachingMaterialRepository.get);
const mockProgressList = vi.mocked(LessonProgressRepository.list);
const mockComplete = vi.mocked(LessonProgressRepository.complete);
const mockRecordView = vi.mocked(DashboardRepository.recordChapterView);

function course(): Course {
  return {
    id: 5,
    companyId: 10,
    createdByUserId: 1,
    title: 'Git 入門',
    description: '',
    category: 'dev-basics',
    sortOrder: 20,
    isPublished: true,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
  };
}

function material(id: number, content = ''): TeachingMaterial {
  return {
    id,
    companyId: 10,
    courseId: 5,
    createdByUserId: 1,
    title: `章 ${id}`,
    content,
    orderInCourse: id,
    isPublished: true,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
  };
}

function view(teachingMaterialId: number): UserChapterView {
  return {
    userId: 7,
    teachingMaterialId,
    courseId: 5,
    firstViewedAt: '2026-07-01T00:00:00Z',
    lastViewedAt: '2026-07-08T00:00:00Z',
    viewCount: 2,
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
        <MemoryRouter initialEntries={['/courses/5']}>
          <Routes>
            <Route path="/courses/:id" element={<CourseDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    </Provider>,
  );
}

describe('CourseDetailPage 続きから表示 + 完了トグル (FRESTYLE-99 / FRESTYLE-100)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCourse.mockResolvedValue(course());
    mockListMaterials.mockResolvedValue([material(11), material(12)]);
    mockLastViewed.mockResolvedValue(view(12));
    mockGetMaterial.mockImplementation(async (id: number) => material(id, '本文テキスト'));
    mockProgressList.mockResolvedValue([]);
    mockComplete.mockResolvedValue(undefined);
    mockRecordView.mockResolvedValue(undefined);
  });

  it('受講者が開くと最後に閲覧した章が自動表示され、閲覧が記録される', async () => {
    renderPage('trainee');
    await waitFor(() =>
      expect(screen.getByRole('heading', { level: 1, name: '章 12' })).toBeInTheDocument(),
    );
    expect(mockLastViewed).toHaveBeenCalledWith(5);
    await waitFor(() => expect(mockRecordView).toHaveBeenCalledWith(12));
  });

  it('完了トグルは sticky なメタ行と本文末尾の 2 箇所に表示される', async () => {
    renderPage('trainee');
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: '完了にする' })).toHaveLength(2),
    );
    // 先頭(メタ行)のトグルはスクロールしても画面に残るよう sticky な行に入っている(FRESTYLE-100)。
    const [metaToggle] = screen.getAllByRole('button', { name: '完了にする' });
    expect(metaToggle.closest('.sticky')).not.toBeNull();
  });

  it('メタ行の完了トグルをクリックすると完了 API を呼ぶ', async () => {
    renderPage('trainee');
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: '完了にする' })).toHaveLength(2),
    );
    fireEvent.click(screen.getAllByRole('button', { name: '完了にする' })[0]);
    await waitFor(() => expect(mockComplete).toHaveBeenCalledWith(12));
  });

  it('閲覧履歴が無い場合は先頭の章が表示される', async () => {
    mockLastViewed.mockResolvedValue(null);
    renderPage('trainee');
    await waitFor(() =>
      expect(screen.getByRole('heading', { level: 1, name: '章 11' })).toBeInTheDocument(),
    );
  });
});
