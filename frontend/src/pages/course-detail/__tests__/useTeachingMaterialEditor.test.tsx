import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { AxiosError, AxiosHeaders } from 'axios';
import type { TeachingMaterial } from '@/entities/course';
import type { RichDocContent } from '@/shared/ui/RichTextEditor';
import { useTeachingMaterialEditor } from '../model/useTeachingMaterialEditor';

vi.mock('@/entities/course/api/teachingMaterialRepository', () => ({
  default: {
    get: vi.fn(),
    updateDoc: vi.fn(),
  },
}));
import TeachingMaterialRepository from '@/entities/course/api/teachingMaterialRepository';

const mockUpdateDoc = vi.mocked(TeachingMaterialRepository.updateDoc);
const mockGet = vi.mocked(TeachingMaterialRepository.get);

const doc = (text: string): RichDocContent => ({
  type: 'doc',
  content: [{ type: 'paragraph', content: [{ type: 'text', text }] }],
});

function material(overrides: Partial<TeachingMaterial> = {}): TeachingMaterial {
  return {
    id: 1,
    companyId: 10,
    courseId: 5,
    createdByUserId: 1,
    title: '章',
    doc: doc('サーバの本文'),
    revision: 3,
    orderInCourse: 1,
    isPublished: true,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-08T00:00:00Z',
    ...overrides,
  };
}

function conflict409(): AxiosError {
  return new AxiosError('conflict', 'ERR_BAD_REQUEST', undefined, undefined, {
    status: 409,
    statusText: 'Conflict',
    headers: {},
    config: { headers: new AxiosHeaders() },
    data: { error: 'conflict' },
  });
}

beforeEach(() => {
  vi.useFakeTimers();
  mockUpdateDoc.mockReset();
  mockGet.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
});

function renderEditor(selected: TeachingMaterial, extra: { onConflict?: () => void; onDocSynced?: (m: TeachingMaterial) => void } = {}) {
  return renderHook(() =>
    useTeachingMaterialEditor({
      selectedId: selected.id,
      selected,
      update: vi.fn().mockResolvedValue(undefined),
      ...extra,
    }),
  );
}

describe('useTeachingMaterialEditor の doc 編集 (FRESTYLE-339 / FRESTYLE-347)', () => {
  it('doc がある章は初期値がサーバの doc で始まる', () => {
    const { result } = renderEditor(material());
    expect(result.current.editDoc).toEqual(doc('サーバの本文'));
  });

  it('doc が null の章（本文未保存の新規章）は空 doc で始まる', () => {
    const { result } = renderEditor(material({ doc: null, revision: 1 }));
    expect(result.current.editDoc).toEqual({ type: 'doc', content: [{ type: 'paragraph' }] });
  });

  it('handleDocChange の 800ms 後に expectedRevision 付きで保存され、revision が進む', async () => {
    mockUpdateDoc.mockResolvedValue(material({ revision: 4 }));
    const synced = vi.fn();
    const { result } = renderEditor(material(), { onDocSynced: synced });

    act(() => {
      result.current.handleDocChange(doc('編集した本文'));
    });
    expect(result.current.saveStatus).toBe('unsaved');
    await act(async () => {
      await vi.advanceTimersByTimeAsync(800);
    });
    expect(mockUpdateDoc).toHaveBeenCalledWith(1, {
      doc: doc('編集した本文'),
      expectedRevision: 3,
    });
    expect(result.current.saveStatus).toBe('saved');
    expect(synced).toHaveBeenCalledWith(material({ revision: 4 }));

    // 次の保存は進んだ revision（4）を送る。
    mockUpdateDoc.mockResolvedValue(material({ revision: 5 }));
    act(() => {
      result.current.handleDocChange(doc('さらに編集'));
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(800);
    });
    expect(mockUpdateDoc).toHaveBeenLastCalledWith(1, {
      doc: doc('さらに編集'),
      expectedRevision: 4,
    });
  });

  it('409 のときはサーバ最新版を取り直してエディタへ反映し、onConflict が呼ばれる', async () => {
    mockUpdateDoc.mockRejectedValue(conflict409());
    mockGet.mockResolvedValue(material({ doc: doc('他の人の最新版'), revision: 9, title: '更新後タイトル' }));
    const onConflict = vi.fn();
    const { result } = renderEditor(material(), { onConflict });

    act(() => {
      result.current.handleDocChange(doc('競合する編集'));
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(800);
    });
    expect(mockGet).toHaveBeenCalledWith(1);
    expect(result.current.editDoc).toEqual(doc('他の人の最新版'));
    expect(result.current.editTitle).toBe('更新後タイトル');
    expect(result.current.saveStatus).toBe('saved');
    expect(onConflict).toHaveBeenCalledTimes(1);

    // 取り直した revision（9）で次の保存が走る。
    mockUpdateDoc.mockReset();
    mockUpdateDoc.mockResolvedValue(material({ revision: 10 }));
    act(() => {
      result.current.handleDocChange(doc('やり直しの編集'));
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(800);
    });
    expect(mockUpdateDoc).toHaveBeenCalledWith(1, {
      doc: doc('やり直しの編集'),
      expectedRevision: 9,
    });
  });

  it('409 以外の保存失敗は unsaved に戻す（再試行を促す）', async () => {
    mockUpdateDoc.mockRejectedValue(new Error('network'));
    const { result } = renderEditor(material());
    act(() => {
      result.current.handleDocChange(doc('保存に失敗する編集'));
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(800);
    });
    expect(result.current.saveStatus).toBe('unsaved');
  });
});
