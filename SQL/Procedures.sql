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
	ORDER BY p_upvotes DESC, p_downvotes, p_reposts DESC
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


DROP PROCEDURE IF EXISTS GetPostsBySearch;
DELIMITER @@
CREATE PROCEDURE GetPostsBySearch(IN user_id INT, IN search_content VARCHAR(100), IN _offset INT, IN _length INT)
BEGIN
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v
	WHERE p_visibility = "all" AND INSTR(p_text_content, search_content) > 0
	UNION
	SELECT v.*,
		p_upvotes - p_downvotes AS p_score,
		HasVoted(user_id, v.p_id) AS p_voted,
		(SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = v.p_id) >= 1 OR 
		(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = v.p_id) >= 1 AS p_has_reposted,
		IF(v.p_is_repost = TRUE, v.origin_upvotes - v.origin_downvotes, NULL) AS origin_score,
		IF(v.p_is_repost = TRUE, HasVoted(user_id,v.p_id_origin_post), NULL) AS origin_voted
	FROM posts_view v, users_follows
	WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = user_id AND INSTR(p_text_content, search_content) > 0
	ORDER BY p_publish_date DESC
	LIMIT _offset, _length;
END@@
DELIMITER ;

CALL GetPostsBySearch(1,'Lorem',0,10);
CALL GetPostsBySearch(23,'uuyg',0,5);

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

CALL GetCommentsByTime(1, 228, 0, 50);



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


DROP PROCEDURE IF EXISTS DeletePost;
DELIMITER @@
CREATE PROCEDURE DeletePost(IN post_id INT)
BEGIN
	# (trigger 'before_post_delete' handles stuff before deleting)
	# delete all reposts
	DELETE FROM posts WHERE p_id_origin_post = post_id;
	# delete post
	DELETE FROM posts WHERE p_id = post_id;
END @@ 
DELIMITER ;

CALL DeletePost(1);

SELECT p_publisher_id FROM posts WHERE p_id = 222;






# messages
# messages_view


DROP PROCEDURE IF EXISTS GetMessagesByReceiverID;
DELIMITER @@
CREATE PROCEDURE GetMessagesByReceiverID(IN user_id INT, IN _offset INT, IN _length INT)
BEGIN
	SELECT 
		*
	FROM messages
	WHERE m_receiver = user_id OR (m_sender = user_id )
	LIMIT _offset, _length;
END @@
DELIMITER ;

CALL GetMessagesByReceiverID(2,0,10);


DROP PROCEDURE IF EXISTS GetMessagesByContact;
DELIMITER @@
CREATE PROCEDURE GetMessagesByContact(IN user_id INT, IN sender_id INT, IN _offset INT, IN _length INT)
BEGIN
	SELECT 
		*
	FROM messages
	WHERE (m_receiver = user_id AND m_sender = sender_id) 
		OR (m_receiver = sender_id AND m_sender = user_id)
	ORDER BY m_datetime DESC
	LIMIT _offset, _length;
END @@
DELIMITER ;

CALL GetMessagesByContact(2,1,0,10);


DROP PROCEDURE IF EXISTS FlagHasReceived;
DELIMITER @@
CREATE PROCEDURE FlagHasReceived(IN user_id INT, IN sender_id INT)
BEGIN
	UPDATE messages 
	SET m_has_received = TRUE
	WHERE m_sender = sender_id AND m_receiver = user_id;
END @@
DELIMITER ;


CALL FlagHasReceived(2,1);


DROP PROCEDURE IF EXISTS FlagUnread;
DELIMITER @@
CREATE PROCEDURE FlagUnread(IN user_id INT, IN sender_id INT)
BEGIN
	UPDATE messages 
	SET m_has_received = FALSE
	WHERE m_sender = sender_id AND m_receiver = user_id
	ORDER BY m_datetime DESC
	LIMIT 1;
END @@
DELIMITER ;


CALL FlagUnread(2,1);

DROP PROCEDURE IF EXISTS GetMessageContacts;
DELIMITER @@
CREATE PROCEDURE GetMessageContacts(IN user_id INT)
BEGIN
	SELECT 
		m_sender, u_username, 
		(SELECT a.m_text_content 
			FROM messages_view a 
			WHERE (m_receiver = user_id AND m_sender = b.m_sender) 
				OR (m_receiver = b.m_sender AND m_sender = user_id)
			ORDER BY a.m_datetime DESC
			LIMIT 1
		) AS m_text_content,
		(SELECT a.m_datetime 
			FROM messages_view a 
			WHERE (m_receiver = user_id AND m_sender = b.m_sender) 
				OR (m_receiver = b.m_sender AND m_sender = user_id)
			ORDER BY a.m_datetime DESC
			LIMIT 1
		) AS m_datetime,
		(SELECT COUNT(m_has_received) 
			FROM messages 
			WHERE m_sender = 1 AND m_has_received = 0
		) AS m_has_received
	FROM messages_view b
	WHERE m_receiver = user_id AND m_datetime > DATE(NOW() - INTERVAL 7 DAY)
	GROUP BY m_sender;
END @@
DELIMITER ;

CALL GetMessageContacts(2);

CALL FlagHasReceived(2,5);











