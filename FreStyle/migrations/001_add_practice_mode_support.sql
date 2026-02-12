-- ==========================================
-- 本番DB用マイグレーション: 練習モード機能追加
-- 実行日: 2026-02-12
-- ==========================================

-- 1. ai_chat_sessions テーブルに練習モード用カラムを追加
ALTER TABLE ai_chat_sessions
ADD COLUMN IF NOT EXISTS session_type VARCHAR(20) DEFAULT 'normal' COMMENT 'セッション種別（normal, practice）',
ADD COLUMN IF NOT EXISTS scenario_id INT DEFAULT NULL COMMENT '練習モードのシナリオID';

-- 2. practice_scenarios テーブルを作成
CREATE TABLE IF NOT EXISTS practice_scenarios (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    role_name VARCHAR(100) NOT NULL,
    difficulty VARCHAR(20) DEFAULT 'intermediate',
    system_prompt TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 3. 初期シナリオデータを投入（ITエンジニア向け12シナリオ）

-- カテゴリ1: 顧客折衝（5シナリオ）
INSERT IGNORE INTO practice_scenarios (id, name, description, category, role_name, difficulty, system_prompt) VALUES
(1, '本番障害の緊急報告', '本番環境で重大な障害が発生。顧客への第一報と状況説明を行う', 'customer', '怒っている顧客（SIer企業のPM）', 'intermediate', 'デプロイ直後に本番環境で障害が発生しました。顧客のシステムが30分以上停止しており、顧客PMから緊急連絡が入っています。顧客は非常に怒っており、原因究明と復旧見込みを求めています。'),
(2, '要件変更の影響説明', 'スプリント中盤での要件変更依頼。スケジュール・コストへの影響を説明する', 'customer', '要件追加を求める顧客（事業部長）', 'intermediate', '開発中のWebアプリケーションについて、顧客の事業部長から「やっぱりこの機能も追加してほしい」と要望がありました。追加するとスケジュールが2週間延びる見込みです。顧客は「簡単でしょ？」と考えています。'),
(3, '見積もり提案・交渉', '新規開発案件の見積もりを提示し、予算内に収まるよう交渉する', 'customer', '予算を削りたい顧客（情シス部長）', 'advanced', '新規システム開発の見積もりを提示する場面です。顧客の予算は1000万円ですが、こちらの見積もりは1500万円です。顧客は「もっと安くならないか」と値引きを求めています。'),
(4, 'リリース延期の報告', 'テストで重大バグが発見され、予定通りのリリースが困難になった状況を報告する', 'customer', '納期を重視する顧客（プロジェクトオーナー）', 'advanced', 'リリース予定日の1週間前に重大なセキュリティバグが発見されました。修正には最低2週間必要です。顧客はこのリリースに合わせてマーケティング施策を準備しており、延期は大きな損失になります。'),
(5, '技術提案プレゼン', '顧客にクラウド移行を提案するプレゼンテーション', 'customer', '技術に詳しくない経営層（CEO）', 'intermediate', '顧客企業のオンプレミスシステムをAWSに移行する提案を行います。CEOは技術には詳しくないですが、コスト削減と事業のスピードアップに関心があります。');

-- カテゴリ2: シニアエンジニア・上司とのコミュニケーション（5シナリオ）
INSERT IGNORE INTO practice_scenarios (id, name, description, category, role_name, difficulty, system_prompt) VALUES
(6, '設計レビューでの意見対立', 'マイクロサービス vs モノリスで意見が分かれた設計レビュー', 'senior', '経験豊富なテックリード', 'advanced', '新規プロジェクトのアーキテクチャ設計レビューです。あなたはマイクロサービスを推していますが、テックリードはモノリスファーストを主張しています。チームの規模は5人で、まだMVP段階です。テックリードは10年以上の経験があり、論理的に反論してきます。'),
(7, 'コードレビューのフィードバック受け入れ', '厳しいコードレビューコメントへの対応', 'senior', '厳格なシニアエンジニア', 'beginner', 'あなたが書いたPull Requestに対して、シニアエンジニアから厳しいコードレビューコメントが20件以上付きました。設計パターンの選択、命名規則、テストカバレッジなど多岐にわたる指摘です。'),
(8, '進捗遅延の上司への報告', 'タスクが予定より大幅に遅れていることを上司に報告する', 'senior', '進捗を気にするエンジニアリングマネージャー', 'intermediate', '見積もり3日のタスクが1週間経っても完了していません。技術的な難しさを事前に見積もれなかったことが原因です。マネージャーは今週中の完了を期待しています。'),
(9, '技術負債の改善提案', 'レガシーコードのリファクタリングを上司に提案する', 'senior', 'ビジネス優先のプロダクトマネージャー', 'intermediate', 'プロダクトのコードベースが肥大化し、新機能の開発速度が低下しています。あなたは2スプリント分のリファクタリング期間を確保したいと考えていますが、PMは新機能のリリースを優先したいと考えています。'),
(10, '1on1での成長相談', 'キャリアの方向性について上司と相談する', 'senior', '面倒見の良いエンジニアリングマネージャー', 'beginner', '入社2年目のあなたが、今後のキャリアパスについてマネージャーと1on1で相談する場面です。バックエンドを中心にやってきましたが、インフラやフロントエンドにも興味があり、どう成長していくべきか悩んでいます。');

-- カテゴリ3: チーム内コミュニケーション（2シナリオ）
INSERT IGNORE INTO practice_scenarios (id, name, description, category, role_name, difficulty, system_prompt) VALUES
(11, 'チーム内での意見の食い違い', 'ライブラリ選定でチームメンバーと意見が対立した場面', 'team', '異なる意見を持つチームメンバー', 'intermediate', 'フロントエンドのステート管理ライブラリを選定する場面です。あなたはReduxを推していますが、同僚はRecoilを推しています。それぞれに技術的なメリット・デメリットがあり、チーム内で意見が分かれています。'),
(12, '後輩エンジニアへのメンタリング', '困っている後輩にアドバイスをする', 'team', '困っている後輩エンジニア', 'beginner', '後輩エンジニアがバグの原因を3時間探しても見つからず、困っています。あなたに助けを求めてきました。ただし、あなたも今日中に終わらせたいタスクを抱えています。');

-- 4. インデックスを追加（パフォーマンス最適化）
CREATE INDEX IF NOT EXISTS idx_practice_scenarios_category ON practice_scenarios(category);
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_scenario_id ON ai_chat_sessions(scenario_id);

-- マイグレーション完了
SELECT 'マイグレーション完了: 練習モード機能のテーブル・カラムを追加しました' AS status;
