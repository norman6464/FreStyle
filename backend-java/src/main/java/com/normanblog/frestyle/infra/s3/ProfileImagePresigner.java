package com.normanblog.frestyle.infra.s3;

import com.normanblog.frestyle.dto.ProfileImageUploadUrl;

/** プロフィールアイコン用の S3 PUT 署名付き URL を発行する境界。 */
public interface ProfileImagePresigner {

  /**
   * {@code profiles/{userId}/{epochNanos}{ext}} キーへの PUT 署名 URL と、表示用 URL を発行する。
   *
   * @param userId 宛先ユーザー(キーの名前空間に使う)
   * @param fileName 任意。拡張子の推定に使う
   * @param contentType 任意。空なら image/png とみなす(署名に焼き込むため PUT 時のヘッダと一致が必要)
   */
  ProfileImageUploadUrl generate(Long userId, String fileName, String contentType);
}
