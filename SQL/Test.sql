DELETE FROM `email_validations` 
			WHERE ev_datetime < CURRENT_TIMESTAMP - INTERVAL 1 HOUR;
			
			
INSERT INTO email_validations (ev_email, ev_code, ev_datetime) VALUES ('1', '2', NOW())


INSERT INTO email_validations (ev_email, ev_code, ev_datetime) VALUES ('1519787190@qq.com', 'rainbowwolfer123456789:rw1519787190@qq.com��ُ ', NOW())


SELECT u_id FROM users WHERE u_email = '1519787190@qq.com'

INSERT INTO email_validations (ev_email, ev_code, ev_datetime) VALUES ('1519787190@qq.com', '31265967ffaf9d7a03fcfd568357784f', NOW())


SELECT ev_code, ev_email, ev_username, ev_password FROM email_validations WHERE ev_email = '1519787190@qq.com' AND ev_datetime <= NOW()

DELETE FROM email_validations
	WHERE ev_email = 'rw@qq.com'
	
	
	
	
SELECT * FROM posts ORDER BY p_publish_date DESC;
	


INSERT INTO users (u_username, u_password, u_email) VALUES ();

UPDATE users SET u_username = 'btest' WHERE u_id = '1' AND u_username = 'forcingsmile' AND u_password = '123456789:rw';


SELECT u_id FROM users WHERE u_username = 'bt2est'

SELECT ev_id FROM email_validations WHERE ev_email = '%s' AND ev_password = '%s'



#get followers
SELECT * FROM users,  users_follows
WHERE u_id = uf_id_follower AND uf_id_target = 1;

#get followings
SELECT * FROM users,  users_follows
WHERE u_id = uf_id_target AND uf_id_follower = 2;


SELECT p_id, p_publisher_id, p_publish_date, p_edit_date,
	p_edit_times, p_text_content, p_deleted, p_images_count,
	p_tags, p_upvotes, p_downvotes, p_repost, p_comments,
	p_visibility, p_reply, p_is_repost, p_id_reposter, p_id_origin_post
FROM posts, users_follows
WHERE p_publisher_id = uf_id_target AND uf_id_follower = 2 AND p_visibility = 'all'
UNION
SELECT p_id, p_publisher_id, p_publish_date, p_edit_date,
	p_edit_times, p_text_content, p_deleted, p_images_count,
	p_tags, p_upvotes, p_downvotes, p_repost, p_comments,
	p_visibility, p_reply, p_is_repost, p_id_reposter, p_id_origin_post
FROM posts, users_follows
WHERE p_visibility = 'follower' AND p_publisher_id = uf_id_target AND uf_id_follower = 2
ORDER BY p_publish_date DESC
LIMIT 0, 1000;


SELECT RAND()



INSERT INTO comments (c_id_user, c_id_post, c_text_content, c_datetime) VALUES ();


SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_post = 1 AND pl_vote = -1;

SELECT COUNT(p_id) FROM posts WHERE p_is_repost = TRUE AND p_id_origin_post = 212

SELECT pl_vote FROM post_likes WHERE pl_id_user = 1 AND pl_id_post = 1;

SELECT pl_id FROM post_likes WHERE pl_id_user = 1 AND pl_id_post = 112

SELECT
	a.*,
	(SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_post = a.p_id AND pl_vote = 1) AS p_upvotes, 
	(SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_post = a.p_id AND pl_vote = 0) AS p_downvotes,
	(SELECT COUNT(c_id) FROM comments WHERE c_id_post = a.p_id) AS p_comments,
	(SELECT COUNT(b.p_id) FROM posts b WHERE b.p_is_repost = TRUE AND b.p_id_origin_post = a.p_id) AS p_reposts,
	CASE
		WHEN (SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_user = 1 AND pl_id_post = a.p_id) = 0 THEN -1
		ELSE (SELECT pl_vote FROM post_likes WHERE pl_id_user = 1 AND pl_id_post = a.p_id)
	END AS p_voted,
	u.u_id,
	u.u_username,
	u.u_email,
	u.u_profileDescription,
	r.u_id AS reposter_id,
	r.u_username AS reposter_username,
	r.u_email AS reposter_email,
	r.u_profileDescription AS reposter_profileDescription,
	p.*
FROM posts a
LEFT JOIN users u ON u.u_id = a.p_publisher_id
LEFT JOIN posts p ON a.p_id_origin_post = p.p_id
LEFT JOIN users r ON r.u_id = p.p_publisher_id AND a.p_is_repost = TRUE
WHERE a.p_visibility = "all"
ORDER BY a.p_publish_date DESC
LIMIT 0, 1000;


#repost
INSERT INTO posts 
()
VALUES
();

INSERT INTO posts 
	(p_publisher_id, p_publish_date, p_edit_date, p_text_content, 
	p_visibility, p_reply, p_images_count, p_tags, p_is_repost, p_id_origin_post) 
VALUES 
	('%s',NOW(),NOW(),'%s','%s','%s','%d','%s',TRUE,'%s')























