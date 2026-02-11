interface UserAvatarProps {
  name: string;
  size?: 'sm' | 'md';
}

export default function UserAvatar({ name, size = 'md' }: UserAvatarProps) {
  const initial = name ? name.charAt(0).toUpperCase() : '?';
  const sizeClasses = size === 'sm' ? 'w-7 h-7 text-xs' : 'w-8 h-8 text-sm';

  return (
    <div
      className={`${sizeClasses} rounded-full bg-primary-100 text-primary-700 flex items-center justify-center font-semibold`}
    >
      {initial}
    </div>
  );
}
