package com.normanblog.frestyle.dto;

import com.fasterxml.jackson.annotation.JsonInclude;
import java.util.List;

/**
 * AI チャットメッセージ 1 件のクライアント向け表現。Go 版 domain.AiChatMessage と互換のフィールド名。
 *
 * <p>実体は DynamoDB に保存され、API ではこの形で返す。attachments は無い場合 null(省略)。
 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public record AiChatMessageResponse(
    Long sessionId,
    String messageId,
    String role,
    String content,
    List<AiChatAttachmentDto> attachments,
    String createdAt) {}
