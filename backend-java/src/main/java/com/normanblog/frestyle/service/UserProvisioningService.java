package com.normanblog.frestyle.service;

import com.normanblog.frestyle.cognito.IdTokenClaims;
import com.normanblog.frestyle.config.CognitoProperties;
import com.normanblog.frestyle.entity.AdminInvitation;
import com.normanblog.frestyle.entity.InvitationStatus;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.AdminInvitationRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import java.util.Optional;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/**
 * ログイン時の users 行の作成・更新と「招待ゲート」を担う。
 *
 * <p>新規ユーザーは「pending な招待を持つ」か「Cognito の admin グループ」のいずれかでなければ
 * 作成を拒否する(招待制の B2B SaaS のため)。既存ユーザーは role / company / displayName を同期する。
 */
@Service
public class UserProvisioningService {

  private final UserRepository users;
  private final AdminInvitationRepository invitations;
  private final String adminGroup;

  public UserProvisioningService(
      UserRepository users, AdminInvitationRepository invitations, CognitoProperties cognito) {
    this.users = users;
    this.invitations = invitations;
    this.adminGroup = cognito.adminGroup();
  }

  /**
   * id_token のクレームから users 行を upsert する。
   *
   * @return ログインを許可する場合 true。招待も admin グループも無い新規ユーザーは false(403 相当)。
   */
  @Transactional
  public boolean upsertFromIdToken(IdTokenClaims claims, String invitationToken) {
    String sub = claims.sub();
    if (sub == null || sub.isBlank()) {
      return false;
    }
    boolean isCognitoAdmin = claims.groups().contains(adminGroup);
    AdminInvitation invitation = findInvitation(invitationToken, claims.email());

    Optional<User> existing = users.findByCognitoSubAndDeletedAtIsNull(sub);
    if (existing.isPresent()) {
      updateExisting(existing.get(), claims, isCognitoAdmin, invitation);
      return true;
    }
    return createNew(sub, claims, isCognitoAdmin, invitation);
  }

  // 招待検索: invitationToken 優先(期限内 pending)、無ければ email でフォールバック(pending の最新)。
  private AdminInvitation findInvitation(String token, String email) {
    if (token != null && !token.isBlank()) {
      Optional<AdminInvitation> byToken =
          invitations.findFirstByTokenAndStatusAndExpiresAtAfter(
              token, InvitationStatus.PENDING, Instant.now());
      if (byToken.isPresent()) {
        return byToken.get();
      }
    }
    if (email != null && !email.isBlank()) {
      return invitations
          .findFirstByEmailAndStatusOrderByCreatedAtDesc(email, InvitationStatus.PENDING)
          .orElse(null);
    }
    return null;
  }

  private void updateExisting(
      User user, IdTokenClaims claims, boolean isCognitoAdmin, AdminInvitation invitation) {
    boolean changed = false;

    // displayName が email のまま(旧フローの仮値)なら OIDC name で補正する。
    if (claims.name() != null
        && !claims.name().isBlank()
        && claims.email() != null
        && claims.email().equals(user.getDisplayName())) {
      user.setDisplayName(claims.name());
      changed = true;
    }
    // Cognito admin は super_admin に昇格(最優先・降格はしない)。
    if (isCognitoAdmin && !Role.SUPER_ADMIN.equals(user.getRole())) {
      user.setRole(Role.SUPER_ADMIN);
      changed = true;
    }
    // 招待による昇格は super_admin でないユーザーにのみ適用する。
    if (invitation != null && !Role.SUPER_ADMIN.equals(user.getRole())) {
      if (Role.TRAINEE.equals(user.getRole()) && Role.COMPANY_ADMIN.equals(invitation.getRole())) {
        user.setRole(Role.COMPANY_ADMIN);
        changed = true;
      }
      if (hasCompany(invitation)
          && (user.getCompanyId() == null
              || !user.getCompanyId().equals(invitation.getCompanyId()))) {
        user.setCompanyId(invitation.getCompanyId());
        changed = true;
      }
      markAccepted(invitation);
    }
    if (changed) {
      user.setUpdatedAt(Instant.now());
      users.save(user);
    }
  }

  private boolean createNew(
      String sub, IdTokenClaims claims, boolean isCognitoAdmin, AdminInvitation invitation) {
    // 招待も admin グループも無ければ新規ユーザーは作らない(ログイン拒否)。
    if (!isCognitoAdmin && invitation == null) {
      return false;
    }
    String role = isCognitoAdmin ? Role.SUPER_ADMIN : Role.TRAINEE;
    Long companyId = null;
    String displayName =
        (claims.name() != null && !claims.name().isBlank()) ? claims.name() : claims.email();
    if (invitation != null) {
      // 招待の role(company_admin / trainee)を優先する。
      if (Role.COMPANY_ADMIN.equals(invitation.getRole())
          || Role.TRAINEE.equals(invitation.getRole())) {
        role = invitation.getRole();
      }
      companyId = invitation.getCompanyId();
      if (invitation.getDisplayName() != null && !invitation.getDisplayName().isBlank()) {
        displayName = invitation.getDisplayName();
      }
    }
    Instant now = Instant.now();
    users.save(
        User.builder()
            .cognitoSub(sub)
            .email(claims.email())
            .displayName(displayName)
            .role(role)
            .companyId(companyId)
            .createdAt(now)
            .updatedAt(now)
            .build());
    if (invitation != null) {
      markAccepted(invitation);
    }
    return true;
  }

  private boolean hasCompany(AdminInvitation invitation) {
    return invitation.getCompanyId() != null && invitation.getCompanyId() != 0L;
  }

  // 招待を accepted にして再利用を防ぐ(監査履歴にも残る)。
  private void markAccepted(AdminInvitation invitation) {
    invitation.setStatus(InvitationStatus.ACCEPTED);
    invitations.save(invitation);
  }
}
