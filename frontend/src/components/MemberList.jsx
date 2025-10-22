import MemberItem from './MemberItem';

export default function MemberList({ members }) {
  <div className="space-y-2">
    {members.map((name, index) => {
      <MemberItem key={index} name={name} />;
    })}
  </div>;
}
