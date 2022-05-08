DELETE FROM `email_validations` 
			WHERE ev_datetime < CURRENT_TIMESTAMP - INTERVAL 1 HOUR;
			
			
INSERT INTO email_validations (ev_email,ev_code,ev_datetime) VALUES ('1','2',NOW())


INSERT INTO email_validations (ev_email,ev_code,ev_datetime) VALUES ('1519787190@qq.com','rainbowwolfer123456789:rw1519787190@qq.com��ُ ',NOW())


SELECT u_id FROM users WHERE u_email = '1519787190@qq.com'

INSERT INTO email_validations (ev_email,ev_code,ev_datetime) VALUES ('1519787190@qq.com','31265967ffaf9d7a03fcfd568357784f',NOW())


SELECT ev_code,ev_email,ev_username,ev_password FROM email_validations WHERE ev_email = '1519787190@qq.com' AND ev_datetime <= NOW()

DELETE FROM email_validations
	WHERE ev_email = 'rw@qq.com'


insert into users (u_username,u_password,u_email) values ();

update users set u_username = 'btest' where u_id = '1' and u_username = 'forcingsmile' and u_password = '123456789:rw';


select u_id from users where u_username = 'bt2est'

select ev_id from email_validations where ev_email = '%s' and ev_password = '%s'

