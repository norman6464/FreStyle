import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import NotePageTitle from '../NotePageTitle';

describe('NotePageTitle', () => {
  it('編集できない人には見出しとして出す（入力欄を出さない）', () => {
    render(<NotePageTitle title="設計メモ" canEdit={false} onRename={vi.fn()} />);

    expect(screen.getByRole('heading', { name: '設計メモ' })).toBeInTheDocument();
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
  });

  it('書き換えて Enter で確定すると onRename が呼ばれる', async () => {
    const onRename = vi.fn().mockResolvedValue(undefined);
    render(<NotePageTitle title="無題" canEdit onRename={onRename} />);

    const input = screen.getByRole('textbox', { name: 'ページの題名' });
    fireEvent.change(input, { target: { value: '設計メモ' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => expect(onRename).toHaveBeenCalledWith('設計メモ'));
  });

  it('欄外へ出ても（blur）確定する', async () => {
    const onRename = vi.fn().mockResolvedValue(undefined);
    render(<NotePageTitle title="無題" canEdit onRename={onRename} />);

    const input = screen.getByRole('textbox', { name: 'ページの題名' });
    fireEvent.change(input, { target: { value: '議事録' } });
    fireEvent.blur(input);

    await waitFor(() => expect(onRename).toHaveBeenCalledWith('議事録'));
  });

  it('空にして確定しても改名しない（元の題名に戻す）', async () => {
    const onRename = vi.fn();
    render(<NotePageTitle title="設計メモ" canEdit onRename={onRename} />);

    const input = screen.getByRole('textbox', { name: 'ページの題名' });
    fireEvent.change(input, { target: { value: '   ' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => expect(input).toHaveValue('設計メモ'));
    expect(onRename).not.toHaveBeenCalled();
  });

  it('同じ題名のまま確定しても呼ばない（無駄な往復をしない）', async () => {
    const onRename = vi.fn();
    render(<NotePageTitle title="設計メモ" canEdit onRename={onRename} />);

    const input = screen.getByRole('textbox', { name: 'ページの題名' });
    fireEvent.change(input, { target: { value: '設計メモ' } });
    fireEvent.blur(input);

    expect(onRename).not.toHaveBeenCalled();
  });

  it('Escape は打ちかけを捨てて元に戻し、直後の blur でも確定しない', async () => {
    const onRename = vi.fn();
    render(<NotePageTitle title="設計メモ" canEdit onRename={onRename} />);

    const input = screen.getByRole('textbox', { name: 'ページの題名' });
    fireEvent.change(input, { target: { value: '打ちかけ' } });
    fireEvent.keyDown(input, { key: 'Escape' });
    // Escape はフォーカスを外す実装なので、続けて起きる blur を再現する。
    fireEvent.blur(input);

    await waitFor(() => expect(input).toHaveValue('設計メモ'));
    expect(onRename).not.toHaveBeenCalled();
  });

  it('確定が終わるまで欄は無効になり、重ねて確定しても 1 回しか呼ばない', async () => {
    let resolveRename: () => void = () => {};
    const onRename = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveRename = resolve;
        }),
    );
    render(<NotePageTitle title="無題" canEdit onRename={onRename} />);

    const input = screen.getByRole('textbox', { name: 'ページの題名' });
    fireEvent.change(input, { target: { value: '設計メモ' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => expect(input).toBeDisabled());
    fireEvent.keyDown(input, { key: 'Enter' });
    fireEvent.blur(input);
    expect(onRename).toHaveBeenCalledTimes(1);

    resolveRename();
    await waitFor(() => expect(input).not.toBeDisabled());
  });

  it('日本語入力の変換確定 Enter では確定しない（isComposing を見る）', () => {
    const onRename = vi.fn();
    render(<NotePageTitle title="無題" canEdit onRename={onRename} />);

    const input = screen.getByRole('textbox', { name: 'ページの題名' });
    fireEvent.change(input, { target: { value: '設計' } });
    // 変換確定の Enter（isComposing=true）。ここで確定すると打ちかけの題名で改名が飛ぶ。
    fireEvent.keyDown(input, { key: 'Enter', isComposing: true });

    expect(onRename).not.toHaveBeenCalled();
  });

  it('失敗しても入力は消さない（打ち直しにさせない）', async () => {
    const onRename = vi.fn().mockRejectedValue(new Error('boom'));
    render(<NotePageTitle title="無題" canEdit onRename={onRename} />);

    const input = screen.getByRole('textbox', { name: 'ページの題名' });
    fireEvent.change(input, { target: { value: '設計メモ' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => expect(onRename).toHaveBeenCalled());
    expect(input).toHaveValue('設計メモ');
  });
});

describe('Enter で本文へ移る合図（onEnter）', () => {
  it('Enter で確定したとき onEnter が呼ばれる（題名が変わっていなくても）', () => {
    const onEnter = vi.fn();
    render(
      <NotePageTitle title="設計メモ" canEdit onRename={vi.fn()} onEnter={onEnter} />,
    );

    fireEvent.keyDown(screen.getByRole('textbox', { name: 'ページの題名' }), { key: 'Enter' });

    expect(onEnter).toHaveBeenCalledTimes(1);
  });

  it('改名が走る場合も、失敗した場合も onEnter は 1 回だけ呼ばれる', async () => {
    // 題名を変えずに Enter を押すと commit は何もせずに終わる。改名が実際に走る
    // 経路（そして失敗する経路）でも、本文へ移る合図は同じように 1 回出る。
    const onRename = vi.fn().mockRejectedValue(new Error('down'));
    const onEnter = vi.fn();
    render(
      <NotePageTitle title="設計メモ" canEdit onRename={onRename} onEnter={onEnter} />,
    );
    const input = screen.getByRole('textbox', { name: 'ページの題名' });

    fireEvent.change(input, { target: { value: '設計メモ（改訂）' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => expect(onRename).toHaveBeenCalledWith('設計メモ（改訂）'));
    expect(onEnter).toHaveBeenCalledTimes(1);
    // 失敗しても打ちかけは残す（打ち直しにさせない）。
    expect(input).toHaveValue('設計メモ（改訂）');
  });

  it('変換中の Enter は keyCode だけでも確定にしない（isComposing を持たない環境）', () => {
    // Safari は変換中に isComposing を立てず keyCode 229 だけを送る。
    const onRename = vi.fn();
    const onEnter = vi.fn();
    render(
      <NotePageTitle title="設計メモ" canEdit onRename={onRename} onEnter={onEnter} />,
    );

    fireEvent.keyDown(screen.getByRole('textbox', { name: 'ページの題名' }), {
      key: 'Enter',
      isComposing: false,
      keyCode: 229,
    });

    expect(onEnter).not.toHaveBeenCalled();
    expect(onRename).not.toHaveBeenCalled();
  });

  it('日本語入力の変換確定 Enter では呼ばれない', () => {
    const onEnter = vi.fn();
    render(
      <NotePageTitle title="設計メモ" canEdit onRename={vi.fn()} onEnter={onEnter} />,
    );

    fireEvent.keyDown(screen.getByRole('textbox', { name: 'ページの題名' }), {
      key: 'Enter',
      isComposing: true,
      keyCode: 229,
    });

    expect(onEnter).not.toHaveBeenCalled();
  });

  it('Escape では呼ばれない', () => {
    const onEnter = vi.fn();
    render(
      <NotePageTitle title="設計メモ" canEdit onRename={vi.fn()} onEnter={onEnter} />,
    );

    fireEvent.keyDown(screen.getByRole('textbox', { name: 'ページの題名' }), { key: 'Escape' });

    expect(onEnter).not.toHaveBeenCalled();
  });
});
