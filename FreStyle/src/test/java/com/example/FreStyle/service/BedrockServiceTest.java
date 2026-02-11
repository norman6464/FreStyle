package com.example.FreStyle.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Spy;
import org.mockito.junit.jupiter.MockitoExtension;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import software.amazon.awssdk.core.SdkBytes;
import software.amazon.awssdk.services.bedrockruntime.BedrockRuntimeClient;
import software.amazon.awssdk.services.bedrockruntime.model.InvokeModelRequest;
import software.amazon.awssdk.services.bedrockruntime.model.InvokeModelResponse;

@ExtendWith(MockitoExtension.class)
class BedrockServiceTest {

    @Mock
    private BedrockRuntimeClient bedrockClient;

    @Spy
    private SystemPromptBuilder systemPromptBuilder = new SystemPromptBuilder();

    @InjectMocks
    private BedrockService bedrockService;

    private final ObjectMapper objectMapper = new ObjectMapper();

    private InvokeModelResponse createMockResponse(String aiReply) throws Exception {
        String responseJson = String.format(
                "{\"content\":[{\"type\":\"text\",\"text\":\"%s\"}]}", aiReply);
        return InvokeModelResponse.builder()
                .body(SdkBytes.fromUtf8String(responseJson))
                .build();
    }

    @BeforeEach
    void setUp() throws Exception {
        // BedrockServiceのbedrockClientフィールドをモックに差し替え
        java.lang.reflect.Field clientField = BedrockService.class.getDeclaredField("bedrockClient");
        clientField.setAccessible(true);
        clientField.set(bedrockService, bedrockClient);
    }

    @Test
    @DisplayName("chat()がコールセンター式コーチのシステムプロンプト付きでリクエストを構築する")
    void chatShouldIncludeCoachSystemPrompt() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("テスト応答"));

        bedrockService.chat("こんにちは");

        ArgumentCaptor<InvokeModelRequest> captor = ArgumentCaptor.forClass(InvokeModelRequest.class);
        verify(bedrockClient).invokeModel(captor.capture());

        String requestBody = captor.getValue().body().asUtf8String();
        JsonNode json = objectMapper.readTree(requestBody);

        // システムプロンプトが含まれていることを確認
        assertThat(json.has("system")).isTrue();
        String systemPrompt = json.get("system").asText();
        assertThat(systemPrompt).contains("コールセンター");
        assertThat(systemPrompt).contains("コミュニケーションコーチ");
    }

    @Test
    @DisplayName("chatWithUserProfile()がQA評価軸を含むフィードバックプロンプトを使用する")
    void chatWithUserProfileShouldIncludeFeedbackPrompt() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("フィードバック応答"));

        bedrockService.chatWithUserProfile(
                "テストメッセージ", "太郎", "自己紹介", "丁寧", "真面目", "目標", "懸念", "優しく");

        ArgumentCaptor<InvokeModelRequest> captor = ArgumentCaptor.forClass(InvokeModelRequest.class);
        verify(bedrockClient).invokeModel(captor.capture());

        String requestBody = captor.getValue().body().asUtf8String();
        JsonNode json = objectMapper.readTree(requestBody);

        String systemPrompt = json.get("system").asText();
        assertThat(systemPrompt).contains("共感力");
        assertThat(systemPrompt).contains("クッション言葉");
        assertThat(systemPrompt).contains("結論ファースト");
        assertThat(systemPrompt).contains("ポジティブ変換");
        assertThat(systemPrompt).contains("傾聴姿勢");
        assertThat(systemPrompt).contains("コールセンター");
        assertThat(systemPrompt).contains("太郎");
    }

    @Test
    @DisplayName("chat()がAIの応答テキストを正しく返す")
    void chatShouldReturnAiReply() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("AI応答テスト"));

        String result = bedrockService.chat("質問です");

        assertThat(result).isEqualTo("AI応答テスト");
    }
}
