package com.example.FreStyle.infrastructure;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;

import lombok.extern.slf4j.Slf4j;
import software.amazon.awssdk.services.sqs.SqsClient;
import software.amazon.awssdk.services.sqs.model.SendMessageRequest;

@Component
@Slf4j
public class SqsMessageProducer {

    private final SqsClient sqsClient;
    private final String queueUrl;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public SqsMessageProducer(SqsClient sqsClient,
                               @Value("${aws.sqs.report-queue-url}") String queueUrl) {
        this.sqsClient = sqsClient;
        this.queueUrl = queueUrl;
    }

    public void sendReportGenerationMessage(Integer userId, Integer year, Integer month) {
        ObjectNode message = objectMapper.createObjectNode();
        message.put("userId", userId);
        message.put("year", year);
        message.put("month", month);

        String messageBody = message.toString();
        log.info("SQSメッセージ送信: queue={}, body={}", queueUrl, messageBody);

        sqsClient.sendMessage(SendMessageRequest.builder()
                .queueUrl(queueUrl)
                .messageBody(messageBody)
                .build());
    }
}
