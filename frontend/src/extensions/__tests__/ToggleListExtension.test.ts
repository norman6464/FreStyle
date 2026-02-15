import { describe, it, expect } from 'vitest';
import { ToggleList, ToggleSummary, ToggleContent } from '../ToggleListExtension';

describe('ToggleListExtension', () => {
  it('ToggleListの名前がtoggleList', () => {
    expect(ToggleList.name).toBe('toggleList');
  });

  it('ToggleSummaryの名前がtoggleSummary', () => {
    expect(ToggleSummary.name).toBe('toggleSummary');
  });

  it('ToggleContentの名前がtoggleContent', () => {
    expect(ToggleContent.name).toBe('toggleContent');
  });

  it('ToggleListのデフォルトopen属性がtrue', () => {
    const config = ToggleList.config;
    const attrs = config.addAttributes?.call(null as never);
    expect(attrs?.open?.default).toBe(true);
  });
});
