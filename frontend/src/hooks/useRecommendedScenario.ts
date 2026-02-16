import { useEffect, useState } from 'react';
import type { AxisScore } from '../types';
import type { PracticeScenario } from '../repositories/PracticeRepository';
import PracticeRepository from '../repositories/PracticeRepository';
import { AXIS_SCENARIO_MAP } from '../constants/axisScenarioMap';

export function useRecommendedScenario(weakAxis: AxisScore | null) {
  const [scenario, setScenario] = useState<PracticeScenario | null>(null);
  const [loading, setLoading] = useState(false);
  const [scenarios, setScenarios] = useState<PracticeScenario[] | null>(null);

  // シナリオ一覧を一度だけ取得してキャッシュ
  useEffect(() => {
    if (scenarios !== null) return;

    setLoading(true);
    PracticeRepository.getScenarios()
      .then((fetched) => {
        setScenarios(fetched);
      })
      .catch(() => {
        setScenarios([]);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [scenarios]);

  // 弱点軸が変わった時だけ推薦シナリオを再計算
  useEffect(() => {
    if (!weakAxis || !scenarios || scenarios.length === 0) {
      setScenario(null);
      return;
    }

    const mapping = AXIS_SCENARIO_MAP[weakAxis.axis];
    if (!mapping) {
      setScenario(null);
      return;
    }

    const matched = scenarios.filter(
      (s) => mapping.categories.includes(s.category) && mapping.difficulties.includes(s.difficulty)
    );

    if (matched.length > 0) {
      const randomIndex = Math.floor(Math.random() * matched.length);
      setScenario(matched[randomIndex]);
    } else {
      setScenario(null);
    }
  }, [weakAxis?.axis, scenarios]);

  return { scenario, loading };
}
