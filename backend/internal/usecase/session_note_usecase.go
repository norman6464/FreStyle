package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetSessionNoteUseCase は AI チャットセッション紐付きのノートを取得する。
type GetSessionNoteUseCase struct {
	repo repository.SessionNoteRepository
}

func NewGetSessionNoteUseCase(r repository.SessionNoteRepository) *GetSessionNoteUseCase {
	return &GetSessionNoteUseCase{repo: r}
}

// GetSessionNoteInput は取得対象と、それを要求している利用者を表す。
// UserID は所有者検証に使う（他人のノートを sessionID の総当たりで読めないようにする）。
type GetSessionNoteInput struct {
	SessionID uint64
	UserID    uint64
}

// Execute は所有者本人のノートのみ返す。
// 他人のノートは「存在しない」と同じ扱い（nil, nil）にして、
// そのセッションにノートがあること自体を漏らさない。
//
// 読み出しでは書き込み側（UpsertSessionNoteUseCase）のようなセッション所有者の照会を行わない。
// 理由は 3 つ:
//
//  1. 秘匿の観点で足りている。ここが返すのは session_notes.user_id が呼び出し元と一致する行だけで、
//     一致しなければ nil を返す。セッションの所有者を追加で見ても、他人のメモが読めなくなる
//     ケースは 1 つも増えない（読めるのは元から自分名義の行だけ）。
//  2. 所有者が食い違う行はもう作られない。書き込み側がセッション所有者を権威として弾くので、
//     以後 session_notes.user_id はセッションの所有者と必ず一致する。読み出しでの二重確認は、
//     同じ不変条件をもう一度確かめるだけで新しい保証を生まない。
//  3. 照会を足すと存在オラクルの取り扱いが増える。AiChatSessionRepository.FindByID は未存在を
//     domain.ErrNotFound で返すので、それをここで「見つからない（nil, nil）」へ翻訳し直す分岐が
//     要る。取り違えると「セッションが無い」と「メモが無い」を応答差で区別できてしまう。
//     読み出し 1 回あたり DB 往復も 1 つ増える。
//
// 書き込み側は逆で、そこが session_notes.user_id を「決める」場所なので、セッション所有者の
// 照会が要る（詳しくは UpsertSessionNoteUseCase.Execute のコメント）。
func (u *GetSessionNoteUseCase) Execute(ctx context.Context, in GetSessionNoteInput) (*domain.SessionNote, error) {
	if in.SessionID == 0 || in.UserID == 0 {
		return nil, errors.New("sessionID and userID are required")
	}
	n, err := u.repo.FindBySessionID(ctx, in.SessionID)
	if err != nil || n == nil {
		return nil, err
	}
	if n.UserID != in.UserID {
		return nil, nil
	}
	return n, nil
}

// UpsertSessionNoteUseCase はセッションノートを upsert する。
//
// セッションメモの所有権は「メモの行」ではなく「セッション」に従属する。
// メモは常にあるセッションに 1 件ぶら下がる従属物なので、誰がそれを書いてよいかを決めるのは
// ai_chat_sessions.user_id（＝セッションの所有者）であって、session_notes.user_id ではない。
// そのため書き込み経路は sessions を権威として参照する。
type UpsertSessionNoteUseCase struct {
	repo     repository.SessionNoteRepository
	sessions repository.AiChatSessionRepository
}

// NewUpsertSessionNoteUseCase は書き込み先のメモ repository と、所有者の権威である
// セッション repository を受け取る。
func NewUpsertSessionNoteUseCase(
	r repository.SessionNoteRepository,
	s repository.AiChatSessionRepository,
) *UpsertSessionNoteUseCase {
	return &UpsertSessionNoteUseCase{repo: r, sessions: s}
}

type UpsertSessionNoteInput struct {
	SessionID uint64
	UserID    uint64
	Content   string
}

// Execute は自分が所有するセッションに限りメモを作成 / 更新する。
//
// # 検証は 2 層ある
//
//	(1) この usecase … セッションの所有者（ai_chat_sessions.user_id）と呼び出し元を突き合わせる。
//	    新規作成・更新の両方を止める、唯一の入口側の関門。
//	(2) SQL の ON CONFLICT ... WHERE … 既存行の user_id と書き込もうとした user_id を突き合わせる。
//	    衝突（＝既に行がある）ときだけ効く後衛で、(1) が消えたときの最後の砦。
//
// # なぜ SQL の ON CONFLICT ... WHERE だけでは足りないのか
//
// ON CONFLICT ... DO UPDATE ... WHERE は、**衝突が起きて UPDATE に進んだときにしか評価されない**。
// つまり守れるのは「既にメモがある行の上書き」だけで、衝突しない初回の INSERT は素通りする。
// 被害者のセッションにまだメモが無ければ、攻撃者は自分の user_id のままメモを新規作成できた。
//
// # 新規作成が通ると何が起きるか（実 PostgreSQL で再現した壊れ方）
//
//	攻撃者が書く → 被害者のセッションに user_id = 攻撃者 の行ができる
//	被害者が読む → 行の所有者が違うので nil（自分のセッションにメモが無いように見える）
//	被害者が書く → 既存行と衝突し、ON CONFLICT の WHERE で user_id が合わず 0 行 → not-found
//	           → 被害者は自分のセッションでメモを取れなくなる（サービス拒否）
//
// 「他人のメモを上書きされる」より悪い結果になるため、入口でセッション所有者を見るのが本筋になる。
//
// # 見つからない・他人のものは同じ応答にする
//
// セッションが存在しない場合も、他人のセッションだった場合も domain.ErrNotFound を返し、
// repo.Upsert は呼ばない。応答を分けると、その差から「その session_id が存在するか」を
// 総当たりで判別できてしまう（handler もこの 2 つを同じ 404・同じ本文にしている）。
func (u *UpsertSessionNoteUseCase) Execute(ctx context.Context, in UpsertSessionNoteInput) (*domain.SessionNote, error) {
	if in.SessionID == 0 || in.UserID == 0 {
		return nil, errors.New("sessionID and userID are required")
	}
	// (1) 所有権の権威はセッション。未存在（FindByID は domain.ErrNotFound を返す）も
	// 他人のセッションも同じ not-found に畳む。
	s, err := u.sessions.FindByID(ctx, in.SessionID)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if s == nil || s.UserID != in.UserID {
		return nil, domain.ErrNotFound
	}
	// 既存メモの所有者も念のため確かめる。セッション所有者の検証を通った以上、
	// 本来ここで弾かれる行は存在しない（メモの user_id はセッションの所有者と必ず一致する）。
	// 残しているのは、この修正より前に作られてしまった「所有者が食い違う行」に当たったとき、
	// SQL の 0 行を待たずに usecase 側で not-found として扱えるようにするため。
	if existing, err := u.repo.FindBySessionID(ctx, in.SessionID); err != nil {
		return nil, err
	} else if existing != nil && existing.UserID != in.UserID {
		return nil, domain.ErrNotFound
	}
	n := &domain.SessionNote{SessionID: in.SessionID, UserID: in.UserID, Content: in.Content}
	if err := u.repo.Upsert(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}
