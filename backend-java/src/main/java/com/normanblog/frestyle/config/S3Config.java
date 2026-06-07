package com.normanblog.frestyle.config;

import com.normanblog.frestyle.infra.s3.AiChatAttachmentPresigner;
import com.normanblog.frestyle.infra.s3.ProfileImagePresigner;
import com.normanblog.frestyle.infra.s3.S3AiChatAttachmentPresigner;
import com.normanblog.frestyle.infra.s3.S3ProfileImagePresigner;
import com.normanblog.frestyle.infra.s3.StubAiChatAttachmentPresigner;
import com.normanblog.frestyle.infra.s3.StubProfileImagePresigner;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.s3.presigner.S3Presigner;

/**
 * 画像アップロード用 presigner の組み立て。
 *
 * <p>bucket が設定されていれば AWS SDK 実装、未設定ならローカル / テスト用 stub を返す。
 * これにより AWS 資格情報の無い環境でも起動・テストできる。
 */
@Configuration
public class S3Config {

  private static final Logger log = LoggerFactory.getLogger(S3Config.class);

  @Bean
  public ProfileImagePresigner profileImagePresigner(S3Properties props) {
    if (props.bucket() == null || props.bucket().isBlank()) {
      log.info("frestyle.s3.bucket 未設定のため stub presigner を使用します");
      return new StubProfileImagePresigner(props.bucket(), props.cdnBase());
    }
    S3Presigner presigner =
        S3Presigner.builder().region(Region.of(props.regionOrDefault())).build();

    return new S3ProfileImagePresigner(presigner, props.bucket(), props.cdnBase());
  }

  @Bean
  public AiChatAttachmentPresigner aiChatAttachmentPresigner(S3Properties props) {
    if (props.bucket() == null || props.bucket().isBlank()) {
      log.info("frestyle.s3.bucket 未設定のため AI チャット添付は stub presigner を使用します");
      return new StubAiChatAttachmentPresigner(props.bucket());
    }
    S3Presigner presigner =
        S3Presigner.builder().region(Region.of(props.regionOrDefault())).build();

    return new S3AiChatAttachmentPresigner(presigner, props.bucket());
  }
}
