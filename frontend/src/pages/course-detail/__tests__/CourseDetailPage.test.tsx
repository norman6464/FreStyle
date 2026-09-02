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
import type { RichDocContent } from '@/shared/ui/RichTextEditor';

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

function material(id: number, doc?: RichDocContent | null): TeachingMaterial {
  return {
    id,
    courseId: 5,
    createdByUserId: 1,
    title: `章 ${id}`,
    doc,
    revision: doc ? 1 : undefined,
    orderInCourse: id,
    isPublished: true,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
  };
}

// 見出し付き doc にすると TOC(DocTableOfContents) の IntersectionObserver(jsdom 未実装)が
// 動くため、テストの本文は見出しなし（段落 / 画像のみ）で組み立てる。
function textDoc(text: string): RichDocContent {
  return { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text }] }] };
}

function imageDoc(src: string, alt: string): RichDocContent {
  return {
    type: 'doc',
    content: [{ type: 'paragraph', content: [{ type: 'image', attrs: { src, alt } }] }],
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

/**
 * renderPage は画面を描く。
 *
 * **編集できるかはロールでは決まらない。** サーバーが返す canEdit / canManage で決まるので、
 * 管理者として見たいテストは course() の側を差し替える（renderPage の引数では変わらない）。
 * ロールはまだ store に置かれているが、この画面はもう見ていない。
 */
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
    mockGetMaterial.mockImplementation(async (id: number) => material(id, textDoc('本文テキスト')));
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

  it('編集できる人には自動選択も閲覧記録もされない', async () => {
    // 付与を持つ人（canEdit=true）。ロールではなくサーバーの答えで決まる。
    // canManage は false。ここで true にすると、対象コードが canManage を見ていても通る。
    mockGetCourse.mockResolvedValue({ ...course(), canEdit: true, canManage: false });
    renderPage('trainee');
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

describe('CourseDetailPage 左パネルの章一覧 (FRESTYLE-341)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCourse.mockResolvedValue(course());
    mockCourseList.mockResolvedValue([]);
    mockListMaterials.mockResolvedValue([material(11), material(12)]);
    mockLastViewed.mockResolvedValue(view(11));
    mockGetMaterial.mockImplementation(async (id: number) => material(id, textDoc('本文テキスト')));
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

  it('受講者ビューでも左パネル(コース名・章数つき)に章一覧が表示される', async () => {
    renderPage('trainee');
    // SecondaryPanel のタイトルがコース名。章はリスト項目として並ぶ。
    await waitFor(() => expect(screen.getAllByText('Git 入門').length).toBeGreaterThan(0));
    const items = await screen.findAllByRole('button', { name: /章 11/ });
    expect(items.length).toBeGreaterThan(0);
    expect(screen.getAllByRole('button', { name: /章 12/ }).length).toBeGreaterThan(0);
    // 現在表示中の章が aria-current でハイライトされる。
    await waitFor(() => {
      const current = document.querySelector('[aria-current="page"]');
      expect(current).toHaveTextContent('章 11');
    });
  });

  it('章一覧の章をクリックすると本文が切り替わる', async () => {
    renderPage('trainee');
    const targets = await screen.findAllByRole('button', { name: /章 12/ });
    fireEvent.click(targets[0]);
    await waitFor(() => expect(mockGetMaterial).toHaveBeenCalledWith(12));
  });

  it('完了済みの章にはチェックアイコン、未完了の章には番号が出る', async () => {
    renderPage('trainee');
    await screen.findAllByRole('button', { name: /章 12/ });
    // 11 は完了(チェック)、12 は未完了(番号 2)。デスクトップ + モバイルドロワーで二重描画されるため
    // 「1 つ以上」を確認する。
    await waitFor(() => {
      expect(document.querySelectorAll('[aria-label="完了"]').length).toBeGreaterThan(0);
    });
    const numbered = Array.from(document.querySelectorAll('span')).filter((el) => el.textContent === '2');
    expect(numbered.length).toBeGreaterThan(0);
  });
});

describe('CourseDetailPage 本文内の画像 (FRESTYLE-125)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCourse.mockResolvedValue(course());
    mockCourseList.mockResolvedValue([]);
    mockListMaterials.mockResolvedValue([material(11)]);
    mockLastViewed.mockResolvedValue(null);
    mockGetMaterial.mockImplementation(async (id: number) =>
      material(id, imageDoc('https://example.com/diagram.png', '構成図')),
    );
    mockProgressList.mockResolvedValue([]);
    mockRecordView.mockResolvedValue(undefined);
  });

  it('画像はリンクで包まれない（クリックで別タブに原寸が開かない）', async () => {
    renderPage('trainee');
    const img = await screen.findByRole('img', { name: '構成図' });
    expect(img.closest('a')).toBeNull();
    // tiptap の描画(ProseMirror)内に出る。
    expect(img.closest('.ProseMirror')).not.toBeNull();
  });
});

