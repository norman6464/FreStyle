package com.normanblog.frestyle.s3;

import com.normanblog.frestyle.dto.AiChatAttachmentUploadUrl;

/** AI チャット添付用の S3 PUT 署名付き URL を発行する境界。 */
public interface AiChatAttachmentPresigner {

  /**
   * {@code ai-chat/{userId}/{uuid}.{ext}} キーへの PUT 署名 URL を発行する。
   *
   * @param userId 宛先ユーザー(キーの名前空間に使う)
   * @param filename 拡張子の推定にだけ使う(キーには埋めない / 衝突・インジェクション回避)
   * @param contentType 署名に焼き込むため PUT 時のヘッダと一致が必要
   */
  AiChatAttachmentUploadUrl generate(Long userId, String filename, String contentType);
}
