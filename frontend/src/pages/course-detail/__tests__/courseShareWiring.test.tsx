import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import CourseDetailPage from '../ui/CourseDetailPage';
import authReducer from '@/entities/user/model/authSlice';
import { ToastProvider } from '@/app/providers/ToastProvider';
import { CourseRepository, LessonProgressRepository, TeachingMaterialRepository } from '@/entities/course';
import { ChapterViewRepository } from '@/entities/user';
import type { Course } from '@/entities/course';

// 深いパスで mock する（barrel を読むとカバレッジの分母が膨らむ）。
vi.mock('@/entities/course/api/courseRepository', () => ({
  default: {
    get: vi.fn(),
    list: vi.fn(),
    listMaterials: vi.fn(),
    lastViewed: vi.fn(),
    listGrants: vi.fn(),
    listGrantablePrincipals: vi.fn(),
    grantRole: vi.fn(),
    revokeRole: vi.fn(),
  },
}));

vi.mock('@/entities/course/api/teachingMaterialRepository', () => ({
  default: { get: vi.fn(), create: vi.fn(), update: vi.fn(), remove: vi.fn() },
}));

vi.mock('@/entities/course/api/lessonProgressRepository', () => ({
  default: { list: vi.fn(), complete: vi.fn(), incomplete: vi.fn() },
}));

vi.mock('@/entities/user/api/chapterViewRepository', () => ({
  default: { recordChapterView: vi.fn() },
}));

const getCourse = vi.mocked(CourseRepository.get);
const listGrants = vi.mocked(CourseRepository.listGrants);
const listPrincipals = vi.mocked(CourseRepository.listGrantablePrincipals);

function course(overrides: Partial<Course> = {}): Course {
  return {
    id: 5,
    createdByUserId: 1,
    title: 'Git 入門',
    description: '',
    category: '',
    language: '',
    sortOrder: 100,
    isPublished: true,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function renderPage() {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { role: 'trainee' } as never },
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

beforeEach(() => {
  vi.clearAllMocks();
  getCourse.mockResolvedValue(course());
  vi.mocked(CourseRepository.list).mockResolvedValue([]);
  vi.mocked(CourseRepository.listMaterials).mockResolvedValue([]);
  vi.mocked(CourseRepository.lastViewed).mockResolvedValue(null);
  vi.mocked(TeachingMaterialRepository.get).mockResolvedValue(null as never);
  vi.mocked(LessonProgressRepository.list).mockResolvedValue([]);
  vi.mocked(ChapterViewRepository.recordChapterView).mockResolvedValue(undefined as never);
  listGrants.mockResolvedValue([]);
  listPrincipals.mockResolvedValue([]);
});

describe('コース詳細の共有', () => {
  it('権限を変えられない人には共有ボタンを出さない', async () => {
    // 押しても 403 が返るだけのボタンは、権限が無いことすら伝えない。
    getCourse.mockResolvedValue(course({ canManage: false }));
    renderPage();

    await waitFor(() => expect(getCourse).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '共有' })).not.toBeInTheDocument();
  });

  it('開くまで権限は取りに行かない', async () => {
    getCourse.mockResolvedValue(course({ canManage: true }));
    renderPage();

    // 章メニューと同じく、モバイル drawer とデスクトップパネルの 2 箇所に描かれる。
    const share = (await screen.findAllByRole('button', { name: '共有' }))[0];
    // 開いていないパネルのために、コースを開くたび 2 本引かない。
    expect(listGrants).not.toHaveBeenCalled();
    expect(listPrincipals).not.toHaveBeenCalled();

    fireEvent.click(share);
    await waitFor(() => expect(listGrants).toHaveBeenCalledWith(5));
    expect(listPrincipals).toHaveBeenCalledWith(5);
    expect((await screen.findAllByRole('region', { name: '共有' })).length).toBeGreaterThan(0);
  });

  it('canEdit が false なら編集の入口を出さない', async () => {
    // ここが緩むと、付与を持たない人にボタンだけ出て保存が全部弾かれる
    // （以前はアプリのロールで出していたので、実際にその状態だった）。
    getCourse.mockResolvedValue(course({ canEdit: false }));
    renderPage();

    await waitFor(() => expect(getCourse).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '新しい教材' })).not.toBeInTheDocument();
  });

  it('canEdit が true なら編集の入口を出す', async () => {
    getCourse.mockResolvedValue(course({ canEdit: true }));
    renderPage();

    expect((await screen.findAllByRole('button', { name: '新しい教材' })).length).toBeGreaterThan(0);
  });
});
