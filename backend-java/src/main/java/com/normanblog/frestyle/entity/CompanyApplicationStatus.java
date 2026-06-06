package com.normanblog.frestyle.entity;

import java.util.Set;

/** company_applications.status の許容値(既存 Go 実装と同値)。 */
public final class CompanyApplicationStatus {

  public static final String PENDING = "pending";
  public static final String APPROVED = "approved";
  public static final String REJECTED = "rejected";

  public static final Set<String> ALL = Set.of(PENDING, APPROVED, REJECTED);

  private CompanyApplicationStatus() {}
}
