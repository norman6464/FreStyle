import { useEffect, useState } from 'react';
import type { AxisScore } from '../types';
import type { PracticeScenario } from '../repositories/PracticeRepository';
import PracticeRepository from '../repositories/PracticeRepository';
import { AXIS_SCENARIO_MAP } from '../constants/axisScenarioMap';

export function useRecommendedScenario(weakAxis: AxisScore | null) {
  const [scenario, setScenario] = useState<PracticeScenario | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!weakAxis) {
      setScenario(null);
      return;
    }

    const mapping = AXIS_SCENARIO_MAP[weakAxis.axis];
    if (!mapping) {
      setScenario(null);
      return;
    }

    setLoading(true);
    PracticeRepository.getScenarios()
      .then((scenarios) => {
        const matched = scenarios.filter(
          (s) => mapping.categories.includes(s.category) && mapping.difficulties.includes(s.difficulty)
        );
        if (matched.length > 0) {
          const randomIndex = Math.floor(Math.random() * matched.length);
          setScenario(matched[randomIndex]);
        } else {
          setScenario(null);
        }
      })
      .catch(() => {
        setScenario(null);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [weakAxis]);

  return { scenario, loading };
}
