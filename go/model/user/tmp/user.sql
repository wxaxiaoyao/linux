
-- 用户表
DROP TABLE IF EXISTS t_user_info; 
CREATE TABLE `t_user_info` (
	  `uid` bigint(20) NOT NULL AUTO_INCREMENT PRIMARY KEY,
	  `password` varchar(33) NOT NULL DEFAULT '', 
	  `phonenum` varchar(16) NOT NULL UNIQUE DEFAULT '', 
	  `register_date` timestamp NOT NULL DEFAULT 0,
	  `update_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
ALTER TABLE `t_user_info` AUTO_INCREMENT = 100000;


