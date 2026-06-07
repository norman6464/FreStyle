package com.normanblog.frestyle.infra.bedrock;

import java.util.List;
import java.util.function.Consumer;

/**
 * Bedrock を呼ばない stub。model 未設定のローカル / テスト環境で使う。
 *
 * <p>固定の応答を数チャンクに分けて onDelta へ流し、全文を返す(ストリーミングの体裁を保つ)。
 */
public class StubBedrockChatClient implements BedrockChatClient {

  private static final List<String> CHUNKS =
      List.of("(stub) ", "AI ", "応答は", "現在", "利用できません。");

  @Override
  public String converseStream(
      String systemPrompt, List<BedrockMessage> history, Consumer<String> onDelta) {
    StringBuilder full = new StringBuilder();
    for (String chunk : CHUNKS) {
      full.append(chunk);
      onDelta.accept(chunk);
    }

    return full.toString();
  }
}
