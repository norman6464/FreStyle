import { describe, expect, it, beforeEach, afterEach, vi } from 'vitest';
import { createMockStorage } from '@/test/mockStorage';
import { rememberVisitedPage, getLastVisitedPageId, forgetVisitedPageIfMatches } from '../lastVisitedPage';

describe('lastVisitedPage', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('覚えたページを読み出せる', () => {
    expect(getLastVisitedPageId()).toBeNull();
    rememberVisitedPage('p1');
    expect(getLastVisitedPageId()).toBe('p1');
  });

  it('後から開いたページで上書きされる', () => {
    rememberVisitedPage('p1');
    rememberVisitedPage('p2');
    expect(getLastVisitedPageId()).toBe('p2');
  });

  it('一致するときだけ忘れる', () => {
    rememberVisitedPage('p1');
    forgetVisitedPageIfMatches('unrelated');
    expect(getLastVisitedPageId()).toBe('p1');

    forgetVisitedPageIfMatches('p1');
    expect(getLastVisitedPageId()).toBeNull();
  });

  it('localStorage が使えなくても例外を投げない', () => {
    vi.stubGlobal('localStorage', {
      getItem: () => {
        throw new Error('blocked');
      },
      setItem: () => {
        throw new Error('blocked');
      },
      removeItem: () => {
        throw new Error('blocked');
      },
    });

    expect(() => rememberVisitedPage('p1')).not.toThrow();
    expect(getLastVisitedPageId()).toBeNull();
    expect(() => forgetVisitedPageIfMatches('p1')).not.toThrow();
  });
});
