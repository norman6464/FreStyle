package com.example.FreStyle.config;

import static org.mockito.Mockito.verify;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.repository.UserRepository;

@DisplayName("WarmupListener テスト")
@ExtendWith(MockitoExtension.class)
class WarmupListenerTest {

    @Mock
    private UserRepository userRepository;

    @InjectMocks
    private WarmupListener warmupListener;

    @Test
    @DisplayName("onApplicationReady実行時にDBウォームアップクエリが実行される")
    void onApplicationReadyExecutesWarmupQuery() {
        warmupListener.onApplicationReady();

        verify(userRepository).count();
    }
}
