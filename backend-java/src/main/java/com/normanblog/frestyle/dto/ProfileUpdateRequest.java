package com.normanblog.frestyle.dto;

/**
 * プロフィール更新のリクエストボディ。
 *
 * <p>name / iconUrl は旧フロント互換の別名。それぞれ displayName / avatarUrl を優先する。
 * 省略(null)されたフィールドは既存値を保持する(誤って空にしないため)。
 */
public record ProfileUpdateRequest(
    String displayName, String name, String bio, String avatarUrl, String iconUrl, String status) {

  /** displayName 優先、無ければ name。 */
  public String resolvedDisplayName() {
    return displayName != null && !displayName.isBlank() ? displayName : name;
  }

  /** avatarUrl 優先、無ければ iconUrl。 */
  public String resolvedAvatarUrl() {
    return avatarUrl != null ? avatarUrl : iconUrl;
  }
}
