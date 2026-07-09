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
import type { CourseWithProgress } from '../../types';

vi.mock('../../repositories/CourseRepository', () => ({
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
    sortOrder: 100,
    isPublished: true,
    materialCount: 8,
    completedCount: 0,
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

  it('カテゴリ付きコースはセクション見出し(バッジ + 件数)の下に表示される', async () => {
    mockList.mockResolvedValue([makeCourse({ category: 'database' })]);
    renderPage();
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    // セクション見出し = 「データベース 1 件」の折りたたみボタン
    expect(screen.getByRole('button', { name: /データベース\s*1 件/ })).toBeInTheDocument();
  });

  it('未分類コースは「未分類」セクションに入り、カテゴリ名は表示されない', async () => {
    mockList.mockResolvedValue([makeCourse({ category: '' })]);
    renderPage();
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /未分類\s*1 件/ })).toBeInTheDocument();
    for (const c of COURSE_CATEGORIES) {
      expect(screen.queryByText(c.label)).not.toBeInTheDocument();
    }
  });

  it('セクション見出しクリックで閉じ開きできる', async () => {
    mockList.mockResolvedValue([makeCourse({ category: 'database' })]);
    renderPage();
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    const header = screen.getByRole('button', { name: /データベース\s*1 件/ });
    expect(header).toHaveAttribute('aria-expanded', 'true');

    fireEvent.click(header);
    expect(header).toHaveAttribute('aria-expanded', 'false');
    expect(screen.queryByText('PostgreSQL 徹底入門')).not.toBeInTheDocument();

    fireEvent.click(header);
    expect(header).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument();
  });

  it('カテゴリチップで絞り込み・再クリックで解除できる', async () => {
    mockList.mockResolvedValue([
      makeCourse({ id: 1, title: 'PostgreSQL 徹底入門', category: 'database' }),
      makeCourse({ id: 2, title: 'Terraform 入門', category: 'infra' }),
    ]);
    renderPage();
    await waitFor(() => expect(screen.getByText('Terraform 入門')).toBeInTheDocument());

    const chip = screen.getByRole('button', { name: 'データベース' });
    fireEvent.click(chip);
    expect(chip).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument();
    expect(screen.queryByText('Terraform 入門')).not.toBeInTheDocument();

    fireEvent.click(chip);
    expect(chip).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByText('Terraform 入門')).toBeInTheDocument();
  });

  it('絞り込みで 0 件になったら該当なしの EmptyState を出す', async () => {
    mockList.mockResolvedValue([
      makeCourse({ id: 1, title: 'PostgreSQL 徹底入門', category: 'database' }),
      makeCourse({ id: 2, title: 'Terraform 入門', category: 'infra' }),
    ]);
    renderPage();
    await waitFor(() => expect(screen.getByText('Terraform 入門')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'データベース' }));
    fireEvent.change(screen.getByLabelText('コースを検索'), { target: { value: 'terraform' } });
    expect(screen.getByText('該当するコースがありません')).toBeInTheDocument();
  });

  it('管理者の作成フォームにカテゴリ選択（未分類 + 全カテゴリ）が表示される', async () => {
    mockList.mockResolvedValue([]);
    renderPage('company_admin');
    // 空一覧時はヘッダーと EmptyState の 2 箇所に「新しいコース」が出るため getAllByRole で扱う
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: /新しいコース/ }).length).toBeGreaterThan(0),
    );
    fireEvent.click(screen.getAllByRole('button', { name: /新しいコース/ })[0]);
    const select = screen.getByRole('combobox');
    expect(select).toBeInTheDocument();
    expect(screen.getByRole('option', { name: '未分類' })).toBeInTheDocument();
    for (const c of COURSE_CATEGORIES) {
      expect(screen.getByRole('option', { name: c.label })).toBeInTheDocument();
    }
  });
});

