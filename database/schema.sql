-- 围棋复盘教学系统数据库脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS go_replay DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE go_replay;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码(哈希)',
    nickname VARCHAR(50) COMMENT '昵称',
    role ENUM('student', 'teacher', 'admin') NOT NULL DEFAULT 'student' COMMENT '角色',
    avatar VARCHAR(255) COMMENT '头像URL',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 棋谱表
CREATE TABLE IF NOT EXISTS games (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL COMMENT '上传用户ID',
    title VARCHAR(255) NOT NULL COMMENT '棋谱标题',
    black_player VARCHAR(100) COMMENT '黑方',
    white_player VARCHAR(100) COMMENT '白方',
    board_size TINYINT NOT NULL DEFAULT 19 COMMENT '棋盘大小(9/13/19)',
    komi DECIMAL(4,1) DEFAULT 6.5 COMMENT '贴目',
    result VARCHAR(50) COMMENT '结果',
    date_played DATE COMMENT '对局日期',
    sgf_content LONGTEXT NOT NULL COMMENT 'SGF原始内容',
    description TEXT COMMENT '描述',
    is_public TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否公开',
    view_count INT NOT NULL DEFAULT 0 COMMENT '浏览次数',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='棋谱表';

-- 问题手/标记表
CREATE TABLE IF NOT EXISTS move_markers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    game_id BIGINT NOT NULL COMMENT '棋谱ID',
    move_number INT NOT NULL COMMENT '手数(第几手)',
    node_path VARCHAR(255) COMMENT 'SGF节点路径(如0-1-2)',
    marker_type ENUM('black_adv', 'white_adv', 'key', 'question', 'good') NOT NULL COMMENT '标记类型:黑优/白优/关键手/疑问手/好棋',
    user_id BIGINT COMMENT '标记用户ID',
    note TEXT COMMENT '标记备注',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_game_move (game_id, move_number),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='手标记表';

-- 点评表
CREATE TABLE IF NOT EXISTS comments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    game_id BIGINT NOT NULL COMMENT '棋谱ID',
    move_number INT NOT NULL COMMENT '起始手数',
    node_path VARCHAR(255) COMMENT 'SGF节点路径',
    user_id BIGINT NOT NULL COMMENT '点评用户ID(老师)',
    content TEXT NOT NULL COMMENT '点评内容',
    variation_sgf TEXT COMMENT '参考变化图SGF',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_game_move (game_id, move_number),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点评表';

-- 死活题库表
CREATE TABLE IF NOT EXISTS problems (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL COMMENT '出题用户ID',
    title VARCHAR(255) NOT NULL COMMENT '题目名称',
    board_size TINYINT NOT NULL DEFAULT 19 COMMENT '棋盘大小',
    goal ENUM('black_kill', 'black_live', 'white_kill', 'white_live') NOT NULL COMMENT '目标:黑先杀/黑先活/白先杀/白先活',
    initial_sgf TEXT NOT NULL COMMENT '初始局面SGF',
    solution_sgf TEXT NOT NULL COMMENT '正解SGF(包含正确路径)',
    description TEXT COMMENT '题目说明',
    difficulty ENUM('easy', 'medium', 'hard', 'expert') NOT NULL DEFAULT 'medium' COMMENT '难度',
    is_public TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否公开',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_difficulty (difficulty),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='死活题库表';

-- 作答记录表
CREATE TABLE IF NOT EXISTS problem_attempts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    problem_id BIGINT NOT NULL COMMENT '题目ID',
    user_id BIGINT NOT NULL COMMENT '作答用户ID',
    user_moves TEXT COMMENT '用户落子序列JSON',
    is_correct TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否正确',
    time_spent INT COMMENT '用时(秒)',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_problem_user (problem_id, user_id),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='作答记录表';

-- 棋谱收藏表
CREATE TABLE IF NOT EXISTS game_favorites (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    game_id BIGINT NOT NULL COMMENT '棋谱ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_game (user_id, game_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='棋谱收藏表';

-- 插入初始测试数据
-- 默认管理员用户 (密码: admin123, 实际需要bcrypt哈希)
INSERT INTO users (username, password, nickname, role) VALUES
('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '管理员', 'admin'),
('teacher1', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '张老师', 'teacher'),
('student1', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '小明', 'student')
ON DUPLICATE KEY UPDATE username=username;
