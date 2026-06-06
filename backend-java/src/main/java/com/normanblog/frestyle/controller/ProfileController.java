package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.ProfileImageUploadUrl;
import com.normanblog.frestyle.dto.ProfileResponse;
import com.normanblog.frestyle.dto.ProfileUpdateRequest;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.ProfileImageService;
import com.normanblog.frestyle.service.ProfileService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

/**
 * プロフィール API。自分のプロフィールのみ操作できる(IDOR 対策で他ユーザーは 403)。
 *
 * <p>userId は "me" または自分の数値 id。
 */
@RestController
@RequestMapping("/api/v2/profile")
public class ProfileController {

  private final ProfileService profileService;
  private final ProfileImageService profileImageService;
  private final CurrentUserProvider currentUser;

  public ProfileController(
      ProfileService profileService,
      ProfileImageService profileImageService,
      CurrentUserProvider currentUser) {
    this.profileService = profileService;
    this.profileImageService = profileImageService;
    this.currentUser = currentUser;
  }

  /** プロフィール画像アップロード URL 発行のリクエスト(body は任意)。 */
  public record IssueImageUrlRequest(String fileName, String contentType) {}

  @GetMapping("/{userId}")
  public ProfileResponse get(@PathVariable String userId) {
    return profileService.view(requireSelf(userId));
  }

  @PutMapping("/{userId}")
  public ProfileResponse update(
      @PathVariable String userId, @RequestBody ProfileUpdateRequest request) {
    return profileService.update(requireSelf(userId), request);
  }

  // 旧フロント互換の別パス(正規は PUT /profile/{userId})。
  @PutMapping("/{userId}/update")
  public ProfileResponse updateAlias(
      @PathVariable String userId, @RequestBody ProfileUpdateRequest request) {
    return profileService.update(requireSelf(userId), request);
  }

  @PostMapping("/me/onboarding/complete")
  public ResponseEntity<Void> completeOnboarding() {
    profileService.completeOnboarding(currentUser.require());

    return ResponseEntity.noContent().build();
  }

  /**
   * プロフィールアイコン用の S3 PUT 署名付き URL を発行する。
   *
   * <p>"me" か自分の id のみ許可(IDOR 対策)。body は任意で、無ければ既定値で処理する。
   */
  @PostMapping("/{userId}/image/presigned-url")
  public ProfileImageUploadUrl issueImageUploadUrl(
      @PathVariable String userId, @RequestBody(required = false) IssueImageUrlRequest request) {
    Long actorId = requireSelf(userId).getId();
    IssueImageUrlRequest body = request == null ? new IssueImageUrlRequest(null, null) : request;

    return profileImageService.issueUploadUrl(actorId, body.fileName(), body.contentType());
  }

  // userId が "me" か自分の id のときだけ許可。他人の id は 403(IDOR 対策)。
  private User requireSelf(String userId) {
    User actor = currentUser.require();
    if ("me".equals(userId)) {
      return actor;
    }
    try {
      if (actor.getId().equals(Long.valueOf(userId))) {
        return actor;
      }
    } catch (NumberFormatException e) {
      throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "invalid_user_id");
    }
    throw new ResponseStatusException(HttpStatus.FORBIDDEN, "forbidden");
  }
}
