import { useCallback, useEffect, useRef, useState } from 'react';

type SetValue<T> = T | ((prev: T) => T);

// 同じ key を複数箇所（例: ヘッダーとサイドバー本体）で使うとき、片方の更新がもう片方に
// 即座に反映されるようにするための同一タブ内の購読レジストリ。localStorage の
// 'storage' イベントは他タブ向けで同一タブ内では発火しないため、ここで補う。
const subscribers = new Map<string, Set<(value: unknown) => void>>();

function notify(key: string, value: unknown) {
  subscribers.get(key)?.forEach((fn) => fn(value));
}

function subscribe(key: string, fn: (value: unknown) => void) {
  let set = subscribers.get(key);
  if (!set) {
    set = new Set();
    subscribers.set(key, set);
  }
  set.add(fn);
  return () => {
    set!.delete(fn);
    if (set!.size === 0) subscribers.delete(key);
  };
}

export function useLocalStorage<T>(key: string, defaultValue: T): [T, (value: SetValue<T>) => void, () => void] {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = localStorage.getItem(key);
      return item !== null ? JSON.parse(item) : defaultValue;
    } catch {
      return defaultValue;
    }
  });

  // 関数形式の setValue が最新値を参照できるようにする（購読側からの更新も反映される）。
  const valueRef = useRef(storedValue);
  valueRef.current = storedValue;

  useEffect(() => subscribe(key, (value) => setStoredValue(value as T)), [key]);

  const setValue = useCallback((value: SetValue<T>) => {
    const newValue = value instanceof Function ? (value as (prev: T) => T)(valueRef.current) : value;
    try {
      localStorage.setItem(key, JSON.stringify(newValue));
    } catch {
      // QuotaExceededError等 - stateは更新する
    }
    setStoredValue(newValue);
    notify(key, newValue);
  }, [key]);

  const removeValue = useCallback(() => {
    try {
      localStorage.removeItem(key);
    } catch {
      // localStorage unavailable
    }
    setStoredValue(defaultValue);
    notify(key, defaultValue);
  }, [key, defaultValue]);

  return [storedValue, setValue, removeValue];
}
