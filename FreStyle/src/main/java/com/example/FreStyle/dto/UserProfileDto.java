package com.example.FreStyle.dto;

import java.util.List;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class UserProfileDto {
    private Integer id;
    private Integer userId;
    
    // 基本的な自己紹介
    private String displayName;
    private String selfIntroduction;
    
    // コミュニケーションスタイル
    private String communicationStyle;
    private List<String> personalityTraits;
    
    // AIフィードバック用の追加情報
    private String goals;
    private String concerns;
    private String preferredFeedbackStyle;
}
