DROP VIEW IF EXISTS user_1_posts;
CREATE VIEW user_1_posts
AS 
SELECT 
p_id,p_publisher_id,p_publish_date,p_edit_date
,p_edit_times,p_text_content,p_deleted,p_images_count
,p_tags,p_upvotes,p_downvotes,p_repost,p_comments
,p_visibility,p_reply,p_is_repost,p_id_reposter,p_id_origin_post,
p_upvotes - p_downvotes AS p_score
FROM posts
LIMIT 1000;












