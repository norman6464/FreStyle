package com.normanblog.frestyle.controller;

import java.util.Map;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

/** API 全体で共通の例外→レスポンス変換。 */
@RestControllerAdvice
public class GlobalExceptionHandler {

  /**
   * Bean Validation(@Valid)違反を 400 JSON で返す。
   *
   * <p>これが無いと Spring Boot が /error へフォワードし、resource server の認証境界に当たって
   * 400 が 401 に化ける。明示的にハンドルして本来の 400 を返す。
   */
  @ExceptionHandler(MethodArgumentNotValidException.class)
  public ResponseEntity<Map<String, String>> handleValidation(MethodArgumentNotValidException e) {
    String detail =
        e.getBindingResult().getFieldErrors().stream()
            .findFirst()
            .map(fe -> fe.getField() + ": " + fe.getDefaultMessage())
            .orElse("invalid request");
    return ResponseEntity.status(HttpStatus.BAD_REQUEST)
        .body(Map.of("error", "validation_failed", "message", detail));
  }
}
