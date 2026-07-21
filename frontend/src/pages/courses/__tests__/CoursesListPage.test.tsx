import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import CoursesListPage from '../ui/CoursesListPage';
import { ToastProvider } from '@/app/providers/ToastProvider';
import authReducer from '@/entities/user/model/authSlice';
import { CourseRepository } from '@/entities/course';
import { COURSE_CATEGORIES, findCourseCategory } from '@/entities/course';
import { COURSE_LANGUAGES } from '@/entities/course';
import type { CourseWithProgress } from '@/entities/course';

vi.mock('@/entities/course/api/courseRepository', () => ({
  default: {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

const mockList = vi.mocked(CourseRepository.list);

function makeCourse(overrides: Partial<CourseWithProgress> = {}): CourseWithProgress {
  return {
    id: 1,
    companyId: 10,
    createdByUserId: 1,
    title: 'PostgreSQL 徹底入門',
    description: 'DB の基礎',
    category: 'database',
    language: '',
    sortOrder: 100,
    isPublished: true,
    materialCount: 8,
    completedCount: 0,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
    ...overrides,
  };
}

// CoursesListPage は /courses/category/:category にマウントされる。
// slug の領域だけに絞って表示するので、テストは対象コースと同じ slug で描画する。
function renderPage(role = 'trainee', slug = 'database') {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { role } as never },
  });
  return render(
    <Provider store={store}>
      <ToastProvider>
        <MemoryRouter initialEntries={[`/courses/category/${slug}`]}>
          <Routes>
            <Route path="/courses/category/:category" element={<CoursesListPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    </Provider>,
  );
}

describe('CoursesListPage 領域スコープ (FRESTYLE-177)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('URL の領域名を見出しに出し、その領域のコースを表示する', async () => {
    mockList.mockResolvedValue([makeCourse({ category: 'database' })]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.getByRole('heading', { name: 'データベース', level: 1 })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /コース一覧に戻る/ })).toHaveAttribute('href', '/courses');
  });

  it('別領域のコースは表示しない（URL の領域だけに絞る）', async () => {
    mockList.mockResolvedValue([
      makeCourse({ id: 1, title: 'PostgreSQL 徹底入門', category: 'database' }),
      makeCourse({ id: 2, title: 'Terraform 入門', category: 'infra' }),
    ]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.queryByText('Terraform 入門')).not.toBeInTheDocument();
  });

  it('uncategorized は「未分類」見出しで未分類コースを表示する', async () => {
    mockList.mockResolvedValue([makeCourse({ category: '' })]);
    renderPage('trainee', 'uncategorized');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.getByRole('heading', { name: '未分類', level: 1 })).toBeInTheDocument();
  });

  it('領域内検索でタイトルを絞り込める', async () => {
    mockList.mockResolvedValue([
      makeCourse({ id: 1, title: 'PostgreSQL 徹底入門', category: 'database' }),
      makeCourse({ id: 2, title: 'MySQL 入門', category: 'database' }),
    ]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('MySQL 入門')).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('コースを検索'), { target: { value: 'mysql' } });
    expect(screen.queryByText('PostgreSQL 徹底入門')).not.toBeInTheDocument();
    expect(screen.getByText('MySQL 入門')).toBeInTheDocument();
  });

  it('検索で 0 件になったら該当なしの EmptyState を出す', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, title: 'PostgreSQL 徹底入門', category: 'database' })]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('コースを検索'), { target: { value: 'terraform' } });
    expect(screen.getByText('該当するコースがありません')).toBeInTheDocument();
  });
});

