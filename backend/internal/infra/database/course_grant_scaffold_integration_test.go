//go:build integration

package database_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// コースと章の骨格が、対象ごとの権限を張れる形になっていることを固定する。
//
// # なぜ骨格の検査が要るのか
//
// コース・教材の編集可否を「対象ごと」に決めるには、権限の行から対象を**テナントごと**
// 指す必要がある。それを DB に守らせるのが (workspace_id, id) の一意制約で、これが無いと
// 複合外部キーの参照先にできず、テナントを跨いだ付与はアプリの検査だけが頼りになる。
//
// 章とコースの外部キーは、ORM が作っていた頃から無かった。コースを消しても章を残せる状態で、
// 消す側のコードが明示的に消していただけ。対象ごとの権限を張ると、親の居ない章に権限だけが
// 残り、誰の目にも触れないまま生き続ける。
func TestCourseScaffold_権限を張れる形になっている_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	resetPublicSchema(t, db)
	require.NoError(t, database.Migrate(t.Context(), db))

	const ws = "00000000-0000-4000-8000-000000000001"
	// 同じ DB をパッケージ内の他のテストと共有するので、入れた行は必ず片付ける。
	// 残すと、スキーマを作り直さないテストがこの行を拾って別の理由で落ちる。
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM course_chapters WHERE workspace_id = $1`, ws)
		_, _ = db.Exec(`DELETE FROM courses WHERE workspace_id = $1`, ws)
		_, _ = db.Exec(`DELETE FROM workspaces WHERE id = $1`, ws)
	})
	_, err := db.Exec(
		`INSERT INTO workspaces (id, slug, name, is_active, created_at, updated_at)
		 VALUES ($1, 'ws-scaffold', '足場', true, now(), now())`, ws,
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO courses (id, workspace_id, created_by_user_id, title, created_at, updated_at)
		 VALUES (9001, $1, 1, 'コース', now(), now())`, ws,
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO course_chapters (id, workspace_id, course_id, created_by_user_id, title, created_at, updated_at)
		 VALUES (9101, $1, 9001, 1, '章', now(), now())`, ws,
	)
	require.NoError(t, err)

	t.Run("実在しないコースには章をぶら下げられない", func(t *testing.T) {
		_, err := db.Exec(
			`INSERT INTO course_chapters (id, workspace_id, course_id, created_by_user_id, title, created_at, updated_at)
			 VALUES (9199, $1, 999999, 1, '孤児', now(), now())`, ws,
		)
		assert.Error(t, err, "親の居ない章が入ってしまう")
	})

	t.Run("コースを消すと章も消える", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO courses (id, workspace_id, created_by_user_id, title, created_at, updated_at)
			 VALUES (9002, $1, 1, '消すコース', now(), now())`, ws)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO course_chapters (id, workspace_id, course_id, created_by_user_id, title, created_at, updated_at)
			 VALUES (9201, $1, 9002, 1, '消える章', now(), now())`, ws)
		require.NoError(t, err)

		_, err = db.Exec(`DELETE FROM courses WHERE id = 9002`)
		require.NoError(t, err)

		var left int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM course_chapters WHERE course_id = 9002`).Scan(&left))
		assert.Zero(t, left, "コースが消えたのに章が残っている")
	})

	// 権限の表はまだ無いので、同じ形の表をその場で作って参照先として使えることを見る。
	// ここが通らなければ、権限の表そのものが作れない。
	for _, c := range []struct {
		name   string
		target string
		id     int
	}{
		{"コース", "courses", 9001},
		{"章", "course_chapters", 9101},
	} {
		t.Run(c.name+"はテナントごと指せる", func(t *testing.T) {
			probe := "probe_" + c.target
			_, err := db.Exec(`CREATE TABLE ` + probe + ` (
				workspace_id uuid NOT NULL,
				target_id    bigint NOT NULL,
				CONSTRAINT fk_probe FOREIGN KEY (workspace_id, target_id)
					REFERENCES ` + c.target + ` (workspace_id, id) ON DELETE CASCADE
			)`)
			require.NoError(t, err, "(workspace_id, id) を参照先にできない ＝ 権限の表が作れない")
			t.Cleanup(func() { _, _ = db.Exec(`DROP TABLE IF EXISTS ` + probe) })

			_, err = db.Exec(`INSERT INTO `+probe+` VALUES ($1, $2)`, ws, c.id)
			require.NoError(t, err, "同じワークスペースからは指せる")

			_, err = db.Exec(`INSERT INTO `+probe+` VALUES ($1, $2)`,
				"00000000-0000-4000-8000-0000000000ff", c.id)
			assert.Error(t, err, "別ワークスペースから指せてしまう（テナント跨ぎの付与を DB が塞げていない）")
		})
	}
}
