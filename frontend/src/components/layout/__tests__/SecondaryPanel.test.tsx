import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import SecondaryPanel from '../SecondaryPanel';

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
});
