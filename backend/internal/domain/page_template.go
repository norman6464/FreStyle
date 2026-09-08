package domain

import (
	"errors"
	"strings"
	"time"
)

// ErrPageTemplateNotFound は対象の雛形が存在しない（無い ID・別ワークスペースのもの）
// ときに返す。
var ErrPageTemplateNotFound = errors.New("page template not found")

// ErrInvalidTemplateName は雛形名が保存してよい形でないときに返す。
var ErrInvalidTemplateName = errors.New("invalid page template name")

// PageTemplateNameMaxBytes は雛形名に許す最大バイト数（TrimSpace 後）。
const PageTemplateNameMaxBytes = 100

// PageTemplate はページの雛形。「雛形として保存」で作られ、「雛形から作る」の元になる。
//
// SpaceID が nil ならワークスペース全体で見える雛形、値があればそのスペース限定
// （schema.hcl の page_templates.doc コメント参照）。
type PageTemplate struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspaceId"`
	SpaceID     *string `json:"spaceId,omitempty"`
	Name        string  `json:"name"`
	// Icon は元ページのアイコンをそのままコピーする。未設定は nil。
	Icon *PageIcon `json:"icon,omitempty"`
	// Doc は ProseMirror ドキュメント（JSON 文字列）。pageRef・画像ノードを取り除いた形で
	// 保存する（usecase/kb の stripPageRefAndImageNodesForTemplate 参照）。API へは応答形で
	// 別途 json.RawMessage へ変換するため、ここでは持ち出さない。
	Doc             string    `json:"-"`
	CreatedByUserID uint64    `json:"createdByUserId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ValidateTemplateName は雛形名が保存してよい形かを検証し、保存用に正規化した値（TrimSpace 後）
// を返す。
//
//   - TrimSpace 後に空文字なら ErrInvalidTemplateName（空白だけの名前を許さない）
//   - TrimSpace 後 PageTemplateNameMaxBytes バイトを超えるものは ErrInvalidTemplateName
//
// 重複（同じワークスペース内で同名）はここでは見ない。DB の UNIQUE 制約
// （uq_page_templates_workspace_name）に任せ、違反を repository が ErrDuplicateTemplateName へ
// 翻訳する（検査してから INSERT するまでの間に別の要求が同じ名前を取り得るため、
// 一意制約を唯一の判定にする）。
func ValidateTemplateName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ErrInvalidTemplateName
	}
	if len(trimmed) > PageTemplateNameMaxBytes {
		return "", ErrInvalidTemplateName
	}
	return trimmed, nil
}
