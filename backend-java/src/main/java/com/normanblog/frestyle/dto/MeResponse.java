package com.normanblog.frestyle.dto;

import com.normanblog.frestyle.entity.User;
import java.time.Instant;
import java.util.List;

/** GET /api/v2/auth/me のレスポンス。現在ログイン中のユーザー + 認証由来の派生情報。 */
public record MeResponse(
    Long id,
    String cognitoSub,
    String email,
    String displayName,
    Long companyId,
    String role,
    Instant createdAt,
    Instant updatedAt,
    List<String> groups,
    boolean isAdmin,
    boolean onboarded,
    boolean aiChatEnabledForTrainees) {

  public static MeResponse of(
      User user, List<String> groups, boolean isAdmin, boolean aiChatEnabledForTrainees) {
    return new MeResponse(
        user.getId(),
        user.getCognitoSub(),
        user.getEmail(),
        user.getDisplayName(),
        user.getCompanyId(),
        user.getRole(),
        user.getCreatedAt(),
        user.getUpdatedAt(),
        groups,
        isAdmin,
        user.getOnboardedAt() != null,
        aiChatEnabledForTrainees);
  }
}
