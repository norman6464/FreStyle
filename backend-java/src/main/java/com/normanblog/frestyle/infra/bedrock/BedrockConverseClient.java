package com.normanblog.frestyle.infra.bedrock;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;
import software.amazon.awssdk.core.SdkBytes;
import software.amazon.awssdk.services.bedrockruntime.BedrockRuntimeAsyncClient;
import software.amazon.awssdk.services.bedrockruntime.model.ContentBlock;
import software.amazon.awssdk.services.bedrockruntime.model.ConversationRole;
import software.amazon.awssdk.services.bedrockruntime.model.ConverseStreamRequest;
import software.amazon.awssdk.services.bedrockruntime.model.ConverseStreamResponseHandler;
import software.amazon.awssdk.services.bedrockruntime.model.ImageBlock;
import software.amazon.awssdk.services.bedrockruntime.model.ImageFormat;
import software.amazon.awssdk.services.bedrockruntime.model.ImageSource;
import software.amazon.awssdk.services.bedrockruntime.model.Message;
import software.amazon.awssdk.services.bedrockruntime.model.SystemContentBlock;

/**
 * AWS Bedrock Converse(ストリーミング)を呼び出す本番実装。
 *
 * <p>非同期 SDK の EventStream を購読し、token を {@code onDelta} に転送しつつ全文を組み立てる。
 * 呼び出しスレッドは {@code future.join()} でストリーム完了までブロックする。
 */
public class BedrockConverseClient implements BedrockChatClient, AutoCloseable {

  // 応答が来ない Bedrock 呼び出しでスレッドが永久にブロックしないよう全体タイムアウトを設ける。
  private static final long STREAM_TIMEOUT_SECONDS = 120;

  private final BedrockRuntimeAsyncClient client;
  private final String modelId;

  public BedrockConverseClient(BedrockRuntimeAsyncClient client, String modelId) {
    this.client = client;
    this.modelId = modelId;
  }

  @Override
  public String converseStream(
      String systemPrompt, List<BedrockMessage> history, Consumer<String> onDelta) {
    List<Message> messages = buildMessages(history);

    ConverseStreamRequest.Builder request =
        ConverseStreamRequest.builder().modelId(modelId).messages(messages);
    if (systemPrompt != null && !systemPrompt.isBlank()) {
      request.system(SystemContentBlock.fromText(systemPrompt));
    }

    StringBuilder full = new StringBuilder();
    ConverseStreamResponseHandler.Visitor visitor =
        ConverseStreamResponseHandler.Visitor.builder()
            .onContentBlockDelta(
                event -> {
                  if (event.delta() != null && event.delta().text() != null) {
                    String text = event.delta().text();
                    full.append(text);
                    onDelta.accept(text);
                  }
                })
            .build();
    ConverseStreamResponseHandler handler =
        ConverseStreamResponseHandler.builder().subscriber(visitor).build();

    // join() は EventStream 完了までブロックし、失敗時は例外を伝播する(呼び出し側で扱う)。
    // orTimeout で上限を設け、ハングした呼び出しがスレッドを占有し続けないようにする。
    client
        .converseStream(request.build(), handler)
        .orTimeout(STREAM_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .join();

    return full.toString();
  }

  // domain の会話を Bedrock の Message 列に変換する。Bedrock は user/assistant 交互かつ非空を要求するため、
  // 空メッセージには placeholder を入れ、ブロックが無いものは skip する。
  private List<Message> buildMessages(List<BedrockMessage> history) {
    List<Message> messages = new ArrayList<>(history.size());
    for (BedrockMessage m : history) {
      ConversationRole role =
          "assistant".equals(m.role()) ? ConversationRole.ASSISTANT : ConversationRole.USER;

      List<ContentBlock> blocks = new ArrayList<>();
      if (m.images() != null) {
        for (BedrockMessage.BedrockImage img : m.images()) {
          if (img.bytes() != null && img.bytes().length > 0) {
            blocks.add(
                ContentBlock.fromImage(
                    ImageBlock.builder()
                        .format(ImageFormat.fromValue(img.format()))
                        .source(
                            ImageSource.fromBytes(SdkBytes.fromByteArray(img.bytes())))
                        .build()));
          }
        }
      }

      String text = m.text();
      if (text == null || text.isEmpty()) {
        // 画像のみのメッセージは text 無しでも可。画像も無ければ役割交互制約のため placeholder。
        if (blocks.isEmpty()) {
          text = role == ConversationRole.ASSISTANT ? "(応答が空でした)" : "(空のメッセージ)";
        } else {
          text = null;
        }
      }
      if (text != null && !text.isEmpty()) {
        blocks.add(ContentBlock.fromText(text));
      }
      if (blocks.isEmpty()) {
        continue;
      }
      messages.add(Message.builder().role(role).content(blocks).build());
    }

    return messages;
  }

  @Override
  public void close() {
    client.close();
  }
}
