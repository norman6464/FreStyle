package com.example.FreStyle.usecase;

import static org.mockito.Mockito.verify;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.infrastructure.SqsMessageProducer;

@ExtendWith(MockitoExtension.class)
@DisplayName("EnqueueReportGenerationUseCase テスト")
class EnqueueReportGenerationUseCaseTest {

    @Mock
    private SqsMessageProducer sqsMessageProducer;

    @InjectMocks
    private EnqueueReportGenerationUseCase enqueueReportGenerationUseCase;

    @Test
    @DisplayName("SQSにレポート生成メッセージを送信する")
    void enqueuesMessage() {
        enqueueReportGenerationUseCase.execute(1, 2026, 2);

        verify(sqsMessageProducer).sendReportGenerationMessage(1, 2026, 2);
    }
}
