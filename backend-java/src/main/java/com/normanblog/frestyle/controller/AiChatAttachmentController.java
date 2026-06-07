package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.AiChatAttachmentUploadUrl;
import com.normanblog.frestyle.dto.AiChatAttachmentUploadUrlRequest;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.AiChatAttachmentService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * AI チャット添付の S3 PUT 署名付き URL を発行するコントローラ。
 *
 * <p>許容 MIME(画像)以外は 415、サイズ上限超過は 413、必須欠落は 400。
 */
@RestController
@RequestMapping("/api/v2/ai-chat/attachments")
public class AiChatAttachmentController {

  private final AiChatAttachmentService attachments;
  private final CurrentUserProvider currentUser;

  public AiChatAttachmentController(
      AiChatAttachmentService attachments, CurrentUserProvider currentUser) {
    this.attachments = attachments;
    this.currentUser = currentUser;
  }

  @PostMapping("/upload-url")
  public AiChatAttachmentUploadUrl issueUploadUrl(
      @Valid @RequestBody AiChatAttachmentUploadUrlRequest request) {
    Long userId = currentUser.require().getId();

    return attachments.issueUploadUrl(
        userId, request.filename(), request.contentType(), request.sizeBytes());
  }
}
