package usecase

// このファイルだけ package usecase（内部）に置く。maxDocBytes / maxTitleLen は
// unexported なので、上限ちょうど／+1 超過の隣接境界を固定するには内部テストが要る。
// 振る舞い・認可・楽観ロックの検証は外部パッケージ（rich_document_usecase_test.go）側にある。

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// stubRichDocRepo は境界テスト用の最小 stub。Create は常に成功させ、上限“ちょうど”が
// バリデーションを通過して repo まで到達できることを確認する（呼び出し内容は問わない）。
type stubRichDocRepo struct{}

var _ repository.RichDocumentRepository = stubRichDocRepo{}

func (stubRichDocRepo) Create(context.Context, *domain.RichDocument) error { return nil }
func (stubRichDocRepo) FindByID(context.Context, string) (*domain.RichDocument, error) {
	return nil, nil
}
func (stubRichDocRepo) UpdateWithRevision(context.Context, *domain.RichDocument, int) error {
	return nil
}
func (stubRichDocRepo) SoftDelete(context.Context, string, uint64) error { return nil }

const boundsValidDoc = `{"type":"doc","content":[{"type":"paragraph"}]}`

// docOfSize は type='doc' を保ったまま長さちょうど n バイトの doc JSON を作る。
func docOfSize(n int) string {
	const prefix = `{"type":"doc","x":"`
	const suffix = `"}`
	return prefix + strings.Repeat("a", n-len(prefix)-len(suffix)) + suffix
}

func Test_CreateRichDocument_サイズと長さの境界(t *testing.T) {
	uc := NewCreateRichDocumentUseCase(stubRichDocRepo{})
	cases := []struct {
		name    string
		in      CreateRichDocumentInput
		wantErr bool
	}{
		{
			"docちょうどmaxDocBytesは受理",
			CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: docOfSize(maxDocBytes)},
			false,
		},
		{
			"docがmaxDocBytes+1は拒否",
			CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: docOfSize(maxDocBytes + 1)},
			true,
		},
		{
			"titleちょうどmaxTitleLen文字は受理",
			CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: strings.Repeat("あ", maxTitleLen), Doc: boundsValidDoc},
			false,
		},
		{
			"titleがmaxTitleLen+1文字は拒否",
			CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: strings.Repeat("あ", maxTitleLen+1), Doc: boundsValidDoc},
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), tc.in)
			if tc.wantErr {
				if !errors.Is(err, ErrRichDocumentInvalid) {
					t.Fatalf("err = %v, want ErrRichDocumentInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("want ok, got %v", err)
			}
		})
	}
}
