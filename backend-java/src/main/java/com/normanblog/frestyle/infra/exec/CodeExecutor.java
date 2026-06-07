package com.normanblog.frestyle.infra.exec;

import com.normanblog.frestyle.dto.CodeExecuteResponse;

/** ユーザーが書いたコードをサーバ側サンドボックスで実行する境界。 */
public interface CodeExecutor {

  /**
   * 指定言語のコードを stdin 付きで実行し、stdout / stderr / 終了コードを返す。
   *
   * <p>子プロセスには backend のシークレットを渡さず(環境クリーン化)、timeout / 出力上限を課す。
   *
   * @param language "java" 等
   * @param code 実行するソース
   * @param stdin 標準入力に流す内容(null/空なら無し)。演習採点でテストケースの入力に使う
   */
  CodeExecuteResponse execute(String language, String code, String stdin);

  /** stdin 無しで実行する(コード実行プレイグラウンド用)。 */
  default CodeExecuteResponse execute(String language, String code) {
    return execute(language, code, null);
  }
}
