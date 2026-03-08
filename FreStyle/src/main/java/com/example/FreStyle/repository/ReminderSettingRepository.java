package com.example.FreStyle.repository;

import java.util.Optional;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import com.example.FreStyle.entity.ReminderSetting;

@Repository
public interface ReminderSettingRepository extends JpaRepository<ReminderSetting, Integer> {
    Optional<ReminderSetting> findByUserId(Integer userId);
}
