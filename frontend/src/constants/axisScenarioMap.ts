/**
 * 評価軸とシナリオカテゴリ・難易度のマッピング
 *
 * 弱点軸に基づくシナリオ推薦で使用
 */

export interface AxisScenarioMapping {
  categories: string[];
  difficulties: string[];
}

export const AXIS_SCENARIO_MAP: Record<string, AxisScenarioMapping> = {
  '論理的構成力': { categories: ['customer', 'senior'], difficulties: ['beginner', 'intermediate'] },
  '配慮表現': { categories: ['customer', 'senior'], difficulties: ['intermediate', 'advanced'] },
  '要約力': { categories: ['senior', 'team'], difficulties: ['intermediate', 'advanced'] },
  '提案力': { categories: ['customer'], difficulties: ['intermediate', 'advanced'] },
  '質問・傾聴力': { categories: ['team', 'senior'], difficulties: ['beginner', 'intermediate'] },
};
