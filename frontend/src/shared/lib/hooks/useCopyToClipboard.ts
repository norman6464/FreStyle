import { useState, useCallback, useRef, useEffect } from 'react';

export function useCopyToClipboard() {
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  const copyToClipboard = useCallback(async (id: string, text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      if (timerRef.current) clearTimeout(timerRef.current);
      setCopiedId(id);
      timerRef.current = setTimeout(() => {
        setCopiedId(null);
      }, 2000);
    } catch {
      // クリップボードAPI失敗時は何もしない
    }
  }, []);

  return { copiedId, copyToClipboard };
}
