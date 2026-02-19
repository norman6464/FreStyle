package com.example.FreStyle.config;

import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

import com.example.FreStyle.repository.UserRepository;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@Component
@RequiredArgsConstructor
public class WarmupListener {

    private final UserRepository userRepository;

    @EventListener(ApplicationReadyEvent.class)
    public void onApplicationReady() {
        log.info("ウォームアップ開始: DB接続プール・Hibernateの事前初期化");
        try {
            long start = System.currentTimeMillis();
            userRepository.count();
            long elapsed = System.currentTimeMillis() - start;
            log.info("ウォームアップ完了: {}ms", elapsed);
        } catch (Exception e) {
            log.warn("ウォームアップ中にエラーが発生しましたが、起動は継続します: {}", e.getMessage());
        }
    }
}
