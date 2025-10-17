package com.example.FreStyle.service;

import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.cognitoidentityprovider.CognitoIdentityProviderClient;
import software.amazon.awssdk.services.cognitoidentityprovider.model.AuthFlowType;
import software.amazon.awssdk.services.cognitoidentityprovider.model.AuthenticationResultType;
import software.amazon.awssdk.services.cognitoidentityprovider.model.InitiateAuthRequest;
import software.amazon.awssdk.services.cognitoidentityprovider.model.InitiateAuthResponse;
import software.amazon.awssdk.services.cognitoidentityprovider.model.NotAuthorizedException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UserNotConfirmedException;

@Service
public class CognitoLoginService {
  
  private final CognitoIdentityProviderClient cognitoClient;
  
  @Value("${cognito.client-id}")
  private String clientId;
  
  @Value("${cognito.client-secret}")
  private String clientSecret;
  
  public CognitoLoginService(
    @Value("${aws.access-key}") String accessKey,
    @Value("${aws.secret-key}") String secretKey,
    @Value("${aws.region}") String region
  ) {
    this.cognitoClient = CognitoIdentityProviderClient.builder()
      .credentialsProvider(StaticCredentialsProvider.create(
          AwsBasicCredentials.create(accessKey, secretKey)))
          .region(Region.of(region))
          .build();
  }
  
  public Map<String, String> login(String email, String password) {
    
    // System.out.println(email);
    // System.out.println(password);
    // System.out.println(clientId);
    // System.out.println(clientSecret);
    
    Map<String,String> authParams = new HashMap<>();
    authParams.put("USERNAME", email);
    authParams.put("PASSWORD", password);
    authParams.put("SECRET_HASH", calculateHash(email));
    
    InitiateAuthRequest authRequest = InitiateAuthRequest.builder()
        .authFlow(AuthFlowType.USER_PASSWORD_AUTH)
        .clientId(clientId)
        .authParameters(authParams)
        .build();
        
    
      try {
        InitiateAuthResponse response = cognitoClient.initiateAuth(authRequest);
        
        AuthenticationResultType result = response.authenticationResult();
        
        Map<String, String> tokens = new HashMap<>();
        tokens.put("access_token", result.accessToken());
        tokens.put("id_token", result.idToken());
        tokens.put("refresh_token", result.refreshToken());
        
        return tokens;
        
      } catch (NotAuthorizedException e) {
        throw new RuntimeException("メールアドレスまたはパスワードが間違っています。");
      } catch (UserNotConfirmedException e) {
        throw new RuntimeException("メール確認は完了していません。");
      } catch (Exception e) {
        throw new RuntimeException("ログインエラー:" + e.getMessage());
      }
      
  }
  
    // クライアントシークレットでの計算
  // SECRET_HASHを計算してサインアップリクエストに付与するには、HMAC-SHA256アルゴリズムで
  public String calculateHash(String username) {
    try {
      String message = username + clientId;
      Mac mac = Mac.getInstance("HmacSHA256");
      SecretKeySpec keySpec = new SecretKeySpec(clientSecret.getBytes("UTF-8"),"HmacSHA256");
      mac.init(keySpec);
      byte[] rawHmac = mac.doFinal(message.getBytes("UTF-8"));
      return Base64.getEncoder().encodeToString(rawHmac);
    } catch (Exception e) {
      throw new RuntimeException("Error while calculating SECRET_HASH", e);
    }
  }
  
}
