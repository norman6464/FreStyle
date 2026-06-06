package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.ProfileResponse;
import com.normanblog.frestyle.dto.ProfileUpdateRequest;
import com.normanblog.frestyle.entity.Profile;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.ProfileRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/** 認証ユーザー自身のプロフィール(displayName + bio/avatar/status)と onboarding を担う。 */
@Service
public class ProfileService {

  private final ProfileRepository profiles;
  private final UserRepository users;

  public ProfileService(ProfileRepository profiles, UserRepository users) {
    this.profiles = profiles;
    this.users = users;
  }

  /** プロフィール表示。profile 行が無ければ空の値で返す。 */
  public ProfileResponse view(User actor) {
    Profile profile = profiles.findById(actor.getId()).orElse(null);

    return ProfileResponse.of(actor, profile);
  }

  /**
   * プロフィールを更新する。displayName は users、bio/avatar/status は profiles に保存する。
   * 省略(null)されたフィールドは既存値を保持する。
   */
  @Transactional
  public ProfileResponse update(User actor, ProfileUpdateRequest request) {
    String displayName = request.resolvedDisplayName();
    if (displayName != null && !displayName.isBlank()) {
      actor.setDisplayName(displayName);
      actor.setUpdatedAt(Instant.now());
      users.save(actor);
    }

    Profile profile = profiles.findById(actor.getId()).orElseGet(() -> newProfile(actor.getId()));
    if (request.bio() != null) {
      profile.setBio(request.bio());
    }
    String avatarUrl = request.resolvedAvatarUrl();
    if (avatarUrl != null) {
      profile.setAvatarUrl(avatarUrl);
    }
    if (request.status() != null) {
      profile.setStatus(request.status());
    }
    profile.setUpdatedAt(Instant.now());
    profiles.save(profile);

    return ProfileResponse.of(actor, profile);
  }

  /**
   * Welcome 完了時に users.onboarded_at をセットする。
   *
   * <p>すでに値があれば何もしない(冪等。再押下でも初回日時を保持する)。
   */
  @Transactional
  public void completeOnboarding(User actor) {
    if (actor.getOnboardedAt() == null) {
      actor.setOnboardedAt(Instant.now());
      users.save(actor);
    }
  }

  private Profile newProfile(Long userId) {
    return Profile.builder().userId(userId).status("").build();
  }
}
