package com.example.FreStyle.service;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.entity.UserIdentity;
import com.example.FreStyle.repository.UserIdentityRepository;

@Service
public class UserIdentityService {

    private final UserIdentityRepository userIdentityRepository;

    public UserIdentityService(UserIdentityRepository userIdentityRepository) {
        this.userIdentityRepository = userIdentityRepository;
    }

    // ------------------------
    // UserIdentity を登録
    // ------------------------
    @Transactional
    public void registerUserIdentity(User user, String provider, String sub) {

        boolean exists = userIdentityRepository
                .findByProviderAndProviderSub(provider, sub)
                .isPresent();

        if (!exists) {
            UserIdentity identity = new UserIdentity();
            identity.setUser(user);
            identity.setProvider(provider);
            identity.setProviderSub(sub);
            userIdentityRepository.save(identity);
        }
    }

    // ------------------------
    // sub から User を取得
    // ------------------------
    @Transactional(readOnly = true)
    public User findUserBySub(String sub) {
        UserIdentity identity = userIdentityRepository
                .findByProviderSub(sub)
                .orElseThrow(() -> new RuntimeException("ユーザーが存在しません。"));

        return identity.getUser();
    }
}
