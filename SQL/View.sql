DROP VIEW IF EXISTS posts_view;
CREATE VIEW posts_view AS
SELECT
	# Post Info
	a.*,
	(SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_post = a.p_id AND pl_vote = 1) AS p_upvotes, 
	(SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_post = a.p_id AND pl_vote = 0) AS p_downvotes,
	(SELECT COUNT(c_id) FROM comments WHERE c_id_post = a.p_id) AS p_comments,
	(SELECT COUNT(b.p_id) FROM posts b WHERE b.p_is_repost = TRUE AND b.p_id_origin_post = a.p_id) AS p_reposts,
	# Publisher Info
	u.u_username,
	u.u_email,
	u.u_profileDescription AS u_profile,
	# Origin Publisher Info (If Reposted)
	r.u_id AS origin_user_id,
	r.u_username AS origin_user_username,
	r.u_email AS origin_user_email,
	r.u_profileDescription AS origin_user_profile,
	# Origin Post Info (If Reposted)
	p.p_publish_date AS origin_publish_date,
	p.p_edit_date AS origin_edit_date,
	p.p_edit_times AS origin_edit_times,
	p.p_text_content AS origin_text_content,
	p.p_deleted AS origin_deleted,
	p.p_images_count AS origin_images_count,
	p.p_tags AS origin_tags,
	p.p_visibility AS origin_visibility,
	p.p_reply AS origin_reply,
	p.p_is_repost AS origin_is_repost,
	p.p_id_origin_post AS origin_id_origin_post,
	(SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_post = p.p_id AND pl_vote = 1) AS origin_upvotes, 
	(SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_post = p.p_id AND pl_vote = 0) AS origin_downvotes,
	(SELECT COUNT(c_id) FROM comments WHERE c_id_post = p.p_id) AS origin_comments,
	(SELECT COUNT(c.p_id) FROM posts c WHERE c.p_is_repost = TRUE AND c.p_id_origin_post = p.p_id) AS origin_reposts
FROM posts a
LEFT JOIN users u ON u.u_id = a.p_publisher_id
LEFT JOIN posts p ON a.p_id_origin_post = p.p_id
LEFT JOIN users r ON r.u_id = p.p_publisher_id AND a.p_is_repost = TRUE;

SELECT * FROM posts_view;


SELECT * FROM post_likes 
WHERE pl_id_post = 1 AND pl_vote = 1





SELECT v.*,
	p_upvotes - p_downvotes AS p_score,
	CASE
		WHEN (SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_user = 1 AND pl_id_post = v.p_id) = 0 THEN -1
		ELSE (SELECT pl_vote FROM post_likes WHERE pl_id_user = 1 AND pl_id_post = v.p_id)
	END AS p_voted
FROM posts_view v
WHERE p_visibility = "all" 
UNION
SELECT v.*,
	p_upvotes - p_downvotes AS p_score,
	CASE
		WHEN (SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_user = 2 AND pl_id_post = v.p_id) = 0 THEN -1
		ELSE (SELECT pl_vote FROM post_likes WHERE pl_id_user = 2 AND pl_id_post = v.p_id)
	END AS p_voted
FROM posts_view v, users_follows
WHERE p_visibility = "follower" AND p_publisher_id = uf_id_target AND uf_id_follower = 2
ORDER BY p_publish_date DESC
LIMIT 0, 1000;






DROP VIEW IF EXISTS comments_view;
CREATE VIEW comments_view AS
SELECT
	a.*,	
	u.u_username,
	u.u_email,
	u.u_profileDescription AS u_profile,
	(SELECT COUNT(cl_id) FROM comment_likes WHERE cl_id_comment = a.c_id AND cl_vote = 1) AS c_upvotes, 
	(SELECT COUNT(cl_id) FROM comment_likes WHERE cl_id_comment = a.c_id AND cl_vote = 0) AS c_downvotes
FROM comments a
LEFT JOIN users u ON u.u_id = a.c_id_user;

SELECT * FROM comments_view;



#message type to be completed
DROP VIEW IF EXISTS collections_view;
CREATE VIEW collections_view AS
SELECT
	a.uc_id,
	a.uc_id_user,
	a.uc_id_target,
	a.uc_type,
	a.uc_time,
	p.p_publisher_id,
	u.u_username,
	p.p_text_content,
	p.p_images_count,
	p.p_is_repost
FROM user_collections a
LEFT JOIN posts p ON uc_type = 'POST' AND a.uc_id_target = p.p_id
LEFT JOIN users u ON p.p_publisher_id = u.u_id;



SELECT * FROM collections_view;


SELECT * FROM collections_view WHERE uc_id_user = 1 LIMIT 0,15;










DROP VIEW IF EXISTS messages_view;
CREATE VIEW messages_view AS
SELECT
	m.*,
	sender.*,
	IF(IFNULL(uf_id, 0) = 0, 0, 1) AS u_is_following
FROM messages m
LEFT JOIN users sender ON sender.u_id = m.m_sender
LEFT JOIN users_follows uf ON uf_id_target = m.m_receiver AND uf_id_follower = sender.u_id;


SELECT * FROM messages_view;



