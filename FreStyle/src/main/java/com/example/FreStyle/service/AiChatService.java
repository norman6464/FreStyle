package com.example.FreStyle.service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.AiChatMessageDto;

import jakarta.annotation.PostConstruct;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.AttributeValue;
import software.amazon.awssdk.services.dynamodb.model.QueryRequest;
import software.amazon.awssdk.services.dynamodb.model.QueryResponse;


// AIとのやり取りをするためのSDKの設定
@Service
public class AiChatService {
    @Value("${aws.access-key}")
    private String accessKey;
    
    @Value("${aws.secret-key}")
    private String secretKey;
    
    @Value("${aws.region}")
    private String region;
    
    @Value("${aws.dynamodb.table-name.ai-chat}")
    private String tableName;
    
    private DynamoDbClient dynamoDbClient;
    
    // PostConstructではBeanが生成されて依存性注入（DI）が完了した直後に実行されるメソッドを定義
    @PostConstruct
    
    public void init() {
      dynamoDbClient = DynamoDbClient.builder()
              .region(Region.of(region))
              .credentialsProvider(
                  StaticCredentialsProvider.create(
                    AwsBasicCredentials.create(accessKey, secretKey) 
                  )
              )
              .build();
    }
    
    
    // sender_idとcognito_subが紐づいて保存されている。
    public List<AiChatMessageDto> getChatHistory(String senderId) {
      
      // sender_idに基づいてScanリクエストをし、条件一致した項目を全部取得をしている
      QueryRequest queryRequest = QueryRequest.builder()
          .tableName(tableName)
          .keyConditionExpression("sender_id = :sender_id")
          .expressionAttributeValues(Map.of(
            ":sender_id", AttributeValue.builder().s(senderId).build()
          ))
          .scanIndexForward(true) // 昇順
          .build();
      QueryResponse response = dynamoDbClient.query(queryRequest);
      
      List<AiChatMessageDto> result = new ArrayList<>();
      for (Map<String, AttributeValue> item : response.items()) {
        AiChatMessageDto dto = new AiChatMessageDto(
          item.get("content").s(),
          item.get("is_user").bool(),
          Long.parseLong(item.get("timestamp").n())
        );
        result.add(dto);
      }
      
      return result;
      
    }
    
    
    
}
