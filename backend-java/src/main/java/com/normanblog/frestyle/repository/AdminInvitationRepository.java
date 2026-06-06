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

  // email フォールバックも期限切れを弾く。複数 pending は新しい順に 1 件。
  Optional<AdminInvitation> findFirstByEmailAndStatusAndExpiresAtAfterOrderByCreatedAtDesc(
      String email, String status, Instant now);
}
