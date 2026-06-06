package com.normanblog.frestyle.dto;

import com.normanblog.frestyle.entity.Profile;
import com.normanblog.frestyle.entity.User;
import java.time.Instant;

/** プロフィール表示用。users(displayName / email) と profiles(bio / avatar / status)の合成。 */
public record ProfileResponse(
    Long userId,
    String displayName,
    String email,
    String bio,
    String avatarUrl,
    String status,
    Instant updatedAt) {

  /** profile が未作成(null)でも空の値で組み立てる。 */
  public static ProfileResponse of(User user, Profile profile) {
    return new ProfileResponse(
        user.getId(),
        user.getDisplayName(),
        user.getEmail(),
        profile != null ? profile.getBio() : null,
        profile != null ? profile.getAvatarUrl() : null,
        profile != null ? profile.getStatus() : "",
        profile != null ? profile.getUpdatedAt() : null);
  }
}
