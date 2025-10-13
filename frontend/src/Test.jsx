import { useEffect, useState } from 'react';

export default function Test() {
  const [message, setMessage] = useState('読み込み中...');

  useEffect(() => {
    fetch('http://localhost:8080/api/hello')
      .then((res) => {
        if (!res.ok) throw new Error('ネットワークエラー');
        return res.json();
      })
      .then((data) => setMessage(data))
      .catch((err) => setMessage(err));
  }, []);

  return (
    <div>
      <h1>Spring bootからのメッセージ</h1>
      <p>{message.name}</p>
    </div>
  );
}
