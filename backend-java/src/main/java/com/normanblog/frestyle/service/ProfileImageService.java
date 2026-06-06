package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.ProfileImageUploadUrl;
import com.normanblog.frestyle.s3.ProfileImagePresigner;
import org.springframework.stereotype.Service;

/** プロフィールアイコン用の S3 PUT 署名付き URL 発行を担うサービス。 */
@Service
public class ProfileImageService {

  private final ProfileImagePresigner presigner;

  public ProfileImageService(ProfileImagePresigner presigner) {
    this.presigner = presigner;
  }

  /** 指定ユーザーのプロフィール画像アップロード URL を発行する。 */
  public ProfileImageUploadUrl issueUploadUrl(Long userId, String fileName, String contentType) {
    return presigner.generate(userId, fileName, contentType);
  }
}
