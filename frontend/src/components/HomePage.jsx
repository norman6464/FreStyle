import { useSelector } from 'react-redux';

export default function HomePage() {
  const accessToken = useSelector((state) => state.auth.name);
  return <div>ようこそ{accessToken}！</div>;
}
