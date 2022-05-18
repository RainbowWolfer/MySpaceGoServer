DELIMITER $$

CREATE TRIGGER after_posts_view_delete
AFTER DELETE
ON record_posts_views FOR EACH ROW
BEGIN
	CALL DeletePostView(old.rpv_name_posts_view);
END$$   
DELIMITER ;


SET @@log_bin_trust_function_creators =1