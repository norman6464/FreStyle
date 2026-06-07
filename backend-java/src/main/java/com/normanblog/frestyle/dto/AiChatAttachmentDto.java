package com.normanblog.frestyle.dto;

/**
 * AI チャットメッセージに添付されたファイルのメタ情報。Go 版 domain.Attachment と互換
 * (BlobData は永続化しないため含まない)。
 */
public record AiChatAttachmentDto(
    String key,
    String filename,
    String contentType,
    String format,
    String kind,
    long sizeBytes) {}
