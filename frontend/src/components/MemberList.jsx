import MemberItem from './MemberItem';

export default function MemberList({ users }) {
  return (
    <div className="space-y-2">
      {users.map((user) => (
        <MemberItem
          key={user.id}
          id={user.id}
          name={user.email}
          roomId={user.roomId}
        />
      ))}
    </div>
  );
}