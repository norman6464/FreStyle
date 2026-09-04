package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ── ノートの「権限そのものを変える」API に共通する認可 ──
//
// # なぜ handler で認可を判定するのか
//
// 権限を書き換える usecase（GrantWorkspaceRoleUseCase / GrantSpaceRoleUseCase /
// GrantPageRoleUseCase / IssueShareLinkUseCase …）は、認可を一切見ない。
// 受け取った workspaceID / spaceID / principalID をそのまま書くだけで、検査するのは
// 入力の妥当性（空文字・役割名やケイパビリティが既知の値か・主体が実在するか）に限られる。
//
// これは段 1-b のページ操作と同じ分担で、認可は handler が Check*PermissionUseCase を
// 呼んで先に決める（CheckPagePermissionUseCase / CreateSpaceUseCase の doc も同じことを言う）。
// 判定を usecase へ持ち込まないのは、ワークスペースを確定させるのが middleware
// （URL の slug + principals）で、呼び出し元（＝ 誰が叩いたか）を知っているのが
// handler だけだから。usecase は *gin.Context を受け取らない（CLAUDE.md §3.5）。
//
// 裏を返すと、このファイルの関数を通さずにルートを生やした時点で
// 「ログインさえしていれば誰でも自分を admin にできる」状態になる。権限操作のルートは
// 必ず requireWorkspaceAdmin / requireSpaceAdmin / requirePageAdmin のいずれかを
// 最初に通すこと。
//
// # なぜ特権ロールを特別扱いしないのか
//
// ノートの役割（domain.GrantRole の admin / editor / commenter / viewer）は
// per-workspace の grant だけで閉じており、アプリ全体のグローバルなロール概念は
// 存在しない（domain/grant.go のコメント）。「この入れ物で何ができるか」を
// 入れ物ごとに持つ、それだけがこのアプリの権限モデル。
//
// 「特権ロールなら全部できる」という分岐をここに 1 つでも足すと、権限の出どころが
// principals / grants とそれ以外の 2 系統になり、ワークスペースの admin が
// 知らないところで自分のテナントを読み書きされる（しかも grant を全部見ても
// その事実が説明できない）。したがってこの gate が見るのは grant だけである。
//
// # なぜ拒否をすべて 404 not_found で揃えるのか（存在オラクル対策）
//
// 「存在しないから断られた」と「権限が無いから断られた」を撃ち分けると、対象の ID を
// 総当たりするだけで、中身を 1 バイトも読めないまま実在を数え上げられる（存在の有無
// そのものが他人の情報）。このリポジトリは直近で PUT /notes/:id の 403/400 撃ち分けと
// セッションメモの越権書き込みでこの穴を 2 件塞いだところで、同じ穴を新設しない。
//
// そのため権限操作 API の拒否は、対象の種類（ワークスペース / スペース / ページ / 主体 /
// 共有リンク）にも理由（不在か無権限か）にもよらず、すべて 404 + {"error":"not_found"} に
// 揃える。middleware.KnowledgeBaseWorkspace が非メンバーへ返すものと同じ応答なので、
// viewer / editor / commenter / 非メンバー / 別ワークスペースの admin のどれで叩いても
// 返るバイト列は完全に一致する。
//
// 手順としても「認可を先に、対象に触るのは後」を守る。認可に落ちた要求は対象を
// 一度も読まないので、応答の内容が対象の状態に依存しようがない。
//
// 引き換えに、権限を持たない相手には 403 という手掛かりも返らない（UI からは
// そもそも操作を出さない前提）。ページ CRUD 側が「閲覧できる相手には 403 を返す」
// （実在を既に知っているため）としているのとは扱いが違うが、権限操作は対象の ID を
// 呼び出し側が自由に指定できる（＝ 総当たりできる）ので、こちらは一律で閉じる。

// kbPermissionGate は権限操作 API の認可判定をまとめた入口。
// 各 handler はこれを埋め込んで使う（同じ判定を handler ごとに写経しないため）。
type kbPermissionGate struct {
	checkWorkspace *usecase.CheckWorkspacePermissionUseCase
	checkSpace     *usecase.CheckSpacePermissionUseCase
	checkPage      *usecase.CheckPagePermissionUseCase
}

// newKbPermissionGate は権限操作 API 共通の認可判定を組み立てる。
func newKbPermissionGate(
	checkWorkspace *usecase.CheckWorkspacePermissionUseCase,
	checkSpace *usecase.CheckSpacePermissionUseCase,
	checkPage *usecase.CheckPagePermissionUseCase,
) *kbPermissionGate {
	return &kbPermissionGate{checkWorkspace: checkWorkspace, checkSpace: checkSpace, checkPage: checkPage}
}

// respondKbPermissionDenied は権限操作 API の唯一の拒否応答。
//
// 拒否の理由（不在 / 無権限）でも対象の種類でも分けない。ここを 1 関数に閉じているのは、
// あとから「この場合だけ 403 にする」を足しにくくするため（足した瞬間に撃ち分けが復活する）。
func respondKbPermissionDenied(c *gin.Context) {
	c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
}

