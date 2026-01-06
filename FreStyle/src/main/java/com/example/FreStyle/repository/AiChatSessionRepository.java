package com.example.FreStyle.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;

@Repository
public interface AiChatSessionRepository extends JpaRepository<AiChatSession, Integer> {

    /**
     * 指定ユーザーのセッション一覧を作成日時降順で取得
     */
    List<AiChatSession> findByUserOrderByCreatedAtDesc(User user);

    /**
     * 指定ユーザーのセッション一覧をユーザーIDで取得（作成日時降順）
     */
    @Query("SELECT s FROM AiChatSession s WHERE s.user.id = :userId ORDER BY s.createdAt DESC")
    List<AiChatSession> findByUserIdOrderByCreatedAtDesc(@Param("userId") Integer userId);

    /**
     * 指定ルームに関連するセッションを取得
     */
    @Query("SELECT s FROM AiChatSession s WHERE s.relatedRoom.id = :roomId ORDER BY s.createdAt DESC")
    List<AiChatSession> findByRelatedRoomId(@Param("roomId") Integer roomId);

    /**
     * 指定ユーザーの特定セッションを取得（権限チェック用）
     */
    @Query("SELECT s FROM AiChatSession s WHERE s.id = :sessionId AND s.user.id = :userId")
    Optional<AiChatSession> findByIdAndUserId(@Param("sessionId") Integer sessionId, @Param("userId") Integer userId);

}
