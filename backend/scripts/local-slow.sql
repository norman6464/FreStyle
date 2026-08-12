-- 遅いクエリの上位 20 件を pg_stat_statements から取り出す(make local-slow)。
--
-- どこが遅いかは体感ではなくここを根拠にする。total_exec_time 順にするのは、
-- 1 回あたりは速くても呼ばれる回数が多いクエリのほうが全体では効くため
-- (mean_exec_time 順だと「たまにしか呼ばれない重いクエリ」に目を奪われる)。
--
-- 計測をやり直すときは make local-slow-reset で統計をリセットする。
SELECT
    calls,
    round(mean_exec_time::numeric, 2)  AS mean_ms,
    round(total_exec_time::numeric, 2) AS total_ms,
    left(query, 90)                    AS query
FROM pg_stat_statements
-- 自分自身の計測でノイズを増やさない。
WHERE query NOT LIKE '%pg_stat_statements%'
ORDER BY total_exec_time DESC
LIMIT 20;
