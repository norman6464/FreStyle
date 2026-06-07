package com.normanblog.frestyle.config;

import com.normanblog.frestyle.dynamo.AiChatMessageReader;
import com.normanblog.frestyle.dynamo.DynamoAiChatMessageReader;
import com.normanblog.frestyle.dynamo.StubAiChatMessageReader;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;

/**
 * AI チャットメッセージ読み取り(DynamoDB)の組み立て。
 *
 * <p>aiChatTable が設定されていれば AWS SDK 実装、未設定ならローカル / テスト用 stub を返す。
 * これにより AWS 資格情報の無い環境でも起動・テストできる。
 */
@Configuration
public class DynamoConfig {

  private static final Logger log = LoggerFactory.getLogger(DynamoConfig.class);

  @Bean
  public AiChatMessageReader aiChatMessageReader(DynamoProperties props) {
    if (props.aiChatTable() == null || props.aiChatTable().isBlank()) {
      log.info("frestyle.dynamo.ai-chat-table 未設定のため stub message reader を使用します");
      return new StubAiChatMessageReader();
    }
    DynamoDbClient client =
        DynamoDbClient.builder().region(Region.of(props.regionOrDefault())).build();

    return new DynamoAiChatMessageReader(client, props.aiChatTable());
  }
}
