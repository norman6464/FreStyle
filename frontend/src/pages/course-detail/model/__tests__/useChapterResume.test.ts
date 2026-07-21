import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useChapterResume } from '../useChapterResume';
import { CourseRepository } from '@/entities/course';
import type { TeachingMaterial, UserChapterView } from '@/entities/course';

vi.mock('@/entities/course/api/courseRepository', () => ({
  default: {
    lastViewed: vi.fn(),
  },
}));

const mockLastViewed = vi.mocked(CourseRepository.lastViewed);

function material(id: number, courseId = 5): TeachingMaterial {
  return {
    id,
    companyId: 10,
    courseId,
    createdByUserId: 1,
    title: `章 ${id}`,
    content: '',
    orderInCourse: id,
    isPublished: true,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
  };
}

function view(teachingMaterialId: number, courseId = 5): UserChapterView {
  return {
    userId: 7,
    teachingMaterialId,
    courseId,
    firstViewedAt: '2026-07-01T00:00:00Z',
    lastViewedAt: '2026-07-08T00:00:00Z',
    viewCount: 3,
  };
}

type Params = Parameters<typeof useChapterResume>[0];

function baseParams(overrides: Partial<Params> = {}): Params {
  return {
    enabled: true,
    courseId: 5,
    materials: [material(11), material(12), material(13)],
    loading: false,
    selectedId: null,
    selectMaterial: vi.fn(),
    ...overrides,
  };
}

describe('useChapterResume', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('最後に閲覧した章が一覧にあればそれを自動選択する', async () => {
    mockLastViewed.mockResolvedValue(view(12));
    const params = baseParams();
    renderHook(() => useChapterResume(params));
    await waitFor(() => expect(params.selectMaterial).toHaveBeenCalledWith(12));
    expect(params.selectMaterial).toHaveBeenCalledTimes(1);
  });

  it('履歴なし(204/null)は先頭の章を選択する', async () => {
    mockLastViewed.mockResolvedValue(null);
    const params = baseParams();
    renderHook(() => useChapterResume(params));
    await waitFor(() => expect(params.selectMaterial).toHaveBeenCalledWith(11));
  });

  it('履歴の章が一覧に無い(非公開化等)場合は先頭へフォールバック', async () => {
    mockLastViewed.mockResolvedValue(view(999));
    const params = baseParams();
    renderHook(() => useChapterResume(params));
    await waitFor(() => expect(params.selectMaterial).toHaveBeenCalledWith(11));
  });

  it('履歴の取得に失敗しても先頭の章を選択する', async () => {
    mockLastViewed.mockRejectedValue(new Error('network'));
    const params = baseParams();
    renderHook(() => useChapterResume(params));
    await waitFor(() => expect(params.selectMaterial).toHaveBeenCalledWith(11));
  });

  it('enabled=false(管理ロール)では API も自動選択もしない', async () => {
    const params = baseParams({ enabled: false });
    renderHook(() => useChapterResume(params));
    await new Promise((r) => setTimeout(r, 10));
    expect(mockLastViewed).not.toHaveBeenCalled();
    expect(params.selectMaterial).not.toHaveBeenCalled();
  });

  it('loading 中は発火せず、ロード完了後に一度だけ発火する', async () => {
    mockLastViewed.mockResolvedValue(view(12));
    const params = baseParams({ loading: true });
    const { rerender } = renderHook((p: Params) => useChapterResume(p), {
      initialProps: params,
    });
    expect(mockLastViewed).not.toHaveBeenCalled();

    const loaded = { ...params, loading: false };
    rerender(loaded);
    await waitFor(() => expect(params.selectMaterial).toHaveBeenCalledWith(12));

    // 再レンダーしても再発火しない(同一コースで一度だけ)。
    rerender({ ...loaded });
    await new Promise((r) => setTimeout(r, 10));
    expect(mockLastViewed).toHaveBeenCalledTimes(1);
    expect(params.selectMaterial).toHaveBeenCalledTimes(1);
  });

  it('すでに章が選択済みなら API も自動選択もしない', async () => {
    const params = baseParams({ selectedId: 13 });
    renderHook(() => useChapterResume(params));
    await new Promise((r) => setTimeout(r, 10));
    expect(mockLastViewed).not.toHaveBeenCalled();
    expect(params.selectMaterial).not.toHaveBeenCalled();
  });

  it('履歴取得中にユーザーが手動選択したら上書きしない', async () => {
    let resolveFetch: (v: UserChapterView | null) => void = () => {};
    mockLastViewed.mockReturnValue(
      new Promise<UserChapterView | null>((resolve) => {
        resolveFetch = resolve;
      }),
    );
    const params = baseParams();
    const { rerender } = renderHook((p: Params) => useChapterResume(p), {
      initialProps: params,
    });
    await waitFor(() => expect(mockLastViewed).toHaveBeenCalledTimes(1));

    // fetch 未解決のままユーザーが章 13 を手動選択した状態を再現。
    rerender({ ...params, selectedId: 13 });
    resolveFetch(view(12));
    await new Promise((r) => setTimeout(r, 10));
    expect(params.selectMaterial).not.toHaveBeenCalled();
  });

  it('一覧がまだ前のコースのものである間は発火しない', async () => {
    // courseId=6 に切り替わったが materials は courseId=5 のものが残っている commit。
    const params = baseParams({ courseId: 6 });
    renderHook(() => useChapterResume(params));
    await new Promise((r) => setTimeout(r, 10));
    expect(mockLastViewed).not.toHaveBeenCalled();
  });

  it('取得中に別コースへ切り替わったら前コースの応答では選択しない(stale ガード)', async () => {
    let resolveCourse5: (v: UserChapterView | null) => void = () => {};
    mockLastViewed.mockImplementationOnce(
      () =>
        new Promise<UserChapterView | null>((resolve) => {
          resolveCourse5 = resolve;
        }),
    );
    const selectMaterial = vi.fn();
    const { rerender } = renderHook((p: Params) => useChapterResume(p), {
      initialProps: baseParams({ selectMaterial }), // courseId=5
    });
    await waitFor(() => expect(mockLastViewed).toHaveBeenCalledWith(5));

    // コース 6 へ遷移(選択リセット + 新コースの一覧ロード完了)。コース 6 の履歴は即 null。
    mockLastViewed.mockResolvedValue(null);
    rerender(
      baseParams({
        courseId: 6,
        materials: [material(61, 6), material(62, 6)],
        selectedId: null,
        selectMaterial,
      }),
    );
    await waitFor(() => expect(mockLastViewed).toHaveBeenCalledWith(6));

    // 遅れて届いたコース 5 の応答はコース 6 の画面を上書きしない。
    resolveCourse5(view(12, 5));
    await new Promise((r) => setTimeout(r, 10));
    expect(selectMaterial).not.toHaveBeenCalledWith(12);
    // コース 6 側のレジューム(履歴なし → 先頭章)は正常に動く。
    await waitFor(() => expect(selectMaterial).toHaveBeenCalledWith(61));
  });
});
