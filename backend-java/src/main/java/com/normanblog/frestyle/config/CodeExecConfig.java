package com.normanblog.frestyle.config;

import com.normanblog.frestyle.infra.exec.CodeExecutor;
import com.normanblog.frestyle.infra.exec.ProcessCodeExecutor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/** コード実行 executor の組み立て(timeout / 出力上限は設定から)。 */
@Configuration
public class CodeExecConfig {

  @Bean
  public CodeExecutor codeExecutor(CodeExecProperties props) {
    return new ProcessCodeExecutor(props.timeoutSecondsOrDefault(), props.maxOutputBytesOrDefault());
  }
}
