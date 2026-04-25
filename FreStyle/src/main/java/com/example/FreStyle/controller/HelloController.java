package com.example.FreStyle.controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirements;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequestMapping("/api")
@Slf4j
@Tag(name = "Health", description = "ヘルスチェック / 疎通確認エンドポイント")
public class HelloController {

  @GetMapping("/hello")
  @Operation(summary = "疎通確認", description = "認証不要。固定文字列 'hello' を返す。")
  @SecurityRequirements
  public String hello() {
    log.debug("Hello health check endpoint called");
    return "hello";
  }

}
