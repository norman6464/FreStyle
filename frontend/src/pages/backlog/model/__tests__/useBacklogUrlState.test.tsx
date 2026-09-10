import { act, renderHook } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { MemoryRouter, useLocation } from 'react-router-dom';
import type { ReactNode } from 'react';
import { useBacklogUrlState } from '../useBacklogUrlState';

function wrapperAt(initial: string) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <MemoryRouter initialEntries={[initial]}>{children}</MemoryRouter>;
  };
}

function renderAt(initial: string) {
  return renderHook(
    () => ({ state: useBacklogUrlState(), search: useLocation().search }),
    { wrapper: wrapperAt(initial) },
  );
}

describe('useBacklogUrlState', () => {
  it('何も付いていない URL は既定（チケットの面・現役・未選択）', () => {
    const { result } = renderAt('/kb/backlog/s-1');
    expect(result.current.state).toMatchObject({ tab: 'tickets', archived: false, selectedId: null });
  });

  it('URL から面・アーカイブ・選択を読む', () => {
    const { result } = renderAt('/kb/backlog/s-1?tab=statuses&archived=1&ticket=t-9');
    expect(result.current.state).toMatchObject({ tab: 'statuses', archived: true, selectedId: 't-9' });
  });

  it('知らない面の名前は既定に落とす', () => {
    const { result } = renderAt('/kb/backlog/s-1?tab=nonsense');
    expect(result.current.state.tab).toBe('tickets');
  });

  it('選んだチケットを URL に載せ、他の項目は残す', () => {
    const { result } = renderAt('/kb/backlog/s-1?archived=1');
    act(() => result.current.state.selectTicket('t-9'));
    expect(result.current.search).toContain('ticket=t-9');
    expect(result.current.search).toContain('archived=1');
  });

  it('既定の値は URL に書かない', () => {
    const { result } = renderAt('/kb/backlog/s-1?tab=types&archived=1&ticket=t-9');
    act(() => result.current.state.setTab('tickets'));
    act(() => result.current.state.setArchived(false));
    act(() => result.current.state.selectTicket(null));
    expect(result.current.search).toBe('');
  });

  it('スペースを移ったときは文脈ごと捨てる', () => {
    const { result } = renderAt('/kb/backlog/s-1?tab=statuses&archived=1&ticket=t-9');
    act(() => result.current.state.reset());
    expect(result.current.search).toBe('');
  });

  it('チケット以外のクエリには触らない', () => {
    const { result } = renderAt('/kb/backlog/s-1?from=notification');
    act(() => result.current.state.selectTicket('t-9'));
    expect(result.current.search).toContain('from=notification');
  });
});
