package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotBlank;

/** Cognito callback のリクエストボディ。code は必須、invitationToken はマジックリンク経由で任意。 */
public record CallbackRequest(@NotBlank String code, String invitationToken) {}
