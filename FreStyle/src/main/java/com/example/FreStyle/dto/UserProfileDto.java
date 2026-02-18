package com.example.FreStyle.dto;

import java.util.List;

public record UserProfileDto(
        Integer id,
        Integer userId,
        String displayName,
        String selfIntroduction,
        String communicationStyle,
        List<String> personalityTraits,
        String goals,
        String concerns,
        String preferredFeedbackStyle) {
}
