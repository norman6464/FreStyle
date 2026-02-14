import { useState, useCallback } from 'react';

type SetValue<T> = T | ((prev: T) => T);

export function useLocalStorage<T>(key: string, defaultValue: T): [T, (value: SetValue<T>) => void, () => void] {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = localStorage.getItem(key);
      return item !== null ? JSON.parse(item) : defaultValue;
    } catch {
      return defaultValue;
    }
  });

  const setValue = useCallback((value: SetValue<T>) => {
    setStoredValue((prev) => {
      const newValue = value instanceof Function ? value(prev) : value;
      localStorage.setItem(key, JSON.stringify(newValue));
      return newValue;
    });
  }, [key]);

  const removeValue = useCallback(() => {
    localStorage.removeItem(key);
    setStoredValue(defaultValue);
  }, [key, defaultValue]);

  return [storedValue, setValue, removeValue];
}
