import { useState, useEffect, useCallback } from 'react';
import { ReminderRepository } from '../repositories/ReminderRepository';
import { ReminderSetting } from '../types';

const DAY_OPTIONS = [
  { key: 'mon', label: '月' },
  { key: 'tue', label: '火' },
  { key: 'wed', label: '水' },
  { key: 'thu', label: '木' },
  { key: 'fri', label: '金' },
  { key: 'sat', label: '土' },
  { key: 'sun', label: '日' },
];

export function useReminder() {
  const [setting, setSetting] = useState<ReminderSetting>({
    enabled: true,
    reminderTime: '20:00',
    daysOfWeek: 'mon,tue,wed,thu,fri',
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let cancelled = false;
    ReminderRepository.fetchSetting()
      .then((data) => {
        if (!cancelled) setSetting(data);
      })
      .catch(() => {})
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, []);

  const toggleEnabled = useCallback(() => {
    setSetting((prev) => ({ ...prev, enabled: !prev.enabled }));
  }, []);

  const setTime = useCallback((time: string) => {
    setSetting((prev) => ({ ...prev, reminderTime: time }));
  }, []);

  const toggleDay = useCallback((day: string) => {
    setSetting((prev) => {
      const days = prev.daysOfWeek.split(',').filter(Boolean);
      const newDays = days.includes(day)
        ? days.filter((d) => d !== day)
        : [...days, day];
      return { ...prev, daysOfWeek: newDays.join(',') };
    });
  }, []);

  const save = useCallback(async () => {
    setSaving(true);
    try {
      const saved = await ReminderRepository.saveSetting(setting);
      setSetting(saved);
    } finally {
      setSaving(false);
    }
  }, [setting]);

  const selectedDays = setting.daysOfWeek.split(',').filter(Boolean);

  return {
    setting,
    loading,
    saving,
    toggleEnabled,
    setTime,
    toggleDay,
    save,
    dayOptions: DAY_OPTIONS,
    selectedDays,
  };
}
