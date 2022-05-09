
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