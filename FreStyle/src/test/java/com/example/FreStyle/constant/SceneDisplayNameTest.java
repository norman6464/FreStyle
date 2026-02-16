package com.example.FreStyle.constant;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.Test;

class SceneDisplayNameTest {

    @Test
    void 全シーンの表示名を取得できる() {
        assertThat(SceneDisplayName.of("meeting")).isEqualTo("会議");
        assertThat(SceneDisplayName.of("one_on_one")).isEqualTo("1on1");
        assertThat(SceneDisplayName.of("email")).isEqualTo("メール");
        assertThat(SceneDisplayName.of("presentation")).isEqualTo("プレゼン");
        assertThat(SceneDisplayName.of("negotiation")).isEqualTo("商談");
        assertThat(SceneDisplayName.of("code_review")).isEqualTo("コードレビュー");
        assertThat(SceneDisplayName.of("incident")).isEqualTo("障害対応");
        assertThat(SceneDisplayName.of("daily_report")).isEqualTo("日報・週報");
    }

    @Test
    void nullの場合は空文字を返す() {
        assertThat(SceneDisplayName.of(null)).isEmpty();
    }

    @Test
    void 未知のシーンの場合は空文字を返す() {
        assertThat(SceneDisplayName.of("unknown")).isEmpty();
    }
}
