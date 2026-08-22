//go:build integration

package persistence_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

const rdDoc = `{"type":"doc","content":[{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"見出し"}]}]}`

// TestRichDocumentRepository_Integration は rich_documents の CRUD・jsonb 往復・楽観ロック・制約を
// 実 Postgres で固定する。
func TestRichDocumentRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	userRepo := persistence.NewUserRepository(db)
	repo := persistence.NewRichDocumentRepository(db)
	ctx := context.Background()

	// FK（owner_id → users.id）を満たすため作成者を用意する。
	mkOwner := func(t *testing.T, sub, email string) uint64 {
		t.Helper()
		u := &domain.User{Email: email, Role: domain.RoleTrainee}
		require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
		return u.ID
	}

	t.Run("Create → FindByID で jsonb が意味的に往復する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-1", "rd1@example.com")

		doc := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "メモ", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, doc))
		require.NotEmpty(t, doc.ID) // UUID 採番済み

		got, err := repo.FindByID(ctx, doc.ID)
		require.NoError(t, err)
		require.Equal(t, "メモ", got.Title)
		require.Equal(t, owner, got.OwnerID)
		require.Equal(t, 1, got.Revision)

		// jsonb は正規化されるので文字列一致ではなく構造で比較する。
		var parsed struct {
			Type    string `json:"type"`
			Content []struct {
				Type  string         `json:"type"`
				Attrs map[string]int `json:"attrs"`
			} `json:"content"`
		}
		require.NoError(t, json.Unmarshal([]byte(got.Doc), &parsed))
		require.Equal(t, "doc", parsed.Type)
		require.Equal(t, "heading", parsed.Content[0].Type)
		require.Equal(t, 2, parsed.Content[0].Attrs["level"])
	})

	t.Run("UpdateWithRevision は版一致時のみ更新し revision を +1 する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-2", "rd2@example.com")
		doc := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "v1", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, doc))

		upd := &domain.RichDocument{ID: doc.ID, Title: "v2", IsPublic: true, SchemaVersion: 1, Doc: `{"type":"doc","content":[]}`}
		require.NoError(t, repo.UpdateWithRevision(ctx, upd, 1))
		require.Equal(t, 2, upd.Revision) // 反映済み
		require.Equal(t, "v2", upd.Title)
		require.True(t, upd.IsPublic)

		got, err := repo.FindByID(ctx, doc.ID)
		require.NoError(t, err)
		require.Equal(t, 2, got.Revision)
		require.Equal(t, "v2", got.Title)
	})

	t.Run("UpdateWithRevision は版不一致で ErrRichDocumentConflict", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-3", "rd3@example.com")
		doc := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "v1", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, doc))

		// いま revision=1 なのに 5 を期待して更新 → 競合。
		stale := &domain.RichDocument{ID: doc.ID, Title: "x", SchemaVersion: 1, Doc: rdDoc}
		err := repo.UpdateWithRevision(ctx, stale, 5)
		require.ErrorIs(t, err, repository.ErrRichDocumentConflict)
	})

	t.Run("UpdateWithRevision は存在しない ID で ErrRichDocumentNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "rich_documents")
		ghost := &domain.RichDocument{ID: "31400a07-297e-8057-884b-c05dbdf9fa53", Title: "x", SchemaVersion: 1, Doc: rdDoc}
		err := repo.UpdateWithRevision(ctx, ghost, 1)
		require.ErrorIs(t, err, repository.ErrRichDocumentNotFound)
	})

	t.Run("SoftDelete は所有者のみ・以後 FindByID から外れる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-4", "rd4@example.com")
		doc := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "d", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, doc))

		// 他人は消せない（0 行 → NotFound）。
		require.ErrorIs(t, repo.SoftDelete(ctx, doc.ID, owner+999), repository.ErrRichDocumentNotFound)

		// 所有者は消せる。
		require.NoError(t, repo.SoftDelete(ctx, doc.ID, owner))
		_, err := repo.FindByID(ctx, doc.ID)
		require.ErrorIs(t, err, repository.ErrRichDocumentNotFound)
	})

	t.Run("DB 制約: doc が type='doc' でなければ CHECK で拒否される", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-5", "rd5@example.com")
		err := db.Exec(
			`INSERT INTO rich_documents (id, owner_id, kind, title, doc, revision, created_at, updated_at)
			 VALUES (gen_random_uuid(), ?, 'note', 't', '{"type":"paragraph"}'::jsonb, 1, now(), now())`, owner,
		).Error
		require.ErrorContains(t, err, "ck_rich_documents_doc")
	})

	t.Run("DB 制約: 存在しない owner_id は FK で拒否される", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "rich_documents", "users", "user_oidc_identities")
		err := db.Exec(
			`INSERT INTO rich_documents (id, owner_id, kind, title, doc, revision, created_at, updated_at)
			 VALUES (gen_random_uuid(), 424242, 'note', 't', '{"type":"doc"}'::jsonb, 1, now(), now())`,
		).Error
		require.ErrorContains(t, err, "fk_rich_documents_owner")
	})
}
