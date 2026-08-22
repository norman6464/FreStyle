import { describe, it, expect } from 'vitest';
import { toRichDocumentSummary } from '../toSummary';
import type { RichDocument } from '../../model/types';

describe('toRichDocumentSummary', () => {
  it('doc 本体を除いたサマリを返す', () => {
    const document: RichDocument = {
      id: 'a',
      ownerId: 7,
      kind: 'note',
      title: 'メモ',
      isPublic: false,
      schemaVersion: 1,
      revision: 3,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-02T00:00:00Z',
      doc: { type: 'doc', content: [{ type: 'paragraph' }] },
    };
    const summary = toRichDocumentSummary(document);
    expect(summary).not.toHaveProperty('doc');
    expect(summary).toEqual({
      id: 'a',
      ownerId: 7,
      kind: 'note',
      title: 'メモ',
      isPublic: false,
      schemaVersion: 1,
      revision: 3,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-02T00:00:00Z',
    });
  });
});
