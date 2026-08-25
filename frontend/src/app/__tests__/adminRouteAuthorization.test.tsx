import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import type { ReactElement } from 'react';

/**
 * admin 各ルートの「誰が入れて、誰がダッシュボードへ戻されるか」を固定する特性テスト。
 *
 * 認可ゲートがページ内にあってもルート側にあっても、この表を同じように通ることが
 * 「移設で認可が緩んでいない / 締まっていない」ことの根拠になる。
 * 通過条件そのものを変えるときは、この表を先に更新すること。
 */

const mockState = {
  auth: { isAdmin: true, loading: false, role: 'super_admin' as string | null },
};
vi.mock('react-redux', () => ({
  useSelector: (sel: (s: typeof mockState) => unknown) => sel(mockState),
  useDispatch: () => vi.fn(),
}));

// 以下は「認可の判定だけ」を見るためのデータ取得スタブ。API/axios は呼ばない。
vi.mock('@/entities/company', () => ({
  CompanyRepository: {
    listStats: () => Promise.resolve([]),
    list: () => Promise.resolve([]),
    updateActive: () => Promise.resolve(),
  },
}));
vi.mock('@/entities/user', () => ({
  AuthRepository: {
    getCurrentUser: () => Promise.resolve({ id: 1, role: 'super_admin', companyId: null }),
  },
}));
vi.mock('@/entities/invitation', () => ({
  AdminInvitationRepository: {
    list: () => Promise.resolve([]),
    create: vi.fn(),
    createWithTemporaryPassword: vi.fn(),
    cancel: vi.fn(),
  },
}));
vi.mock('@/shared/lib/hooks/useToast', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('@/pages/admin-audit-log/model/useAuditLog', () => ({
  useAuditLog: () => ({ events: [], loading: false, error: null }),
}));
vi.mock('@/pages/admin-dashboard/model/useAdminDashboard', () => ({
  useAdminDashboard: () => ({ summary: null, loading: false, error: null }),
}));
vi.mock('@/pages/admin-company-applications/model/useCompanyApplications', () => ({
  useCompanyApplications: () => ({
    applications: [],
    pendingCount: 0,
    loading: false,
    error: null,
    updatingId: null,
    setStatus: vi.fn(),
    reload: vi.fn(),
  }),
}));
vi.mock('@/pages/admin-members/model/useAdminMembers', () => ({
  useAdminMembers: () => ({
    members: [],
    loading: false,
    error: null,
    updatingId: null,
    setAiAccess: vi.fn(),
    setActive: vi.fn(),
    remove: vi.fn(),
  }),
}));

import AdminDashboardPage from '@/pages/admin-dashboard/ui/AdminDashboardPage';
import AdminCompaniesPage from '@/pages/admin-companies/ui/AdminCompaniesPage';
import AdminCompanyApplicationsPage from '@/pages/admin-company-applications/ui/AdminCompanyApplicationsPage';
import AdminMembersPage from '@/pages/admin-members/ui/AdminMembersPage';
import AdminAuditLogPage from '@/pages/admin-audit-log/ui/AdminAuditLogPage';
import AdminInvitationsPage from '@/pages/admin-invitations/ui/AdminInvitationsPage';

/**
 * App.tsx の admin ルート定義と同じ組み立て。
 * 認可ゲートの置き場所を変えたら、変えるのはこの配線だけで、下の表は動かさない。
 */
const ADMIN_ROUTES: { path: string; title: string; element: ReactElement }[] = [
  { path: '/admin/dashboard', title: '運営ダッシュボード', element: <AdminDashboardPage /> },
  { path: '/admin/companies', title: '管理: 会社一覧', element: <AdminCompaniesPage /> },
  { path: '/admin/applications', title: '管理: 利用申請', element: <AdminCompanyApplicationsPage /> },
  { path: '/admin/members', title: '管理: 従業員一覧', element: <AdminMembersPage /> },
  { path: '/admin/audit', title: '監査ログ', element: <AdminAuditLogPage /> },
  { path: '/admin/invitations', title: '管理: メンバー招待', element: <AdminInvitationsPage /> },
];

const REDIRECT_MARKER = 'ダッシュボード（リダイレクト先）';

function renderRoute(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        {ADMIN_ROUTES.map((r) => (
          <Route key={r.path} path={r.path} element={r.element} />
        ))}
        <Route path="/dashboard" element={<div>{REDIRECT_MARKER}</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

/** 各認証状態で通過できる admin パス。ここに無いパスはダッシュボードへ戻される。 */
const CASES: { label: string; auth: { isAdmin: boolean; role: string | null }; pass: string[] }[] = [
  {
    label: 'role=super_admin / isAdmin=true',
    auth: { isAdmin: true, role: 'super_admin' },
    pass: [
      '/admin/dashboard',
      '/admin/companies',
      '/admin/applications',
      '/admin/members',
      '/admin/audit',
      '/admin/invitations',
    ],
  },
  {
    // role だけを見るページ（会社一覧 / 監査ログ）は isAdmin フラグを要求しない。
    label: 'role=super_admin / isAdmin=false',
    auth: { isAdmin: false, role: 'super_admin' },
    pass: ['/admin/companies', '/admin/audit'],
  },
  {
    label: 'role=company_admin / isAdmin=true',
    auth: { isAdmin: true, role: 'company_admin' },
    pass: ['/admin/members', '/admin/invitations'],
  },
  {
    label: 'role=company_admin / isAdmin=false',
    auth: { isAdmin: false, role: 'company_admin' },
    pass: [],
  },
  {
    // 従業員一覧 / 招待は role を見ず isAdmin フラグだけで通す（backend は Cognito の
    // admin グループにも isAdmin=true を返すため、role=trainee でも到達しうる）。
    label: 'role=trainee / isAdmin=true',
    auth: { isAdmin: true, role: 'trainee' },
    pass: ['/admin/members', '/admin/invitations'],
  },
  {
    label: 'role=trainee / isAdmin=false',
    auth: { isAdmin: false, role: 'trainee' },
    pass: [],
  },
  {
    label: 'role=null（未確定）/ isAdmin=true',
    auth: { isAdmin: true, role: null },
    pass: ['/admin/members', '/admin/invitations'],
  },
];

describe('admin ルートの認可', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockState.auth = { isAdmin: true, loading: false, role: 'super_admin' };
  });

  for (const c of CASES) {
    describe(c.label, () => {
      for (const route of ADMIN_ROUTES) {
        const shouldPass = c.pass.includes(route.path);

        it(`${route.path} は ${shouldPass ? '表示される' : 'ダッシュボードへリダイレクトされる'}`, async () => {
          mockState.auth = { ...c.auth, loading: false };
          renderRoute(route.path);

          if (shouldPass) {
            expect(await screen.findByText(route.title)).toBeInTheDocument();
            expect(screen.queryByText(REDIRECT_MARKER)).not.toBeInTheDocument();
          } else {
            expect(await screen.findByText(REDIRECT_MARKER)).toBeInTheDocument();
            expect(screen.queryByText(route.title)).not.toBeInTheDocument();
          }
        });
      }
    });
  }

  describe('認証情報の確認中（loading=true）', () => {
    for (const route of ADMIN_ROUTES) {
      it(`${route.path} はローディングを出し、判定を保留する`, async () => {
        mockState.auth = { isAdmin: false, loading: true, role: null };
        renderRoute(route.path);

        expect(await screen.findByRole('status')).toBeInTheDocument();
        expect(screen.queryByText(route.title)).not.toBeInTheDocument();
        expect(screen.queryByText(REDIRECT_MARKER)).not.toBeInTheDocument();
      });
    }
  });
});
