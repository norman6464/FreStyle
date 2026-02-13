import apiClient from '../lib/axios';
import type { MemberUser } from '../types';

const UserSearchRepository = {
  async searchUsers(query?: string): Promise<MemberUser[]> {
    const params: Record<string, string> = {};
    if (query) params.query = query;
    const res = await apiClient.get('/api/chat/users', { params });
    return res.data.users || [];
  },
};

export default UserSearchRepository;
