import { describe, it, expect } from 'vitest';
import { computeSuggestionDiff, extractPlainText } from '../suggestionDiff';

describe('extractPlainText', () => {
  it('トップレベルの各ブロックを1行にして改行で繋ぐ', () => {
    const doc = {
      type: 'doc',
      content: [
        { type: 'heading', content: [{ type: 'text', text: '見出し' }] },
        { type: 'paragraph', content: [{ type: 'text', text: '本文の' }, { type: 'text', text: '段落' }] },
      ],
    };
    expect(extractPlainText(doc)).toBe('見出し\n本文の段落');
  });

  it('doc でない・content が無い値は空文字列', () => {
    expect(extractPlainText(null)).toBe('');
    expect(extractPlainText({ type: 'doc' })).toBe('');
    expect(extractPlainText('not a doc')).toBe('');
  });

  it('パースできないブロックはプレースホルダ行になる', () => {
    const doc = { type: 'doc', content: [null, { type: 'paragraph', content: [{ type: 'text', text: 'ok' }] }] };
    expect(extractPlainText(doc)).toBe('[変更あり]\nok');
  });

  it('テキストを持たないブロック（type だけ）は空行になる', () => {
    const doc = { type: 'doc', content: [{ type: 'horizontalRule' }] };
    expect(extractPlainText(doc)).toBe('');
  });
});

describe('computeSuggestionDiff', () => {
  it('baseDoc が無ければ追加行だけの差分になる', () => {
    const doc = { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: '新規の本文' }] }] };
    const diff = computeSuggestionDiff(undefined, doc);
    expect(diff).toEqual([{ type: 'added', text: '新規の本文' }]);
  });

  it('変わった行だけが追加・削除として色分けされ、同じ行は unchanged になる', () => {
    const base = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '変わらない行' }] },
        { type: 'paragraph', content: [{ type: 'text', text: '古い行' }] },
      ],
    };
    const doc = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '変わらない行' }] },
        { type: 'paragraph', content: [{ type: 'text', text: '新しい行' }] },
      ],
    };
    const diff = computeSuggestionDiff(base, doc);
    expect(diff).toContainEqual({ type: 'unchanged', text: '変わらない行' });
    expect(diff).toContainEqual({ type: 'removed', text: '古い行' });
    expect(diff).toContainEqual({ type: 'added', text: '新しい行' });
  });

  it('完全に同じ内容なら unchanged だけになる', () => {
    const doc = { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: '同じ' }] }] };
    const diff = computeSuggestionDiff(doc, doc);
    expect(diff).toEqual([{ type: 'unchanged', text: '同じ' }]);
  });
});
