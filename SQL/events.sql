DELIMITER $$

-- SET GLOBAL event_scheduler = ON$$     -- required for event to execute but not create    

CREATE	/*[DEFINER = { user | CURRENT_USER }]*/	EVENT `wjx`.`auto_delete_email_validations`

ON SCHEDULE
	 /* uncomment the example below you want to use */

	-- scheduleexample 1: test once

	   --  AT 'YYYY-MM-DD HH:MM.SS'/CURRENT_TIMESTAMP { + INTERVAL 1 [HOUR|MONTH|WEEK|DAY|MINUTE|...] }

	-- scheduleexample 2: test at intervals forever after creation

	   EVERY 1 HOUR

	-- scheduleexample 3: specified start time, end time and interval for execution
	   /*EVERY 1  [HOUR|MONTH|WEEK|DAY|MINUTE|...]

	   STARTS CURRENT_TIMESTAMP/'YYYY-MM-DD HH:MM.SS' { + INTERVAL 1[HOUR|MONTH|WEEK|DAY|MINUTE|...] }

	   ENDS CURRENT_TIMESTAMP/'YYYY-MM-DD HH:MM.SS' { + INTERVAL 1 [HOUR|MONTH|WEEK|DAY|MINUTE|...] } */

/*[ON COMPLETION [NOT] PRESERVE]
[ENABLE | DISABLE]
[COMMENT 'comment']*/

DO
	BEGIN
		DELETE FROM `email_validations` 
			WHERE ev_datetime < CURRENT_TIMESTAMP - INTERVAL 1 HOUR;
	END$$

DELIMITER ;





DELIMITER $$

CREATE EVENT `wjx`.`auto_delete_email_password_resets`

ON SCHEDULE
	   EVERY 1 HOUR

DO
	BEGIN
		DELETE FROM `email_password_resets` 
			WHERE epr_datetime < CURRENT_TIMESTAMP - INTERVAL 1 HOUR;
	END$$

DELIMITER ;