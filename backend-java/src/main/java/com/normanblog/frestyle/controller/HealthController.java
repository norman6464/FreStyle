package com.normanblog.frestyle.controller;

import java.util.Map;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

/** ロードバランサ等が死活確認に使うヘルスチェックエンドポイント。 */
@RestController
public class HealthController {

  // DB や外部サービスに依存させない。デプロイ切替(Blue/Green)の判定で頻繁に叩かれるため、
  // 依存先の遅延・障害でヘルスが落ちると正常なデプロイまで巻き込んで失敗してしまうため。
  @GetMapping("/api/v2/health")
  public Map<String, String> health() {
    return Map.of("status", "UP");
  }
}
