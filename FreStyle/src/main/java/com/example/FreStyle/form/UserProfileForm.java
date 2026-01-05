package com.example.FreStyle.form;

import java.util.List;

import jakarta.validation.constraints.Size;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class UserProfileForm {
    
    // 基本的な自己紹介
    @Size(max = 100, message = "表示名は100文字以内で入力してください")
    private String displayName;
    
    private String selfIntroduction;
    
    // コミュニケーションスタイル
    @Size(max = 50, message = "コミュニケーションスタイルは50文字以内で入力してください")
    private String communicationStyle;
    
    private List<String> personalityTraits;
    
    // AIフィードバック用の追加情報
    private String goals;
    
    private String concerns;
    
    @Size(max = 50, message = "フィードバックスタイルは50文字以内で入力してください")
    private String preferredFeedbackStyle;
}
