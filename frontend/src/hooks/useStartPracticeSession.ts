import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import type { PracticeScenario } from '../repositories/PracticeRepository';
import PracticeRepository from '../repositories/PracticeRepository';

export function useStartPracticeSession() {
  const navigate = useNavigate();
  const [starting, setStarting] = useState(false);

  const startSession = useCallback(async (scenario: PracticeScenario) => {
    setStarting(true);
    try {
      const session = await PracticeRepository.createPracticeSession({ scenarioId: scenario.id });
      navigate(`/chat/ask-ai/${session.id}`, {
        state: {
          sessionType: 'practice',
          scenarioId: scenario.id,
          scenarioName: scenario.name,
          initialPrompt: '練習開始',
        },
      });
    } catch {
      navigate('/practice');
    } finally {
      setStarting(false);
    }
  }, [navigate]);

  return { startSession, starting };
}
