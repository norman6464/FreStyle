import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import HeaderWorkspaceSwitcher from '../ui/HeaderWorkspaceSwitcher';

const hoisted = vi.hoisted(() => ({
  fetchWorkspaces: vi.fn(),
}));

vi.mock('@/entities/note/api/noteRepository', () => ({
  default: { fetchWorkspaces: hoisted.fetchWorkspaces },
}));

// 遷移先の state を検査するための踏み台。
function LocationProbe() {
  const location = useLocation();
  return <div data-testid="location-state">{JSON.stringify(location.state)}</div>;
}

function renderSwitcher() {
  return render(
    <MemoryRouter initialEntries={['/dashboard']}>
      <Routes>
        <Route
          path="/dashboard"
          element={<HeaderWorkspaceSwitcher />}
        />
        <Route path="/notes" element={<LocationProbe />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('HeaderWorkspaceSwitcher', () => {
  beforeEach(() => vi.clearAllMocks());

  it('所属ワークスペースが無ければ何も出さない', async () => {
    hoisted.fetchWorkspaces.mockResolvedValue([]);
    const { container } = renderSwitcher();
    await waitFor(() => expect(hoisted.fetchWorkspaces).toHaveBeenCalled());
    expect(container).toBeEmptyDOMElement();
  });

  it('選ぶと /notes へ選んだ slug を渡して遷移する', async () => {
    hoisted.fetchWorkspaces.mockResolvedValue([{ slug: 'acme', name: 'Acme', createdAt: '' }]);
    renderSwitcher();

    const trigger = await screen.findByRole('button', { name: /ワークスペースを選択/ });
    fireEvent.click(trigger);
    fireEvent.click(await screen.findByRole('button', { name: 'Acme' }));

    const probe = await screen.findByTestId('location-state');
    expect(probe.textContent).toBe(JSON.stringify({ workspaceSlug: 'acme' }));
  });
});