describe('CoursesListPage CRUD フロー', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('trainee には「新しいコース」ボタンを表示しない', async () => {
    mockList.mockResolvedValue([makeCourse()]);
    renderPage('trainee');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /新しいコース/ })).not.toBeInTheDocument();
  });

  it('取得失敗時はエラーメッセージを表示する', async () => {
    mockList.mockRejectedValue(new Error('network'));
    renderPage();
    await waitFor(() =>
      expect(screen.getByText('コースの取得に失敗しました')).toBeInTheDocument(),
    );
  });

  it('検索でタイトルを絞り込める', async () => {
    mockList.mockResolvedValue([
      makeCourse({ id: 1, title: 'PostgreSQL 徹底入門' }),
      makeCourse({ id: 2, title: 'Terraform 入門', category: 'infra' }),
    ]);
    renderPage();
    await waitFor(() => expect(screen.getByText('Terraform 入門')).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('コースを検索'), { target: { value: 'terraform' } });
    expect(screen.queryByText('PostgreSQL 徹底入門')).not.toBeInTheDocument();
    expect(screen.getByText('Terraform 入門')).toBeInTheDocument();
  });

  it('作成フォームの送信でカテゴリ込みの payload が送られる', async () => {
    const mockCreate = vi.mocked(CourseRepository.create);
    mockList.mockResolvedValue([]);
    mockCreate.mockResolvedValue(makeCourse({ id: 99, title: '新コース', category: 'security' }));
    renderPage('company_admin');
    await waitFor(() => expect(screen.getAllByRole('button', { name: /新しいコース/ }).length).toBeGreaterThan(0));
    fireEvent.click(screen.getAllByRole('button', { name: /新しいコース/ })[0]);

    // モーダル内のタイトル入力 = 検索ボックス以外の <input type="text">。
    const titleInput = screen
      .getAllByRole('textbox')
      .find((el) => el.tagName === 'INPUT' && el.getAttribute('aria-label') !== 'コースを検索');
    expect(titleInput).toBeDefined();
    fireEvent.change(titleInput!, { target: { value: '新コース' } });
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'security' } });
    fireEvent.click(screen.getByRole('button', { name: '作成' }));

    await waitFor(() => expect(mockCreate).toHaveBeenCalledTimes(1));
    expect(mockCreate).toHaveBeenCalledWith(
      expect.objectContaining({ title: '新コース', category: 'security' }),
    );
  });

  it('編集フォームは既存カテゴリが初期選択され、更新 payload に含まれる', async () => {
    const mockUpdate = vi.mocked(CourseRepository.update);
    mockList.mockResolvedValue([makeCourse({ id: 5, category: 'database' })]);
    mockUpdate.mockResolvedValue(makeCourse({ id: 5, category: 'infra' }));
    renderPage('company_admin');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コースを編集' }));
    const select = screen.getByRole('combobox') as HTMLSelectElement;
    expect(select.value).toBe('database');
    fireEvent.change(select, { target: { value: 'infra' } });
    fireEvent.click(screen.getByRole('button', { name: '更新' }));

    await waitFor(() => expect(mockUpdate).toHaveBeenCalledTimes(1));
    expect(mockUpdate).toHaveBeenCalledWith(
      5,
      expect.objectContaining({ category: 'infra' }),
    );
  });

  it('削除は確認モーダル経由で remove を呼ぶ', async () => {
    const mockRemove = vi.mocked(CourseRepository.remove);
    mockList.mockResolvedValue([makeCourse({ id: 7 })]);
    mockRemove.mockResolvedValue(undefined);
    renderPage('company_admin');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コースを削除' }));
    expect(screen.getByText(/このコースを削除しますか/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '削除' }));

    await waitFor(() => expect(mockRemove).toHaveBeenCalledWith(7));
    // 削除後は一覧から消える
    await waitFor(() =>
      expect(screen.queryByText('PostgreSQL 徹底入門')).not.toBeInTheDocument(),
    );
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

describe('API が null を返す場合の防御 (FRESTYLE-70)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('コース一覧 API が null でもクラッシュせず EmptyState を表示する', async () => {
    mockList.mockResolvedValue(null as unknown as CourseWithProgress[]);
    renderPage();
    await waitFor(() => expect(screen.getByText('コースがありません')).toBeInTheDocument());
  });
});

describe('CoursesListPage カード進捗表示 (FRESTYLE-98)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('受講者のカードに完了章数/全章数と進捗バーが表示される', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 8, completedCount: 2 })]);
    renderPage('trainee');
    await waitFor(() => expect(screen.getByText('2/8（25%）')).toBeInTheDocument());
    expect(screen.getByRole('progressbar', { name: '学習の進捗' })).toHaveAttribute(
      'aria-valuenow',
      '25',
    );
  });

  it('完了記録が無いコースは 0/N と表示される', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 8, completedCount: 0 })]);
    renderPage('trainee');
    await waitFor(() => expect(screen.getByText('0/8（0%）')).toBeInTheDocument());
  });

  it('管理ロールには進捗バーを表示しない', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 8, completedCount: 3 })]);
    renderPage('company_admin');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
  });

  it('章が 0 件のコースには進捗バーを出さない', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 0 })]);
    renderPage('trainee');
    await waitFor(() => expect(screen.getByText('PostgreSQL 徹底入門')).toBeInTheDocument());
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
  });

  it('全章完了は 100% と表示される', async () => {
    mockList.mockResolvedValue([makeCourse({ id: 1, materialCount: 5, completedCount: 5 })]);
    renderPage('trainee');
    await waitFor(() => expect(screen.getByText('5/5（100%）')).toBeInTheDocument());
  });
});
