package com.example.FreStyle.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.Invitation;

@Repository
public interface InvitationRepository extends JpaRepository<Invitation, Long> {

    /** 会社別の未承諾招待を新しい順で取得 */
    List<Invitation> findByCompanyIdAndAcceptedAtIsNullOrderByCreatedAtDesc(Long companyId);

    /** トークンから招待を取得 */
    Optional<Invitation> findByToken(String token);

    /** 同一会社・同一メールの未承諾招待が存在するか */
    Optional<Invitation> findByCompanyIdAndEmailAndAcceptedAtIsNull(Long companyId, String email);
}
