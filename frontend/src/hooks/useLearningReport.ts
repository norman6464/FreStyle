import { useState, useCallback, useEffect, useRef } from 'react';
import { LearningReportRepository } from '../repositories/LearningReportRepository';
import type { LearningReport } from '../types';

export function useLearningReport() {
  const [reports, setReports] = useState<LearningReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const reportCountRef = useRef(0);
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const stopPolling = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current);
      pollingRef.current = null;
    }
  }, []);

  const fetchReports = useCallback(async () => {
    try {
      const data = await LearningReportRepository.getAll();
      setReports(data);
      return data.length;
    } catch {
      setReports([]);
      return 0;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchReports().then((count) => {
      reportCountRef.current = count;
    });
  }, [fetchReports]);

  useEffect(() => {
    return () => stopPolling();
  }, [stopPolling]);

  const generateReport = useCallback(async (year: number, month: number) => {
    try {
      await LearningReportRepository.generate(year, month);
      setGenerating(true);

      // ポーリング開始（5秒間隔）
      stopPolling();
      pollingRef.current = setInterval(async () => {
        const newCount = await fetchReports();
        if (newCount > reportCountRef.current) {
          reportCountRef.current = newCount;
          setGenerating(false);
          stopPolling();
        }
      }, 5000);
    } catch {
      await fetchReports();
    }
  }, [fetchReports, stopPolling]);

  return { reports, loading, generating, generateReport, refresh: fetchReports };
}
