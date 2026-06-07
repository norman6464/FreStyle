package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.AiChatMessageResponse;
import com.normanblog.frestyle.dto.AiChatStreamRequest;
import com.normanblog.frestyle.dto.AiChatStreamRequest.AiChatStreamAttachment;
import com.normanblog.frestyle.entity.AiChatSession;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.AiChatAttachmentRules;
import com.normanblog.frestyle.service.AiChatSessionService;
import com.normanblog.frestyle.service.AiMessageStreamListener;
import com.normanblog.frestyle.service.SendAiMessageUseCase;
import jakarta.annotation.PreDestroy;
import jakarta.validation.Valid;
import java.io.IOException;
import java.time.Duration;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicBoolean;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

/**
 * AI チャットの SSE ストリーミング。Bedrock の token を session / token / done / error の 4 イベントで配信する。
 *
 * <p>事前検証(本文/添付/所有者)は同期的に行い、不正は通常の JSON エラー(400/403/404)で返す。
 * 検証通過後に SSE を開始し、{@link SendAiMessageUseCase} を別スレッドで走らせて listener 経由で書き込む。
 */
@RestController
@RequestMapping("/api/v2/ai-chat")
public class AiChatStreamController {

  private static final Logger log = LoggerFactory.getLogger(AiChatStreamController.class);
  private static final long STREAM_TIMEOUT_MS = Duration.ofMinutes(5).toMillis();

  private final SendAiMessageUseCase sendMessage;
  private final AiChatSessionService sessions;
  private final CurrentUserProvider currentUser;
  // SSE は接続を保持し続けるため、リクエストスレッドとは別の専用プールで usecase を走らせる。
  private final ExecutorService executor = Executors.newFixedThreadPool(16);

  public AiChatStreamController(
      SendAiMessageUseCase sendMessage,
      AiChatSessionService sessions,
      CurrentUserProvider currentUser) {
    this.sendMessage = sendMessage;
    this.sessions = sessions;
    this.currentUser = currentUser;
  }

  @PostMapping(value = "/stream", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
  public SseEmitter stream(@Valid @RequestBody AiChatStreamRequest request) {
    User actor = currentUser.require();
    validate(actor, request);

    SseEmitter emitter = new SseEmitter(STREAM_TIMEOUT_MS);
    AtomicBoolean closed = new AtomicBoolean(false);
    AiMessageStreamListener listener = newListener(emitter, closed);

    executor.execute(
        () -> {
          try {
            sendMessage.execute(actor, request, listener);
          } catch (RuntimeException e) {
            log.warn("ai-chat stream task failed", e);
            sendEvent(emitter, closed, "error", Map.of("message", "メッセージの送信に失敗しました"));
            complete(emitter, closed);
          }
        });

    return emitter;
  }

  // 本文/添付/所有者の事前検証。違反は通常の HTTP エラー(SSE 開始前なので JSON で返る)。
  private void validate(User actor, AiChatStreamRequest request) {
    boolean noContent = request.content() == null || request.content().isBlank();
    boolean noAttachments = request.attachments() == null || request.attachments().isEmpty();
    if (noContent && noAttachments) {
      throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "content_required");
    }
    if (request.attachments() != null) {
      String expectedPrefix = "ai-chat/" + actor.getId() + "/";
      for (AiChatStreamAttachment a : request.attachments()) {
        if (a.key() == null || a.contentType() == null) {
          throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "attachment_invalid");
        }
        if (!a.key().startsWith(expectedPrefix)) {
          throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "attachment_key_not_allowed");
        }
        AiChatAttachmentRules.Rule rule = AiChatAttachmentRules.of(a.contentType());
        if (rule == null) {
          throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "attachment_unsupported_type");
        }
        if (a.sizeBytes() <= 0 || a.sizeBytes() > rule.maxBytes()) {
          throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "attachment_too_large");
        }
      }
    }
    // 既存セッション指定時は所有者検証(他人 403 / 不在 404)を SSE 開始前に済ませる。
    if (request.sessionId() != null && request.sessionId() != 0L) {
      sessions.get(request.sessionId(), actor);
    }
  }

  private AiMessageStreamListener newListener(SseEmitter emitter, AtomicBoolean closed) {
    return new AiMessageStreamListener() {
      @Override
      public void onSession(AiChatSession session) {
        Map<String, Object> payload = new HashMap<>();
        payload.put("id", session.getId());
        payload.put("title", session.getTitle());
        payload.put("sessionType", session.getSessionType());
        payload.put("scenarioId", session.getScenarioId());
        payload.put("createdAt", session.getCreatedAt() == null ? null : session.getCreatedAt().toString());
        sendEvent(emitter, closed, "session", payload);
      }

      @Override
      public void onToken(String delta) {
        sendEvent(emitter, closed, "token", Map.of("delta", delta));
      }

      @Override
      public void onDone(AiChatMessageResponse finalMessage) {
        Map<String, Object> payload = new HashMap<>();
        payload.put("sessionId", finalMessage.sessionId());
        payload.put("id", finalMessage.messageId());
        payload.put("role", finalMessage.role());
        payload.put("content", finalMessage.content());
        payload.put("createdAt", finalMessage.createdAt());
        sendEvent(emitter, closed, "done", payload);
        complete(emitter, closed);
      }

      @Override
      public void onError(String message) {
        sendEvent(emitter, closed, "error", Map.of("message", message));
        complete(emitter, closed);
      }
    };
  }

  private void sendEvent(SseEmitter emitter, AtomicBoolean closed, String event, Object data) {
    if (closed.get()) {
      return;
    }
    try {
      emitter.send(SseEmitter.event().name(event).data(data, MediaType.APPLICATION_JSON));
    } catch (IOException | IllegalStateException e) {
      // クライアント切断などで書き込めない。以降の送信を止める。
      closed.set(true);
    }
  }

  private void complete(SseEmitter emitter, AtomicBoolean closed) {
    if (closed.compareAndSet(false, true)) {
      try {
        emitter.complete();
      } catch (RuntimeException ignored) {
        // 既に閉じている場合は無視。
      }
    }
  }

  @PreDestroy
  void shutdown() {
    executor.shutdown();
  }
}
