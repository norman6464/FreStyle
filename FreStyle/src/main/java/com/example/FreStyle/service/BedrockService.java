package com.example.FreStyle.service;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import jakarta.annotation.PostConstruct;
import lombok.extern.slf4j.Slf4j;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.core.SdkBytes;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.bedrockruntime.BedrockRuntimeClient;
import software.amazon.awssdk.services.bedrockruntime.model.InvokeModelRequest;
import software.amazon.awssdk.services.bedrockruntime.model.InvokeModelResponse;

@Service
@Slf4j
public class BedrockService {

    @Value("${aws.access-key}")
    private String accessKey;

    @Value("${aws.secret-key}")
    private String secretKey;

    @Value("${aws.region}")
    private String region;

    // Bedrock モデルID（Claude 3 Haiku）
    private static final String MODEL_ID = "anthropic.claude-3-haiku-20240307-v1:0";

    private BedrockRuntimeClient bedrockClient;
    private final ObjectMapper objectMapper = new ObjectMapper();

    @PostConstruct
    public void init() {
        log.info("🚀 Bedrock Runtime Client を初期化中...");
        bedrockClient = BedrockRuntimeClient.builder()
                .region(Region.of(region))
                .credentialsProvider(
                        StaticCredentialsProvider.create(
                                AwsBasicCredentials.create(accessKey, secretKey)
                        )
                )
                .build();
        log.info("✅ Bedrock Runtime Client 初期化完了 - Region: {}", region);
    }

    /**
     * ユーザーのメッセージをBedrockに送信し、AIの応答を取得
     *
     * @param userMessage ユーザーからのメッセージ
     * @return AIからの応答テキスト
     */
    public String chat(String userMessage) {
        log.info("📤 Bedrock にメッセージ送信中...");
        log.debug("   - userMessage: {}", userMessage);

        try {
            // リクエストボディを構築
            ObjectNode requestBody = objectMapper.createObjectNode();
            requestBody.put("anthropic_version", "bedrock-2023-05-31");
            requestBody.put("max_tokens", 1024);
            requestBody.put("temperature", 0.7);

            // messages配列を構築
            ArrayNode messagesArray = objectMapper.createArrayNode();
            ObjectNode userMessageNode = objectMapper.createObjectNode();
            userMessageNode.put("role", "user");
            
            ArrayNode contentArray = objectMapper.createArrayNode();
            ObjectNode textContent = objectMapper.createObjectNode();
            textContent.put("type", "text");
            textContent.put("text", userMessage);
            contentArray.add(textContent);
            
            userMessageNode.set("content", contentArray);
            messagesArray.add(userMessageNode);
            
            requestBody.set("messages", messagesArray);

            String requestBodyJson = objectMapper.writeValueAsString(requestBody);
            log.debug("   - Request Body: {}", requestBodyJson);

            // Bedrockにリクエストを送信
            InvokeModelRequest request = InvokeModelRequest.builder()
                    .modelId(MODEL_ID)
                    .contentType("application/json")
                    .accept("application/json")
                    .body(SdkBytes.fromUtf8String(requestBodyJson))
                    .build();

            InvokeModelResponse response = bedrockClient.invokeModel(request);

            // レスポンスをパース
            String responseBody = response.body().asUtf8String();
            log.debug("   - Response Body: {}", responseBody);

            JsonNode responseJson = objectMapper.readTree(responseBody);
            String aiReply = responseJson.path("content").get(0).path("text").asText();

            log.info("✅ Bedrock からの応答を取得しました");
            log.debug("   - AI Reply: {}", aiReply);

            return aiReply;

        } catch (Exception e) {
            log.error("❌ Bedrock 呼び出しエラー: {}", e.getMessage());
            e.printStackTrace();
            throw new RuntimeException("AI応答の取得に失敗しました: " + e.getMessage(), e);
        }
    }

    /**
     * 会話履歴を含めたチャット（コンテキストを維持）
     *
     * @param conversationHistory 会話履歴のJSON文字列
     * @param userMessage         新しいユーザーメッセージ
     * @return AIからの応答テキスト
     */
    public String chatWithHistory(String conversationHistory, String userMessage) {
        log.info("📤 Bedrock に会話履歴付きメッセージ送信中...");

        try {
            ObjectNode requestBody = objectMapper.createObjectNode();
            requestBody.put("anthropic_version", "bedrock-2023-05-31");
            requestBody.put("max_tokens", 1024);
            requestBody.put("temperature", 0.7);

            // 会話履歴がある場合はパースして使用
            ArrayNode messagesArray;
            if (conversationHistory != null && !conversationHistory.isEmpty()) {
                messagesArray = (ArrayNode) objectMapper.readTree(conversationHistory);
            } else {
                messagesArray = objectMapper.createArrayNode();
            }

            // 新しいユーザーメッセージを追加
            ObjectNode userMessageNode = objectMapper.createObjectNode();
            userMessageNode.put("role", "user");
            
            ArrayNode contentArray = objectMapper.createArrayNode();
            ObjectNode textContent = objectMapper.createObjectNode();
            textContent.put("type", "text");
            textContent.put("text", userMessage);
            contentArray.add(textContent);
            
            userMessageNode.set("content", contentArray);
            messagesArray.add(userMessageNode);

            requestBody.set("messages", messagesArray);

            String requestBodyJson = objectMapper.writeValueAsString(requestBody);

            InvokeModelRequest request = InvokeModelRequest.builder()
                    .modelId(MODEL_ID)
                    .contentType("application/json")
                    .accept("application/json")
                    .body(SdkBytes.fromUtf8String(requestBodyJson))
                    .build();

            InvokeModelResponse response = bedrockClient.invokeModel(request);

            String responseBody = response.body().asUtf8String();
            JsonNode responseJson = objectMapper.readTree(responseBody);
            String aiReply = responseJson.path("content").get(0).path("text").asText();

            log.info("✅ Bedrock からの応答を取得しました（履歴付き）");

            return aiReply;

        } catch (Exception e) {
            log.error("❌ Bedrock 呼び出しエラー（履歴付き）: {}", e.getMessage());
            e.printStackTrace();
            throw new RuntimeException("AI応答の取得に失敗しました: " + e.getMessage(), e);
        }
    }
}
