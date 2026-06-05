package com.normanblog.frestyle;

import static org.assertj.core.api.Assertions.assertThat;

import org.flywaydb.core.Flyway;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

/** Flyway のマイグレーションが起動時に適用されることを検証する。 */
@SpringBootTest
class MigrationTest {

  @Autowired private Flyway flyway;

  @Test
  void flyway_appliedV1() {
    // 起動時に適用済みの migration が 1 件以上あり、最新版が V1 であること。
    assertThat(flyway.info().applied()).isNotEmpty();
    assertThat(flyway.info().current().getVersion().getVersion()).isEqualTo("1");
  }
}
