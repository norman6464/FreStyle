package com.normanblog.frestyle.service;

import static org.assertj.core.api.Assertions.assertThat;

import com.normanblog.frestyle.infra.cognito.IdTokenClaims;
import com.normanblog.frestyle.entity.AdminInvitation;
import com.normanblog.frestyle.entity.InvitationStatus;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.AdminInvitationRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.List;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

/** 招待ゲート(新規ユーザーの許可/拒否)と role/company 反映を検証する。 */
@SpringBootTest
class UserProvisioningServiceTest {

  @Autowired private UserProvisioningService provisioning;
  @Autowired private UserRepository users;
  @Autowired private AdminInvitationRepository invitations;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    invitations.deleteAll();
  }

  private IdTokenClaims claims(String sub, String email, List<String> groups) {
    return new IdTokenClaims(sub, email, null, groups);
  }

  @Test
  void newUser_withoutInvitationOrAdmin_isBlocked() {
    boolean allowed = provisioning.upsertFromIdToken(claims("s1", "a@x.com", List.of()), null);
    assertThat(allowed).isFalse();
    assertThat(users.findByCognitoSubAndDeletedAtIsNull("s1")).isEmpty();
  }

  @Test
  void newUser_asCognitoAdmin_isCreatedAsSuperAdmin() {
    boolean allowed = provisioning.upsertFromIdToken(claims("s2", "a@x.com", List.of("admin")), null);
    assertThat(allowed).isTrue();
    User u = users.findByCognitoSubAndDeletedAtIsNull("s2").orElseThrow();
    assertThat(u.getRole()).isEqualTo(Role.SUPER_ADMIN);
  }

  @Test
  void newUser_withPendingInvitation_isCreatedWithInvitedRoleAndCompany() {
    invitations.save(
        AdminInvitation.builder()
            .email("inv@x.com")
            .role(Role.COMPANY_ADMIN)
            .companyId(42L)
            .displayName("Invited")
            .status(InvitationStatus.PENDING)
            .expiresAt(Instant.now().plus(1, ChronoUnit.DAYS))
            .createdAt(Instant.now())
            .build());

    boolean allowed = provisioning.upsertFromIdToken(claims("s3", "inv@x.com", List.of()), null);

    assertThat(allowed).isTrue();
    User u = users.findByCognitoSubAndDeletedAtIsNull("s3").orElseThrow();
    assertThat(u.getRole()).isEqualTo(Role.COMPANY_ADMIN);
    assertThat(u.getCompanyId()).isEqualTo(42L);
    assertThat(u.getDisplayName()).isEqualTo("Invited");
    // 招待は accepted にマークされ再利用できない。
    assertThat(
            invitations.findFirstByEmailAndStatusAndExpiresAtAfterOrderByCreatedAtDesc(
                "inv@x.com", InvitationStatus.PENDING, Instant.now()))
        .isEmpty();
  }

  @Test
  void newUser_withExpiredTokenInvitation_isBlocked() {
    invitations.save(
        AdminInvitation.builder()
            .email("exp@x.com")
            .role(Role.TRAINEE)
            .token("tok-expired")
            .status(InvitationStatus.PENDING)
            .expiresAt(Instant.now().minus(1, ChronoUnit.DAYS))
            .createdAt(Instant.now())
            .build());

    // token は期限切れで弾かれ、email フォールバックも無い(別 email)ので拒否。
    boolean allowed =
        provisioning.upsertFromIdToken(claims("s4", "other@x.com", List.of()), "tok-expired");
    assertThat(allowed).isFalse();
  }

  @Test
  void newUser_withTokenInvitationButMismatchedEmail_isBlocked() {
    invitations.save(
        AdminInvitation.builder()
            .email("invited@x.com")
            .role(Role.TRAINEE)
            .token("tok1")
            .status(InvitationStatus.PENDING)
            .expiresAt(Instant.now().plus(1, ChronoUnit.DAYS))
            .createdAt(Instant.now())
            .build());

    // トークンは有効だがログイン email が招待先と違う → 横取り防止で拒否。
    boolean allowed =
        provisioning.upsertFromIdToken(claims("s9", "attacker@x.com", List.of()), "tok1");
    assertThat(allowed).isFalse();
  }

  @Test
  void existingUser_withAdminGroup_isPromotedToSuperAdmin() {
    users.save(User.builder().cognitoSub("s5").email("e@x.com").role(Role.TRAINEE).build());

    boolean allowed = provisioning.upsertFromIdToken(claims("s5", "e@x.com", List.of("admin")), null);

    assertThat(allowed).isTrue();
    assertThat(users.findByCognitoSubAndDeletedAtIsNull("s5").orElseThrow().getRole())
        .isEqualTo(Role.SUPER_ADMIN);
  }
}
