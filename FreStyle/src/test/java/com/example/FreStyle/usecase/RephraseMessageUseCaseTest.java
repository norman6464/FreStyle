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
}
