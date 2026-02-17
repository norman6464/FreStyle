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
    private final SystemPromptBuilder systemPromptBuilder;

    public BedrockService(SystemPromptBuilder systemPromptBuilder) {
        this.systemPromptBuilder = systemPromptBuilder;
    }

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

        String coachPrompt = systemPromptBuilder.buildCoachPrompt();
        return invokeSingleMessage(coachPrompt, userMessage, 1024, 0.7,
                "Bedrock", "AI応答の取得に失敗しました");
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
            ArrayNode messagesArray;
            if (conversationHistory != null && !conversationHistory.isEmpty()) {
                messagesArray = (ArrayNode) objectMapper.readTree(conversationHistory);
            } else {
                messagesArray = objectMapper.createArrayNode();
            }
            messagesArray.add(buildUserMessageNode(userMessage));

            ObjectNode requestBody = buildRequestBody(null, messagesArray, 1024, 0.7);
            return invokeAndParseResponse(requestBody, "Bedrock（履歴付き）");

        } catch (Exception e) {
            log.error("Bedrock 呼び出しエラー（履歴付き）: {}", e.getMessage(), e);
            throw new RuntimeException("AI応答の取得に失敗しました: " + e.getMessage(), e);
        }
    }

    /**
     * 練習モード用のAI応答取得
     * シナリオに基づいたロールプレイ相手役として応答する
     *
     * @param userMessage ユーザーからのメッセージ
     * @param practicePrompt 練習用システムプロンプト
     * @return AIからの応答テキスト
     */
    public String chatInPracticeMode(String userMessage, String practicePrompt) {
        log.info("📤 Bedrock に練習モードメッセージ送信中...");

        return invokeSingleMessage(practicePrompt, userMessage, 1024, 0.8,
                "Bedrock（練習モード）", "練習モードのAI応答取得に失敗しました");
    }

    /**
     * メッセージの言い換え提案を取得（3パターン: フォーマル版/ソフト版/簡潔版）
     *
     * @param originalMessage 元のメッセージ
     * @param scene シーン識別子（null可）
     * @return AIからの応答テキスト（JSON形式）
     */
    public String rephrase(String originalMessage, String scene) {
        log.info("📤 Bedrock に言い換えリクエスト送信中... scene={}", scene);

        String systemPrompt = systemPromptBuilder.buildRephrasePrompt(scene);
        return invokeSingleMessage(systemPrompt, originalMessage, 1024, 0.7,
                "Bedrock（言い換え）", "言い換え提案の取得に失敗しました");
    }

    private String invokeSingleMessage(String systemPrompt, String userMessage,
                                       int maxTokens, double temperature,
                                       String logContext, String errorMessage) {
        try {
            ArrayNode messagesArray = objectMapper.createArrayNode();
            messagesArray.add(buildUserMessageNode(userMessage));

            ObjectNode requestBody = buildRequestBody(systemPrompt, messagesArray, maxTokens, temperature);
            return invokeAndParseResponse(requestBody, logContext);

        } catch (Exception e) {
            log.error("{} エラー: {}", logContext, e.getMessage(), e);
            throw new RuntimeException(errorMessage + ": " + e.getMessage(), e);
        }
    }

    private ObjectNode buildUserMessageNode(String text) {
        ObjectNode userMessageNode = objectMapper.createObjectNode();
        userMessageNode.put("role", "user");

        ArrayNode contentArray = objectMapper.createArrayNode();
        ObjectNode textContent = objectMapper.createObjectNode();
        textContent.put("type", "text");
        textContent.put("text", text);
        contentArray.add(textContent);

        userMessageNode.set("content", contentArray);
        return userMessageNode;
    }

    private ObjectNode buildRequestBody(String systemPrompt, ArrayNode messages, int maxTokens, double temperature) {
        ObjectNode requestBody = objectMapper.createObjectNode();
        requestBody.put("anthropic_version", "bedrock-2023-05-31");
        requestBody.put("max_tokens", maxTokens);
        requestBody.put("temperature", temperature);
        if (systemPrompt != null) {
            requestBody.put("system", systemPrompt);
        }
        requestBody.set("messages", messages);
        return requestBody;
    }

    private String invokeAndParseResponse(ObjectNode requestBody, String logContext) throws Exception {
        String requestBodyJson = objectMapper.writeValueAsString(requestBody);
        log.debug("   - Request Body: {}", requestBodyJson);

        InvokeModelRequest request = InvokeModelRequest.builder()
                .modelId(MODEL_ID)
                .contentType("application/json")
                .accept("application/json")
                .body(SdkBytes.fromUtf8String(requestBodyJson))
                .build();

        InvokeModelResponse response = bedrockClient.invokeModel(request);

        String responseBody = response.body().asUtf8String();
        log.debug("   - Response Body: {}", responseBody);

        JsonNode responseJson = objectMapper.readTree(responseBody);
        String aiReply = responseJson.path("content").get(0).path("text").asText();

        log.info("✅ {} からの応答を取得しました", logContext);
        return aiReply;
    }
}
