type AvatarSize = 'sm' | 'md' | 'lg' | 'xl';

interface AvatarProps {
  name: string;
  src?: string;
  size?: AvatarSize;
}

const SIZE_STYLES: Record<AvatarSize, { container: string; text: string }> = {
  sm: { container: 'w-8 h-8', text: 'text-xs' },
  md: { container: 'w-10 h-10', text: 'text-sm' },
  lg: { container: 'w-14 h-14', text: 'text-xl' },
  xl: { container: 'w-16 h-16', text: 'text-2xl' },
};

export default function Avatar({ name, src, size = 'md' }: AvatarProps) {
  const initial = name ? name.charAt(0).toUpperCase() : '?';
  const styles = SIZE_STYLES[size];

  if (src) {
    return (
      <div className={`${styles.container} rounded-full flex-shrink-0 overflow-hidden`}>
        <img src={src} alt={name} className="w-full h-full object-cover rounded-full" />
      </div>
    );
  }

  return (
    <div
      className={`${styles.container} rounded-full bg-primary-500 flex items-center justify-center flex-shrink-0`}
    >
      <span className={`text-white ${styles.text} font-bold`}>{initial}</span>
    </div>
  );
}
