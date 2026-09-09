package pgtext_test

import (
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/pgtext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pgx の stdlib 互換層は date 列を宛先の Go 型に関わらず time.Time として渡してくる
// （null_date.go の doc 参照。simple / extended protocol のどちらでも同じ）。この型は
// time.Time と string の両方を受けて 'YYYY-MM-DD' に正規化する。
func Test_NullDate_TimeTimeを正規化する(t *testing.T) {
	var d pgtext.NullDate
	require.NoError(t, d.Scan(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)))
	assert.True(t, d.Valid)
	assert.Equal(t, "2026-09-01", d.String)
}

func Test_NullDate_stringもそのまま受ける(t *testing.T) {
	var d pgtext.NullDate
	require.NoError(t, d.Scan("2026-09-01"))
	assert.True(t, d.Valid)
	assert.Equal(t, "2026-09-01", d.String)
}

func Test_NullDate_バイト列も受ける(t *testing.T) {
	var d pgtext.NullDate
	require.NoError(t, d.Scan([]byte("2026-09-01")))
	assert.True(t, d.Valid)
	assert.Equal(t, "2026-09-01", d.String)
}

func Test_NullDate_NULLはValidがfalseになる(t *testing.T) {
	d := pgtext.NullDate{String: "2026-09-01", Valid: true}
	require.NoError(t, d.Scan(nil))
	assert.False(t, d.Valid)
	assert.Empty(t, d.String)
}

func Test_NullDate_未対応の型は拒否する(t *testing.T) {
	var d pgtext.NullDate
	err := d.Scan(123)
	require.Error(t, err)
	assert.False(t, d.Valid)
}

func Test_NullDate_Value(t *testing.T) {
	v, err := pgtext.NullDate{String: "2026-09-01", Valid: true}.Value()
	require.NoError(t, err)
	assert.Equal(t, "2026-09-01", v)

	v, err = pgtext.NullDate{}.Value()
	require.NoError(t, err)
	assert.Nil(t, v)
}
