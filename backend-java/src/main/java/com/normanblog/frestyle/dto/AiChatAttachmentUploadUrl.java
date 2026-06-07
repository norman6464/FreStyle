package com.normanblog.frestyle.dto;

/**
 * AI チャット添付の S3 PUT 署名付き URL 発行結果。Go 版の JSON 形と互換のフィールド名。
 *
 * <p>クライアントは {@code uploadUrl} に PUT 後、{@code key} を SSE 送信ペイロードの
 * attachments[].key に詰める。
 */
public record AiChatAttachmentUploadUrl(String uploadUrl, String key, int expiresIn) {}
