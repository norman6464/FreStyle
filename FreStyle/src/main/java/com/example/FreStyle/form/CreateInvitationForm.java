package com.example.FreStyle.form;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Size;

/**
 * 招待作成リクエストの入力フォーム。
 *
 * @param email 招待先メールアドレス
 * @param role  招待後のロール (trainee / company_admin)
 * @param displayName 任意の表示名（無ければ email から推定）
 */
public record CreateInvitationForm(
        @NotBlank @Email @Size(max = 254) String email,
        @NotBlank @Pattern(regexp = "trainee|company_admin",
                message = "role は trainee または company_admin")
        String role,
        @Size(max = 100) String displayName
) {}
