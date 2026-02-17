package com.example.FreStyle.dto;

import java.util.List;

public record AiChatSessionStatsDto(
        int totalSessions,
        List<TypeCount> sessionsByType,
        List<SceneCount> sessionsByScene) {

    public record TypeCount(String type, long count) {
    }

    public record SceneCount(String scene, long count) {
    }
}
