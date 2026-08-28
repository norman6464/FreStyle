import { describe, it, expect, vi, beforeEach } from 'vitest';
import NoteRepository from '../noteRepository';
import apiClient from '@/shared/api/axios';

vi.mock('@/shared/api/axios');

const mockGet = vi.mocked(apiClient.get);
const mockPost = vi.mocked(apiClient.post);
const mockPatch = vi.mocked(apiClient.patch);
const mockPut = vi.mocked(apiClient.put);

beforeEach(() => {
  vi.clearAllMocks();
});

describe('NoteRepository', () => {
  it('fetchWorkspaces は GET /kb/workspaces で配列を返す', async () => {
    mockGet.mockResolvedValue({ data: [{ slug: 'acme', name: 'Acme 社', createdAt: '2026-08-01T00:00:00Z' }] });

    const list = await NoteRepository.fetchWorkspaces();

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces');
    expect(list).toHaveLength(1);
  });

  it('fetchSpaces は slug を URL に埋める', async () => {
    mockGet.mockResolvedValue({ data: [] });

    await NoteRepository.fetchSpaces('acme');

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces');
  });

  it('一覧が null で返っても空配列にする', async () => {
    // 0 件を null で返されると map / for-of が落ちて画面が開けなくなる。
    mockGet.mockResolvedValue({ data: null });

    await expect(NoteRepository.fetchWorkspaces()).resolves.toEqual([]);
    await expect(NoteRepository.fetchSpaces('acme')).resolves.toEqual([]);
  });

  describe('fetchPageTree', () => {
    it('slug と spaceId を URL に埋める', async () => {
      mockGet.mockResolvedValue({ data: { pages: [], hasHiddenChildren: false } });

      await NoteRepository.fetchPageTree('acme', 'space-1');

      expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/space-1/pages', {
        params: undefined,
      });
    });

    it('archived を渡すとスコープを切り替える（別の口ではなく同じ口）', async () => {
      mockGet.mockResolvedValue({ data: { pages: [], hasHiddenChildren: false } });

      await NoteRepository.fetchPageTree('acme', 's1', { archived: true });

      expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s1/pages', {
        params: { archived: 'true' },
      });
    });

    it('hasHiddenChildren をそのまま通す', async () => {
      mockGet.mockResolvedValue({ data: { pages: [], hasHiddenChildren: true } });

      await expect(NoteRepository.fetchPageTree('acme', 's1')).resolves.toEqual({
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

      const tree = await NoteRepository.fetchPageTree('acme', 's1');

      expect(tree.hasHiddenChildren).toBe(false);
    });

    it('pages が欠けていても空配列にする', async () => {
      mockGet.mockResolvedValue({ data: { hasHiddenChildren: false } });

      await expect(NoteRepository.fetchPageTree('acme', 's1')).resolves.toEqual({
        pages: [],
        hasHiddenChildren: false,
      });
    });
  });

  describe('createPage', () => {
    it('parentId を省くと空文字で送る（backend は空文字を「親なし」として扱う）', async () => {
      mockPost.mockResolvedValue({ data: { id: 'new-1' } });

      await NoteRepository.createPage('acme', 's1', { title: '無題' });

      expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s1/pages', {
        title: '無題',
        parentId: '',
      });
    });

    it('parentId を渡すとそのまま送る', async () => {
      mockPost.mockResolvedValue({ data: { id: 'new-1' } });

      await NoteRepository.createPage('acme', 's1', { title: '無題', parentId: 'p1' });

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
        NoteRepository.createPage('acme', 's1', { title: '無題' }),
      ).rejects.toThrow();
    });
  });

  describe('renamePage', () => {
    it('PATCH で題名だけを送る', async () => {
      mockPatch.mockResolvedValue({ data: { id: 'p1', title: '新しい名前' } });

      await NoteRepository.renamePage('acme', 'p1', '新しい名前');

      expect(mockPatch).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/pages/p1', {
        title: '新しい名前',
      });
    });

    it('失敗は握り潰さず投げる', async () => {
      mockPatch.mockRejectedValue(new Error('boom'));

      await expect(NoteRepository.renamePage('acme', 'p1', 'x')).rejects.toThrow();
    });
  });

  describe('archivePage / unarchivePage', () => {
    it('archivePage は POST /archive を叩く', async () => {
      mockPost.mockResolvedValue({ data: undefined });

      await NoteRepository.archivePage('acme', 'p1');

      expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/pages/p1/archive');
    });

    it('unarchivePage は POST /unarchive を叩き、戻ったページを返す', async () => {
      mockPost.mockResolvedValue({ data: { id: 'p1', title: '戻った' } });

      const restored = await NoteRepository.unarchivePage('acme', 'p1');

      expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/pages/p1/unarchive');
      expect(restored.title).toBe('戻った');
    });

    it('失敗は握り潰さず投げる', async () => {
      mockPost.mockRejectedValue(new Error('boom'));

      await expect(NoteRepository.archivePage('acme', 'p1')).rejects.toThrow();
      await expect(NoteRepository.unarchivePage('acme', 'p1')).rejects.toThrow();
    });
  });

  it('fetchPage は slug と pageId を URL に埋める', async () => {
    mockGet.mockResolvedValue({ data: { page: { id: 'p1' }, doc: { type: 'doc' } } });

    await NoteRepository.fetchPage('acme', 'p1');

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/pages/p1');
  });

  it('searchPages は GET /search に q を渡し、null 応答でも空配列にする', async () => {
    mockGet.mockResolvedValue({ data: null });

    await expect(NoteRepository.searchPages('acme', 'docker')).resolves.toEqual([]);
    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/search', {
      params: { q: 'docker' },
    });
  });

  it('searchPages は limit を渡したときだけ params に載せる', async () => {
    mockGet.mockResolvedValue({ data: [] });

    await NoteRepository.searchPages('acme', 'docker', 5);

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/search', {
      params: { q: 'docker', limit: 5 },
    });
  });

  it('renameSpace は PATCH /spaces/:id に name だけを送る', async () => {
    mockPatch.mockResolvedValue({
      data: { id: 'sp-1', key: 'eng', name: '技術部', createdAt: '2026-08-01T00:00:00Z' },
    });

    const space = await NoteRepository.renameSpace('acme', 'sp-1', '技術部');

    expect(mockPatch).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/sp-1', {
      name: '技術部',
    });
    expect(space.name).toBe('技術部');
  });

  it('renameSpace の失敗は投げる（握り潰すと成功の表示だけが残る）', async () => {
    mockPatch.mockRejectedValue(new Error('forbidden'));

    await expect(NoteRepository.renameSpace('acme', 'sp-1', 'x')).rejects.toThrow();
  });

  it('resolvePage は GET /kb/pages/:id で解決結果をそのまま返す', async () => {
    const resolved = {
      workspaceSlug: 'w-3f2a9c',
      page: { id: 'p-1', spaceId: 'sp-1', title: '設計メモ' },
      doc: { type: 'doc', content: [] },
      canEdit: true,
    };
    mockGet.mockResolvedValue({ data: resolved });

    const got = await NoteRepository.resolvePage('p-1');

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/pages/p-1');
    expect(got).toEqual(resolved);
  });

  it('resolvePage の失敗は投げる（404 は「無い」と「見えない」の両方）', async () => {
    mockGet.mockRejectedValue(new Error('not found'));

    await expect(NoteRepository.resolvePage('p-x')).rejects.toThrow();
  });

  it('replaceContent は PUT /pages/:id/content に { doc } を送り、正規形を返す', async () => {
    const normalized = { doc: { type: 'doc', content: [] }, builtAt: '2026-08-28T00:00:00Z' };
    mockPut.mockResolvedValue({ data: normalized });

    const got = await NoteRepository.replaceContent('acme', 'p-1', {
      type: 'doc',
      content: [{ type: 'paragraph' }],
    });

    expect(mockPut).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/pages/p-1/content', {
      doc: { type: 'doc', content: [{ type: 'paragraph' }] },
    });
    expect(got).toEqual(normalized);
  });

  it('replaceContent の失敗は投げる（呼び出し側が未保存へ戻すため）', async () => {
    mockPut.mockRejectedValue(new Error('conflict'));

    await expect(NoteRepository.replaceContent('acme', 'p-1', {})).rejects.toThrow();
  });
});
