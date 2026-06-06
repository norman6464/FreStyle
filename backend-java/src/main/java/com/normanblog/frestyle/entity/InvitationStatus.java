package com.normanblog.frestyle.entity;

/** invitations.status の許容値(既存 Go 実装と同値)。 */
public final class InvitationStatus {

  public static final String PENDING = "pending";
  public static final String ACCEPTED = "accepted";
  public static final String CANCELED = "canceled";

  private InvitationStatus() {}
}
