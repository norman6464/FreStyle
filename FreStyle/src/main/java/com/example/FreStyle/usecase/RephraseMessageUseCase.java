package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;

import com.example.FreStyle.service.BedrockService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class RephraseMessageUseCase {
    private final BedrockService bedrockService;

    public String execute(String originalMessage, String scene) {
        return bedrockService.rephrase(originalMessage, scene);
    }
}
