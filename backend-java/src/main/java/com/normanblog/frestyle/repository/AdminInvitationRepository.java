package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.AdminInvitation;
import java.time.Instant;
import java.util.Optional;
import org.springframework.data.jpa.repository.JpaRepository;

/** invitations テーブルへのアクセスを担うリポジトリ。 */
public interface AdminInvitationRepository extends JpaRepository<AdminInvitation, Long> {

  // token は期限切れも弾く(マジックリンクの寿命を尊重)。
  Optional<AdminInvitation> findFirstByTokenAndStatusAndExpiresAtAfter(
      String token, String status, Instant now);

  // email フォールバックは複数 pending を考慮して新しい順に 1 件。
  Optional<AdminInvitation> findFirstByEmailAndStatusOrderByCreatedAtDesc(String email, String status);
}
