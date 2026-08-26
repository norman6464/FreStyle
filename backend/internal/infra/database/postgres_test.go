package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIsPgBouncerDSN は pooler 経由の判定を固定する。
//
// この関数だけが「本番は simple query protocol で喋る」を決めている。false に倒れると
// 本番でだけ prepared statement does not exist が出て、ローカルと CI では再現しない。
// 逆に紛れ込んだ文字列で true に倒れると、直結でも simple protocol になる。
func TestIsPgBouncerDSN(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want bool
	}{
		// --- pooler 経由と判定すべきもの ---
		{
			name: "本番の pooler ホスト",
			dsn:  "postgres://postgres.abcdefghijklmn:pw@aws-0-ap-northeast-1.pooler.supabase.com:6543/postgres",
			want: true,
		},
		{
			name: "postgresql スキームとクエリ付き",
			dsn:  "postgresql://postgres.abcdefghijklmn:pw@aws-0-ap-northeast-1.pooler.supabase.com:6543/postgres?sslmode=require",
			want: true,
		},
		{
			name: "ホスト名が大文字でも拾う",
			dsn:  "postgres://user:pw@AWS-0-AP-NORTHEAST-1.POOLER.SUPABASE.COM:6543/postgres",
			want: true,
		},
		{
			name: "pgbouncer=true の明示",
			dsn:  "postgres://user:pw@localhost:5432/frestyle?pgbouncer=true",
			want: true,
		},
		{
			name: "pgbouncer=true が他のクエリと並んでいる",
			dsn:  "postgres://user:pw@localhost:5432/frestyle?sslmode=disable&pgbouncer=true",
			want: true,
		},
		{
			name: "前後に空白があっても拾う",
			dsn:  "  postgres://user:pw@aws-0-ap-northeast-1.pooler.supabase.com:6543/postgres\n",
			want: true,
		},
		{
			name: "key=value 形式の host",
			dsn:  "host=aws-0-ap-northeast-1.pooler.supabase.com port=6543 user=postgres dbname=postgres",
			want: true,
		},
		{
			name: "key=value 形式はキーも値も大文字小文字を問わない",
			dsn:  "HOST=AWS-0-AP-NORTHEAST-1.POOLER.SUPABASE.COM port=6543",
			want: true,
		},

		// --- 直結と判定すべきもの ---
		{
			name: "空文字",
			dsn:  "",
			want: false,
		},
		{
			name: "ローカルへの直結",
			dsn:  "postgres://frestyle:frestyle@localhost:5432/frestyle?sslmode=disable",
			want: false,
		},
		{
			name: "パスワードに pooler.supabase.com が紛れていても host は localhost",
			dsn:  "postgres://user:pooler.supabase.com@localhost:5432/frestyle",
			want: false,
		},
		{
			name: "path に紛れていても host は localhost",
			dsn:  "postgres://user:pw@localhost:5432/pooler.supabase.com",
			want: false,
		},
		{
			name: "クエリの値に紛れていても host は localhost",
			dsn:  "postgres://user:pw@localhost:5432/frestyle?options=pooler.supabase.com",
			want: false,
		},
		{
			name: "pgbouncer=false",
			dsn:  "postgres://user:pw@localhost:5432/frestyle?pgbouncer=false",
			want: false,
		},
		{
			// pooler の接続文字列が出すのは小文字の true。値は厳密一致で見る。
			name: "pgbouncer の値が true 以外なら host 判定へ落ちる",
			dsn:  "postgres://user:pw@localhost:5432/frestyle?pgbouncer=1",
			want: false,
		},
		{
			name: "URL として壊れている",
			dsn:  "postgres://user:pw@[::1:5432/frestyle",
			want: false,
		},
		{
			name: "key=value 形式で host 以外のキーに紛れている",
			dsn:  "host=localhost port=5432 password=pooler.supabase.com",
			want: false,
		},
		{
			name: "key=value 形式で dbname に紛れている",
			dsn:  "host=localhost port=5432 dbname=pooler.supabase.com",
			want: false,
		},
		{
			name: "key=value ですらない裸のホスト名",
			dsn:  "pooler.supabase.com",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isPgBouncerDSN(tt.dsn))
		})
	}
}

// TestOpenSQLDB_不正なDSNは開かない は、解釈できない DSN を握りつぶさないことを固定する。
func TestOpenSQLDB_不正なDSNは開かない(t *testing.T) {
	db, err := OpenSQLDB("://not-a-dsn", false)
	require.Error(t, err)
	require.Nil(t, db)
}
