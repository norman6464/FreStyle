package com.example.FreStyle.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.UserProfile;
import com.example.FreStyle.entity.User;

@Repository
public interface UserProfileRepository extends JpaRepository<UserProfile, Integer> {

    /**
     * ユーザーIDでプロファイルを検索
     */
    Optional<UserProfile> findByUserId(Integer userId);

    /**
     * ユーザーエンティティでプロファイルを検索
     */
    Optional<UserProfile> findByUser(User user);

    /**
     * ユーザーIDでプロファイルが存在するか確認
     */
    boolean existsByUserId(Integer userId);

    /**
     * ユーザーIDでプロファイルを削除
     */
    void deleteByUserId(Integer userId);
}
