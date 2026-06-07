package com.normanblog.frestyle.config;

import com.normanblog.frestyle.infra.bedrock.BedrockChatClient;
import com.normanblog.frestyle.infra.bedrock.BedrockConverseClient;
import com.normanblog.frestyle.infra.bedrock.StubBedrockChatClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.bedrockruntime.BedrockRuntimeAsyncClient;

/**
 * AI チャットの Bedrock client の組み立て。
 *
 * <p>modelId が設定されていれば AWS SDK 実装、未設定ならローカル / テスト用 stub を返す。
 * これにより AWS 資格情報の無い環境でも起動・テストできる。
 */
@Configuration
public class BedrockConfig {

  private static final Logger log = LoggerFactory.getLogger(BedrockConfig.class);

  @Bean
  public BedrockChatClient bedrockChatClient(BedrockProperties props) {
    if (props.modelId() == null || props.modelId().isBlank()) {
      log.info("frestyle.bedrock.model-id 未設定のため stub Bedrock client を使用します");
      return new StubBedrockChatClient();
    }
    BedrockRuntimeAsyncClient client =
        BedrockRuntimeAsyncClient.builder().region(Region.of(props.regionOrDefault())).build();

    return new BedrockConverseClient(client, props.modelId());
  }
}
