package com.example.FreStyle.service;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import jakarta.annotation.PostConstruct;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;


// AIとのやり取りをするためのSDKの設定]
@Service
public class ChatService {
    @Value("${aws.access-key}")
    private String accessKey;
    
    @Value("${aws.secret-key}")
    private String secretKey;
    
    @Value("${aws.region}")
    private String region;
    
    @Value("${aws.dynamodb.table-name.chat}")
    private String tableName;
    
    private DynamoDbClient dynamoDbClient;
    
    // PostConstructではBeanが生成されて依存製注入（DI）が完了した直後に実行されるメソッドを定義する
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
    
    // 
    
}
