package com.normanblog.frestyle.infra.bedrock;

import java.util.List;
import java.util.function.Consumer;

/** Bedrock Converse(ストリーミング)を呼び出す境界。 */
public interface BedrockChatClient {

  /**
   * 会話履歴を Bedrock に送り、トークンを逐次 {@code onDelta} へ渡しつつ、完成した全文を返す。
   *
   * <p>呼び出しスレッドをブロックし、ストリーム完了時に全文を返す(SSE 配信は onDelta 側で行う)。
   *
   * @param systemPrompt system プロンプト(空なら付けない)
   * @param history user/assistant 交互の会話。末尾が最新ユーザー発話
   * @param onDelta トークン追加分ごとに呼ばれる
   * @return 連結したアシスタント返答の全文
   */
  String converseStream(String systemPrompt, List<BedrockMessage> history, Consumer<String> onDelta);
}
