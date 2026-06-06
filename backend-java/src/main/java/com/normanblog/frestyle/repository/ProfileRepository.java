package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.Profile;
import org.springframework.data.jpa.repository.JpaRepository;

/** profiles テーブルへのアクセス。主キーは user_id。 */
public interface ProfileRepository extends JpaRepository<Profile, Long> {}
