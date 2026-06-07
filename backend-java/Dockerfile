# --- Build stage ---
# Gradle Wrapper(9.5.1) で bootJar を生成する。Java 21 toolchain。
FROM eclipse-temurin:21-jdk-noble AS builder

WORKDIR /app

# 依存解決を layer cache に乗せるため、まず wrapper と build script だけ COPY。
COPY gradlew settings.gradle build.gradle ./
COPY gradle ./gradle
RUN ./gradlew --no-daemon dependencies > /dev/null 2>&1 || true

COPY src ./src
RUN ./gradlew --no-daemon clean bootJar

# --- Runtime stage ---
# JDK image を使う。理由: コード演習サンドボックスが `java Main.java`(単一ファイル実行)で
# ユーザーの Java コードをコンパイル+実行するため、JRE ではなく javac を含む JDK が要る。
# ECS 0.5GB 制約に合わせ container-aware な JVM 設定で起動する。
FROM eclipse-temurin:21-jdk-noble

WORKDIR /

# root で実行するとコンテナ侵害時の被害が大きいため、非 root ユーザーで動かす。
# UID=65532 は distroless 由来の慣例値。
RUN groupadd -g 65532 nonroot && useradd -u 65532 -g nonroot -s /sbin/nologin nonroot

COPY --from=builder /app/build/libs/*-SNAPSHOT.jar /app.jar
RUN chown nonroot:nonroot /app.jar

USER nonroot:nonroot
EXPOSE 8080

# MaxRAMPercentage=70: コンテナ memory limit(ECS task) からヒープ上限を自動計算。
# ローカル実測では 0.5GB 制約下で RSS ~270MB / heap 360MB cap。
ENTRYPOINT ["java", "-XX:MaxRAMPercentage=70", "-XX:+UseG1GC", "-Xss512k", "-jar", "/app.jar"]
