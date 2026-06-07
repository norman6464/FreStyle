package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.AiChatSession;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;

/** ai_chat_sessions テーブルへのアクセスを担うリポジトリ。 */
public interface AiChatSessionRepository extends JpaRepository<AiChatSession, Long> {

  List<AiChatSession> findByUserIdOrderByCreatedAtDesc(Long userId);
}
