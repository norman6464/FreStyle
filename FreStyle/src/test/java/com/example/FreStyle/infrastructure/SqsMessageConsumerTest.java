package com.example.FreStyle.infrastructure;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Collections;
import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.UserRepository;
import com.example.FreStyle.usecase.CreateNotificationUseCase;
import com.example.FreStyle.usecase.GenerateMonthlyReportUseCase;

import software.amazon.awssdk.services.sqs.SqsClient;
import software.amazon.awssdk.services.sqs.model.DeleteMessageRequest;
import software.amazon.awssdk.services.sqs.model.Message;
import software.amazon.awssdk.services.sqs.model.ReceiveMessageRequest;
import software.amazon.awssdk.services.sqs.model.ReceiveMessageResponse;

import static org.assertj.core.api.Assertions.assertThat;

@ExtendWith(MockitoExtension.class)
@DisplayName("SqsMessageConsumer テスト")
class SqsMessageConsumerTest {

    @Mock
    private SqsClient sqsClient;

    @Mock
    private GenerateMonthlyReportUseCase generateMonthlyReportUseCase;

    @Mock
    private CreateNotificationUseCase createNotificationUseCase;

    @Mock
    private UserRepository userRepository;

    private SqsMessageConsumer sqsMessageConsumer;

    private static final String QUEUE_URL = "https://sqs.ap-northeast-1.amazonaws.com/123456789/test-queue";

    @BeforeEach
    void setUp() {
        sqsMessageConsumer = new SqsMessageConsumer(
                sqsClient, generateMonthlyReportUseCase, createNotificationUseCase, userRepository, QUEUE_URL);
    }

    @Test
    @DisplayName("メッセージ受信→レポート生成→通知作成→メッセージ削除")
    void processesMessageSuccessfully() {
        User user = new User();
        user.setId(1);
        user.setName("テストユーザー");

        String messageBody = "{\"userId\":1,\"year\":2026,\"month\":2}";
        Message message = Message.builder()
                .body(messageBody)
                .receiptHandle("receipt-123")
                .build();

        when(sqsClient.receiveMessage(any(ReceiveMessageRequest.class)))
                .thenReturn(ReceiveMessageResponse.builder().messages(List.of(message)).build());
        when(userRepository.findById(1)).thenReturn(Optional.of(user));

        sqsMessageConsumer.pollMessages();

        verify(generateMonthlyReportUseCase).execute(user, 2026, 2);
        verify(createNotificationUseCase).execute(
                eq(user), eq("report"), eq("月次レポート生成完了"),
                eq("2026年2月の月次レポートが生成されました。"), any());

        ArgumentCaptor<DeleteMessageRequest> deleteCaptor = ArgumentCaptor.forClass(DeleteMessageRequest.class);
        verify(sqsClient).deleteMessage(deleteCaptor.capture());
        assertThat(deleteCaptor.getValue().receiptHandle()).isEqualTo("receipt-123");
    }

    @Test
    @DisplayName("キューにメッセージがない場合は何もしない")
    void doesNothingWhenNoMessages() {
        when(sqsClient.receiveMessage(any(ReceiveMessageRequest.class)))
                .thenReturn(ReceiveMessageResponse.builder().messages(Collections.emptyList()).build());

        sqsMessageConsumer.pollMessages();

        verify(generateMonthlyReportUseCase, never()).execute(any(), any(), any());
        verify(createNotificationUseCase, never()).execute(any(), any(), any(), any(), any());
    }

    @Test
    @DisplayName("ユーザーが見つからない場合でもメッセージを削除する")
    void deletesMessageWhenUserNotFound() {
        String messageBody = "{\"userId\":999,\"year\":2026,\"month\":2}";
        Message message = Message.builder()
                .body(messageBody)
                .receiptHandle("receipt-456")
                .build();

        when(sqsClient.receiveMessage(any(ReceiveMessageRequest.class)))
                .thenReturn(ReceiveMessageResponse.builder().messages(List.of(message)).build());
        when(userRepository.findById(999)).thenReturn(Optional.empty());

        sqsMessageConsumer.pollMessages();

        verify(generateMonthlyReportUseCase, never()).execute(any(), any(), any());
        verify(sqsClient).deleteMessage(any(DeleteMessageRequest.class));
    }
}
