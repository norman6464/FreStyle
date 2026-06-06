package com.normanblog.frestyle.s3;

import java.time.Instant;

/** S3 オブジェクトキーと拡張子の組み立てを担う小道具(presigner 実装で共有する)。 */
final class ImageKeys {

  private ImageKeys() {}

  /** {@code profiles/{userId}/{epochNanos}{ext}} 形式のキーを作る(Go 版と同じ命名規則)。 */
  static String profileKey(Long userId, String fileName, String contentType) {
    Instant now = Instant.now();
    long epochNanos = now.getEpochSecond() * 1_000_000_000L + now.getNano();
    return "profiles/" + userId + "/" + epochNanos + guessExt(fileName, contentType);
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
