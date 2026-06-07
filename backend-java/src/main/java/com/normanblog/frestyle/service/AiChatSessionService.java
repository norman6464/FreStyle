package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.AiChatSession;
import com.normanblog.frestyle.entity.AiChatSessionType;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.AiChatSessionRepository;
import java.time.Instant;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.server.ResponseStatusException;

/**
 * AI チャットセッション(メタデータ)の一覧・作成・取得・タイトル更新・削除を担う。
 *
 * <p>取得 / 更新 / 削除は所有者検証を行い、他人のセッションは 403 にする(Go 版に無かった IDOR 対策を追加)。
 */
@Service
public class AiChatSessionService {

  private final AiChatSessionRepository sessions;

  public AiChatSessionService(AiChatSessionRepository sessions) {
    this.sessions = sessions;
  }

  /** current user のセッションを作成日降順で返す。 */
  public List<AiChatSession> list(Long userId) {
    return sessions.findByUserIdOrderByCreatedAtDesc(userId);
  }

  /** セッションを作成する。sessionType 省略時は free。 */
  @Transactional
  public AiChatSession create(Long userId, String title, String sessionType, Long scenarioId) {
    Instant now = Instant.now();
    String type = (sessionType == null || sessionType.isBlank()) ? AiChatSessionType.FREE : sessionType;

    return sessions.save(
        AiChatSession.builder()
            .userId(userId)
            .title(title)
            .sessionType(type)
            .scenarioId(scenarioId)
            .createdAt(now)
            .updatedAt(now)
            .build());
  }

  /** 自分のセッションを取得する。他人のものは 403、存在しなければ 404。 */
  public AiChatSession get(Long id, User actor) {
    return requireOwned(id, actor);
  }

  /** 自分のセッションのタイトルを更新する。所有者検証つき。 */
  @Transactional
  public AiChatSession updateTitle(Long id, User actor, String title) {
    AiChatSession session = requireOwned(id, actor);
    session.setTitle(title);
    session.setUpdatedAt(Instant.now());

    return sessions.save(session);
  }

  /** 自分のセッションを削除する。所有者検証つき。 */
  @Transactional
  public void delete(Long id, User actor) {
    sessions.delete(requireOwned(id, actor));
  }

  // 指定セッションが存在し、かつ actor の所有であることを保証する。
  private AiChatSession requireOwned(Long id, User actor) {
    AiChatSession session =
        sessions
            .findById(id)
            .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "not_found"));
    if (!session.getUserId().equals(actor.getId())) {
      throw new ResponseStatusException(HttpStatus.FORBIDDEN, "forbidden");
    }

    return session;
  }
}
