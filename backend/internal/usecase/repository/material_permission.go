package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// CourseWithFacts は 1 コースと、その実効権限を決める事実の組。
// ListCourseFactsForUser が返す（ふるい落としは domain.ResolveMaterialPermission が行う）。
//
// **返ってきた時点ではまだ「見せてよいコース」に絞られていない。** 下書きも、
// 付与の無いコースも含まれる。絞るのは呼び出し側で、ここで絞らないのは判定規則を
// domain の 1 箇所に閉じるため。
type CourseWithFacts struct {
	Course domain.Course
	Facts  domain.MaterialFacts
}

// MaterialPermissionRepository は教材（コース / 章）の権限モデルへのアクセスを提供する。
//
// TeachingMaterialRepository / CourseRepository と分けているのは、境界が違うため。
// あちらは教材そのものの読み書きで、こちらは「誰が何をしてよいか」。同じ interface に
// 足すと、教材を扱うだけの実装や fake が権限のメソッドまで実装することになる
// （ノートで KnowledgeBaseRepository と KnowledgeBasePermissionRepository を
// 分けているのと同じ理由）。
//
// 事実を集めるだけで、そこから何ができるかは domain.ResolveMaterialPermission が決める。
// 規則をここへ写さないこと。
type MaterialPermissionRepository interface {
	// CourseFactsForUser はコース 1 つの実効権限を決める事実を 1 回のクエリで集める。
	// コースが無い・別ワークスペースなら domain.ErrNotFound。
	CourseFactsForUser(ctx context.Context, workspaceID string, courseID uint64, userID uint64) (*domain.MaterialFacts, error)
	// ChapterFactsForUser は章 1 つについて同じ事実を集める。
	// コースに張られた付与も見る（章へ降りてくるため）。
	ChapterFactsForUser(ctx context.Context, workspaceID string, chapterID uint64, userID uint64) (*domain.MaterialFacts, error)

	// ListCourseFactsForUser はワークスペース内のコース全件と、それぞれの事実を
	// 1 回のクエリで返す（sort_order 順）。コースごとに引く（N+1）ことはしない。
	ListCourseFactsForUser(ctx context.Context, workspaceID string, userID uint64) ([]CourseWithFacts, error)

	// UpsertCourseGrant はコースでの既定の役割を与える（同じ主体には 1 行だけ）。
	//
	// **既存の行があれば役割を置き換える**ので、この行だけを見れば弱めることもできる。
	// 弱められないのは**段をまたいだとき**で、章に弱い役割を張ってもコースの役割は
	// 下がらない（合成は最も強いものを採る）。「この人だけこの章では外す」は
	// この層では表せない。
	UpsertCourseGrant(ctx context.Context, workspaceID string, courseID uint64, principalID string, role domain.GrantRole) (*domain.CourseGrant, error)
	// DeleteCourseGrant はコースでの既定の役割を剥がす（冪等）。
	DeleteCourseGrant(ctx context.Context, workspaceID string, courseID uint64, principalID string) error
	// ListCourseGrants はそのコース自身に張られた付与を返す（章に張った分は含まない）。
	ListCourseGrants(ctx context.Context, workspaceID string, courseID uint64) ([]domain.CourseGrant, error)

	// UpsertChapterGrant は章 1 つでの既定の役割を与える。
	UpsertChapterGrant(ctx context.Context, workspaceID string, chapterID uint64, principalID string, role domain.GrantRole) (*domain.ChapterGrant, error)
	// DeleteChapterGrant は章での既定の役割を剥がす（冪等）。
	DeleteChapterGrant(ctx context.Context, workspaceID string, chapterID uint64, principalID string) error
	// ListChapterGrants はその章自身に張られた付与を返す（コースから降りてくる分は含まない）。
	//
	// **「この教材を編集できる人の一覧」ではない。** コースに張られた付与で編集できる人は
	// 含まれず、空でも「誰も編集できない」の意味にならない。
	ListChapterGrants(ctx context.Context, workspaceID string, chapterID uint64) ([]domain.ChapterGrant, error)
}
