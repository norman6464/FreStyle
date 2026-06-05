package com.normanblog.frestyle.handler;

import java.util.Map;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

/** HealthController は ALB / Blue/Green が叩く死活エンドポイント。Go 版 /api/v2/health と互換。 */
@RestController
public class HealthController {

  @GetMapping("/api/v2/health")
  public Map<String, String> health() {
    return Map.of("status", "UP");
  }
}
