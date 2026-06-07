package com.normanblog.frestyle.infra.s3;

/**
 * S3 を呼ばない stub。bucket 未設定のローカル / テスト環境で使う。
 *
 * <p>常に null を返す(添付なし扱い)。本番では {@link S3AttachmentDownloader} に差し替わる。
 */
public class StubAttachmentDownloader implements AttachmentDownloader {

  @Override
  public byte[] download(String key) {
    return null;
  }
}
