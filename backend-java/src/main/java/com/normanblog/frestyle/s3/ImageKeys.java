package com.normanblog.frestyle.s3;

import java.time.Instant;
import java.util.UUID;

/** S3 オブジェクトキーと拡張子の組み立てを担う小道具(presigner 実装で共有する)。 */
final class ImageKeys {

  private ImageKeys() {}

  /** {@code profiles/{userId}/{epochNanos}{ext}} 形式のキーを作る(Go 版と同じ命名規則)。 */
  static String profileKey(Long userId, String fileName, String contentType) {
    Instant now = Instant.now();
    long epochNanos = now.getEpochSecond() * 1_000_000_000L + now.getNano();
    return "profiles/" + userId + "/" + epochNanos + guessExt(fileName, contentType);
  }

  /**
   * {@code ai-chat/{userId}/{uuid}.{ext}} 形式のキーを作る(Go 版と同じ命名規則)。
   *
   * <p>filename は拡張子推定にだけ使い、キーには埋めない(衝突・インジェクション回避)。
   */
  static String aiChatAttachmentKey(Long userId, String fileName, String contentType) {
    return "ai-chat/" + userId + "/" + UUID.randomUUID() + attachmentExt(fileName, contentType);
  }

  /** 添付の拡張子。MIME を優先逆引きし、不明なら filename 末尾の拡張子、両方無ければ ""。 */
  static String attachmentExt(String fileName, String contentType) {
    String ext =
        switch (contentType == null ? "" : contentType) {
          case "image/png" -> ".png";
          case "image/jpeg", "image/jpg" -> ".jpg";
          case "image/gif" -> ".gif";
          case "image/webp" -> ".webp";
          case "application/pdf" -> ".pdf";
          case "text/csv" -> ".csv";
          default -> "";
        };
    if (!ext.isEmpty()) {
      return ext;
    }
    if (fileName != null) {
      int dot = fileName.lastIndexOf('.');
      if (dot >= 0 && dot < fileName.length() - 1) {
        return fileName.substring(dot).toLowerCase();
      }
    }
    return "";
  }

  /** fileName の拡張子を優先し、無ければ contentType から推定する。 */
  static String guessExt(String fileName, String contentType) {
    if (fileName != null) {
      int dot = fileName.lastIndexOf('.');
      if (dot != -1 && dot < fileName.length() - 1) {
        return fileName.substring(dot).toLowerCase();
      }
    }
    return switch (contentType == null ? "" : contentType) {
      case "image/jpeg", "image/jpg" -> ".jpg";
      case "image/png" -> ".png";
      case "image/gif" -> ".gif";
      case "image/webp" -> ".webp";
      default -> ".bin";
    };
  }
}
