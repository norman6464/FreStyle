/** 指定日が属する週の月曜日（00:00:00）を返す */
export function getMonday(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  d.setDate(d.getDate() - (day === 0 ? 6 : day - 1));
  d.setHours(0, 0, 0, 0);
  return d;
}

/** N週前の月曜〜翌月曜の範囲を返す */
export function getWeekRange(weeksAgo: number): { start: Date; end: Date } {
  const start = getMonday(new Date());
  start.setDate(start.getDate() - weeksAgo * 7);

  const end = new Date(start);
  end.setDate(start.getDate() + 7);

  return { start, end };
}