describe('CoursesListPage CRUD フロー', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('trainee には「新しいコース」ボタンを表示しない', async () => {
    mockList.mockResolvedValue([makeCourse()]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /新しいコース/ })).not.toBeInTheDocument();
  });

  it('取得失敗時はエラーメッセージを表示する', async () => {
    mockList.mockRejectedValue(new Error('network'));
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('コースの取得に失敗しました')).toBeInTheDocument());
  });

  it('管理者の作成フォームにカテゴリ選択（未分類 + 全カテゴリ）が表示される', async () => {
    mockList.mockResolvedValue([]);
    renderPage('company_admin', 'database');
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: /新しいコース/ }).length).toBeGreaterThan(0),
    );
    fireEvent.click(screen.getAllByRole('button', { name: /新しいコース/ })[0]);
    expect(screen.getByRole('combobox', { name: /カテゴリ（学習領域）/ })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: '未分類' })).toBeInTheDocument();
    for (const c of COURSE_CATEGORIES) {
      expect(screen.getByRole('option', { name: c.label })).toBeInTheDocument();
    }
  });

  it('管理者の作成フォームは現在の領域を初期選択にする', async () => {
    mockList.mockResolvedValue([]);
    renderPage('company_admin', 'security');
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: /新しいコース/ }).length).toBeGreaterThan(0),
    );
    fireEvent.click(screen.getAllByRole('button', { name: /新しいコース/ })[0]);
    const select = screen.getByRole('combobox', { name: /カテゴリ（学習領域）/ }) as HTMLSelectElement;
    expect(select.value).toBe('security');
  });

  it('管理者の作成フォームに言語選択（未設定 + 言語一覧）が表示される', async () => {
    mockList.mockResolvedValue([]);
    renderPage('company_admin', 'database');
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: /新しいコース/ }).length).toBeGreaterThan(0),
    );
    fireEvent.click(screen.getAllByRole('button', { name: /新しいコース/ })[0]);
    expect(screen.getByRole('combobox', { name: /主に扱う言語・技術/ })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: '未設定（言語が主題でないコース）' })).toBeInTheDocument();
    for (const l of COURSE_LANGUAGES) {
      expect(screen.getByRole('option', { name: l.label })).toBeInTheDocument();
    }
  });

  it('作成フォームの送信でカテゴリ込みの payload が送られる', async () => {
    const mockCreate = vi.mocked(CourseRepository.create);
    mockList.mockResolvedValue([]);
    mockCreate.mockResolvedValue(makeCourse({ id: 99, title: '新コース', category: 'security' }));
    renderPage('company_admin', 'database');
    await waitFor(() => expect(screen.getAllByRole('button', { name: /新しいコース/ }).length).toBeGreaterThan(0));
    fireEvent.click(screen.getAllByRole('button', { name: /新しいコース/ })[0]);

    const titleInput = screen
      .getAllByRole('textbox')
      .find((el) => el.tagName === 'INPUT' && el.getAttribute('aria-label') !== 'コースを検索');
    expect(titleInput).toBeDefined();
    fireEvent.change(titleInput!, { target: { value: '新コース' } });
    fireEvent.change(screen.getByRole('combobox', { name: /カテゴリ（学習領域）/ }), {
      target: { value: 'security' },
    });
    fireEvent.change(screen.getByRole('combobox', { name: /主に扱う言語・技術/ }), {
      target: { value: 'go' },
    });
    fireEvent.click(screen.getByRole('button', { name: '作成' }));

    await waitFor(() => expect(mockCreate).toHaveBeenCalledTimes(1));
    expect(mockCreate).toHaveBeenCalledWith(
      expect.objectContaining({ title: '新コース', category: 'security', language: 'go' }),
    );
  });

  it('編集フォームは既存カテゴリ・言語が初期選択され、更新 payload に含まれる', async () => {
    const mockUpdate = vi.mocked(CourseRepository.update);
    mockList.mockResolvedValue([makeCourse({ id: 5, category: 'database', language: 'postgresql' })]);
    mockUpdate.mockResolvedValue(makeCourse({ id: 5, category: 'infra', language: 'terraform' }));
    renderPage('company_admin', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コースを編集' }));
    const select = screen.getByRole('combobox', { name: /カテゴリ（学習領域）/ }) as HTMLSelectElement;
    expect(select.value).toBe('database');
    const langSelect = screen.getByRole('combobox', { name: /主に扱う言語・技術/ }) as HTMLSelectElement;
    expect(langSelect.value).toBe('postgresql');
    fireEvent.change(select, { target: { value: 'infra' } });
    fireEvent.change(langSelect, { target: { value: 'terraform' } });
    fireEvent.click(screen.getByRole('button', { name: '更新' }));

    await waitFor(() => expect(mockUpdate).toHaveBeenCalledTimes(1));
    expect(mockUpdate).toHaveBeenCalledWith(
      5,
      expect.objectContaining({ category: 'infra', language: 'terraform' }),
    );
  });

  it('削除は確認モーダル経由で remove を呼ぶ', async () => {
    const mockRemove = vi.mocked(CourseRepository.remove);
    mockList.mockResolvedValue([makeCourse({ id: 7 })]);
    mockRemove.mockResolvedValue(undefined);
    renderPage('company_admin', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コースを削除' }));
    expect(screen.getByText(/このコースを削除しますか/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '削除' }));

    await waitFor(() => expect(mockRemove).toHaveBeenCalledWith(7));
    await waitFor(() => expect(screen.queryByText('PostgreSQL 徹底入門')).not.toBeInTheDocument());
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

describe('CoursesListPage カード進捗表示 (FRESTYLE-98)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('受講者のカードに完了章数/全章数・残り章数と進捗バーが表示される', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 8, completedCount: 2 })]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('2/8（25%・残り 6 章）')).toBeInTheDocument());
    expect(screen.getByRole('progressbar', { name: '学習の進捗' })).toHaveAttribute('aria-valuenow', '25');
  });

  it('完了記録が無いコースは 0/N と表示される', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 8, completedCount: 0 })]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('0/8（0%・残り 8 章）')).toBeInTheDocument());
  });

  it('管理ロールには進捗バーを表示しない', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 8, completedCount: 3 })]);
    renderPage('company_admin', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
  });

  it('章が 0 件のコースには進捗バーを出さない', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 0 })]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
  });

  it('全章完了は「すべて完了」+ 完了バッジ + 完了デザインで表示される (FRESTYLE-114)', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 5, completedCount: 5 })]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('すべて完了（5 章）')).toBeInTheDocument());
    expect(screen.getByText('完了')).toBeInTheDocument();
  });

  it('未完了のコースには完了バッジを出さない', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 5, completedCount: 4 })]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('4/5（80%・残り 1 章）')).toBeInTheDocument());
    expect(screen.queryByText('完了')).not.toBeInTheDocument();
  });

  it('管理ロールには完了バッジを出さない（進捗は受講者個人のもの）', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 5, completedCount: 5 })]);
    renderPage('company_admin', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.queryByText('完了')).not.toBeInTheDocument();
  });
});

describe('CoursesListPage 言語バッジ (FRESTYLE-114)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('language が設定されたコースは言語バッジを表示する', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, language: 'go', title: 'Go 言語徹底攻略', category: 'backend' })]);
    renderPage('trainee', 'backend');
    await waitFor(() => expect(screen.getByText('Go 言語徹底攻略')).toBeInTheDocument());
    expect(screen.getByText('Go')).toBeInTheDocument();
  });

  it('language が空のコースはバッジを出さない', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, language: '' })]);
    renderPage('trainee', 'database');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.queryByText('Go')).not.toBeInTheDocument();
  });
});
