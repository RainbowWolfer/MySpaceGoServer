DROP TRIGGER IF EXISTS before_post_delete;
DELIMITER $$
CREATE TRIGGER before_post_delete
BEFORE DELETE
ON posts FOR EACH ROW
BEGIN
	#delete all comments
	DELETE FROM comments WHERE c_id_post = old.p_id;
	#delete all votes
	DELETE FROM post_likes WHERE pl_id_post = old.p_id;
END$$
DELIMITER ;