import { describe, it, expect } from 'vitest';

import { extractApiErrorMessage } from '../lib/extractApiErrorMessage';

describe('extractApiErrorMessage', () => {
  it('message があればそれを返す', () => {
    const err = { response: { data: { message: '本文', error: 'code' } } };
    expect(extractApiErrorMessage(err, '既定')).toBe('本文');
  });

  it('message が無いときは error を返す', () => {
    const err = { response: { data: { error: 'code' } } };
    expect(extractApiErrorMessage(err, '既定')).toBe('code');
  });

  it('空文字は本文とみなさず次の候補へ倒す', () => {
    const err = { response: { data: { message: '', error: 'code' } } };
    expect(extractApiErrorMessage(err, '既定')).toBe('code');
  });

  it('レスポンス本文が無いときは fallback を返す', () => {
    expect(extractApiErrorMessage(new Error('boom'), '既定')).toBe('既定');
    expect(extractApiErrorMessage(undefined, '既定')).toBe('既定');
    expect(extractApiErrorMessage({ response: {} }, '既定')).toBe('既定');
  });
});
