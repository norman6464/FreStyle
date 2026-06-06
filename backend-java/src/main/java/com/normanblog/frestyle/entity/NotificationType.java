package com.normanblog.frestyle.entity;

/** notifications.type の既知の値。DB には文字列で保存する(既存 Go 実装と同値)。 */
public final class NotificationType {

  // 企業の利用申請が届いたことを super_admin に知らせる通知。
  public static final String COMPANY_APPLICATION = "company_application";

  private NotificationType() {}
}
