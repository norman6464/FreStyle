package com.normanblog.frestyle.entity;

/** users.role の許容値。DB には文字列で保存する(既存 Go 実装と同値)。 */
public final class Role {

  public static final String SUPER_ADMIN = "super_admin";
  public static final String COMPANY_ADMIN = "company_admin";
  public static final String TRAINEE = "trainee";

  private Role() {}
}
