DROP PROCEDURE IF EXISTS add_tags;
DELIMITER @@
CREATE PROCEDURE add_tags(IN tags_str VARCHAR(700))
BEGIN
	DECLARE full_length INT;
	DECLARE str_length INT;
	DECLARE loop_index INT;
	DECLARE item_str VARCHAR(40);
	SET full_length = LENGTH(tags_str);
	SET str_length = LENGTH(REPLACE(tags_str,',',''));
	SET loop_index = 1;
	WHILE loop_index <= full_length - str_length + 1 DO
		SET item_str = REPLACE(SUBSTRING(SUBSTRING_INDEX(tags_str, ',', loop_index), LENGTH(SUBSTRING_INDEX(tags_str, ',', loop_index -1)) + 1),',', '');
		SET loop_index = loop_index + 1;
		SELECT item_str;
		IF (SELECT COUNT(t_id) FROM tags WHERE t_tag = item_str) = 0 THEN
			INSERT INTO tags(t_tag) VALUES(item_str);
		END IF;
	END WHILE;
END@@
DELIMITER ;

CALL add_tags('yiff,furry');

## RANDOM

DROP PROCEDURE IF EXISTS CreateUserRandomPosts;
DELIMITER @@
CREATE PROCEDURE CreateUserRandomPosts(IN user_id INT)
BEGIN
	SET @table_name := CONCAT('user_',user_id,'_random_posts');
	SET @sql := CONCAT('DROP VIEW IF EXISTS ',@table_name);
	PREPARE delete_sql FROM @sql;
	EXECUTE delete_sql;
	SET @sql := CONCAT('CREATE VIEW ',@table_name,' AS SELECT 
	p_id,p_publisher_id,p_publish_date,p_edit_date
	,p_edit_times,p_text_content,p_deleted,p_images_count
	,p_tags,p_upvotes,p_downvotes,p_repost,p_comments
	,p_visibility,p_reply,p_is_repost,p_id_reposter,p_id_origin_post,
	p_upvotes - p_downvotes AS p_score,
	RAND() as random FROM posts order by random LIMIT 0,1000;');
	PREPARE create_sql FROM @sql;
	EXECUTE create_sql;
	IF (SELECT COUNT(rpv_id) FROM `record_posts_views` WHERE rpv_id_user = user_id AND rpv_type = 'random') = 0 THEN
		INSERT INTO `record_posts_views` (rpv_time,rpv_id_user,rpv_name_posts_view,rpv_type) 
		VALUES (NOW(), user_id, @table_name,'random');
	ELSE
		UPDATE `record_posts_views` SET rpv_time = NOW() WHERE rpv_id_user = user_id AND rpv_type = 'random';
	END IF;
END@@

CALL CreateUserRandomPosts(2);

DROP PROCEDURE IF EXISTS GetUserRandomPosts;
DELIMITER @@
CREATE PROCEDURE GetUserRandomPosts(IN user_id INT)
BEGIN
	IF (SELECT COUNT(u_id) FROM users WHERE u_id = user_id) = 0 THEN
		SELECT 'no user found' AS result;
	ELSE
		IF (SELECT COUNT(rpv_id) FROM `record_posts_views` WHERE rpv_id_user = user_id AND rpv_type = 'random') = 0 THEN
			CALL CreateUserRandomPosts(user_id);
		ELSE
			IF (SELECT rpv_time FROM `record_posts_views` WHERE rpv_id_user = user_id AND rpv_type = 'random') < CURRENT_TIMESTAMP - INTERVAL 1 HOUR THEN
				CALL CreateUserRandomPosts(user_id);
			END IF;
		END IF;
		SET @table_name := CONCAT('user_',user_id,'_random_posts');
		SET @query := CONCAT('select * from ',@table_name);
		PREPARE stmt FROM @query;
		EXECUTE stmt;
	END IF;
END@@
DELIMITER ;

CALL GetUserRandomPosts(2);


## FOLLOW

DROP PROCEDURE IF EXISTS CreateUserFollowPosts;
DELIMITER @@
CREATE PROCEDURE CreateUserFollowPosts(IN user_id INT)
BEGIN
	SET @table_name := CONCAT('user_',user_id,'_follow_posts');
	SET @sql := CONCAT('DROP VIEW IF EXISTS ', @table_name);
	PREPARE delete_sql FROM @sql;
	EXECUTE delete_sql;
	SET @sql := CONCAT('CREATE VIEW ',@table_name,' AS
		SELECT p_id, p_publisher_id, p_publish_date, p_edit_date,
			p_edit_times, p_text_content, p_deleted, p_images_count,
			p_tags, p_upvotes, p_downvotes, p_repost, p_comments,
			p_visibility, p_reply, p_is_repost, p_id_reposter, p_id_origin_post
		FROM posts, users_follows
		WHERE p_publisher_id = uf_id_target AND uf_id_follower = ',user_id,' AND p_visibility = "all"
		UNION
		SELECT p_id, p_publisher_id, p_publish_date, p_edit_date,
			p_edit_times, p_text_content, p_deleted, p_images_count,
			p_tags, p_upvotes, p_downvotes, p_repost, p_comments,
			p_visibility, p_reply, p_is_repost, p_id_reposter, p_id_origin_post
		FROM posts, users_follows
		WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = ',user_id,'
		ORDER BY p_publish_date DESC
		LIMIT 0, 1000;'
	);
	PREPARE stmt FROM @sql;
	EXECUTE stmt;
	IF (SELECT COUNT(rpv_id) FROM record_posts_views WHERE rpv_id_user = user_id AND rpv_type = 'follow') = 0 THEN
		INSERT INTO record_posts_views (rpv_time,rpv_id_user,rpv_name_posts_view,rpv_type)
		VALUES (NOW(), user_id, @table_name, 'follow');
	ELSE
		UPDATE record_posts_views SET rpv_time = NOW() WHERE rpv_id_user = user_id AND rpv_type = 'follow';
	END IF;
END@@

DELIMITER ;

CALL CreateUserFollowPosts(2);


DROP PROCEDURE IF EXISTS GetUserFollowPosts;
DELIMITER @@
CREATE PROCEDURE GetUserFollowPosts(IN user_id INT)
BEGIN
	DECLARE EXIT HANDLER FOR 1146
	BEGIN
		CALL CreateUserFollowPosts(user_id);
		SET @table_name := CONCAT('user_',user_id,'_follow_posts');
		SET @query := CONCAT('select * from ',@table_name);
		PREPARE stmt FROM @query;
		EXECUTE stmt;
	END;
	IF (SELECT COUNT(u_id) FROM users WHERE u_id = user_id) = 0 THEN
		SELECT 'no user found' AS result;
	ELSE
		IF (SELECT COUNT(rpv_id) FROM record_posts_views WHERE rpv_id_user = user_id AND rpv_type = 'follow') = 0 THEN
			CALL CreateUserFollowPosts(user_id);
		ELSE
			IF (SELECT rpv_time FROM record_posts_views WHERE rpv_id_user = user_id AND rpv_type = 'follow') < CURRENT_TIMESTAMP - INTERVAL 1 HOUR THEN
				CALL CreateUserFollowPosts(user_id);
			END IF;
		END IF;
		SET @table_name := CONCAT('user_',user_id,'_follow_posts');
		SET @query := CONCAT('select * from ',@table_name);
		PREPARE stmt FROM @query;
		EXECUTE stmt;
	END IF;
END@@

DELIMITER ;

CALL GetUserFollowPosts(2);


DROP PROCEDURE IF EXISTS DeletePostView;
DELIMITER @@
CREATE PROCEDURE DeletePostView(IN view_name VARCHAR(100))
BEGIN
	SET @sql_delete := CONCAT('drop table if exists ', view_name);
	PREPARE stmt FROM @sql_delete;
	EXECUTE stmt;
END@@

DELIMITER ;


## Start Here

DROP PROCEDURE IF EXISTS GetPostsByTime;
DELIMITER @@
CREATE PROCEDURE GetPostsByTime(IN user_id INT,IN _offset INT,IN _length INT)
BEGIN
	IF user_id = 0 OR user_id = -1 THEN
		SELECT *
		FROM posts
		WHERE p_visibility = "all"
		LIMIT _offset, _length;
	ELSE 
		SELECT *
		FROM posts
		WHERE p_visibility = "all"
		UNION
		SELECT posts.*
		FROM posts, users_follows
		WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = user_id
		ORDER BY p_publish_date DESC
		LIMIT _offset, _length;
	END IF;
END@@

DELIMITER ;

CALL GetPostsByTime(1,0,10);


DROP PROCEDURE IF EXISTS GetPostsByScore;
DELIMITER @@
CREATE PROCEDURE GetPostsByScore(IN user_id INT,IN _offset INT,IN _length INT)
BEGIN
	IF user_id = 0 OR user_id = -1 THEN
		SELECT posts.*, p_upvotes - p_downvotes AS p_score
		FROM posts
		WHERE p_visibility = "all"
		ORDER BY p_score DESC
		LIMIT _offset, _length;
	ELSE
		SELECT posts.*, p_upvotes - p_downvotes AS p_score
		FROM posts
		WHERE p_visibility = "all"
		UNION
		SELECT posts.*, p_upvotes - p_downvotes AS p_score
		FROM posts, users_follows
		WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = user_id
		ORDER BY p_score DESC
		LIMIT _offset, _length;
	END IF;
END@@
DELIMITER ;

CALL GetPostsByScore(1,0,10);


DROP PROCEDURE IF EXISTS GetPostsByFollow;
DELIMITER @@
CREATE PROCEDURE GetPostsByFollow(IN user_id INT,IN _offset INT,IN _length INT)
BEGIN
	IF user_id = 0 OR user_id = -1 THEN
		SELECT 'error' AS 'error';
	ELSE
		SELECT posts.*
		FROM posts, users_follows
		WHERE p_publisher_id = uf_id_target AND uf_id_follower = user_id AND p_visibility = "all"
		UNION
		SELECT posts.*
		FROM posts, users_follows
		WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = user_id
		ORDER BY p_publish_date DESC
		LIMIT _offset, _length;
	END IF;
END@@
DELIMITER ;

CALL GetPostsByFollow(1,0,10);


DROP PROCEDURE IF EXISTS GetPostsByRandom;
DELIMITER @@
CREATE PROCEDURE GetPostsByRandom(IN user_id INT,IN _offset INT,IN _length INT,IN seed INT)
BEGIN
	IF user_id = 0 OR user_id = -1 THEN
		SELECT *
		FROM posts
		WHERE p_visibility = "all"
		ORDER BY RAND(seed)
		LIMIT _offset, _length;
	ELSE
		SELECT *
		FROM posts
		WHERE p_visibility = "all"
		UNION
		SELECT posts.*
		FROM posts, users_follows
		WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = user_id
		ORDER BY RAND(seed)
		LIMIT _offset, _length;
	END IF;
END@@
DELIMITER ;

CALL GetPostsByRandom(1,0,10,10);






