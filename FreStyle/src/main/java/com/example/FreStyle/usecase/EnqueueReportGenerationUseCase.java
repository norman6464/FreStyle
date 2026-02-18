package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;

import com.example.FreStyle.infrastructure.SqsMessageProducer;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class EnqueueReportGenerationUseCase {

    private final SqsMessageProducer sqsMessageProducer;

    public void execute(Integer userId, Integer year, Integer month) {
        sqsMessageProducer.sendReportGenerationMessage(userId, year, month);
    }
}
