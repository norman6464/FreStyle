package com.normanblog.frestyle.infra.exec;

import com.normanblog.frestyle.dto.CodeExecuteResponse;

/** ユーザーが書いたコードをサーバ側サンドボックスで実行する境界。 */
public interface CodeExecutor {

  /**
   * 指定言語のコードを実行し、stdout / stderr / 終了コードを返す。
   *
   * <p>子プロセスには backend のシークレットを渡さず(環境クリーン化)、timeout / 出力上限を課す。
   *
   * @param language "java" 等
   * @param code 実行するソース
   */
  CodeExecuteResponse execute(String language, String code);
}
