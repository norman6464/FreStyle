package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.AiChatAttachmentDto;
import com.normanblog.frestyle.dto.AiChatMessageResponse;
import com.normanblog.frestyle.dto.AiChatStreamRequest;
import com.normanblog.frestyle.dto.AiChatStreamRequest.AiChatStreamAttachment;
import com.normanblog.frestyle.entity.AiChatSession;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.infra.bedrock.BedrockChatClient;
import com.normanblog.frestyle.infra.bedrock.BedrockMessage;
import com.normanblog.frestyle.infra.bedrock.BedrockMessage.BedrockImage;
import com.normanblog.frestyle.infra.dynamo.AiChatMessageReader;
import com.normanblog.frestyle.infra.dynamo.AiChatMessageWriter;
import com.normanblog.frestyle.infra.s3.AttachmentDownloader;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * AI チャット送信(SSE streaming)のオーケストレーション。
 *
 * <p>「セッション作成 → ユーザーメッセージ保存 → 履歴ロード → 添付取得 → Bedrock streaming →
 * 返答保存」を 1 つのユースケースとして束ねる。複数のポート(session/message/S3/Bedrock)を跨ぐ複雑な
 * 操作のため、service メソッドではなく専用 UseCase クラスに切り出している。
 */
@Component
public class SendAiMessageUseCase {

  private static final Logger log = LoggerFactory.getLogger(SendAiMessageUseCase.class);
  private static final String SYSTEM_PROMPT =
      "あなたは新卒 IT エンジニア向け学習プラットフォーム『FreStyle』に組み込まれた汎用 AI アシスタントです。"
          + "ユーザーからの質問・要約依頼・コードレビュー・概念説明など、幅広いトピックに答えます。"
          + "日本語で簡潔・丁寧に応答してください。"
          + "回答は **Markdown 形式** で返してください（見出し / 箇条書き / 表 / コードブロックなど、内容に応じて適切に使い分けます）。"
          + "コードを示すときは ```言語名 で囲み、シンタックスハイライトが効くようにしてください。";

  private final AiChatSessionService sessions;
  private final AiChatMessageReader reader;
  private final AiChatMessageWriter writer;
  private final AttachmentDownloader downloader;
  private final BedrockChatClient bedrock;

  public SendAiMessageUseCase(
      AiChatSessionService sessions,
      AiChatMessageReader reader,
      AiChatMessageWriter writer,
      AttachmentDownloader downloader,
      BedrockChatClient bedrock) {
    this.sessions = sessions;
    this.reader = reader;
    this.writer = writer;
    this.downloader = downloader;
    this.bedrock = bedrock;
  }

  /**
   * メッセージ送信〜返答保存を実行し、進行を listener に通知する。呼び出しスレッドをブロックする
   * (controller が別スレッドで実行し、listener が SSE へ書き込む)。
   *
   * <p>添付の事前検証(キー prefix / MIME / サイズ)と所有者検証は controller 側で済ませている前提。
   */
  public void execute(User actor, AiChatStreamRequest request, AiMessageStreamListener listener) {
    try {
      Long sessionId = request.sessionId();
      if (sessionId == null || sessionId == 0L) {
        AiChatSession session =
            sessions.create(actor.getId(), title(request), request.sessionType(), request.scenarioId());
        sessionId = session.getId();
        listener.onSession(session);
      }

      List<AiChatAttachmentDto> attachmentMeta = toAttachmentMeta(request.attachments());
      AiChatMessageResponse userMessage =
          new AiChatMessageResponse(
              sessionId,
              UUID.randomUUID().toString(),
              "user",
              request.content() == null ? "" : request.content(),
              attachmentMeta.isEmpty() ? null : attachmentMeta,
              Instant.now().toString());
      writer.save(userMessage);

      List<AiChatMessageResponse> history = reader.listBySession(sessionId);
      List<BedrockImage> blobs = downloadImages(request.attachments());
      List<BedrockMessage> bedrockHistory = toBedrockMessages(history, blobs);

      String full = bedrock.converseStream(SYSTEM_PROMPT, bedrockHistory, listener::onToken);
      // 空応答を保存すると履歴が「連続 user → 空応答」で壊れるため、保存せずエラー通知する。
      if (full == null || full.isBlank()) {
        listener.onError("bedrock returned empty response");
        return;
      }

      AiChatMessageResponse aiMessage =
          new AiChatMessageResponse(
              sessionId,
              UUID.randomUUID().toString(),
              "assistant",
              full,
              null,
              Instant.now().toString());
      writer.save(aiMessage);
      listener.onDone(aiMessage);
    } catch (RuntimeException e) {
      log.warn("ai-chat send failed", e);
      listener.onError("メッセージの送信に失敗しました");
    }
  }

  // 新規セッションのタイトル。本文先頭 30 文字、空なら添付有無でフォールバック。
  private String title(AiChatStreamRequest request) {
    String content = request.content();
    if (content != null && !content.isBlank()) {
      String trimmed = content.strip();
      return trimmed.codePointCount(0, trimmed.length()) > 30
          ? trimmed.substring(0, trimmed.offsetByCodePoints(0, 30)) + "…"
          : trimmed;
    }
    return (request.attachments() != null && !request.attachments().isEmpty())
        ? "添付ファイルを送信"
        : "新しいチャット";
  }

  // リクエスト添付を保存用メタ(format/kind はルールから導出)に変換する。
  private List<AiChatAttachmentDto> toAttachmentMeta(List<AiChatStreamAttachment> attachments) {
    List<AiChatAttachmentDto> out = new ArrayList<>();
    if (attachments == null) {
      return out;
    }
    for (AiChatStreamAttachment a : attachments) {
      AiChatAttachmentRules.Rule rule = AiChatAttachmentRules.of(a.contentType());
      String format = rule == null ? "" : rule.format();
      String kind = rule == null ? "" : rule.kind();
      out.add(
          new AiChatAttachmentDto(
              a.key(), a.filename(), a.contentType(), format, kind, a.sizeBytes()));
    }

    return out;
  }

  // 最新ユーザー発話の添付実体だけを S3 から取得する(過去履歴の画像は再送しない)。
  private List<BedrockImage> downloadImages(List<AiChatStreamAttachment> attachments) {
    List<BedrockImage> out = new ArrayList<>();
    if (attachments == null) {
      return out;
    }
    for (AiChatStreamAttachment a : attachments) {
      AiChatAttachmentRules.Rule rule = AiChatAttachmentRules.of(a.contentType());
      if (rule == null || !"image".equals(rule.kind())) {
        continue;
      }
      byte[] bytes = downloader.download(a.key());
      if (bytes != null && bytes.length > 0) {
        out.add(new BedrockImage(rule.format(), bytes));
      }
    }

    return out;
  }

  // 履歴を Bedrock 用に変換し、末尾の user メッセージにだけ取得した画像を付ける。
  private List<BedrockMessage> toBedrockMessages(
      List<AiChatMessageResponse> history, List<BedrockImage> latestImages) {
    int lastUser = -1;
    for (int i = history.size() - 1; i >= 0; i--) {
      if ("user".equals(history.get(i).role())) {
        lastUser = i;
        break;
      }
    }

    List<BedrockMessage> out = new ArrayList<>(history.size());
    for (int i = 0; i < history.size(); i++) {
      AiChatMessageResponse m = history.get(i);
      List<BedrockImage> images = (i == lastUser) ? latestImages : List.of();
      out.add(new BedrockMessage(m.role(), m.content(), images));
    }

    return out;
  }
}
