package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.AiChatSessionResponse;
import com.normanblog.frestyle.dto.CreateAiChatSessionRequest;
import com.normanblog.frestyle.dto.UpdateSessionTitleRequest;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.AiChatSessionService;
import jakarta.validation.Valid;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * AI チャットセッション(メタデータ)の CRUD。常に認証ユーザー自身のセッションのみ操作できる
 * (取得 / 更新 / 削除は所有者検証つき)。
 *
 * <p>メッセージ本体(DynamoDB)と SSE ストリーム(Bedrock)は別 PR。
 */
@RestController
@RequestMapping("/api/v2/ai-chat/sessions")
public class AiChatSessionController {

  private final AiChatSessionService sessions;
  private final CurrentUserProvider currentUser;

  public AiChatSessionController(
      AiChatSessionService sessions, CurrentUserProvider currentUser) {
    this.sessions = sessions;
    this.currentUser = currentUser;
  }

  @GetMapping
  public List<AiChatSessionResponse> list() {
    Long userId = currentUser.require().getId();
    return sessions.list(userId).stream().map(AiChatSessionResponse::from).toList();
  }

  @PostMapping
  public ResponseEntity<AiChatSessionResponse> create(
      @Valid @RequestBody CreateAiChatSessionRequest request) {
    Long userId = currentUser.require().getId();
    AiChatSessionResponse body =
        AiChatSessionResponse.from(
            sessions.create(userId, request.title(), request.sessionType(), request.scenarioId()));

    return ResponseEntity.status(HttpStatus.CREATED).body(body);
  }

  @GetMapping("/{id}")
  public AiChatSessionResponse get(@PathVariable Long id) {
    User actor = currentUser.require();
    return AiChatSessionResponse.from(sessions.get(id, actor));
  }

  @PutMapping("/{id}")
  public AiChatSessionResponse updateTitle(
      @PathVariable Long id, @Valid @RequestBody UpdateSessionTitleRequest request) {
    User actor = currentUser.require();
    return AiChatSessionResponse.from(sessions.updateTitle(id, actor, request.title()));
  }

  @DeleteMapping("/{id}")
  public ResponseEntity<Void> delete(@PathVariable Long id) {
    User actor = currentUser.require();
    sessions.delete(id, actor);

    return ResponseEntity.noContent().build();
  }
}
