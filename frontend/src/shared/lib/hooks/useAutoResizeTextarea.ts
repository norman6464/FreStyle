import { useRef, useEffect } from 'react';

interface UseAutoResizeTextareaOptions {
  text: string;
  minRows?: number;
  maxRows?: number;
}

export function useAutoResizeTextarea({ text, minRows = 1, maxRows = 8 }: UseAutoResizeTextareaOptions) {
  const ref = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const textarea = ref.current;
    if (!textarea) return;

    textarea.style.height = 'auto';
    const lineHeight = parseFloat(getComputedStyle(textarea).lineHeight);
    const paddingTop = parseFloat(getComputedStyle(textarea).paddingTop);
    const paddingBottom = parseFloat(getComputedStyle(textarea).paddingBottom);

    const maxHeight = maxRows * lineHeight + paddingTop + paddingBottom;
    const newScrollHeight = textarea.scrollHeight;
    const newHeight = Math.min(newScrollHeight, maxHeight);

    if (newHeight > 0) {
      textarea.style.height = `${newHeight}px`;
    } else {
      textarea.style.height = `${minRows * lineHeight + paddingTop + paddingBottom}px`;
    }

    textarea.style.overflowY = newScrollHeight > maxHeight ? 'scroll' : 'hidden';
  }, [text, minRows, maxRows]);

  return ref;
}
