package com.normanblog.frestyle.infra.s3;

import com.normanblog.frestyle.dto.AiChatAttachmentUploadUrl;

/** AWS を呼ばない stub。bucket 未設定のローカル / テスト環境で使う。 */
public class StubAiChatAttachmentPresigner implements AiChatAttachmentPresigner {

  private static final int TTL_SECONDS = 600;

  private final String bucket;

  public StubAiChatAttachmentPresigner(String bucket) {
    this.bucket = (bucket == null || bucket.isBlank()) ? "stub-bucket" : bucket;
  }

  @Override
  public AiChatAttachmentUploadUrl generate(Long userId, String filename, String contentType) {
    String key = ImageKeys.aiChatAttachmentKey(userId, filename, contentType);
    String uploadUrl = "https://" + bucket + ".s3.amazonaws.com/" + key + "?X-Amz-Stub=1";

    return new AiChatAttachmentUploadUrl(uploadUrl, key, TTL_SECONDS);
  }
}
