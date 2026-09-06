interface FormFieldErrorProps {
  name: string;
  error?: string;
}

export default function FormFieldError({ name, error }: FormFieldErrorProps) {
  if (!error) return null;
  return (
    <p id={`${name}-error`} role="alert" className="text-xs text-rose-700 mt-1">
      {error}
    </p>
  );
}
