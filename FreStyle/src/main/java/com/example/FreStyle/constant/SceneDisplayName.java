package com.example.FreStyle.constant;

import java.util.Map;

public final class SceneDisplayName {

    private SceneDisplayName() {}

    private static final Map<String, String> SCENE_NAMES = Map.of(
            "meeting", "会議",
            "one_on_one", "1on1",
            "email", "メール",
            "presentation", "プレゼン",
            "negotiation", "商談",
            "code_review", "コードレビュー",
            "incident", "障害対応",
            "daily_report", "日報・週報"
    );

    public static String of(String scene) {
        if (scene == null) return "";
        return SCENE_NAMES.getOrDefault(scene, "");
    }
}
