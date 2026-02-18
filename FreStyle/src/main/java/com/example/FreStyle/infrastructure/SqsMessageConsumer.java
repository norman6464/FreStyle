package com.example.FreStyle.infrastructure;

import java.util.List;
import java.util.Optional;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.UserRepository;
import com.example.FreStyle.usecase.CreateNotificationUseCase;
import com.example.FreStyle.usecase.GenerateMonthlyReportUseCase;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import lombok.extern.slf4j.Slf4j;
import software.amazon.awssdk.services.sqs.SqsClient;
import software.amazon.awssdk.services.sqs.model.DeleteMessageRequest;
import software.amazon.awssdk.services.sqs.model.Message;
import software.amazon.awssdk.services.sqs.model.ReceiveMessageRequest;

@Component
@Slf4j
public class SqsMessageConsumer {

    private final SqsClient sqsClient;
    private final GenerateMonthlyReportUseCase generateMonthlyReportUseCase;
    private final CreateNotificationUseCase createNotificationUseCase;
    private final UserRepository userRepository;
    private final String queueUrl;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public SqsMessageConsumer(SqsClient sqsClient,
                               GenerateMonthlyReportUseCase generateMonthlyReportUseCase,
                               CreateNotificationUseCase createNotificationUseCase,
                               UserRepository userRepository,
                               @Value("${aws.sqs.report-queue-url}") String queueUrl) {
        this.sqsClient = sqsClient;
        this.generateMonthlyReportUseCase = generateMonthlyReportUseCase;
        this.createNotificationUseCase = createNotificationUseCase;
        this.userRepository = userRepository;
        this.queueUrl = queueUrl;
    }

    @Scheduled(fixedDelay = 5000)
    public void pollMessages() {
        List<Message> messages = sqsClient.receiveMessage(ReceiveMessageRequest.builder()
                .queueUrl(queueUrl)
                .maxNumberOfMessages(10)
                .waitTimeSeconds(0)
                .build())
                .messages();

        for (Message message : messages) {
            try {
                processMessage(message);
            } catch (Exception e) {
                log.error("SQSメッセージ処理失敗: {}", message.body(), e);
            }
        }
    }

    private void processMessage(Message message) throws Exception {
        try {
            JsonNode body = objectMapper.readTree(message.body());
            int userId = body.get("userId").asInt();
            int year = body.get("year").asInt();
            int month = body.get("month").asInt();

            log.info("レポート生成メッセージ受信: userId={}, year={}, month={}", userId, year, month);

            Optional<User> userOpt = userRepository.findById(userId);
            if (userOpt.isEmpty()) {
                log.warn("ユーザーが見つかりません: userId={}", userId);
                return;
            }

            User user = userOpt.get();
            generateMonthlyReportUseCase.execute(user, year, month);

            createNotificationUseCase.execute(
                    user,
                    "report",
                    "月次レポート生成完了",
                    year + "年" + month + "月の月次レポートが生成されました。",
                    null);

            log.info("レポート生成完了: userId={}, year={}, month={}", userId, year, month);
        } finally {
            deleteMessage(message);
        }
    }

    private void deleteMessage(Message message) {
        sqsClient.deleteMessage(DeleteMessageRequest.builder()
                .queueUrl(queueUrl)
                .receiptHandle(message.receiptHandle())
                .build());
    }
}
