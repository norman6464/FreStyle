package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

func Test_PromoteCognitoAdminRole_未昇格ユーザーをSuperAdminにする(t *testing.T) {
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{ID: 9, Role: domain.RoleTrainee},
		},
	}
	uc := NewPromoteCognitoAdminRoleUseCase(users)

	promoted, err := uc.Execute(
		context.Background(),
		PromoteCognitoAdminRoleInput{CognitoSub: "sub-9"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !promoted {
		t.Fatal("昇格されるべき")
	}
	if users.roleUpdateUserID != 9 || users.roleUpdateValue != domain.RoleSuperAdmin {
		t.Fatalf(
			"UpdateRole(%d, %q), want (9, %q)",
			users.roleUpdateUserID,
			users.roleUpdateValue,
			domain.RoleSuperAdmin,
		)
	}
}

// 昇格だけを行う usecase なので、既に管理者ロールのユーザーには触れない（降格させない）。
func Test_PromoteCognitoAdminRole_既に管理者ならロールを触らない(t *testing.T) {
	roles := []domain.RoleName{domain.RoleSuperAdmin, domain.RoleCompanyAdmin}
	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			users := &upsertUserRepoSpy{
				stubUserRepo: stubUserRepo{
					user: &domain.User{ID: 3, Role: role},
				},
			}
			uc := NewPromoteCognitoAdminRoleUseCase(users)

			promoted, err := uc.Execute(
				context.Background(),
				PromoteCognitoAdminRoleInput{CognitoSub: "sub-3"},
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if promoted {
				t.Fatal("既に管理者なら昇格扱いにしない")
			}
			if users.roleUpdateCalls != 0 {
				t.Fatalf("UpdateRole が %d 回呼ばれた", users.roleUpdateCalls)
			}
		})
	}
}

func Test_PromoteCognitoAdminRole_ユーザーが居なければ何もしない(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc := NewPromoteCognitoAdminRoleUseCase(users)

	promoted, err := uc.Execute(
		context.Background(),
		PromoteCognitoAdminRoleInput{CognitoSub: "unknown"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if promoted {
		t.Fatal("存在しないユーザーを昇格してはいけない")
	}
	if users.roleUpdateCalls != 0 {
		t.Fatalf("UpdateRole が %d 回呼ばれた", users.roleUpdateCalls)
	}
}

// DB 障害と「見つからない」を同じ無反応に畳まない（呼び元がログに残せるようエラーを返す）。
func Test_PromoteCognitoAdminRole_検索失敗はエラーで返す(t *testing.T) {
	findErr := errors.New("db down")
	users := &upsertUserRepoSpy{stubUserRepo: stubUserRepo{err: findErr}}
	uc := NewPromoteCognitoAdminRoleUseCase(users)

	promoted, err := uc.Execute(
		context.Background(),
		PromoteCognitoAdminRoleInput{CognitoSub: "sub-1"},
	)
	if promoted {
		t.Fatal("検索に失敗したら昇格扱いにしない")
	}
	if !errors.Is(err, findErr) {
		t.Fatalf("err = %v, want wrapped %v", err, findErr)
	}
}

// 未知ロール名などで恒久的に失敗する UpdateRole を握り潰さない。
func Test_PromoteCognitoAdminRole_ロール更新失敗はエラーで返す(t *testing.T) {
	updateErr := errors.New(`unknown role "super_admin"`)
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{ID: 4, Role: domain.RoleTrainee},
		},
		roleUpdateErr: updateErr,
	}
	uc := NewPromoteCognitoAdminRoleUseCase(users)

	promoted, err := uc.Execute(
		context.Background(),
		PromoteCognitoAdminRoleInput{CognitoSub: "sub-4"},
	)
	if promoted {
		t.Fatal("更新に失敗したら昇格扱いにしない")
	}
	if !errors.Is(err, updateErr) {
		t.Fatalf("err = %v, want wrapped %v", err, updateErr)
	}
}

func Test_PromoteCognitoAdminRole_Subが空ならエラー(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc := NewPromoteCognitoAdminRoleUseCase(users)

	if _, err := uc.Execute(
		context.Background(),
		PromoteCognitoAdminRoleInput{},
	); err == nil {
		t.Fatal("sub が空ならエラーを返すべき")
	}
	if users.findByCognitoSubCalls != 0 {
		t.Fatalf("検索が %d 回呼ばれた", users.findByCognitoSubCalls)
	}
}

func Test_PromoteCognitoAdminRole_リポジトリ未配線はエラー(t *testing.T) {
	uc := NewPromoteCognitoAdminRoleUseCase(nil)

	if _, err := uc.Execute(
		context.Background(),
		PromoteCognitoAdminRoleInput{CognitoSub: "sub"},
	); err == nil {
		t.Fatal("repository 未配線ならエラーを返すべき")
	}
}
