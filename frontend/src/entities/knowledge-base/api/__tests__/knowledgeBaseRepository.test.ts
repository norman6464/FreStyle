import { describe, it, expect, vi, beforeEach } from 'vitest';
import KnowledgeBaseRepository from '../knowledgeBaseRepository';
import apiClient from '@/shared/api/axios';

vi.mock('@/shared/api/axios');

const mockGet = vi.mocked(apiClient.get);
const mockPost = vi.mocked(apiClient.post);
const mockPatch = vi.mocked(apiClient.patch);

beforeEach(() => {
  vi.clearAllMocks();
});

describe('KnowledgeBaseRepository', () => {
  it('fetchWorkspaces は GET /kb/workspaces で配列を返す', async () => {
    mockGet.mockResolvedValue({ data: [{ slug: 'acme', name: 'Acme 社', createdAt: '2026-08-01T00:00:00Z' }] });

    const list = await KnowledgeBaseRepository.fetchWorkspaces();

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces');
    expect(list).toHaveLength(1);
  });

  it('fetchSpaces は slug を URL に埋める', async () => {
    mockGet.mockResolvedValue({ data: [] });

    await KnowledgeBaseRepository.fetchSpaces('acme');

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces');
  });

  it('一覧が null で返っても空配列にする', async () => {
    // 0 件を null で返されると map / for-of が落ちて画面が開けなくなる。
    mockGet.mockResolvedValue({ data: null });

    await expect(KnowledgeBaseRepository.fetchWorkspaces()).resolves.toEqual([]);
    await expect(KnowledgeBaseRepository.fetchSpaces('acme')).resolves.toEqual([]);
  });

  describe('fetchPageTree', () => {
    it('slug と spaceId を URL に埋める', async () => {
      mockGet.mockResolvedValue({ data: { pages: [], hasHiddenChildren: false } });

      await KnowledgeBaseRepository.fetchPageTree('acme', 'space-1');

      expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/space-1/pages', {
        params: undefined,
      });
    });

    it('archived を渡すとスコープを切り替える（別の口ではなく同じ口）', async () => {
      mockGet.mockResolvedValue({ data: { pages: [], hasHiddenChildren: false } });

      await KnowledgeBaseRepository.fetchPageTree('acme', 's1', { archived: true });

      expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s1/pages', {
        params: { archived: 'true' },
      });
    });

    it('hasHiddenChildren をそのまま通す', async () => {
      mockGet.mockResolvedValue({ data: { pages: [], hasHiddenChildren: true } });

      await expect(KnowledgeBaseRepository.fetchPageTree('acme', 's1')).resolves.toEqual({
        pages: [],
        hasHiddenChildren: true,
      });
    });

    it.each([
      ['項目が無い', { pages: [] }],
      ['null', { pages: [], hasHiddenChildren: null }],
      ['応答そのものが空', undefined],
    ])('hasHiddenChildren が %s なら false に倒す', async (_label, data) => {
      // 倒す向きが逆だと、印が出るはずのない場面で「表示できないページがあります」が出る。
      mockGet.mockResolvedValue({ data });

      const tree = await KnowledgeBaseRepository.fetchPageTree('acme', 's1');

      expect(tree.hasHiddenChildren).toBe(false);
    });

    it('pages が欠けていても空配列にする', async () => {
      mockGet.mockResolvedValue({ data: { hasHiddenChildren: false } });

      await expect(KnowledgeBaseRepository.fetchPageTree('acme', 's1')).resolves.toEqual({
        pages: [],
        hasHiddenChildren: false,
      });
    });
  });

  describe('createPage', () => {
    it('parentId を省くと空文字で送る（backend は空文字を「親なし」として扱う）', async () => {
      mockPost.mockResolvedValue({ data: { id: 'new-1' } });

      await KnowledgeBaseRepository.createPage('acme', 's1', { title: '無題' });

      expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s1/pages', {
        title: '無題',
        parentId: '',
      });
    });

    it('parentId を渡すとそのまま送る', async () => {
      mockPost.mockResolvedValue({ data: { id: 'new-1' } });

      await KnowledgeBaseRepository.createPage('acme', 's1', { title: '無題', parentId: 'p1' });

      expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s1/pages', {
        title: '無題',
        parentId: 'p1',
      });
    });

    it('失敗は握り潰さず投げる', async () => {
      // 握り潰して null や false を返すと、呼び出し側は失敗を知りようがない。
      // このリポジトリの「失敗したのに成功の表示」はどれもここが原因だった。
      mockPost.mockRejectedValue(new Error('boom'));

      await expect(
        KnowledgeBaseRepository.createPage('acme', 's1', { title: '無題' }),
      ).rejects.toThrow();
    });
  });

  describe('renamePage', () => {
    it('PATCH で題名だけを送る', async () => {
      mockPatch.mockResolvedValue({ data: { id: 'p1', title: '新しい名前' } });

      await KnowledgeBaseRepository.renamePage('acme', 'p1', '新しい名前');

      expect(mockPatch).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/pages/p1', {
        title: '新しい名前',
      });
    });

    it('失敗は握り潰さず投げる', async () => {
      mockPatch.mockRejectedValue(new Error('boom'));

      await expect(KnowledgeBaseRepository.renamePage('acme', 'p1', 'x')).rejects.toThrow();
    });
  });

  it('fetchPage は slug と pageId を URL に埋める', async () => {
    mockGet.mockResolvedValue({ data: { page: { id: 'p1' }, doc: { type: 'doc' } } });

    await KnowledgeBaseRepository.fetchPage('acme', 'p1');

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/pages/p1');
  });
});
