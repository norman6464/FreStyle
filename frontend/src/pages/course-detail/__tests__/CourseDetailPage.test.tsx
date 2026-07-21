import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import CourseDetailPage from '../ui/CourseDetailPage';
import { ToastProvider } from '@/app/providers/ToastProvider';
import authReducer from '@/entities/user/model/authSlice';
import { CourseRepository } from '@/entities/course';
import { TeachingMaterialRepository } from '@/entities/course';
import { LessonProgressRepository } from '@/entities/course';
import { DashboardRepository } from '@/entities/user';
import type { Course, CourseWithProgress, TeachingMaterial, UserChapterView } from '@/entities/course';

vi.mock('@/entities/course/api/courseRepository', () => ({
  default: {
    get: vi.fn(),
    list: vi.fn(),
    listMaterials: vi.fn(),
    lastViewed: vi.fn(),
  },
}));

vi.mock('@/entities/course/api/teachingMaterialRepository', () => ({
  default: {
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

vi.mock('@/entities/course/api/lessonProgressRepository', () => ({
  default: {
    list: vi.fn(),
    complete: vi.fn(),
    incomplete: vi.fn(),
  },
}));

vi.mock('@/entities/user/api/dashboardRepository', () => ({
  default: {
    get: vi.fn(),
    recordChapterView: vi.fn(),
  },
}));

const mockGetCourse = vi.mocked(CourseRepository.get);
const mockCourseList = vi.mocked(CourseRepository.list);
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
    language: 'git',
    sortOrder: 20,
    isPublished: true,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
  };
}

function listedCourse(id: number, title: string): CourseWithProgress {
  return {
    ...course(),
    id,
    title,
    sortOrder: id * 10,
    materialCount: 5,
    completedCount: 0,
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
    mockCourseList.mockResolvedValue([listedCourse(5, 'Git 入門'), listedCourse(6, 'Docker 入門')]);
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

  it('完了トグルはメタ行と本文末尾の 2 箇所に表示される', async () => {
    renderPage('trainee');
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: '完了にする' })).toHaveLength(2),
    );
    // メタ行のトグルは通常の行に入っている(FRESTYLE-119 で sticky 解除。固定表示しない)。
    const [metaToggle] = screen.getAllByRole('button', { name: '完了にする' });
    expect(metaToggle.closest('.sticky')).toBeNull();
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

  it('管理ロールでは自動選択も閲覧記録もされない', async () => {
    renderPage('company_admin');
    await waitFor(() => expect(mockListMaterials).toHaveBeenCalled());
    // 章メニューはモバイル drawer とデスクトップパネルの 2 箇所に描画される。
    await waitFor(() => expect(screen.getAllByText('章 11').length).toBeGreaterThan(0));
    expect(mockLastViewed).not.toHaveBeenCalled();
    expect(mockRecordView).not.toHaveBeenCalled();
    // 自動選択されないため本文見出しは出ない。
    expect(screen.queryByRole('heading', { level: 1, name: /章 1[12]/ })).not.toBeInTheDocument();
  });

  it('「次の章へ」で次の教材に切り替わり、閲覧も記録される', async () => {
    mockLastViewed.mockResolvedValue(view(11));
    renderPage('trainee');
    await waitFor(() =>
      expect(screen.getByRole('heading', { level: 1, name: '章 11' })).toBeInTheDocument(),
    );
    fireEvent.click(screen.getByRole('button', { name: /次の章へ/ }));
    await waitFor(() =>
      expect(screen.getByRole('heading', { level: 1, name: '章 12' })).toBeInTheDocument(),
    );
    await waitFor(() => expect(mockRecordView).toHaveBeenCalledWith(12));
  });

  it('最終章の末尾に「次のコースへ」が表示され、クリックで次のコースへ遷移する (FRESTYLE-102)', async () => {
    // lastViewed = 章 12(最終章)。次の章が無いので「次のコースへ」が出る。
    renderPage('trainee');
    await waitFor(() =>
      expect(screen.getByRole('heading', { level: 1, name: '章 12' })).toBeInTheDocument(),
    );
    expect(screen.queryByRole('button', { name: /次の章へ/ })).not.toBeInTheDocument();
    const nextCourseBtn = screen.getByRole('button', { name: /次のコースへ: Docker 入門/ });
    expect(nextCourseBtn).toBeInTheDocument();

    fireEvent.click(nextCourseBtn);
    // /courses/6 へ遷移し、次のコースのデータ取得が始まる。
    await waitFor(() => expect(mockGetCourse).toHaveBeenCalledWith(6));
    await waitFor(() => expect(mockListMaterials).toHaveBeenCalledWith(6));
  });

  it('最終章以外では「次の章へ」が出て「次のコースへ」は出ない', async () => {
    mockLastViewed.mockResolvedValue(view(11));
    renderPage('trainee');
    await waitFor(() =>
      expect(screen.getByRole('heading', { level: 1, name: '章 11' })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /次の章へ/ })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /次のコースへ/ })).not.toBeInTheDocument();
  });

  it('並び順で最後のコースでは最終章でも「次のコースへ」を出さない', async () => {
    mockCourseList.mockResolvedValue([listedCourse(4, 'Linux'), listedCourse(5, 'Git 入門')]);
    renderPage('trainee'); // lastViewed = 章 12(最終章)
    await waitFor(() =>
      expect(screen.getByRole('heading', { level: 1, name: '章 12' })).toBeInTheDocument(),
    );
    expect(screen.queryByRole('button', { name: /次のコースへ/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /次の章へ/ })).not.toBeInTheDocument();
  });

  it('本文の取得に失敗したらエラーメッセージを表示する', async () => {
    mockGetMaterial.mockRejectedValue(new Error('network'));
    renderPage('trainee');
    await waitFor(() => expect(screen.getByText('教材の取得に失敗しました')).toBeInTheDocument());
  });

  it('コースの取得に失敗したらエラー表示になる', async () => {
    mockGetCourse.mockRejectedValue(new Error('network'));
    renderPage('trainee');
    await waitFor(() =>
      expect(screen.getByText('コースの取得に失敗しました')).toBeInTheDocument(),
    );
  });

  it('完了済みの章は「完了済み」表示になり、クリックで完了解除 API を呼ぶ', async () => {
    const mockIncomplete = vi.mocked(LessonProgressRepository.incomplete);
    mockIncomplete.mockResolvedValue(undefined);
    mockProgressList.mockResolvedValue([
      {
        id: 1,
        userId: 7,
        teachingMaterialId: 12,
        courseId: 5,
        completedAt: '2026-07-08T00:00:00Z',
        createdAt: '2026-07-08T00:00:00Z',
      },
    ]);
    renderPage('trainee'); // lastViewed = 章 12
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: '完了済み' }).length).toBeGreaterThan(0),
    );
    fireEvent.click(screen.getAllByRole('button', { name: '完了済み' })[0]);
    await waitFor(() => expect(mockIncomplete).toHaveBeenCalledWith(12));
  });
});

