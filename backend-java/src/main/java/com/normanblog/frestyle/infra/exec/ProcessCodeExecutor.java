package com.normanblog.frestyle.infra.exec;

import com.normanblog.frestyle.dto.CodeExecuteResponse;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Comparator;
import java.util.List;
import java.util.concurrent.TimeUnit;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * ユーザーコードを子プロセスとして実行する本番サンドボックス。
 *
 * <p>セキュリティの肝:
 *
 * <ul>
 *   <li><b>環境クリーン化</b>: 子プロセスに backend のシークレット(DB/AWS/Cognito 等)を一切渡さない
 *       (PATH/LANG だけ通す)。これが無いと提出コードから getenv で機密が抜ける
 *   <li><b>timeout</b>: 上限秒で destroyForcibly。無限ループを止める
 *   <li><b>出力上限</b>: stdout/stderr を上限バイトで打ち切り、巨大出力でメモリを潰されない
 *   <li><b>JVM ヒープ上限</b>: java は -Xmx で制限。専用 temp ディレクトリで実行し後始末する
 * </ul>
 *
 * <p>ネットワーク egress の遮断は infra(セキュリティグループ等)側で担保する前提。
 */
public class ProcessCodeExecutor implements CodeExecutor {

  private static final Logger log = LoggerFactory.getLogger(ProcessCodeExecutor.class);

  private final long timeoutSeconds;
  private final int maxOutputBytes;

  public ProcessCodeExecutor(long timeoutSeconds, int maxOutputBytes) {
    this.timeoutSeconds = timeoutSeconds;
    this.maxOutputBytes = maxOutputBytes;
  }

  @Override
  public CodeExecuteResponse execute(String language, String code) {
    if (!"java".equals(language)) {
      return new CodeExecuteResponse("", "未対応の言語です: " + language, 1);
    }

    Path dir = null;
    try {
      dir = Files.createTempDirectory("codeexec-");
      // Java 単一ファイル実行(java Main.java)。内部のクラス名は任意で良い。
      Path source = dir.resolve("Main.java");
      Files.writeString(source, code, StandardCharsets.UTF_8);

      ProcessBuilder pb =
          new ProcessBuilder(
              "java", "-Xmx64m", "-XX:+UseSerialGC", "-XX:TieredStopAtLevel=1", source.toString());
      pb.directory(dir.toFile());
      // 子プロセスに親の環境(= シークレット)を継承させない。最小限だけ通す。
      pb.environment().clear();
      String path = System.getenv("PATH");
      pb.environment().put("PATH", path == null ? "/usr/bin:/bin" : path);
      pb.environment().put("LANG", "C.UTF-8");

      return runWithLimits(pb);
    } catch (IOException e) {
      log.warn("code execution setup failed", e);
      return new CodeExecuteResponse("", "実行環境を利用できません", 1);
    } finally {
      deleteRecursively(dir);
    }
  }

  private CodeExecuteResponse runWithLimits(ProcessBuilder pb) {
    Process process;
    try {
      process = pb.start();
    } catch (IOException e) {
      log.warn("failed to start process", e);
      return new CodeExecuteResponse("", "実行環境を利用できません", 1);
    }

    // stdout/stderr を並行に汲む(片方のパイプが詰まると deadlock するため)。
    StringBuilder stdout = new StringBuilder();
    StringBuilder stderr = new StringBuilder();
    Thread outThread = drain(process.getInputStream(), stdout);
    Thread errThread = drain(process.getErrorStream(), stderr);
    outThread.start();
    errThread.start();

    try {
      boolean finished = process.waitFor(timeoutSeconds, TimeUnit.SECONDS);
      if (!finished) {
        killTree(process);
        joinQuietly(outThread);
        joinQuietly(errThread);
        return new CodeExecuteResponse(
            stdout.toString(), append(stderr, "実行がタイムアウトしました").toString(), 124);
      }
      joinQuietly(outThread);
      joinQuietly(errThread);

      return new CodeExecuteResponse(stdout.toString(), stderr.toString(), process.exitValue());
    } catch (InterruptedException e) {
      killTree(process);
      Thread.currentThread().interrupt();
      return new CodeExecuteResponse(stdout.toString(), "実行が中断されました", 1);
    }
  }

  // ユーザーコードが起動した子孫プロセスごと止める(destroyForcibly だけだと孫が残り得る)。
  private static void killTree(Process process) {
    process.descendants().forEach(ProcessHandle::destroyForcibly);
    process.destroyForcibly();
  }

  // 入力ストリームを上限バイトまで buf に読み込むスレッドを作る(start はしない)。
  private Thread drain(InputStream in, StringBuilder buf) {
    return new Thread(
        () -> {
          byte[] chunk = new byte[4096];
          int total = 0;
          try (InputStream is = in) {
            int n;
            while ((n = is.read(chunk)) != -1) {
              // 上限まではバッファに記録。上限後も read は止めず読み捨てる
              // (止めると pipe が詰まり子プロセスの write がブロックして終了できなくなる)。
              if (total < maxOutputBytes) {
                int allowed = Math.min(n, maxOutputBytes - total);
                buf.append(new String(chunk, 0, allowed, StandardCharsets.UTF_8));
                total += allowed;
              }
            }
          } catch (IOException ignored) {
            // プロセス終了でストリームが閉じるのは正常。
          }
        });
  }

  private static StringBuilder append(StringBuilder sb, String s) {
    if (sb.length() > 0) {
      sb.append('\n');
    }
    return sb.append(s);
  }

  private static void joinQuietly(Thread t) {
    try {
      t.join(2000);
    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
    }
  }

  private static void deleteRecursively(Path dir) {
    if (dir == null) {
      return;
    }
    try (var paths = Files.walk(dir)) {
      paths.sorted(Comparator.reverseOrder()).forEach(ProcessCodeExecutor::deleteQuietly);
    } catch (IOException ignored) {
      // 後始末失敗は致命ではない。
    }
  }

  private static void deleteQuietly(Path p) {
    try {
      Files.deleteIfExists(p);
    } catch (IOException ignored) {
      // ignore
    }
  }

  // 対応言語一覧(将来 php/go を足すならここに追加)。
  public static List<String> supportedLanguages() {
    return List.of("java");
  }
}
