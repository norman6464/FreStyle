package com.normanblog.frestyle.service;

import static org.assertj.core.api.Assertions.assertThat;

import com.normanblog.frestyle.dto.AiChatMessageResponse;
import com.normanblog.frestyle.dto.AiChatStreamRequest;
import com.normanblog.frestyle.entity.AiChatSession;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.AiChatSessionRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import java.util.List;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

/**
 * AI チャット送信オーケストレーションの中核を検証する。テスト環境は AWS 未設定のため、Bedrock /
 * DynamoDB / S3 はすべて stub(stub Bedrock は固定応答を数チャンクで返す)。
 */
@SpringBootTest
class SendAiMessageUseCaseTest {

  @Autowired private SendAiMessageUseCase useCase;
  @Autowired private UserRepository users;
  @Autowired private AiChatSessionRepository sessions;

  private User actor;

  @BeforeEach
  void setUp() {
    sessions.deleteAll();
    users.deleteAll();
    actor =
        users.save(
            User.builder().cognitoSub("send-sub").email("me@example.com").role(Role.TRAINEE).build());
  }

  /** イベントを記録する listener。 */
  static class Capturing implements AiMessageStreamListener {
    AiChatSession session;
    final StringBuilder tokens = new StringBuilder();
    AiChatMessageResponse done;
    String error;

    @Override
    public void onSession(AiChatSession s) {
      this.session = s;
    }

    @Override
    public void onToken(String delta) {
      tokens.append(delta);
    }

    @Override
    public void onDone(AiChatMessageResponse finalMessage) {
      this.done = finalMessage;
    }

    @Override
    public void onError(String message) {
      this.error = message;
    }
  }

  @Test
  void execute_newSession_emitsSessionTokensAndDone() {
    AiChatStreamRequest request =
        new AiChatStreamRequest(null, "こんにちは", "free", null, List.of());
    Capturing listener = new Capturing();

    useCase.execute(actor, request, listener);

    // 新規セッションが作られ、onSession で通知される。
    assertThat(listener.session).isNotNull();
    assertThat(listener.session.getUserId()).isEqualTo(actor.getId());
    assertThat(listener.session.getTitle()).isEqualTo("こんにちは");
    assertThat(sessions.findById(listener.session.getId())).isPresent();

    // stub Bedrock の固定応答がトークンとして流れ、done の content と一致する。
    assertThat(listener.error).isNull();
    assertThat(listener.done).isNotNull();
    assertThat(listener.done.role()).isEqualTo("assistant");
    assertThat(listener.tokens.toString()).isNotEmpty();
    assertThat(listener.done.content()).isEqualTo(listener.tokens.toString());
  }

  @Test
  void execute_existingSession_doesNotCreateNewSession() {
    AiChatSession existing =
        sessions.save(
            AiChatSession.builder()
                .userId(actor.getId())
                .title("既存")
                .sessionType("free")
                .createdAt(Instant.now())
                .updatedAt(Instant.now())
                .build());
    AiChatStreamRequest request =
        new AiChatStreamRequest(existing.getId(), "続き", "free", null, List.of());
    Capturing listener = new Capturing();

    useCase.execute(actor, request, listener);

    // 既存セッション指定では onSession は呼ばれない。
    assertThat(listener.session).isNull();
    assertThat(listener.done).isNotNull();
    assertThat(sessions.count()).isEqualTo(1);
  }
}