describe('CourseDetailPage 右サイドバーの章一覧 (FRESTYLE-118)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCourse.mockResolvedValue(course());
    mockCourseList.mockResolvedValue([]);
    mockListMaterials.mockResolvedValue([material(11), material(12)]);
    mockLastViewed.mockResolvedValue(view(11));
    // 見出し付き本文にすると TOC の IntersectionObserver(jsdom 未実装)が動くため見出しなしにする。
    mockGetMaterial.mockImplementation(async (id: number) => material(id, '本文テキスト'));
    mockProgressList.mockResolvedValue([
      {
        id: 1,
        userId: 7,
        teachingMaterialId: 11,
        courseId: 5,
        completedAt: '2026-07-08T00:00:00Z',
      } as never,
    ]);
    mockRecordView.mockResolvedValue(undefined);
  });

  it('受講者ビューでは右サイドバーに章一覧(コース名・章数つき)が表示される', async () => {
    renderPage('trainee');
    await waitFor(() => expect(screen.getByRole('navigation', { name: '章一覧' })).toBeInTheDocument());
    const nav = screen.getByRole('navigation', { name: '章一覧' });
    expect(nav).toHaveTextContent('章 11');
    expect(nav).toHaveTextContent('章 12');
    // 現在表示中の章が aria-current でハイライトされる
    const current = nav.querySelector('[aria-current="page"]');
    expect(current).toHaveTextContent('章 11');
  });

  it('章一覧の章をクリックすると本文が切り替わる', async () => {
    renderPage('trainee');
    await waitFor(() => expect(screen.getByRole('navigation', { name: '章一覧' })).toBeInTheDocument());
    const nav = screen.getByRole('navigation', { name: '章一覧' });
    const target = Array.from(nav.querySelectorAll('button')).find((b) => b.textContent?.includes('章 12'));
    expect(target).toBeDefined();
    fireEvent.click(target!);
    await waitFor(() => expect(mockGetMaterial).toHaveBeenCalledWith(12));
  });

  it('完了済みの章にはチェックアイコン、未完了の章には番号が出る', async () => {
    renderPage('trainee');
    await waitFor(() => expect(screen.getByRole('navigation', { name: '章一覧' })).toBeInTheDocument());
    const nav = screen.getByRole('navigation', { name: '章一覧' });
    // 11 は完了(チェック)、12 は未完了(番号 2)
    expect(nav.querySelectorAll('[aria-label="完了"]')).toHaveLength(1);
    expect(nav).toHaveTextContent('2');
  });
});

