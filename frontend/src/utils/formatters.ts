export function formatTime(dateString: string): string {
  if (!dateString) return '';
  const date = new Date(dateString);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const oneDay = 24 * 60 * 60 * 1000;
  const oneWeek = 7 * oneDay;

  if (diff < oneDay && date.getDate() === now.getDate()) {
    return date.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' });
  } else if (diff < oneDay * 2 && date.getDate() === now.getDate() - 1) {
    return '昨日';
  } else if (diff < oneWeek) {
    const days = ['日', '月', '火', '水', '木', '金', '土'];
    return days[date.getDay()] + '曜日';
  } else {
    return date.toLocaleDateString('ja-JP', { month: 'short', day: 'numeric' });
  }
}

export function formatDate(dateString: string): string {
  if (!dateString) return '';
  return new Date(dateString).toLocaleDateString('ja-JP');
}

export function formatHourMinute(dateString?: string): string {
  if (!dateString) return '';
  return new Date(dateString).toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' });
}

export function truncateMessage(message: string | undefined, maxLength = 30): string {
  if (!message) return 'メッセージはありません';
  if (message.length <= maxLength) return message;
  return message.substring(0, maxLength) + '...';
}
