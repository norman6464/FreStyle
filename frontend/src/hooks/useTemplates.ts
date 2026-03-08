import { useState, useEffect, useCallback } from 'react';
import { TemplateRepository } from '../repositories/TemplateRepository';
import { ConversationTemplate } from '../types';

const CATEGORIES = [
  { key: '', label: 'すべて' },
  { key: 'meeting', label: '会議' },
  { key: 'presentation', label: 'プレゼン' },
  { key: 'negotiation', label: '商談' },
  { key: 'email', label: 'メール' },
  { key: 'customer_support', label: 'サポート' },
];

export function useTemplates() {
  const [templates, setTemplates] = useState<ConversationTemplate[]>([]);
  const [category, setCategory] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    TemplateRepository.fetchTemplates(category || undefined)
      .then((data) => {
        if (!cancelled) setTemplates(data);
      })
      .catch(() => {
        if (!cancelled) setError('テンプレートの取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, [category]);

  const changeCategory = useCallback((newCategory: string) => {
    setCategory(newCategory);
  }, []);

  return { templates, category, categories: CATEGORIES, changeCategory, loading, error };
}
