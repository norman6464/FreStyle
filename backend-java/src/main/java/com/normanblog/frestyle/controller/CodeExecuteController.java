package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.CodeExecuteRequest;
import com.normanblog.frestyle.dto.CodeExecuteResponse;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.CodeExecutionService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * コード実行(サンドボックス)エンドポイント。認証ユーザーが書いたコードをサーバ側で実行し
 * stdout/stderr/exitCode を返す。
 */
@RestController
@RequestMapping("/api/v2/code")
public class CodeExecuteController {

  private final CodeExecutionService codeExecution;
  private final CurrentUserProvider currentUser;

  public CodeExecuteController(
      CodeExecutionService codeExecution, CurrentUserProvider currentUser) {
    this.codeExecution = codeExecution;
    this.currentUser = currentUser;
  }

  @PostMapping("/execute")
  public CodeExecuteResponse execute(@Valid @RequestBody CodeExecuteRequest request) {
    // 認証必須(未ログインは 401)。実行自体はユーザー単位ではないが、濫用防止に認証境界を置く。
    currentUser.require();

    return codeExecution.execute(request.language(), request.code());
  }
}
