-- users テーブルの display_name カラムを name にリネームする
ALTER TABLE users RENAME COLUMN display_name TO name;

-- admin_invitations テーブル（invitations として管理）の display_name カラムを name にリネームする
ALTER TABLE invitations RENAME COLUMN display_name TO name;
