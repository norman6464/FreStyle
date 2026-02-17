import { useState, useCallback, useEffect } from 'react';
import { LearningReportRepository } from '../repositories/LearningReportRepository';
import type { LearningReport } from '../types';

export function useLearningReport() {
  const [reports, setReports] = useState<LearningReport[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchReports = useCallback(async () => {
    try {
      const data = await LearningReportRepository.getAll();
      setReports(data);
    } catch {
      setReports([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchReports();
  }, [fetchReports]);

  const generateReport = useCallback(async (year: number, month: number) => {
    try {
      await LearningReportRepository.generate(year, month);
    } catch {
      // エラー時もUIを最新状態に更新
    } finally {
      await fetchReports();
    }
  }, [fetchReports]);

  return { reports, loading, generateReport, refresh: fetchReports };
}
