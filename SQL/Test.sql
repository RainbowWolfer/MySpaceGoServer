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
select * from users,  users_follows
where u_id = uf_id_follower and uf_id_target = 1;

#get followings
SELECT * FROM users,  users_follows
WHERE u_id = uf_id_target AND uf_id_follower = 2;


select p_id, p_publisher_id, p_publish_date, p_edit_date,
	p_edit_times, p_text_content, p_deleted, p_images_count,
	p_tags, p_upvotes, p_downvotes, p_repost, p_comments,
	p_visibility, p_reply, p_is_repost, p_id_reposter, p_id_origin_post
from posts, users_follows
where p_publisher_id = uf_id_target AND uf_id_follower = 2 and p_visibility = 'all'
union
select p_id, p_publisher_id, p_publish_date, p_edit_date,
	p_edit_times, p_text_content, p_deleted, p_images_count,
	p_tags, p_upvotes, p_downvotes, p_repost, p_comments,
	p_visibility, p_reply, p_is_repost, p_id_reposter, p_id_origin_post
from posts, users_follows
where p_visibility = 'follower' and p_publisher_id = uf_id_target and uf_id_follower = 2
ORDER BY p_publish_date DESC
LIMIT 0, 1000;


select rand()






