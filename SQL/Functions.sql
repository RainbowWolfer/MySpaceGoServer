SELECT @@log_bin_trust_function_creators;
SET GLOBAL log_bin_trust_function_creators=1;

#slow behavoir in select
DROP FUNCTION IF EXISTS HasReposted;
DELIMITER @@ 
CREATE FUNCTION HasReposted(user_id INT, post_id INT) RETURNS INT 
BEGIN
	RETURN (SELECT COUNT(c.p_id) FROM posts_view c WHERE c.origin_user_id = user_id AND c.p_id = post_id) >= 1 OR 
	(SELECT COUNT(c.p_id_origin_post) FROM posts_view c WHERE c.origin_user_id = 1 AND c.p_id_origin_post = post_id) >= 1;
END @@
DELIMITER ;

SELECT HasReposted(1, 220);

SELECT c.p_id FROM posts_view c WHERE c.origin_user_id = 1;
SELECT p_id_origin_post FROM posts_view WHERE origin_user_id = 1 AND p_id_origin_post=220;
SELECT * FROM posts_view;
CALL GetPostsByTime(1,0,5);


DROP FUNCTION IF EXISTS HasVoted;
DELIMITER @@
CREATE FUNCTION HasVoted(user_id INT, post_id INT) RETURNS INT 
BEGIN
	RETURN CASE
		WHEN (SELECT COUNT(pl_id) FROM post_likes WHERE pl_id_user = user_id AND pl_id_post = post_id) = 0 THEN -1
		ELSE (SELECT pl_vote FROM post_likes WHERE pl_id_user = user_id AND pl_id_post = post_id)
	END;
END @@
DELIMITER ;


SELECT HasVoted(1, 213);















