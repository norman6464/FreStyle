import { render, screen, fireEvent } from '@testing-library/react';
import FilterTabs from '../FilterTabs';

const TABS = ['すべて', 'カテゴリA', 'カテゴリB'] as const;

describe('FilterTabs', () => {
  it('すべてのタブが表示される', () => {
    render(<FilterTabs tabs={[...TABS]} selected="すべて" onSelect={() => {}} />);
    TABS.forEach((tab) => {
      expect(screen.getByRole('button', { name: tab })).toBeInTheDocument();
    });
  });

  it('選択中のタブにアクティブスタイルが適用される', () => {
    render(<FilterTabs tabs={[...TABS]} selected="カテゴリA" onSelect={() => {}} />);
    const active = screen.getByRole('button', { name: 'カテゴリA' });
    expect(active.className).toContain('border-primary-500');
    expect(active.className).toContain('text-primary-400');
  });

  it('非選択のタブに非アクティブスタイルが適用される', () => {
    render(<FilterTabs tabs={[...TABS]} selected="カテゴリA" onSelect={() => {}} />);
    const inactive = screen.getByRole('button', { name: 'すべて' });
    expect(inactive.className).toContain('border-transparent');
  });

  it('タブクリックでonSelectが呼ばれる', () => {
    const onSelect = vi.fn();
    render(<FilterTabs tabs={[...TABS]} selected="すべて" onSelect={onSelect} />);
    fireEvent.click(screen.getByRole('button', { name: 'カテゴリB' }));
    expect(onSelect).toHaveBeenCalledWith('カテゴリB');
  });

  it('classNameを追加できる', () => {
    const { container } = render(
      <FilterTabs tabs={[...TABS]} selected="すべて" onSelect={() => {}} className="mb-5" />
    );
    expect(container.firstChild).toHaveClass('mb-5');
  });
});
