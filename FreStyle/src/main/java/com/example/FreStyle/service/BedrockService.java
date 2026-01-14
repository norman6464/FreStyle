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

    /**
     * UserProfile情報を含めたチャットフィードバック用AI応答取得
     * システムプロンプトにユーザーの目標、懸念事項、フィードバックスタイルを含める
     *
     * @param userMessage ユーザーからのメッセージ（チャット履歴を含む）
     * @param displayName ユーザーの表示名
     * @param selfIntroduction 自己紹介
     * @param communicationStyle コミュニケーションスタイル
     * @param personalityTraits 性格特性（カンマ区切り）
     * @param goals ユーザーの目標
     * @param concerns 懸念事項
     * @param preferredFeedbackStyle 希望するフィードバックスタイル
     * @return AIからの応答テキスト
     */
    public String chatWithUserProfile(
            String userMessage,
            String displayName,
            String selfIntroduction,
            String communicationStyle,
            String personalityTraits,
            String goals,
            String concerns,
            String preferredFeedbackStyle) {
        
        log.info("📤 Bedrock にUserProfile付きメッセージ送信中...");

        try {
            // システムプロンプトを構築
            StringBuilder systemPromptBuilder = new StringBuilder();
            systemPromptBuilder.append("あなたはコミュニケーションのフィードバックを行う専門家です。\n");
            systemPromptBuilder.append("以下のユーザープロフィール情報を参考にして、チャットのフィードバックを行ってください。\n\n");
            systemPromptBuilder.append("【ユーザープロフィール】\n");
            
            if (displayName != null && !displayName.isEmpty()) {
                systemPromptBuilder.append("- 名前: ").append(displayName).append("\n");
            }
            if (selfIntroduction != null && !selfIntroduction.isEmpty()) {
                systemPromptBuilder.append("- 自己紹介: ").append(selfIntroduction).append("\n");
            }
            if (communicationStyle != null && !communicationStyle.isEmpty()) {
                systemPromptBuilder.append("- コミュニケーションスタイル: ").append(communicationStyle).append("\n");
            }
            if (personalityTraits != null && !personalityTraits.isEmpty()) {
                systemPromptBuilder.append("- 性格特性: ").append(personalityTraits).append("\n");
            }
            if (goals != null && !goals.isEmpty()) {
                systemPromptBuilder.append("- 目標: ").append(goals).append("\n");
            }
            if (concerns != null && !concerns.isEmpty()) {
                systemPromptBuilder.append("- 懸念事項: ").append(concerns).append("\n");
            }
            if (preferredFeedbackStyle != null && !preferredFeedbackStyle.isEmpty()) {
                systemPromptBuilder.append("- 希望するフィードバックスタイル: ").append(preferredFeedbackStyle).append("\n");
            }
            
            systemPromptBuilder.append("\n上記の情報を踏まえて、ユーザーのコミュニケーションについて建設的で具体的なフィードバックを提供してください。");
            systemPromptBuilder.append("ユーザーの目標達成に役立つアドバイスを、希望するフィードバックスタイルに合わせて行ってください。");

            String systemPrompt = systemPromptBuilder.toString();
            log.debug("   - System Prompt: {}", systemPrompt);

            // リクエストボディを構築
            ObjectNode requestBody = objectMapper.createObjectNode();
            requestBody.put("anthropic_version", "bedrock-2023-05-31");
            requestBody.put("max_tokens", 2048); // フィードバック用に少し長めに
            requestBody.put("temperature", 0.7);
            requestBody.put("system", systemPrompt);

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

            log.info("✅ Bedrock からの応答を取得しました（UserProfile付き）");

            return aiReply;

        } catch (Exception e) {
            log.error("❌ Bedrock 呼び出しエラー（UserProfile付き）: {}", e.getMessage());
            e.printStackTrace();
            throw new RuntimeException("AI応答の取得に失敗しました: " + e.getMessage(), e);
        }
    }
}
