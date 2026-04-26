package com.example.FreStyle.usecase;

import java.security.SecureRandom;
import java.time.LocalDateTime;
import java.util.Base64;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.InvitationDto;
import com.example.FreStyle.entity.Invitation;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.entity.UserIdentity;
import com.example.FreStyle.form.CreateInvitationForm;
import com.example.FreStyle.repository.InvitationRepository;
import com.example.FreStyle.repository.UserIdentityRepository;
import com.example.FreStyle.repository.UserRepository;
import com.example.FreStyle.service.CognitoAdminService;
import com.example.FreStyle.service.CognitoAdminService.AdminCreatedUser;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

/**
 * CompanyAdmin が Trainee を招待するユースケース。
 *
 * <p>処理の流れ:</p>
 * <ol>
 *   <li>Cognito にユーザーを管理者作成（標準の招待メールが送信される）</li>
 *   <li>users テーブルに Trainee を INSERT (既存なら role / company_id 上書き)</li>
 *   <li>user_identities に Cognito sub をマッピング</li>
 *   <li>invitations テーブルに招待トークンと有効期限を記録</li>
 * </ol>
 *
 * <p>SES で独自テンプレ送信に置き換える場合は CognitoAdminService の MessageAction を SUPPRESS
 * にし、本 UseCase で別途メール送信する設計に拡張する（Phase 2）。</p>
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class CreateInvitationUseCase {

    private static final SecureRandom RANDOM = new SecureRandom();
    private static final int EXPIRES_IN_DAYS = 14;

    private final InvitationRepository invitationRepository;
    private final UserRepository userRepository;
    private final UserIdentityRepository userIdentityRepository;
    private final CognitoAdminService cognitoAdminService;

    @Transactional
    public InvitationDto execute(Long companyId, Long inviterUserId, CreateInvitationForm form) {
        String email = form.email().trim().toLowerCase();
        String role = form.role();
        String displayName = form.displayName() != null && !form.displayName().isBlank()
                ? form.displayName()
                : email.substring(0, email.indexOf('@'));

        // 1) Cognito 招待 (既存ユーザーは alreadyExisted=true で返るので冪等)
        AdminCreatedUser cognitoUser = cognitoAdminService.inviteUser(email, displayName);

        // 2) users INSERT or UPDATE
        User user = userRepository.findByEmail(email)
                .orElseGet(() -> {
                    User u = new User();
                    u.setEmail(email);
                    u.setName(displayName);
                    u.setIsActive(true);
                    return u;
                });
        user.setCompanyId(companyId);
        user.setRole(role);
        if (user.getName() == null || user.getName().isBlank()) {
            user.setName(displayName);
        }
        user = userRepository.save(user);

        // 3) Cognito sub のマッピング (既に存在しない場合のみ)
        if (cognitoUser.sub() != null) {
            boolean exists = userIdentityRepository
                    .findByProviderAndProviderSub("cognito", cognitoUser.sub())
                    .isPresent();
            if (!exists) {
                UserIdentity identity = new UserIdentity();
                identity.setUser(user);
                identity.setProvider("cognito");
                identity.setProviderSub(cognitoUser.sub());
                userIdentityRepository.save(identity);
            }
        }

        // 4) invitations 記録
        String token = generateToken();
        Invitation invitation = new Invitation();
        invitation.setCompanyId(companyId);
        invitation.setEmail(email);
        invitation.setRole(role);
        invitation.setToken(token);
        invitation.setInvitedBy(inviterUserId);
        invitation.setExpiresAt(LocalDateTime.now().plusDays(EXPIRES_IN_DAYS));
        invitation = invitationRepository.save(invitation);

        log.info("[CreateInvitationUseCase] companyId={} email={} role={} cognitoExists={}",
                companyId, email, role, cognitoUser.alreadyExisted());

        return toDto(invitation);
    }

    private static String generateToken() {
        byte[] bytes = new byte[32];
        RANDOM.nextBytes(bytes);
        return Base64.getUrlEncoder().withoutPadding().encodeToString(bytes);
    }

    private static InvitationDto toDto(Invitation i) {
        return new InvitationDto(
                i.getId(), i.getCompanyId(), i.getEmail(), i.getRole(),
                i.getInvitedBy(), i.getExpiresAt(), i.getAcceptedAt(),
                i.getAcceptedUserId(), i.getCreatedAt()
        );
    }
}
