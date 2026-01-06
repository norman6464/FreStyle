package com.example.FreStyle.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.AiChatMessage;
import com.example.FreStyle.entity.AiChatSession;

@Repository
public interface AiChatMessageRepository extends JpaRepository<AiChatMessage, Integer> {

    /**
     * 指定セッションのメッセージ一覧を作成日時昇順で取得
     */
    List<AiChatMessage> findBySessionOrderByCreatedAtAsc(AiChatSession session);

    /**
     * 指定セッションIDのメッセージ一覧を作成日時昇順で取得
     */
    @Query("SELECT m FROM AiChatMessage m WHERE m.session.id = :sessionId ORDER BY m.createdAt ASC")
    List<AiChatMessage> findBySessionIdOrderByCreatedAtAsc(@Param("sessionId") Integer sessionId);

    /**
     * 指定ユーザーの全メッセージを取得（ユーザーIDで検索）
     */
    @Query("SELECT m FROM AiChatMessage m WHERE m.user.id = :userId ORDER BY m.createdAt ASC")
    List<AiChatMessage> findByUserIdOrderByCreatedAtAsc(@Param("userId") Integer userId);

    /**
     * 指定セッションのメッセージ数をカウント
     */
    @Query("SELECT COUNT(m) FROM AiChatMessage m WHERE m.session.id = :sessionId")
    Long countBySessionId(@Param("sessionId") Integer sessionId);

}
