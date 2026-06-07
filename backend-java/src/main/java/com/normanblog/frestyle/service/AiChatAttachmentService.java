package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.AiChatAttachmentUploadUrl;
import com.normanblog.frestyle.infra.s3.AiChatAttachmentPresigner;
import java.util.Set;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

/**
 * AI チャット添付の署名 URL 発行を担うサービス。
 *
 * <p>contentType の許容セットとサイズ上限を {@link AiChatAttachmentRules} で検証し、不正な添付は
 * presigned URL 発行前に弾く(未対応 MIME → 415 / サイズ超過 → 413)。
 */
@Service
public class AiChatAttachmentService {

  private final AiChatAttachmentPresigner presigner;

  public AiChatAttachmentService(AiChatAttachmentPresigner presigner) {
    this.presigner = presigner;
  }

  /** 添付の署名 URL を発行する。MIME 許容セットとサイズ上限を検証する。 */
  public AiChatAttachmentUploadUrl issueUploadUrl(
      Long userId, String filename, String contentType, long sizeBytes) {
    AiChatAttachmentRules.Rule rule = AiChatAttachmentRules.of(contentType);
    if (rule == null) {
      throw new ResponseStatusException(
          HttpStatus.UNSUPPORTED_MEDIA_TYPE, "attachment: unsupported content type");
    }
    if (sizeBytes <= 0 || sizeBytes > rule.maxBytes()) {
      throw new ResponseStatusException(
          HttpStatus.CONTENT_TOO_LARGE, "attachment: file too large");
    }

    return presigner.generate(userId, filename, contentType);
  }

  /** 許容 MIME 一覧(テスト等から参照する)。 */
  public Set<String> allowedContentTypes() {
    return AiChatAttachmentRules.allowedContentTypes();
  }
}
