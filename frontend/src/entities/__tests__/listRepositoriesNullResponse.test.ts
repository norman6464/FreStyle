import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/shared/api/axios', () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn(), patch: vi.fn() },
}));

import apiClient from '@/shared/api/axios';
// 各 Slice の Public API（index.ts）経由で参照する（FSD の境界ルール / CLAUDE.md §2.5）。
import { AuditRepository } from '@/entities/audit';
import { CompanyRepository } from '@/entities/company';
import { CourseRepository, LessonProgressRepository } from '@/entities/course';
import { ExerciseRepository } from '@/entities/exercise';
import { AdminInvitationRepository } from '@/entities/invitation';
import { AdminMemberRepository } from '@/entities/member';
import { NotificationRepository } from '@/entities/notification';

const mockGet = vi.mocked(apiClient.get);

/**
 * 一覧 API が 0 件のとき null を返しても、フロントが必ず配列を受け取ることを保証する
 * （FRESTYLE-77）。null がそのまま流れると map / filter / for-of が TypeError で落ち、
 * データがまだ無い新規ユーザーがそのページを開けなくなる。
 *
 * backend 側でも空配列を保証しているが、片側だけの対策では
 * 「もう一方が壊れた瞬間にユーザーへ影響が出る」ため両側で守る。
 *
 * 正規化している 11 経路をすべて網羅する（1 つでも漏れるとそこだけ無防備になる）。
 */
describe('一覧 repository は null 応答でも配列を返す', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const cases: ReadonlyArray<readonly [string, () => Promise<unknown[]>]> = [
    ['監査ログ一覧', () => AuditRepository.list()],
    ['会社一覧', () => CompanyRepository.list()],
    ['会社の統計一覧', () => CompanyRepository.listStats()],
    ['コース一覧', () => CourseRepository.list()],
    ['教材一覧', () => CourseRepository.listMaterials(1)],
    ['章の進捗一覧', () => LessonProgressRepository.list()],
    ['演習の言語別集計', () => ExerciseRepository.listLanguageSummary()],
    ['演習の提出履歴', () => ExerciseRepository.listSubmissions(1)],
    ['招待一覧', () => AdminInvitationRepository.list()],
    ['従業員一覧', () => AdminMemberRepository.listMembers()],
    ['通知一覧', () => NotificationRepository.getAll()],
  ];

  it('正規化対象の全経路を網羅している', () => {
    expect(cases).toHaveLength(11);
  });

  it.each(cases)('%s は null でも空配列になる', async (_name, call) => {
    mockGet.mockResolvedValue({ data: null });

    const rows = await call();

    expect(Array.isArray(rows)).toBe(true);
    expect(rows).toEqual([]);
    // 実際にクラッシュしていた操作を通す。
    expect(() => rows.map((r) => r)).not.toThrow();
  });

  it.each(cases)('%s は undefined でも空配列になる', async (_name, call) => {
    mockGet.mockResolvedValue({ data: undefined });

    const rows = await call();

    expect(rows).toEqual([]);
  });

  it.each(cases)('%s は正常な配列をそのまま通す', async (_name, call) => {
    const payload = [{ id: 1 }, { id: 2 }];
    mockGet.mockResolvedValue({ data: payload });

    const rows = await call();

    // 内容も参照も変えずに素通しすること（余計な変換やコピーをしていない）。
    expect(rows).toBe(payload);
    expect(rows).toEqual([{ id: 1 }, { id: 2 }]);
  });
});
