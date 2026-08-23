import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { createRef } from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { Course, TeachingMaterial } from '@/entities/course';
import type { RichDocContent } from '@/shared/ui/RichTextEditor';
import ReadOnlyDetail from '../ui/ReadOnlyDetail';
import DocTableOfContents from '../ui/DocTableOfContents';
import { stripLeadingDocTitle } from '../lib/stripLeadingDocTitle';
import { extractDocHeadings } from '../model/docHeadings';

// jsdom に IntersectionObserver が無いため、no-op スタブを差し込む
// （目次のハイライトはスクロール依存のためユニットテストでは検証しない）。
class FakeIntersectionObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
const originalIO = globalThis.IntersectionObserver;
beforeAll(() => {
  globalThis.IntersectionObserver = FakeIntersectionObserver as unknown as typeof IntersectionObserver;
});
afterAll(() => {
  globalThis.IntersectionObserver = originalIO;
});

const course: Course = {
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

const richDoc: RichDocContent = {
  type: 'doc',
  content: [
    { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: '章タイトル' }] },
    { type: 'heading', attrs: { level: 2 }, content: [{ type: 'text', text: 'SELECT の基本' }] },
    { type: 'paragraph', content: [{ type: 'text', text: 'リッチ本文の段落です。' }] },
    { type: 'heading', attrs: { level: 2 }, content: [{ type: 'text', text: '演習' }] },
    { type: 'paragraph', content: [{ type: 'image', attrs: { src: 'https://example.com/er.png', alt: 'ER図' } }] },
  ],
};

function material(overrides: Partial<TeachingMaterial> = {}): TeachingMaterial {
  return {
    id: 1,
    companyId: 10,
    courseId: 5,
    createdByUserId: 1,
    title: '章タイトル',
    orderInCourse: 1,
    isPublished: true,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-08T00:00:00Z',
    ...overrides,
  };
}

function renderDetail(m: TeachingMaterial) {
  const articleRef = createRef<HTMLDivElement>();
  return render(
    <MemoryRouter>
      <ReadOnlyDetail
        material={m}
        bodyDoc={m.doc ? stripLeadingDocTitle(m.doc) : null}
        articleRef={articleRef}
        completed={false}
        onToggleComplete={() => {}}
      />
    </MemoryRouter>,
  );
}

describe('ReadOnlyDetail の doc（tiptap）表示 (FRESTYLE-338)', () => {
  it('doc がある章は tiptap の読み取り専用表示になる', async () => {
    renderDetail(material({ doc: richDoc, revision: 1 }));
    // tiptap の描画（ProseMirror）で本文が出る。
    await waitFor(() => {
      expect(screen.getByText('リッチ本文の段落です。')).toBeInTheDocument();
    });
    expect(document.querySelector('.ProseMirror')).not.toBeNull();
    // 読み取り専用（contenteditable=false）。
    expect(document.querySelector('.ProseMirror')).toHaveAttribute('contenteditable', 'false');
  });

  it('先頭 h1 は本文から除かれる（ヘッダーのタイトルと二重にならない）', async () => {
    renderDetail(material({ doc: richDoc, revision: 1 }));
    await waitFor(() => expect(document.querySelector('.ProseMirror')).not.toBeNull());
    // ヘッダー（カード上部）の h1 はある。
    expect(screen.getByRole('heading', { level: 1, name: '章タイトル' })).toBeInTheDocument();
    // 本文（ProseMirror）内には h1 が無い。
    expect(document.querySelector('.ProseMirror h1')).toBeNull();
    expect(document.querySelector('.ProseMirror h2')).not.toBeNull();
  });

  it('目次(DocTableOfContents)は doc の見出しから生成され、本文コンテナの見出しに anchor id を振る', async () => {
    // 目次は左パネル(CourseDetailPage)に移設済みのため、単体で本文コンテナと組み合わせて検証する。
    const articleRef = createRef<HTMLDivElement>();
    const bodyDoc = stripLeadingDocTitle(richDoc);
    render(
      <>
        <div ref={articleRef}>
          <h2>SELECT の基本</h2>
          <h2>演習</h2>
        </div>
        <DocTableOfContents doc={bodyDoc} articleRef={articleRef} />
      </>,
    );
    const toc = await screen.findByRole('navigation', { name: '目次' });
    expect(toc).toHaveTextContent('SELECT の基本');
    expect(toc).toHaveTextContent('演習');
    // 先頭 h1（章タイトル）は除去済みの doc から生成するため目次に出ない。
    expect(toc.textContent).not.toContain('章タイトル');
    await waitFor(() => {
      expect(articleRef.current?.querySelector('h2')?.id).toBe('select-の基本');
    });
    expect(toc.querySelector("a[href='#select-の基本']")).not.toBeNull();
  });

  it('本文内の画像クリックでモーダル拡大表示が開く', async () => {
    renderDetail(material({ doc: richDoc, revision: 1 }));
    await waitFor(() => expect(document.querySelector('.ProseMirror img')).not.toBeNull());
    fireEvent.click(document.querySelector('.ProseMirror img')!);
    // ImageLightbox は dialog として開く。
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });
  });

  it('doc が無い章（本文未保存の新規章）は空 doc として描画する', async () => {
    renderDetail(material({ doc: null }));
    // ヘッダーのタイトルは出るが、本文は空の tiptap 描画になる。
    expect(screen.getByRole('heading', { level: 1, name: '章タイトル' })).toBeInTheDocument();
    await waitFor(() => expect(document.querySelector('.ProseMirror')).not.toBeNull());
    expect(document.querySelector('.ProseMirror')!.textContent).toBe('');
  });
});

describe('stripLeadingDocTitle', () => {
  it('先頭が h1 なら取り除く', () => {
    const out = stripLeadingDocTitle(richDoc);
    expect(out.content?.[0]?.type).toBe('heading');
    expect(out.content?.[0]?.attrs?.level).toBe(2);
  });

  it('先頭が h1 でなければそのまま', () => {
    const doc: RichDocContent = {
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: '本文' }] }],
    };
    expect(stripLeadingDocTitle(doc)).toEqual(doc);
  });

  it('h1 しか無い doc は空の段落 1 つになる', () => {
    const doc: RichDocContent = {
      type: 'doc',
      content: [{ type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: 'タイトル' }] }],
    };
    expect(stripLeadingDocTitle(doc).content).toEqual([{ type: 'paragraph' }]);
  });
});

describe('extractDocHeadings', () => {
  it('h1〜h3 を文書順に抽出し、github-slugger の id を付ける', () => {
    const items = extractDocHeadings(richDoc);
    expect(items.map((i) => [i.level, i.text])).toEqual([
      [1, '章タイトル'],
      [2, 'SELECT の基本'],
      [2, '演習'],
    ]);
    expect(items[1].id).toBe('select-の基本');
  });

  it('同名見出しは連番付き id になる（slugger の重複解決）', () => {
    const doc: RichDocContent = {
      type: 'doc',
      content: [
        { type: 'heading', attrs: { level: 2 }, content: [{ type: 'text', text: '演習' }] },
        { type: 'heading', attrs: { level: 2 }, content: [{ type: 'text', text: '演習' }] },
      ],
    };
    const items = extractDocHeadings(doc);
    expect(items[0].id).toBe('演習');
    expect(items[1].id).toBe('演習-1');
  });
});
