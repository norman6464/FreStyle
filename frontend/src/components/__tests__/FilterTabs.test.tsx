import { render, screen, fireEvent } from '@testing-library/react';
import FilterTabs from '../FilterTabs';

const TABS = ['すべて', 'カテゴリA', 'カテゴリB'] as const;

describe('FilterTabs', () => {
  it('すべてのタブが表示される', () => {
    render(<FilterTabs tabs={[...TABS]} selected="すべて" onSelect={() => {}} />);
    TABS.forEach((tab) => {
      expect(screen.getByRole('tab', { name: tab })).toBeInTheDocument();
    });
  });

  it('選択中のタブにアクティブスタイルが適用される', () => {
    render(<FilterTabs tabs={[...TABS]} selected="カテゴリA" onSelect={() => {}} />);
    const active = screen.getByRole('tab', { name: 'カテゴリA' });
    expect(active.className).toContain('border-primary-500');
    expect(active.className).toContain('text-primary-400');
  });

  it('非選択のタブに非アクティブスタイルが適用される', () => {
    render(<FilterTabs tabs={[...TABS]} selected="カテゴリA" onSelect={() => {}} />);
    const inactive = screen.getByRole('tab', { name: 'すべて' });
    expect(inactive.className).toContain('border-transparent');
  });

  it('タブクリックでonSelectが呼ばれる', () => {
    const onSelect = vi.fn();
    render(<FilterTabs tabs={[...TABS]} selected="すべて" onSelect={onSelect} />);
    fireEvent.click(screen.getByRole('tab', { name: 'カテゴリB' }));
    expect(onSelect).toHaveBeenCalledWith('カテゴリB');
  });

  it('classNameを追加できる', () => {
    const { container } = render(
      <FilterTabs tabs={[...TABS]} selected="すべて" onSelect={() => {}} className="mb-5" />
    );
    expect(container.firstChild).toHaveClass('mb-5');
  });

  it('classNameが未指定でもエラーにならない', () => {
    const { container } = render(
      <FilterTabs tabs={[...TABS]} selected="すべて" onSelect={() => {}} />
    );
    expect(container.firstChild).toHaveClass('flex');
  });

  it('選択中のタブを再クリックしてもonSelectが呼ばれる', () => {
    const onSelect = vi.fn();
    render(<FilterTabs tabs={[...TABS]} selected="すべて" onSelect={onSelect} />);
    fireEvent.click(screen.getByRole('tab', { name: 'すべて' }));
    expect(onSelect).toHaveBeenCalledWith('すべて');
  });

  it('タブが1つでも正常にレンダリングされる', () => {
    render(<FilterTabs tabs={['唯一']} selected="唯一" onSelect={() => {}} />);
    expect(screen.getByRole('tab', { name: '唯一' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '唯一' }).className).toContain('border-primary-500');
  });

  it('ボーダー付きのコンテナがレンダリングされる', () => {
    const { container } = render(
      <FilterTabs tabs={[...TABS]} selected="すべて" onSelect={() => {}} />
    );
    expect(container.firstChild).toHaveClass('border-b');
    expect(container.firstChild).toHaveClass('border-surface-3');
  });

  it('右矢印キーで次のタブにフォーカスが移動する', () => {
    const onSelect = vi.fn();
    render(<FilterTabs tabs={[...TABS]} selected="すべて" onSelect={onSelect} />);
    const firstTab = screen.getByRole('tab', { name: 'すべて' });
    fireEvent.keyDown(firstTab, { key: 'ArrowRight' });
    expect(onSelect).toHaveBeenCalledWith('カテゴリA');
  });

  it('左矢印キーで前のタブにフォーカスが移動する', () => {
    const onSelect = vi.fn();
    render(<FilterTabs tabs={[...TABS]} selected="カテゴリA" onSelect={onSelect} />);
    const tab = screen.getByRole('tab', { name: 'カテゴリA' });
    fireEvent.keyDown(tab, { key: 'ArrowLeft' });
    expect(onSelect).toHaveBeenCalledWith('すべて');
  });

  it('最後のタブで右矢印キーを押すと最初のタブに移動する', () => {
    const onSelect = vi.fn();
    render(<FilterTabs tabs={[...TABS]} selected="カテゴリB" onSelect={onSelect} />);
    const lastTab = screen.getByRole('tab', { name: 'カテゴリB' });
    fireEvent.keyDown(lastTab, { key: 'ArrowRight' });
    expect(onSelect).toHaveBeenCalledWith('すべて');
  });

  it('最初のタブで左矢印キーを押すと最後のタブに移動する', () => {
    const onSelect = vi.fn();
    render(<FilterTabs tabs={[...TABS]} selected="すべて" onSelect={onSelect} />);
    const firstTab = screen.getByRole('tab', { name: 'すべて' });
    fireEvent.keyDown(firstTab, { key: 'ArrowLeft' });
    expect(onSelect).toHaveBeenCalledWith('カテゴリB');
  });

  it('Homeキーで最初のタブに移動する', () => {
    const onSelect = vi.fn();
    render(<FilterTabs tabs={[...TABS]} selected="カテゴリB" onSelect={onSelect} />);
    const lastTab = screen.getByRole('tab', { name: 'カテゴリB' });
    fireEvent.keyDown(lastTab, { key: 'Home' });
    expect(onSelect).toHaveBeenCalledWith('すべて');
  });

  it('Endキーで最後のタブに移動する', () => {
    const onSelect = vi.fn();
    render(<FilterTabs tabs={[...TABS]} selected="すべて" onSelect={onSelect} />);
    const firstTab = screen.getByRole('tab', { name: 'すべて' });
    fireEvent.keyDown(firstTab, { key: 'End' });
    expect(onSelect).toHaveBeenCalledWith('カテゴリB');
  });

  it('選択タブはtabIndex=0で非選択タブはtabIndex=-1', () => {
    render(<FilterTabs tabs={[...TABS]} selected="カテゴリA" onSelect={() => {}} />);
    expect(screen.getByRole('tab', { name: 'カテゴリA' })).toHaveAttribute('tabindex', '0');
    expect(screen.getByRole('tab', { name: 'すべて' })).toHaveAttribute('tabindex', '-1');
    expect(screen.getByRole('tab', { name: 'カテゴリB' })).toHaveAttribute('tabindex', '-1');
  });
});
