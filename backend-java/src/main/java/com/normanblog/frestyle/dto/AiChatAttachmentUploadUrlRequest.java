package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Positive;

/** AI チャット添付の署名 URL 発行リクエスト。 */
public record AiChatAttachmentUploadUrlRequest(
    @NotBlank String filename, @NotBlank String contentType, @Positive long sizeBytes) {}