describe('CourseDetailPage 画像のモーダル拡大表示 (FRESTYLE-191)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCourse.mockResolvedValue(course());
    mockCourseList.mockResolvedValue([]);
    mockListMaterials.mockResolvedValue([material(11)]);
    mockLastViewed.mockResolvedValue(null);
    mockGetMaterial.mockImplementation(async (id: number) =>
      material(id, imageDoc('https://example.com/diagram.png', '構成図')),
    );
    mockProgressList.mockResolvedValue([]);
    mockRecordView.mockResolvedValue(undefined);
  });

  // モーダルを開く。tiptap の描画(ProseMirror)内の img へのクリック委譲で開く。
  // 初期ロード後の再レンダーで取得済みノードが差し替わると click が空振りする(stale node)ため、
  // 「click → dialog 出現」を waitFor で丸ごとリトライする。
  // recordView 待ちだけでは CI の遅い環境で後続の再レンダーと競合してフレークした。
  async function openImageModal() {
    await waitFor(() => expect(mockRecordView).toHaveBeenCalled());
    await waitFor(() => {
      const img = document.querySelector('.ProseMirror img');
      expect(img).not.toBeNull();
      fireEvent.click(img!);
      expect(screen.getByRole('dialog', { name: '構成図' })).toBeInTheDocument();
    });
  }

  it('画像クリックでモーダルが開き、拡大画像と alt が引き継がれる', async () => {
    renderPage('trainee');
    await openImageModal();
    // モーダル内にも同じ src の img が出る（本文内 + モーダルで 2 枚）
    expect(screen.getAllByRole('img', { name: '構成図' })).toHaveLength(2);
  });

  it('閉じるボタンで閉じられる', async () => {
    renderPage('trainee');
    await openImageModal();
    fireEvent.click(screen.getByRole('button', { name: '閉じる' }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('Esc キーで閉じられる', async () => {
    renderPage('trainee');
    await openImageModal();
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('背景クリックで閉じるが、拡大画像自体のクリックでは閉じない', async () => {
    renderPage('trainee');
    await openImageModal();
    const dialog = screen.getByRole('dialog', { name: '構成図' });
    // 画像クリック → 閉じない（誤タップ防止）
    const [, modalImg] = screen.getAllByRole('img', { name: '構成図' });
    fireEvent.click(modalImg);
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    // 背景クリック → 閉じる
    fireEvent.click(dialog);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
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

  it('タイトル h1 は本文カラムの先頭にフラットに置かれる (FRESTYLE-340)', async () => {
    mockGetMaterial.mockImplementation(async (id: number) => material(id, textDoc('本文テキスト')));
    renderPage('trainee');
    const heading = await screen.findByRole('heading', { level: 1, name: '章 11' });
    // ノートと同じ「枠のないインライン文書」。カード(article)には入れない(FRESTYLE-340 で
    // FRESTYLE-178 のカードレイアウトを撤回)。
    expect(heading.closest('article')).toBeNull();
  });

  it('doc 先頭の重複タイトル(h1)は本文に二重表示しない', async () => {
    // 本文 doc が material.title と同じ h1 で始まっても、本文側の h1 は除去される。
    mockGetMaterial.mockImplementation(async (id: number) =>
      material(id, {
        type: 'doc',
        content: [
          { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: `章 ${id}` }] },
          { type: 'paragraph', content: [{ type: 'text', text: '本文テキストです。' }] },
        ],
      }),
    );
    renderPage('trainee');
    await screen.findByText('本文テキストです。');
    // 「章 11」という heading はヘッダーの1つだけ(本文側の重複 h1 は stripLeadingDocTitle で除去済み)。
    expect(screen.getAllByRole('heading', { name: '章 11' })).toHaveLength(1);
  });
});
