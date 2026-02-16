package com.example.FreStyle.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
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
    @DisplayName("chat()がビジネスコミュニケーションコーチのシステムプロンプト付きでリクエストを構築する")
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
        assertThat(systemPrompt).contains("ビジネスコミュニケーション");
        assertThat(systemPrompt).contains("コーチ");
    }

    @Test
    @DisplayName("chatWithUserProfile()がビジネス評価軸を含むフィードバックプロンプトを使用する")
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
        assertThat(systemPrompt).contains("論理的構成力");
        assertThat(systemPrompt).contains("配慮表現");
        assertThat(systemPrompt).contains("要約力");
        assertThat(systemPrompt).contains("提案力");
        assertThat(systemPrompt).contains("質問・傾聴力");
        assertThat(systemPrompt).contains("ビジネス");
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

    @Test
    @DisplayName("chat()でBedrock呼び出しエラー時にRuntimeExceptionをスローする")
    void chatShouldThrowOnBedrockError() {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenThrow(new RuntimeException("Bedrock error"));

        assertThatThrownBy(() -> bedrockService.chat("質問です"))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("AI応答の取得に失敗しました");
    }

    // ============================
    // chatWithHistory
    // ============================
    @Test
    @DisplayName("chatWithHistory()が会話履歴なしでも正常に動作する")
    void chatWithHistoryWithEmptyHistory() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("履歴なし応答"));

        String result = bedrockService.chatWithHistory(null, "新しいメッセージ");

        assertThat(result).isEqualTo("履歴なし応答");
    }

    @Test
    @DisplayName("chatWithHistory()が会話履歴を含めてリクエストを構築する")
    void chatWithHistoryShouldIncludeHistory() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("履歴付き応答"));

        String history = "[{\"role\":\"user\",\"content\":[{\"type\":\"text\",\"text\":\"前の質問\"}]},"
                + "{\"role\":\"assistant\",\"content\":[{\"type\":\"text\",\"text\":\"前の回答\"}]}]";

        bedrockService.chatWithHistory(history, "続きの質問");

        ArgumentCaptor<InvokeModelRequest> captor = ArgumentCaptor.forClass(InvokeModelRequest.class);
        verify(bedrockClient).invokeModel(captor.capture());

        String requestBody = captor.getValue().body().asUtf8String();
        JsonNode json = objectMapper.readTree(requestBody);
        JsonNode messages = json.get("messages");

        // 履歴2件 + 新規メッセージ1件 = 3件
        assertThat(messages.size()).isEqualTo(3);
    }

    @Test
    @DisplayName("chatWithHistory()でBedrock呼び出しエラー時にRuntimeExceptionをスローする")
    void chatWithHistoryShouldThrowOnError() {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenThrow(new RuntimeException("Bedrock error"));

        assertThatThrownBy(() -> bedrockService.chatWithHistory(null, "メッセージ"))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("AI応答の取得に失敗しました");
    }

    // ============================
    // chatInPracticeMode
    // ============================
    @Test
    @DisplayName("chatInPracticeMode()が練習用プロンプトを使用してリクエストを構築する")
    void chatInPracticeModeShouldUsePracticePrompt() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("練習応答"));

        bedrockService.chatInPracticeMode("こんにちは", "あなたは上司です。");

        ArgumentCaptor<InvokeModelRequest> captor = ArgumentCaptor.forClass(InvokeModelRequest.class);
        verify(bedrockClient).invokeModel(captor.capture());

        String requestBody = captor.getValue().body().asUtf8String();
        JsonNode json = objectMapper.readTree(requestBody);

        assertThat(json.get("system").asText()).isEqualTo("あなたは上司です。");
        assertThat(json.get("temperature").asDouble()).isEqualTo(0.8);
    }

    @Test
    @DisplayName("chatInPracticeMode()がAIの応答テキストを正しく返す")
    void chatInPracticeModeShouldReturnAiReply() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("練習モード応答"));

        String result = bedrockService.chatInPracticeMode("テスト", "プロンプト");

        assertThat(result).isEqualTo("練習モード応答");
    }

    @Test
    @DisplayName("chatInPracticeMode()でエラー時にRuntimeExceptionをスローする")
    void chatInPracticeModeShouldThrowOnError() {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenThrow(new RuntimeException("Bedrock error"));

        assertThatThrownBy(() -> bedrockService.chatInPracticeMode("メッセージ", "プロンプト"))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("練習モードのAI応答取得に失敗しました");
    }

    // ============================
    // rephrase
    // ============================
    @Test
    @DisplayName("rephrase()が言い換えプロンプトを使用してリクエストを構築する")
    void rephraseShouldUseRephrasePrompt() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("言い換え結果"));

        bedrockService.rephrase("元のメッセージ", "meeting");

        ArgumentCaptor<InvokeModelRequest> captor = ArgumentCaptor.forClass(InvokeModelRequest.class);
        verify(bedrockClient).invokeModel(captor.capture());

        String requestBody = captor.getValue().body().asUtf8String();
        JsonNode json = objectMapper.readTree(requestBody);

        assertThat(json.has("system")).isTrue();
        assertThat(json.get("messages").get(0).get("content").get(0).get("text").asText())
                .isEqualTo("元のメッセージ");
    }

    @Test
    @DisplayName("rephrase()がAIの応答テキストを正しく返す")
    void rephraseShouldReturnAiReply() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("言い換え応答"));

        String result = bedrockService.rephrase("テストメッセージ", null);

        assertThat(result).isEqualTo("言い換え応答");
    }

    @Test
    @DisplayName("rephrase()でエラー時にRuntimeExceptionをスローする")
    void rephraseShouldThrowOnError() {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenThrow(new RuntimeException("Bedrock error"));

        assertThatThrownBy(() -> bedrockService.rephrase("メッセージ", null))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("言い換え提案の取得に失敗しました");
    }

    // ============================
    // chatWithUserProfileAndScene
    // ============================
    @Test
    @DisplayName("chatWithUserProfileAndScene()がシーン指定付きでリクエストを構築する")
    void chatWithUserProfileAndSceneShouldIncludeScene() throws Exception {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenReturn(createMockResponse("シーン付き応答"));

        bedrockService.chatWithUserProfileAndScene(
                "テストメッセージ", "meeting",
                "太郎", "自己紹介", "丁寧", "真面目", "目標", "懸念", "優しく");

        ArgumentCaptor<InvokeModelRequest> captor = ArgumentCaptor.forClass(InvokeModelRequest.class);
        verify(bedrockClient).invokeModel(captor.capture());

        String requestBody = captor.getValue().body().asUtf8String();
        JsonNode json = objectMapper.readTree(requestBody);

        assertThat(json.get("max_tokens").asInt()).isEqualTo(2048);
        assertThat(json.has("system")).isTrue();
    }

    @Test
    @DisplayName("chatWithUserProfileAndScene()でエラー時にRuntimeExceptionをスローする")
    void chatWithUserProfileAndSceneShouldThrowOnError() {
        when(bedrockClient.invokeModel(any(InvokeModelRequest.class)))
                .thenThrow(new RuntimeException("Bedrock error"));

        assertThatThrownBy(() -> bedrockService.chatWithUserProfileAndScene(
                "テストメッセージ", "meeting",
                "太郎", "自己紹介", "丁寧", "真面目", "目標", "懸念", "優しく"))
                .isInstanceOf(RuntimeException.class)
                .hasMessageContaining("AI応答の取得に失敗しました");
    }
}
