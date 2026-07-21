import { useState, useCallback } from 'react';

export function useFormField<T extends Record<string, string>>(initialValues: T) {
  const [form, setForm] = useState<T>(initialValues);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  }, []);

  return { form, setForm, handleChange };
}