describe('CourseDetailPage 本文内の画像 (FRESTYLE-125)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCourse.mockResolvedValue(course());
    mockCourseList.mockResolvedValue([]);
    mockListMaterials.mockResolvedValue([material(11)]);
    mockLastViewed.mockResolvedValue(null);
    // 見出し付き本文にすると TOC の IntersectionObserver(jsdom 未実装)が動くため画像のみにする。
    mockGetMaterial.mockImplementation(async (id: number) =>
      material(id, '![構成図](https://example.com/diagram.png)'),
    );
    mockProgressList.mockResolvedValue([]);
    mockRecordView.mockResolvedValue(undefined);
  });

  it('画像はリンクで包まれない（クリックで別タブに原寸が開かない）', async () => {
    renderPage('trainee');
    const img = await screen.findByRole('img', { name: '構成図' });
    expect(img.closest('a')).toBeNull();
    // キャプション(figcaption)は維持される
    expect(img.closest('figure')).not.toBeNull();
    expect(screen.getByText('構成図', { selector: 'figcaption' })).toBeInTheDocument();
  });
});

describe('CourseDetailPage タイトルのカード外配置 (FRESTYLE-131)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCourse.mockResolvedValue(course());
    mockCourseList.mockResolvedValue([]);
    mockListMaterials.mockResolvedValue([material(11)]);
    mockLastViewed.mockResolvedValue(null);
    mockProgressList.mockResolvedValue([]);
    mockRecordView.mockResolvedValue(undefined);
  });

  it('タイトル h1 は白カード(article)の外に置かれる', async () => {
    // 見出し付き本文にすると TOC の IntersectionObserver(jsdom 未実装)が動くため見出しなしにする。
    mockGetMaterial.mockImplementation(async (id: number) => material(id, '本文テキスト'));
    renderPage('trainee');
    const heading = await screen.findByRole('heading', { level: 1, name: '章 11' });
    // h1 は article の中に入っていない(カードの外のヘッダーにある)。
    expect(heading.closest('article')).toBeNull();
    expect(heading.closest('header')).not.toBeNull();
  });

  it('本文先頭の重複タイトル(# タイトル)はカード内に二重表示しない', async () => {
    // 本文が material.title と同じ h1 で始まっても、カード内にタイトル h1 を出さない。
    mockGetMaterial.mockImplementation(async (id: number) =>
      material(id, '# 章 11\n\n本文テキストです。'),
    );
    renderPage('trainee');
    await screen.findByText('本文テキストです。');
    // 「章 11」という heading はヘッダーの1つだけ(本文側の重複 h1 は除去済み)。
    expect(screen.getAllByRole('heading', { name: '章 11' })).toHaveLength(1);
  });
});
