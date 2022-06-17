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


SELECT * FROM posts_view;

## Start Here

DROP PROCEDURE IF EXISTS GetPostsByTime;
DELIMITER @@
CREATE PROCEDURE GetPostsByTime(IN user_id INT,IN _offset INT,IN _length INT)
BEGIN
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v
	WHERE p_visibility = "all"
	UNION
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v, users_follows
	WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = user_id
	ORDER BY p_publish_date DESC
	LIMIT _offset, _length;
END@@

DELIMITER ;

CALL GetPostsByTime(1,0,5);

SELECT p_id FROM posts_view WHERE origin_user_id = 1;
SELECT (SELECT COUNT(p_id) FROM posts_view WHERE origin_user_id = 1 AND p_id = 220) = 1 AS p_has_reposted;

DROP PROCEDURE IF EXISTS GetPostsByScore;
DELIMITER @@
CREATE PROCEDURE GetPostsByScore(IN user_id INT,IN _offset INT,IN _length INT)
BEGIN
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v
	WHERE p_visibility = "all" 
	UNION
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v, users_follows
	WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = user_id
	ORDER BY p_upvotes DESC
	LIMIT _offset, _length;
END@@
DELIMITER ;

CALL GetPostsByScore(-1,0,10);


DROP PROCEDURE IF EXISTS GetPostsByFollow;
DELIMITER @@
CREATE PROCEDURE GetPostsByFollow(IN user_id INT,IN _offset INT,IN _length INT)
BEGIN
	IF user_id <= 0 THEN
		SELECT 'error' AS 'error';
	ELSE
		SELECT v.*,
			p_upvotes - p_downvotes AS p_score,
			HasVoted(user_id, v.p_id) AS p_voted,
			(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
			(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
			IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
			IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
		FROM posts_view v, users_follows
		WHERE p_publisher_id = uf_id_target AND uf_id_follower = user_id AND p_visibility = "all" 
		UNION
		SELECT v.*,
			p_upvotes - p_downvotes AS p_score,
			HasVoted(user_id, v.p_id) AS p_voted,
			(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
			(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
			IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
			IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
		FROM posts_view v, users_follows
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
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v
	WHERE p_visibility = "all" 
	UNION
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v, users_follows
	WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = user_id
	ORDER BY RAND(seed)
	LIMIT _offset, _length;
END@@
DELIMITER ;

CALL GetPostsByRandom(1,0,10,10);


DROP PROCEDURE IF EXISTS GetPostsByTargetID;
DELIMITER @@
CREATE PROCEDURE GetPostsByTargetID(IN user_id INT, IN target_id INT, IN _offset INT, IN _length INT)
BEGIN
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v
	WHERE p_visibility = "all" AND p_publisher_id = target_id
	UNION
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v, users_follows
	WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = user_id AND p_publisher_id = target_id
	ORDER BY p_publish_date DESC
	LIMIT _offset, _length;
END@@
DELIMITER ;

CALL GetPostsByTargetID(1,2,10,10);


DROP PROCEDURE IF EXISTS GetPostByID;
DELIMITER @@
CREATE PROCEDURE GetPostByID(IN post_id INT, IN _email VARCHAR(40), IN _password VARCHAR(40))
BEGIN
	DECLARE user_id INT;
	SET user_id = (SELECT u_id FROM users WHERE u_email = _email AND u_password = _password);
	#select user_id, user_id <=> NULL;
	IF user_id <=> NULL THEN
		SELECT 'user not found';
	ELSE
		SELECT v.*,
			p_upvotes - p_downvotes AS p_score,
			HasVoted(user_id, v.p_id) AS p_voted,
			(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
			(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
			IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
			IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
		FROM posts_view v
		WHERE p_id = post_id;
	END IF;
END @@
DELIMITER ;

CALL GetPostByID(1,'2@test.com','123456');


DROP PROCEDURE IF EXISTS VotePost;
DELIMITER @@
CREATE PROCEDURE VotePost(IN user_id INT, IN post_id INT, IN cancel BOOL, IN score BOOL)
BEGIN
	IF cancel = TRUE THEN
		DELETE FROM post_likes WHERE pl_id_user = user_id AND pl_id_post = post_id;
		SELECT 'delete';
	ELSE
		IF (SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_user = user_id AND pl_id_post = post_id) = 0 THEN
			INSERT INTO post_likes (pl_id_user, pl_id_post, pl_vote, pl_datetime) VALUES (user_id,post_id,score,NOW());
			SELECT 'insert';
		ELSE
			UPDATE post_likes SET pl_vote = score WHERE pl_id_user = user_id AND pl_id_post = post_id;
			SELECT 'update';
		END IF;
	END IF;
END@@
DELIMITER ;

CALL VotePost(2, 4, FALSE, FALSE);


DROP PROCEDURE IF EXISTS VoteComment;
DELIMITER @@
CREATE PROCEDURE VoteComment(IN user_id INT, IN comment_id INT, IN cancel BOOL, IN score BOOL)
BEGIN
	IF cancel = TRUE THEN	
	DELETE FROM comment_likes WHERE cl_id_user = user_id AND cl_id_comment = comment_id;
		SELECT 'delete';
	ELSE
		IF (SELECT COUNT(cl_id) FROM comment_likes WHERE cl_id_user = user_id AND cl_id_comment = comment_id) = 0 THEN
			INSERT INTO comment_likes (cl_id_user, cl_id_comment, cl_vote, cl_datetime) VALUES (user_id, comment_id, score, NOW());
			SELECT 'insert';
		ELSE
			UPDATE comment_likes SET cl_vote = score WHERE cl_id_user = user_id AND cl_id_comment = comment_id;
			SELECT 'update';
		END IF;
	END IF;
END @@

DELIMITER ;

CALL VoteComment(2, 4, FALSE, FALSE);




DROP PROCEDURE IF EXISTS GetCommentsByTime;
DELIMITER @@
CREATE PROCEDURE GetCommentsByTime(IN user_id INT, IN post_id INT, IN _offset INT, IN _length INT)
BEGIN
	SELECT v.*,
		CASE
			WHEN user_id <= 0 THEN -1
			WHEN (SELECT COUNT(cl_id) FROM comment_likes WHERE cl_id_user = user_id AND cl_id_comment = c_id) = 0 THEN -1
			ELSE (SELECT cl_vote FROM comment_likes WHERE cl_id_user = user_id AND cl_id_comment = c_id)
		END AS c_voted
	FROM comments_view v
	WHERE v.c_id_post = post_id
	ORDER BY c_datetime DESC
	LIMIT _offset,_length;
END @@
DELIMITER ;

CALL GetCommentsByTime(1, 228, 0, s50);



DROP PROCEDURE IF EXISTS GetRepostRecords;
DELIMITER @@
CREATE PROCEDURE GetRepostRecords(IN post_id INT, IN _offset INT, IN _length INT)
BEGIN
	SELECT
		p_id,
		p_publisher_id,
		u_username,
		p_publish_date,
		p_text_content
	FROM posts_view
	WHERE p_id_origin_post = post_id;
END @@
DELIMITER ;

CALL GetRepostRecords(220, 0, 50);



DROP PROCEDURE IF EXISTS GetScoreRecords;
DELIMITER @@
CREATE PROCEDURE GetScoreRecords(IN post_id INT, IN _offset INT, IN _length INT)
BEGIN
	SELECT
		pl_id,
		pl_id_user,
		u_username,
		pl_datetime,
		pl_vote
	FROM post_likes
	LEFT JOIN users ON pl_id_user = u_id
	WHERE pl_id_post = post_id;
END @@
DELIMITER ;

CALL GetScoreRecords(229, 0, 50);



DROP PROCEDURE IF EXISTS GetUserByID;
DELIMITER @@
CREATE PROCEDURE GetUserByID(IN user_id INT, IN self_id INT)
BEGIN
	SELECT 
		u.*,
		IF(IFNULL(uf_id, 0) = 0, 0, 1) AS u_is_following
	FROM users u
	LEFT JOIN users_follows uf ON uf_id_target = user_id AND uf_id_follower = self_id
	WHERE u_id = user_id;
END @@
DELIMITER ;

CALL GetUserByID(3,1);


DROP PROCEDURE IF EXISTS GetUserByUsername;
DELIMITER @@
CREATE PROCEDURE GetUserByUsername(IN username VARCHAR(40), IN self_id INT)
BEGIN
	SELECT 
		u.*,
		IF(IFNULL(uf_id, 0) = 0, 0, 1) AS u_is_following
	FROM users u
	LEFT JOIN users_follows uf ON uf_id_target = u_id AND uf_id_follower = self_id
	WHERE u_username = username;
END @@
DELIMITER ;

CALL GetUserByUsername('mySpaceOfficial',2);


DROP PROCEDURE IF EXISTS GetUserByLogin;
DELIMITER @@
CREATE PROCEDURE GetUserByLogin(IN _email VARCHAR(40), IN _password VARCHAR(40))
BEGIN
	SELECT 
		u.*,
		FALSE AS u_is_following
	FROM users u
	WHERE u_email = _email AND u_password = _password;
END @@
DELIMITER ;

CALL GetUserByLogin('2@test2.com','123456');



DROP PROCEDURE IF EXISTS GetUserPostAndFollowersCount;
DELIMITER @@
CREATE PROCEDURE GetUserPostAndFollowersCount(IN user_id INT)
BEGIN
	SELECT 
		(SELECT COUNT(p_id) FROM posts WHERE p_publisher_id = user_id) AS posts_count,
		(SELECT COUNT(uf_id) FROM users_follows WHERE uf_id_target = user_id) AS followers_count;
END @@ 
DELIMITER ;

CALL GetUserPostAndFollowersCount(1312);


DROP PROCEDURE IF EXISTS GetUserFollowers;
DELIMITER @@
CREATE PROCEDURE GetUserFollowers(IN user_id INT, IN self_id INT)
BEGIN	
	SELECT
		u.*,
		IF(self_id > 0,
			IF(IFNULL((SELECT uf_id 
				FROM users_follows 
				WHERE uf_id_follower = self_id AND uf_id_target = u.u_id
			), 0) = 0, 0, 1)
		, 0) AS u_is_following
	FROM users_follows uf
	RIGHT JOIN users u ON u_id = uf.uf_id_follower
	WHERE uf_id_target = user_id;
END @@ 
DELIMITER ;

CALL GetUserFollowers(1, -1);









































