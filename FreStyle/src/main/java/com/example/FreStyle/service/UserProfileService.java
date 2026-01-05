package com.example.FreStyle.service;

import java.util.ArrayList;
import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.entity.UserProfile;
import com.example.FreStyle.form.UserProfileForm;
import com.example.FreStyle.repository.UserProfileRepository;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class UserProfileService {

    private final UserProfileRepository userProfileRepository;
    private final ObjectMapper objectMapper;

    // ------------------------
    // プロファイル取得（ユーザーIDで）
    // ------------------------
    public UserProfileDto getProfileByUserId(Integer userId) {
        UserProfile profile = userProfileRepository.findByUserId(userId)
                .orElse(null);
        
        if (profile == null) {
            return null;
        }
        
        return convertToDto(profile);
    }

    // ------------------------
    // プロファイル取得（IDで）
    // ------------------------
    public UserProfileDto getProfileById(Integer id) {
        UserProfile profile = userProfileRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("プロファイルが見つかりません。"));
        
        return convertToDto(profile);
    }

    // ------------------------
    // プロファイル作成
    // ------------------------
    @Transactional
    public UserProfileDto createProfile(User user, UserProfileForm form) {
        // 既存のプロファイルがあるか確認
        if (userProfileRepository.existsByUserId(user.getId())) {
            throw new RuntimeException("プロファイルは既に存在します。");
        }

        UserProfile profile = new UserProfile();
        profile.setUser(user);
        updateProfileFromForm(profile, form);
        
        UserProfile savedProfile = userProfileRepository.save(profile);
        return convertToDto(savedProfile);
    }

    // ------------------------
    // プロファイル更新
    // ------------------------
    @Transactional
    public UserProfileDto updateProfile(Integer userId, UserProfileForm form) {
        UserProfile profile = userProfileRepository.findByUserId(userId)
                .orElseThrow(() -> new RuntimeException("プロファイルが見つかりません。"));
        
        updateProfileFromForm(profile, form);
        
        UserProfile savedProfile = userProfileRepository.save(profile);
        return convertToDto(savedProfile);
    }

    // ------------------------
    // プロファイル作成または更新（upsert）
    // ------------------------
    @Transactional
    public UserProfileDto createOrUpdateProfile(User user, UserProfileForm form) {
        UserProfile profile = userProfileRepository.findByUserId(user.getId())
                .orElse(new UserProfile());
        
        if (profile.getId() == null) {
            profile.setUser(user);
        }
        
        updateProfileFromForm(profile, form);
        
        UserProfile savedProfile = userProfileRepository.save(profile);
        return convertToDto(savedProfile);
    }

    // ------------------------
    // プロファイル削除
    // ------------------------
    @Transactional
    public void deleteProfile(Integer userId) {
        if (!userProfileRepository.existsByUserId(userId)) {
            throw new RuntimeException("プロファイルが見つかりません。");
        }
        userProfileRepository.deleteByUserId(userId);
    }

    // ------------------------
    // ヘルパーメソッド: FormからEntityへの変換
    // ------------------------
    private void updateProfileFromForm(UserProfile profile, UserProfileForm form) {
        profile.setDisplayName(form.getDisplayName());
        profile.setSelfIntroduction(form.getSelfIntroduction());
        profile.setCommunicationStyle(form.getCommunicationStyle());
        profile.setGoals(form.getGoals());
        profile.setConcerns(form.getConcerns());
        profile.setPreferredFeedbackStyle(form.getPreferredFeedbackStyle());
        
        // List<String> を JSON 文字列に変換
        if (form.getPersonalityTraits() != null) {
            try {
                profile.setPersonalityTraits(objectMapper.writeValueAsString(form.getPersonalityTraits()));
            } catch (JsonProcessingException e) {
                throw new RuntimeException("性格特性のJSON変換に失敗しました。", e);
            }
        } else {
            profile.setPersonalityTraits(null);
        }
    }

    // ------------------------
    // ヘルパーメソッド: EntityからDTOへの変換
    // ------------------------
    private UserProfileDto convertToDto(UserProfile profile) {
        UserProfileDto dto = new UserProfileDto();
        dto.setId(profile.getId());
        dto.setUserId(profile.getUser().getId());
        dto.setDisplayName(profile.getDisplayName());
        dto.setSelfIntroduction(profile.getSelfIntroduction());
        dto.setCommunicationStyle(profile.getCommunicationStyle());
        dto.setGoals(profile.getGoals());
        dto.setConcerns(profile.getConcerns());
        dto.setPreferredFeedbackStyle(profile.getPreferredFeedbackStyle());
        
        // JSON 文字列を List<String> に変換
        if (profile.getPersonalityTraits() != null) {
            try {
                List<String> traits = objectMapper.readValue(
                        profile.getPersonalityTraits(),
                        new TypeReference<List<String>>() {});
                dto.setPersonalityTraits(traits);
            } catch (JsonProcessingException e) {
                dto.setPersonalityTraits(new ArrayList<>());
            }
        } else {
            dto.setPersonalityTraits(new ArrayList<>());
        }
        
        return dto;
    }
}
