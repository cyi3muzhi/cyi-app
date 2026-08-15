-- DTM 分布式事务管理器初始化脚本
-- 注意：仅当 MySQL 数据卷首次初始化时由 docker-entrypoint-initdb.d 自动执行；
-- 对已有数据卷的环境，需手动在 MySQL 中执行本脚本。

-- 1. DTM 事务状态存储库（DTM 服务启动后自动在其中建表）
CREATE DATABASE IF NOT EXISTS `dtm` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 2. 业务屏障库：dtmcli 屏障表（保证分支幂等/空补偿/悬挂处理）
--    该库需能被业务服务（order）的 MySQL 账号访问，表结构来自 dtm 官方 sqls/dtmcli.barrier.mysql.sql
CREATE DATABASE IF NOT EXISTS `dtm_barrier` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS `dtm_barrier`.`barrier` (
  `id` bigint(22) PRIMARY KEY AUTO_INCREMENT,
  `trans_type` varchar(45) default '',
  `gid` varchar(128) default '',
  `branch_id` varchar(128) default '',
  `op` varchar(45) default '',
  `barrier_id` varchar(45) default '',
  `reason` varchar(45) default '' comment 'the branch type who insert this record',
  `create_time` datetime DEFAULT now(),
  `update_time` datetime DEFAULT now(),
  KEY(`create_time`),
  KEY(`update_time`),
  UNIQUE KEY(`gid`, `branch_id`, `op`, `barrier_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
