package com.example.FreStyle.infrastructure;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.fasterxml.jackson.databind.ObjectMapper;

import software.amazon.awssdk.services.sqs.SqsClient;
import software.amazon.awssdk.services.sqs.model.SendMessageRequest;
import software.amazon.awssdk.services.sqs.model.SendMessageResponse;

import static org.assertj.core.api.Assertions.assertThat;

@ExtendWith(MockitoExtension.class)
@DisplayName("SqsMessageProducer テスト")
class SqsMessageProducerTest {

    @Mock
    private SqsClient sqsClient;

    @InjectMocks
    private SqsMessageProducer sqsMessageProducer;

    @Test
    @DisplayName("レポート生成メッセージをSQSに送信できる")
    void sendsReportGenerationMessage() throws Exception {
        when(sqsClient.sendMessage(any(SendMessageRequest.class)))
                .thenReturn(SendMessageResponse.builder().messageId("msg-123").build());

        sqsMessageProducer.sendReportGenerationMessage(1, 2026, 2);

        ArgumentCaptor<SendMessageRequest> captor = ArgumentCaptor.forClass(SendMessageRequest.class);
        verify(sqsClient).sendMessage(captor.capture());

        SendMessageRequest request = captor.getValue();
        ObjectMapper mapper = new ObjectMapper();
        var body = mapper.readTree(request.messageBody());

        assertThat(body.get("userId").asInt()).isEqualTo(1);
        assertThat(body.get("year").asInt()).isEqualTo(2026);
        assertThat(body.get("month").asInt()).isEqualTo(2);
    }
}
