package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.AiChatAttachmentUploadUrl;
import com.normanblog.frestyle.s3.AiChatAttachmentPresigner;
import java.util.Map;
import java.util.Set;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

/**
 * AI チャット添付の署名 URL 発行を担うサービス。
 *
 * <p>contentType の許容セットとサイズ上限を検証し、不正な添付は presigned URL 発行前に弾く
 * (未対応 MIME → 415 / サイズ超過 → 413)。
 */
@Service
public class AiChatAttachmentService {

  private static final long MAX_IMAGE_BYTES = 5L * 1024 * 1024;

  // presigned URL 発行 / SSE 送信が許可される MIME とサイズ上限(現状は画像のみ / Go 版と同値)。
  private static final Map<String, Long> ALLOWED_MAX_BYTES =
      Map.of(
          "image/png", MAX_IMAGE_BYTES,
          "image/jpeg", MAX_IMAGE_BYTES,
          "image/jpg", MAX_IMAGE_BYTES,
          "image/gif", MAX_IMAGE_BYTES,
          "image/webp", MAX_IMAGE_BYTES);

  private final AiChatAttachmentPresigner presigner;

  public AiChatAttachmentService(AiChatAttachmentPresigner presigner) {
    this.presigner = presigner;
  }

  /** 添付の署名 URL を発行する。MIME 許容セットとサイズ上限を検証する。 */
  public AiChatAttachmentUploadUrl issueUploadUrl(
      Long userId, String filename, String contentType, long sizeBytes) {
    Long max = ALLOWED_MAX_BYTES.get(contentType);
    if (max == null) {
      throw new ResponseStatusException(
          HttpStatus.UNSUPPORTED_MEDIA_TYPE, "attachment: unsupported content type");
    }
    if (sizeBytes <= 0 || sizeBytes > max) {
      throw new ResponseStatusException(
          HttpStatus.CONTENT_TOO_LARGE, "attachment: file too large");
    }

    return presigner.generate(userId, filename, contentType);
  }

  /** 許容 MIME 一覧(テスト等から参照する)。 */
  public Set<String> allowedContentTypes() {
    return ALLOWED_MAX_BYTES.keySet();
  }
}
