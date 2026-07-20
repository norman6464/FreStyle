import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import SecondaryPanel from '../ui/SecondaryPanel';

describe('SecondaryPanel', () => {
  it('タイトルを表示する', () => {
    render(
      <SecondaryPanel title="チャット">
        <div>コンテンツ</div>
      </SecondaryPanel>
    );
    const titles = screen.getAllByText('チャット');
    expect(titles.length).toBeGreaterThanOrEqual(1);
  });

  it('子要素を表示する', () => {
    render(
      <SecondaryPanel title="テスト">
        <div>子要素コンテンツ</div>
      </SecondaryPanel>
    );
    const elements = screen.getAllByText('子要素コンテンツ');
    expect(elements.length).toBeGreaterThanOrEqual(1);
  });

  it('ヘッダーコンテンツを表示する', () => {
    render(
      <SecondaryPanel title="テスト" headerContent={<input placeholder="検索" />}>
        <div>内容</div>
      </SecondaryPanel>
    );
    const inputs = screen.getAllByPlaceholderText('検索');
    expect(inputs.length).toBeGreaterThanOrEqual(1);
  });

  it('モバイル閉じるボタンを表示する', () => {
    render(
      <SecondaryPanel title="テスト" mobileOpen={true} onMobileClose={() => {}}>
        <div>内容</div>
      </SecondaryPanel>
    );
    expect(screen.getByLabelText('パネルを閉じる')).toBeDefined();
  });

  it('collapsible のとき折りたたみボタンを出し、クリックでトグルを呼ぶ', () => {
    const onToggle = vi.fn();
    render(
      <SecondaryPanel title="テスト" collapsible collapsed={false} onToggleCollapsed={onToggle}>
        <div>内容</div>
      </SecondaryPanel>
    );
    const btn = screen.getByLabelText('パネルを折りたたむ');
    fireEvent.click(btn);
    expect(onToggle).toHaveBeenCalledTimes(1);
  });

  it('折りたたみ中は「開く」ボタンを出し、折りたたみボタンは出さない', () => {
    const onToggle = vi.fn();
    render(
      <SecondaryPanel title="テスト" collapsible collapsed onToggleCollapsed={onToggle}>
        <div>章リスト内容</div>
      </SecondaryPanel>
    );
    // 折りたたみ中はデスクトップの全幅パネル（折りたたむボタン）を描画しない。
    expect(screen.queryByLabelText('パネルを折りたたむ')).not.toBeInTheDocument();
    fireEvent.click(screen.getByLabelText('パネルを開く'));
    expect(onToggle).toHaveBeenCalledTimes(1);
  });
});
