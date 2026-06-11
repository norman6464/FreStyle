import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../lib/axios', () => ({
  default: {
    post: vi.fn(),
  },
}));

import apiClient from '../../lib/axios';
import AdminSqlRepository from '../AdminSqlRepository';

const mockedPost = vi.mocked(apiClient.post);

describe('AdminSqlRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('run: /api/v2/admin/sql に query を POST し結果を返す', async () => {
    mockedPost.mockResolvedValue({
      data: {
        columns: ['id', 'email'],
        rows: [[1, 'a@e.com']],
        rowCount: 1,
        truncated: false,
      },
    });

    const res = await AdminSqlRepository.run('SELECT id, email FROM users LIMIT 1');

    expect(mockedPost).toHaveBeenCalledWith('/api/v2/admin/sql', {
      query: 'SELECT id, email FROM users LIMIT 1',
    });
    expect(res.columns).toEqual(['id', 'email']);
    expect(res.rows[0][1]).toBe('a@e.com');
    expect(res.truncated).toBe(false);
  });
});
