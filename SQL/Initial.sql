USE wjx;

SET FOREIGN_KEY_CHECKS = 0;

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
	p_id INT PRIMARY KEY AUTO_INCREMENT NOT NULL, #Both
	p_publisher_id INT NOT NULL, #Both
	p_publish_date DATETIME NOT NULL, #Both
	p_edit_date DATETIME NOT NULL, #Both
	p_edit_times INT NOT NULL DEFAULT 0, #Both
	p_text_content TEXT NOT NULL DEFAULT(''), #Both
	p_deleted BOOL NOT NULL DEFAULT FALSE, #Both
	p_images_count INT NOT NULL DEFAULT 0, #Post
	p_tags VARCHAR(500) NOT NULL DEFAULT '', #Post
	p_upvotes INT NOT NULL DEFAULT 0, #Both
	p_downvotes INT NOT NULL DEFAULT 0, #Both
	p_repost INT NOT NULL DEFAULT 0, #Post
	p_comments INT NOT NULL DEFAULT 0, #Both
	p_visibility ENUM('all','follower','none') NOT NULL DEFAULT 'all', #Both
	p_reply ENUM('all','follower','none') NOT NULL DEFAULT 'all', #Both
	p_is_repost BOOL NOT NULL DEFAULT FALSE, #Repost
	p_id_origin_post INT NOT NULL DEFAULT 0, #Repost
	p_id_reposter INT NOT NULL DEFAULT 0, #Repost
	CONSTRAINT fk_p_publisher_id FOREIGN KEY (p_publisher_id) REFERENCES users(u_id)
);

DROP TABLE IF EXISTS comments;
CREATE TABLE comments(
	c_id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
	c_id_user INT NOT NULL,
	c_id_post INT NOT NULL,
	c_text_content TEXT NOT NULL,
	c_datetime DATETIME NOT NULL,
	c_upvotes INT NOT NULL DEFAULT 0,
	c_downvote INT NOT NULL DEFAULT 0,
	CONSTRAINT fk_c_id_uesr FOREIGN KEY (c_id_user) REFERENCES users(u_id),
	CONSTRAINT fk_c_id_post FOREIGN KEY (c_id_post) REFERENCES posts(p_id)
);

DROP TABLE IF EXISTS tags;
CREATE TABLE tags(
	t_id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
	t_tag VARCHAR(40) NOT NULL UNIQUE
);

DROP TABLE IF EXISTS post_likes;
CREATE TABLE post_likes(
	pl_id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
	pl_id_user INT NOT NULL,
	pl_id_post INT NOT NULL,
	pl_vote BOOL NOT NULL,
	pl_datetime DATETIME NOT NULL,
	CONSTRAINT fk_pl_id_user FOREIGN KEY (pl_id_user) REFERENCES users(u_id),
	CONSTRAINT fk_pl_id_post FOREIGN KEY (pl_id_post) REFERENCES posts(p_id)
);

DROP TABLE IF EXISTS comment_likes;
CREATE TABLE comment_likes(
	cl_id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
	cl_id_user INT NOT NULL,
	cl_id_comment INT NOT NULL,
	pl_vote BOOL NOT NULL,
	pl_datetime DATETIME NOT NULL,
	CONSTRAINT fk_cl_id_user FOREIGN KEY (cl_id_user) REFERENCES users(u_id),
	CONSTRAINT fk_cl_id_comment FOREIGN KEY (cl_id_comment) REFERENCES comments(c_id)
);

SET FOREIGN_KEY_CHECKS = 1;


INSERT INTO users VALUES
	(1,'myspace','myspace','RainbowWolfer@outlook.com','This is official account for MySpace. Feel free to tell us what improvoments should be made or just come small talking. All are welcomed!');

INSERT INTO posts VALUES
	(1,1,NOW(),NOW(),0,'Welcome to My Space!',FALSE,0,'official,welcome,lucky',0,0,0,0,'all','all',FALSE,-1,-1);
INSERT INTO posts VALUES
	(2,1,NOW(),NOW(),0,'Here, you can post whatever you like and make new friends!',FALSE,0,'official,welcome,lucky',0,0,0,0,'all','all',FALSE,-1,-1);
INSERT INTO posts VALUES
	(3,1,NOW(),NOW(),0,'And don\'t forget to have fun!',FALSE,0,'official,welcome,lucky',0,0,0,0,'all','all',FALSE,-1,-1);
