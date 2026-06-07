package com.normanblog.frestyle.infra.s3;

/** S3 から AI チャット添付の実体(バイト列)を取得する境界。 */
public interface AttachmentDownloader {

  /**
   * 指定キーのオブジェクトを取得する。取得失敗時は null を返し、呼び出し側で「添付なし」に劣化させる。
   */
  byte[] download(String key);
}
