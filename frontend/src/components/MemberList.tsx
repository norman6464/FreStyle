import MemberItem from './MemberItem';
import type { MemberUser } from '../types';

interface MemberListProps {
  users: MemberUser[];
}

export default function MemberList({ users }: MemberListProps) {
  return (
    <div className="space-y-4">
      {users.map((user) => (
        <MemberItem
          key={user.id}
          id={user.id}
          name={user.name}
          email={user.email}
          roomId={user.roomId}
        />
      ))}
    </div>
  );
}
