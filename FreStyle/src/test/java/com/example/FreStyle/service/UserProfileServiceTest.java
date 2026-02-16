package com.example.FreStyle.service;

import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.entity.UserProfile;
import com.example.FreStyle.form.UserProfileForm;
import com.example.FreStyle.repository.UserProfileRepository;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UserProfileServiceTest {

    @Mock
    private UserProfileRepository userProfileRepository;

    @Mock
    private ObjectMapper objectMapper;

    @InjectMocks
    private UserProfileService userProfileService;

    private User createUser(Integer id) {
        User user = new User();
        user.setId(id);
        user.setName("テストユーザー");
        return user;
    }

    private UserProfile createProfile(Integer id, User user) {
        UserProfile profile = new UserProfile();
        profile.setId(id);
        profile.setUser(user);
        profile.setDisplayName("表示名");
        profile.setSelfIntroduction("自己紹介");
        profile.setCommunicationStyle("friendly");
        profile.setPersonalityTraits(null);
        profile.setGoals("目標");
        profile.setConcerns("悩み");
        profile.setPreferredFeedbackStyle("gentle");
        return profile;
    }

    private UserProfileForm createForm() {
        UserProfileForm form = new UserProfileForm();
        form.setDisplayName("新しい名前");
        form.setSelfIntroduction("新しい自己紹介");
        form.setCommunicationStyle("casual");
        form.setPersonalityTraits(List.of("明るい", "積極的"));
        form.setGoals("新しい目標");
        form.setConcerns("新しい悩み");
        form.setPreferredFeedbackStyle("direct");
        return form;
    }

    @Test
    @DisplayName("getProfileByUserId: プロファイルが存在する場合DTOを返す")
    void getProfileByUserId_returnsDto() {
        User user = createUser(1);
        UserProfile profile = createProfile(10, user);
        when(userProfileRepository.findByUserId(1)).thenReturn(Optional.of(profile));

        UserProfileDto result = userProfileService.getProfileByUserId(1);

        assertNotNull(result);
        assertEquals(10, result.getId());
        assertEquals(1, result.getUserId());
        assertEquals("表示名", result.getDisplayName());
    }

    @Test
    @DisplayName("getProfileByUserId: プロファイルが存在しない場合nullを返す")
    void getProfileByUserId_returnsNullWhenNotFound() {
        when(userProfileRepository.findByUserId(999)).thenReturn(Optional.empty());

        UserProfileDto result = userProfileService.getProfileByUserId(999);

        assertNull(result);
    }

    @Test
    @DisplayName("getProfileById: プロファイルをDTOで返す")
    void getProfileById_returnsDto() {
        User user = createUser(1);
        UserProfile profile = createProfile(10, user);
        when(userProfileRepository.findById(10)).thenReturn(Optional.of(profile));

        UserProfileDto result = userProfileService.getProfileById(10);

        assertNotNull(result);
        assertEquals(10, result.getId());
    }

    @Test
    @DisplayName("getProfileById: 存在しない場合例外をスローする")
    void getProfileById_throwsWhenNotFound() {
        when(userProfileRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userProfileService.getProfileById(999));
        assertEquals("プロファイルが見つかりません。", ex.getMessage());
    }

    @Test
    @DisplayName("createProfile: プロファイルを作成してDTOを返す")
    void createProfile_createsAndReturnsDto() throws JsonProcessingException {
        User user = createUser(1);
        UserProfileForm form = createForm();
        when(userProfileRepository.existsByUserId(1)).thenReturn(false);
        when(objectMapper.writeValueAsString(form.getPersonalityTraits()))
                .thenReturn("[\"明るい\",\"積極的\"]");
        when(userProfileRepository.save(any())).thenAnswer(inv -> {
            UserProfile p = inv.getArgument(0);
            p.setId(10);
            return p;
        });

        UserProfileDto result = userProfileService.createProfile(user, form);

        assertNotNull(result);
        assertEquals("新しい名前", result.getDisplayName());
        assertEquals("casual", result.getCommunicationStyle());
        verify(userProfileRepository).save(any());
    }

    @Test
    @DisplayName("createProfile: 既にプロファイルが存在する場合例外をスローする")
    void createProfile_throwsWhenAlreadyExists() {
        User user = createUser(1);
        UserProfileForm form = createForm();
        when(userProfileRepository.existsByUserId(1)).thenReturn(true);

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userProfileService.createProfile(user, form));
        assertEquals("プロファイルは既に存在します。", ex.getMessage());
    }

    @Test
    @DisplayName("updateProfile: プロファイルを更新してDTOを返す")
    void updateProfile_updatesAndReturnsDto() throws JsonProcessingException {
        User user = createUser(1);
        UserProfile profile = createProfile(10, user);
        UserProfileForm form = createForm();
        when(userProfileRepository.findByUserId(1)).thenReturn(Optional.of(profile));
        when(objectMapper.writeValueAsString(form.getPersonalityTraits()))
                .thenReturn("[\"明るい\",\"積極的\"]");
        when(userProfileRepository.save(any())).thenAnswer(inv -> inv.getArgument(0));

        UserProfileDto result = userProfileService.updateProfile(1, form);

        assertEquals("新しい名前", result.getDisplayName());
        assertEquals("新しい目標", result.getGoals());
    }

    @Test
    @DisplayName("updateProfile: 存在しない場合例外をスローする")
    void updateProfile_throwsWhenNotFound() {
        when(userProfileRepository.findByUserId(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userProfileService.updateProfile(999, createForm()));
        assertEquals("プロファイルが見つかりません。", ex.getMessage());
    }

    @Test
    @DisplayName("createOrUpdateProfile: 新規作成の場合プロファイルを作成する")
    void createOrUpdateProfile_createsNew() throws JsonProcessingException {
        User user = createUser(1);
        UserProfileForm form = createForm();
        when(userProfileRepository.findByUserId(1)).thenReturn(Optional.empty());
        when(objectMapper.writeValueAsString(form.getPersonalityTraits()))
                .thenReturn("[\"明るい\",\"積極的\"]");
        when(userProfileRepository.save(any())).thenAnswer(inv -> {
            UserProfile p = inv.getArgument(0);
            p.setId(20);
            return p;
        });

        UserProfileDto result = userProfileService.createOrUpdateProfile(user, form);

        assertNotNull(result);
        assertEquals(1, result.getUserId());
        assertEquals("新しい名前", result.getDisplayName());
    }

    @Test
    @DisplayName("createOrUpdateProfile: 既存の場合プロファイルを更新する")
    void createOrUpdateProfile_updatesExisting() throws JsonProcessingException {
        User user = createUser(1);
        UserProfile existing = createProfile(10, user);
        UserProfileForm form = createForm();
        when(userProfileRepository.findByUserId(1)).thenReturn(Optional.of(existing));
        when(objectMapper.writeValueAsString(form.getPersonalityTraits()))
                .thenReturn("[\"明るい\",\"積極的\"]");
        when(userProfileRepository.save(any())).thenAnswer(inv -> inv.getArgument(0));

        UserProfileDto result = userProfileService.createOrUpdateProfile(user, form);

        assertEquals(10, result.getId());
        assertEquals("新しい名前", result.getDisplayName());
    }

    @Test
    @DisplayName("deleteProfile: プロファイルを削除する")
    void deleteProfile_deletesProfile() {
        when(userProfileRepository.existsByUserId(1)).thenReturn(true);

        userProfileService.deleteProfile(1);

        verify(userProfileRepository).deleteByUserId(1);
    }

    @Test
    @DisplayName("deleteProfile: 存在しない場合例外をスローする")
    void deleteProfile_throwsWhenNotFound() {
        when(userProfileRepository.existsByUserId(999)).thenReturn(false);

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userProfileService.deleteProfile(999));
        assertEquals("プロファイルが見つかりません。", ex.getMessage());
    }

    @Test
    @DisplayName("convertToDto: 不正なJSON personalityTraitsは空リストにフォールバックする")
    void convertToDto_invalidJson_fallsBackToEmptyList() throws JsonProcessingException {
        User user = createUser(1);
        UserProfile profile = createProfile(10, user);
        profile.setPersonalityTraits("invalid-json");
        when(userProfileRepository.findByUserId(1)).thenReturn(Optional.of(profile));
        when(objectMapper.readValue(eq("invalid-json"), any(com.fasterxml.jackson.core.type.TypeReference.class)))
                .thenThrow(new JsonProcessingException("parse error") {});

        UserProfileDto result = userProfileService.getProfileByUserId(1);

        assertNotNull(result);
        assertTrue(result.getPersonalityTraits().isEmpty());
    }

    @Test
    @DisplayName("updateProfileFromForm: personalityTraitsのJSON変換失敗でRuntimeExceptionをスローする")
    void updateProfileFromForm_jsonProcessingException_throwsRuntimeException() throws JsonProcessingException {
        User user = createUser(1);
        UserProfileForm form = createForm();
        when(userProfileRepository.existsByUserId(1)).thenReturn(false);
        when(objectMapper.writeValueAsString(form.getPersonalityTraits()))
                .thenThrow(new JsonProcessingException("write error") {});

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userProfileService.createProfile(user, form));
        assertEquals("性格特性のJSON変換に失敗しました。", ex.getMessage());
    }

    @Test
    @DisplayName("createProfile: personalityTraitsがnullの場合nullが設定される")
    void createProfile_nullPersonalityTraits_setsNull() {
        User user = createUser(1);
        UserProfileForm form = createForm();
        form.setPersonalityTraits(null);
        when(userProfileRepository.existsByUserId(1)).thenReturn(false);
        when(userProfileRepository.save(any())).thenAnswer(inv -> {
            UserProfile p = inv.getArgument(0);
            p.setId(10);
            return p;
        });

        UserProfileDto result = userProfileService.createProfile(user, form);

        assertNotNull(result);
        assertTrue(result.getPersonalityTraits().isEmpty());
    }
}
