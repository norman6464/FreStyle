import type { AdminInvitation } from '@/entities/invitation';

/*
 * 招待のテストデータ。型注釈を付けて共有することで、AdminInvitation にフィールドが
 * 増えたときに hook テストとページテストの両方が tsc で同時に落ちる。
 */
export const pendingInvitation: AdminInvitation = {
  id: 10,
  companyId: 1,
  email: 'member@example.com',
  role: 'trainee',
  invitedBy: 2,
  expiresAt: '2026-08-12T00:00:00Z',
  acceptedAt: null,
  acceptedUserId: null,
  createdAt: '2026-08-05T00:00:00Z',
};
