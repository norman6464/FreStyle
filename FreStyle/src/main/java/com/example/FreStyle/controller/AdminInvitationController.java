package com.example.FreStyle.controller;

import java.util.List;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.InvitationDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.CreateInvitationForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.CancelInvitationUseCase;
import com.example.FreStyle.usecase.CreateInvitationUseCase;
import com.example.FreStyle.usecase.ListInvitationsUseCase;

import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

/**
 * 管理者専用: 招待管理 API。
 *
 * <p>{@code /api/admin/**} は SecurityConfig で {@code hasRole("admin")} 制限済み。</p>
 *
 * <p>招待作成時に Cognito Admin Create User で標準の招待メール（一時パスワード付き）が
 * 招待先に送信される。受信者は Hosted UI でパスワードを変更してログインすると、
 * 既に DB 側で会社マッピング (company_id, role=trainee) が設定されているため、
 * そのまま自社のコース・教材にアクセスできる。</p>
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/api/admin/invitations")
@Slf4j
@Tag(name = "Admin Invitations", description = "管理者専用: 招待メール送信とトラッキング")
public class AdminInvitationController {

    private final CreateInvitationUseCase createUseCase;
    private final ListInvitationsUseCase listUseCase;
    private final CancelInvitationUseCase cancelUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping
    public ResponseEntity<List<InvitationDto>> list(@AuthenticationPrincipal Jwt jwt) {
        Long companyId = resolveCompanyId(jwt);
        return ResponseEntity.ok(listUseCase.execute(companyId));
    }

    @PostMapping
    public ResponseEntity<InvitationDto> create(
            @AuthenticationPrincipal Jwt jwt,
            @Valid @RequestBody CreateInvitationForm form
    ) {
        User inviter = userIdentityService.findUserBySub(jwt.getSubject());
        Long companyId = inviter.getCompanyId();
        if (companyId == null) {
            // super_admin は company_id NULL なので、明示的に form 経由で指定する設計に
            // 拡張する必要がある。MVP では company_admin (自社) のみサポート。
            throw new IllegalStateException("会社に所属していないユーザーは招待を発行できません");
        }
        log.info("[AdminInvitationController] create invitation companyId={} email={} by userId={}",
                companyId, form.email(), inviter.getId());
        return ResponseEntity.ok(createUseCase.execute(companyId, inviter.getId().longValue(), form));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> cancel(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Long id
    ) {
        Long companyId = resolveCompanyId(jwt);
        cancelUseCase.execute(companyId, id);
        return ResponseEntity.noContent().build();
    }

    private Long resolveCompanyId(Jwt jwt) {
        User user = userIdentityService.findUserBySub(jwt.getSubject());
        if (user.getCompanyId() == null) {
            throw new IllegalStateException("会社に所属していないユーザー");
        }
        return user.getCompanyId();
    }
}
