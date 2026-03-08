package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.service.BedrockService;

@ExtendWith(MockitoExtension.class)
@DisplayName("RephraseMessageUseCase テスト")
class RephraseMessageUseCaseTest {

    @Mock
    private BedrockService bedrockService;

    @InjectMocks
    private RephraseMessageUseCase rephraseMessageUseCase;

    @Test
    @DisplayName("メッセージを言い換える")
    void execute_rephrasesMessage() {
        when(bedrockService.rephrase("元のメッセージ", "会議")).thenReturn("言い換え結果");

        String result = rephraseMessageUseCase.execute("元のメッセージ", "会議");

        assertThat(result).isEqualTo("言い換え結果");
        verify(bedrockService).rephrase("元のメッセージ", "会議");
    }

    @Test
    @DisplayName("空のメッセージでも処理する")
    void execute_handlesEmptyMessage() {
        when(bedrockService.rephrase("", "日常")).thenReturn("空の言い換え結果");

        String result = rephraseMessageUseCase.execute("", "日常");

        assertThat(result).isEqualTo("空の言い換え結果");
        verify(bedrockService).rephrase("", "日常");
    }

    @Test
    @DisplayName("BedrockServiceに正しい引数を渡す")
    void execute_passesCorrectArgumentsToBedrockService() {
        when(bedrockService.rephrase("特定のメッセージ", "ビジネス")).thenReturn("結果");

        rephraseMessageUseCase.execute("特定のメッセージ", "ビジネス");

        verify(bedrockService).rephrase("特定のメッセージ", "ビジネス");
    }
}
