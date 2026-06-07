package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.AiChatMessageResponse;
import com.normanblog.frestyle.infra.dynamo.AiChatMessageReader;
import com.normanblog.frestyle.entity.User;
import java.util.List;
import org.springframework.stereotype.Service;

/**
 * AI チャットメッセージ(DynamoDB)の読み取りを担うサービス。
 *
 * <p>セッションの所有者検証を行ってから DynamoDB を引く(他人のセッションのメッセージは読めない)。
 * Go 版にはこの所有者検証が無かったため IDOR 対策として追加。
 */
@Service
public class AiChatMessageService {

  private final AiChatMessageReader reader;
  private final AiChatSessionService sessions;

  public AiChatMessageService(AiChatMessageReader reader, AiChatSessionService sessions) {
    this.reader = reader;
    this.sessions = sessions;
  }

  /** 自分のセッションのメッセージを作成順で返す。他人のセッションは 403、存在しなければ 404。 */
  public List<AiChatMessageResponse> listMessages(Long sessionId, User actor) {
    // 所有者検証(他人 403 / 不在 404)を session 側に委譲してから読む。
    sessions.get(sessionId, actor);

    return reader.listBySession(sessionId);
  }
}
