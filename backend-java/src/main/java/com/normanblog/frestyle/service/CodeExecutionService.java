package com.normanblog.frestyle.service;

import com.normanblog.frestyle.config.CodeExecProperties;
import com.normanblog.frestyle.dto.CodeExecuteResponse;
import com.normanblog.frestyle.infra.exec.CodeExecutor;
import com.normanblog.frestyle.infra.exec.ProcessCodeExecutor;
import java.nio.charset.StandardCharsets;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

/**
 * コード実行のビジネスロジック。実行可否(enabled)・コードサイズ・対応言語を検証してから executor に渡す。
 */
@Service
public class CodeExecutionService {

  private final CodeExecutor executor;
  private final CodeExecProperties props;

  public CodeExecutionService(CodeExecutor executor, CodeExecProperties props) {
    this.executor = executor;
    this.props = props;
  }

  /** コードを実行する。無効環境は 503、サイズ超過・未対応言語は 400。 */
  public CodeExecuteResponse execute(String language, String code) {
    if (!props.isEnabled()) {
      throw new ResponseStatusException(
          HttpStatus.SERVICE_UNAVAILABLE, "code execution disabled");
    }
    String lang = (language == null || language.isBlank()) ? "java" : language;
    if (!ProcessCodeExecutor.supportedLanguages().contains(lang)) {
      throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "unsupported_language");
    }
    if (code == null
        || code.getBytes(StandardCharsets.UTF_8).length > props.maxCodeBytesOrDefault()) {
      throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "code_too_large");
    }

    return executor.execute(lang, code);
  }
}
