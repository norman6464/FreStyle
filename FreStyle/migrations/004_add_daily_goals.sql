CREATE TABLE IF NOT EXISTS daily_goals (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    goal_date DATE NOT NULL,
    target INT NOT NULL DEFAULT 3,
    completed INT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_user_date (user_id, goal_date),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