// requireWorkspaceAdmin はワークスペース全体の admin かを確かめる。
// 満たさなければ応答を書いて false を返す。
//
// ワークスペースの grant は配下の全スペースへ届くので、ここを通る相手は
// テナント全体の管理者。スペースやページを指さない操作（メンバーの出入り・
// グループ・ワークスペース grant）が使う。
func (g *kbPermissionGate) requireWorkspaceAdmin(c *gin.Context, scope kbRequestScope) bool {
	perm, err := g.checkWorkspace.Execute(c.Request.Context(), usecase.CheckWorkspacePermissionInput{
		WorkspaceID: scope.workspaceID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKbPermissionErr(c, err)
		return false
	}
	if !perm.CanManage {
		respondKbPermissionDenied(c)
		return false
	}
	return true
}

// requireSpaceAdmin はスペース 1 つの admin かを確かめる。
// 満たさなければ応答を書いて false を返す。
//
// ワークスペースの admin もここを通る（workspace_grants は配下の全スペースへ届く）。
// スペース単位の grant で降格されることは無い（domain.GrantRole.Rank の合成規則で
// 「最も強いもの」を採るため）。
//
// スペースが存在しない場合も無権限と同じ応答にする。CheckSpacePermissionUseCase は
// 役割を集める前にスペースの実在を確かめて ErrSpaceNotFound を返すので、
// 「他テナントのスペース ID を渡すと自分の役割がそのまま返る」という緩み方はしない。
func (g *kbPermissionGate) requireSpaceAdmin(c *gin.Context, scope kbRequestScope, spaceID string) bool {
	perm, err := g.checkSpace.Execute(c.Request.Context(), usecase.CheckSpacePermissionInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     spaceID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKbPermissionErr(c, err)
		return false
	}
	if !perm.CanManage {
		respondKbPermissionDenied(c)
		return false
	}
	return true
}

// requirePageAdmin はページに対する権限を変えてよいかを確かめる。
// 満たさなければ応答を書いて false を返す。
//
// 判定は「そのページに届いている既定の役割が admin か」。役割は 3 段（ワークスペース /
// スペース / ページ）のどこから来ても構わず、最も強いものが実効になる。ページに admin を
// 張られた相手がそのページの共有設定を触れるのは、この段を数に入れているため。
//
// **以前はスペースの admin かどうかだけを見ていた。** page_grants が入る前は
// 「ページに対する管理者」が存在し得なかったのでそれで足りていたが、いまはページにも
// admin を張れる。スペースだけを見ていると、admin を与えられた本人がその権限を
// 一切行使できない状態になる。
//
// **閲覧できるかは別に確かめない。** admin は必ず閲覧もできる（GrantRole.Rank）ので、
// ここで重ねて問う意味が無い。
//
// **DB への問い合わせは 1 回だけ。** ページが無い場合は ErrPageNotFound が返り、
// 役割が足りない場合と同じ 404 に落ちる。落ちる段によって往復の回数が変わらないので、
// 返るまでの時間から「そのページ ID が実在するか」は読めない。
func (g *kbPermissionGate) requirePageAdmin(c *gin.Context, scope kbRequestScope, pageID string) bool {
	perm, err := g.checkPage.Execute(c.Request.Context(), usecase.CheckPagePermissionInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKbPermissionErr(c, err)
		return false
	}
	if !perm.CanManage {
		respondKbPermissionDenied(c)
		return false
	}
	return true
}

// respondKbPermissionErr は認可判定の途中で起きたエラーを応答へ落とす。
//
// 対象が見つからないことを表すセンチネルは、拒否とまったく同じ 404 not_found にする。
// それ以外（DB 障害など）だけを 500 にする。ここで respondKnowledgeBaseErr を使わないのは、
// あちらが ErrPagePermissionDenied を 403 に落とすなど、ページ CRUD 向けに
// 理由ごとの撃ち分けを持っているため。権限操作 API はその撃ち分けを持たない。
func respondKbPermissionErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrWorkspaceNotFound),
		errors.Is(err, repository.ErrSpaceNotFound),
		errors.Is(err, repository.ErrPageNotFound),
		errors.Is(err, repository.ErrPrincipalNotFound),
		errors.Is(err, repository.ErrShareLinkNotFound),
		errors.Is(err, repository.ErrUserNotFound):
		respondKbPermissionDenied(c)
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
	}
}

// respondKbPermissionOperationErr は認可を通ったあとの操作で起きたエラーを応答へ落とす。
//
// ここまで来た相手は admin なので、入力の誤り（未知の役割・グループ名の重複・
// 主体の種類違い）は理由を返してよい — 対象の実在を知る資格が既にある。
// 逆に「対象が無い」はここでも 404 not_found のまま（拒否と同じ応答）にして、
// 呼び出し側から見た応答の集合を増やさない。
func respondKbPermissionOperationErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidGrantRole),
		errors.Is(err, usecase.ErrInvalidCapability),
		errors.Is(err, usecase.ErrPrincipalKindMismatch):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
	case errors.Is(err, repository.ErrPrincipalGroupNameTaken):
		c.JSON(http.StatusConflict, errorResponse{Error: "group_name_taken"})
	case errors.Is(err, repository.ErrLastWorkspaceAdmin):
		// 手前の検査（requireNotLastWorkspaceAdmin）と同じ 409 に落とす。
		// そちらを通り抜けた競合を repository が最後に断ったときだけここへ来るので、
		// 呼び出し側から見た応答は「先に断られた」ときと区別が付かない。
		c.JSON(http.StatusConflict, errorResponse{Error: "last_workspace_admin"})
	default:
		respondKbPermissionErr(c, err)
	}
}
