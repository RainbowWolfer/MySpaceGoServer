USE wjx;

DROP TABLE IF EXISTS users;
CREATE TABLE users(
	u_id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
	u_username VARCHAR(20) NOT NULL UNIQUE,
	u_password VARCHAR(32) NOT NULL, #MD5
	u_email VARCHAR(40) NOT NULL UNIQUE,
	u_profileDescription TEXT NOT NULL DEFAULT('This is a default description.')
);

DROP TABLE IF EXISTS email_validations;
CREATE TABLE email_validations(
	ev_id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
	ev_email VARCHAR(40) NOT NULL UNIQUE,
	ev_username VARCHAR(20) NOT NULL UNIQUE,
	ev_password VARCHAR(32) NOT NULL,
	ev_code VARCHAR(200) NOT NULL,
	ev_datetime DATETIME NOT NULL
);

DROP TABLE IF EXISTS posts;
CREATE TABLE posts(
	p_id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
	p_publisher_id INT NOT NULL,
	p_publish_date DATETIME NOT NULL,
	p_edit_date DATETIME,
	p_edit_times INT NOT NULL DEFAULT 0,
	p_text_content TEXT NOT NULL DEFAULT(''),
	p_deleted BOOL NOT NULL DEFAULT FALSE,
	p_upvotes INT NOT NULL DEFAULT 0,
	p_downvotes INT NOT NULL DEFAULT 0,
	p_visiblity ENUM('all','follower','none') NOT NULL DEFAULT 'all',
	p_reply ENUM('all','follower','none') NOT NULL DEFAULT 'all'
);

SELECT * FROM users;

SELECT u_avatarPath FROM users WHERE u_id = '1';

SELECT u_id FROM users WHERE u_email = 'rainbowwolfer@outlook.com';

UPDATE users SET u_avatarPath = 'a' WHERE u_id = '1';