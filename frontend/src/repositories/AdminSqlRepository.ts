import api from '../lib/axios';
import { ADMIN } from '../constants/apiRoutes';

/** read-only SQL の実行結果（backend handler.runSQLResponse と 1:1）。 */
export interface SqlResult {
  columns: string[];
  /** 各行は列順のセル配列。セルは string | number | boolean | null。 */
  rows: Array<Array<string | number | boolean | null>>;
  rowCount: number;
  /** 行数上限で打ち切られたら true。 */
  truncated: boolean;
}

/**
 * super_admin 向け read-only SQL コンソールの API ラッパー。
 * 単一の SELECT / WITH クエリを実行し、列・行を取得する。
 */
const AdminSqlRepository = {
  async run(query: string): Promise<SqlResult> {
    const res = await api.post<SqlResult>(ADMIN.sql, { query });
    return res.data;
  },
};

export default AdminSqlRepository;
