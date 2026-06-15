-- whimer.alloc_table definition

CREATE TABLE `alloc_table` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `biz_key` varchar(128) NOT NULL DEFAULT '' COMMENT 'biz key identifier',
  `cur_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT 'max id',
  `step` int unsigned NOT NULL DEFAULT '0' COMMENT 'step',
  `created_at` bigint NOT NULL DEFAULT '0' COMMENT 'created unix ms',
  `updated_at` bigint NOT NULL DEFAULT '0' COMMENT 'updated unix ms',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key` (`biz_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='segment allocation table';


-- whimer.chat definition

CREATE TABLE `chat` (
  `id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '会话id 主键',
  `type` tinyint NOT NULL DEFAULT '0' COMMENT '会话类型 1-单聊 2-群聊',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '会话名称',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '会话状态 0-正常',
  `creator` bigint NOT NULL DEFAULT '0' COMMENT '创建者id',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间',
  `last_msg_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '最后一条消息id',
  `settings` bigint NOT NULL DEFAULT '0' COMMENT '会话设置',
  PRIMARY KEY (`id`),
  KEY `idx_creator` (`creator`),
  KEY `idx_last_msg_id` (`last_msg_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='会话表';


-- whimer.chat_inbox definition

CREATE TABLE `chat_inbox` (
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '用户id',
  `chat_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '会话id',
  `last_msg_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '会话最后一条消息id',
  `last_read_msg_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '最后已读消息id',
  `last_read_time` bigint NOT NULL DEFAULT '0' COMMENT '最后已读时间',
  `unread_count` bigint NOT NULL DEFAULT '0' COMMENT '会话未读数',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '收件箱状态 正常还是删除等',
  `is_pinned` tinyint NOT NULL DEFAULT '0' COMMENT '是否置顶',
  PRIMARY KEY (`uid`,`chat_id`),
  KEY `idx_uid_ispinned_mtime_status` (`uid`,`is_pinned`,`mtime`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户收件箱表';


-- whimer.chat_member_p2p definition

CREATE TABLE `chat_member_p2p` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `chat_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '会话id',
  `uid_a` bigint NOT NULL DEFAULT '0' COMMENT '用户a uid小',
  `uid_b` bigint NOT NULL DEFAULT '0' COMMENT '用户b uid大',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_uk_uid_a_uid_b` (`uid_a`,`uid_b`),
  KEY `idx_chat_id_uid_a` (`chat_id`,`uid_a`),
  KEY `idx_chat_id_uid_b` (`chat_id`,`uid_b`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户收件箱表';


-- whimer.chat_message definition

CREATE TABLE `chat_message` (
  `chat_id` binary(16) NOT NULL DEFAULT '0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '会话id',
  `msg_id` binary(16) NOT NULL DEFAULT '0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '消息id',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  `pos` bigint NOT NULL DEFAULT '0' COMMENT '消息在会话所处位置',
  PRIMARY KEY (`chat_id`,`msg_id`),
  KEY `idx_chat_id_pos` (`chat_id`,`pos`),
  KEY `idx_pos` (`pos`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='会话-消息关联表';


-- whimer.comment definition

CREATE TABLE `comment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `oid` bigint NOT NULL DEFAULT '0' COMMENT '评论对象ID',
  `type` tinyint NOT NULL DEFAULT '0' COMMENT '评论类型，包括文本，或者图文类型',
  `content` text COMMENT '评论内容',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '评论发布者',
  `root` bigint DEFAULT '0' COMMENT '主评论ID，如果该评论为主评论，该值为0',
  `parent` bigint DEFAULT '0' COMMENT '回复的评论ID，如果该评论不是回复评论，该值为0',
  `ruid` bigint NOT NULL DEFAULT '0' COMMENT '被回复的用户ID',
  `state` tinyint NOT NULL DEFAULT '0' COMMENT '评论状态，表示评论的状态，是否合规等状态',
  `like` int NOT NULL DEFAULT '0' COMMENT '点赞数',
  `dislike` int NOT NULL DEFAULT '0' COMMENT '点踩数',
  `report` int NOT NULL DEFAULT '0' COMMENT '被举报数',
  `pin` tinyint NOT NULL DEFAULT '0' COMMENT '是否置顶',
  `ip` varbinary(16) NOT NULL DEFAULT '\0' COMMENT '发布时ip',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '评论创建时间，Unix时间戳',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '评论修改时间，Unix时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_oid` (`oid`),
  KEY `idx_ruid` (`ruid`),
  KEY `idx_uid` (`uid`),
  KEY `idx_ctime` (`ctime`),
  KEY `idx_root_parent` (`root`,`parent`),
  KEY `idx_oid_uid` (`oid`,`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='评论表';


-- whimer.comment_asset definition

CREATE TABLE `comment_asset` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `comment_id` bigint NOT NULL DEFAULT '0' COMMENT '所属评论id',
  `type` tinyint NOT NULL DEFAULT '0' COMMENT '资源类型 1-图片; 2-自定义表情; ',
  `store_key` varchar(255) NOT NULL DEFAULT '0' COMMENT '资源存储key',
  `metadata` varchar(512) NOT NULL DEFAULT '' COMMENT '资源的元数据',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_comment_id` (`comment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='评论资源';


-- whimer.comment_ext definition

CREATE TABLE `comment_ext` (
  `comment_id` bigint NOT NULL DEFAULT '0' COMMENT '对应评论id',
  `at_users` text NOT NULL COMMENT 'at用户 json格式存储 [{"nickname":"xxx","uid":yyy}]',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间，Unix时间戳',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间，Unix时间戳',
  PRIMARY KEY (`comment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='评论额外信息';


-- whimer.conductor_namespace definition

CREATE TABLE `conductor_namespace` (
  `id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '主键id',
  `name` varchar(32) NOT NULL DEFAULT '' COMMENT '命名空间名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_uk` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='命名空间表';


-- whimer.conductor_task definition

CREATE TABLE `conductor_task` (
  `id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '主键id',
  `namespace` varchar(32) NOT NULL DEFAULT '' COMMENT '命名空间',
  `task_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '任务类型（自定义任务名）',
  `task_type_shard` int NOT NULL DEFAULT '0' COMMENT '任务类型所属分片',
  `input_args` blob COMMENT '任务输入参数',
  `output_args` blob COMMENT '任务输出结果',
  `callback_url` varchar(255) NOT NULL DEFAULT '' COMMENT '任务回调地址',
  `state` varchar(16) NOT NULL DEFAULT '' COMMENT '任务当前状态 inited|dispatched|running|success|failure|aborted 等',
  `trace_id` varchar(255) NOT NULL DEFAULT '0' COMMENT '请求链路追踪id',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  `utime` bigint NOT NULL DEFAULT '0' COMMENT '更新时间',
  `max_retry_cnt` bigint NOT NULL DEFAULT '0' COMMENT '最大重试次数, -1表示无限重试直到超时, 0表示重试',
  `expire_time` bigint NOT NULL DEFAULT '0' COMMENT '任务过期时间单位unix ms时间戳',
  `settings` text NOT NULL COMMENT '任务额外设置 可选参数',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '版本',
  `cur_retry_cnt` bigint NOT NULL DEFAULT '0' COMMENT '当前已重试次数',
  PRIMARY KEY (`id`),
  KEY `idx_namespace_task_type` (`namespace`,`task_type`),
  KEY `idx_task_type_shard` (`task_type_shard`),
  KEY `idx_ctime` (`ctime`),
  KEY `idx_expire_time` (`expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='任务表';


-- whimer.conductor_task_history definition

CREATE TABLE `conductor_task_history` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `task_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '任务id',
  `state` varchar(16) NOT NULL DEFAULT '' COMMENT '状态快照',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='任务历史表';


-- whimer.counter_record definition

CREATE TABLE `counter_record` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `biz_code` int NOT NULL DEFAULT '0' COMMENT '所属业务',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '用户id 哪个用户执行的这条记录',
  `oid` bigint NOT NULL DEFAULT '0' COMMENT '对象id 哪个对象被操作',
  `act` tinyint NOT NULL DEFAULT '0' COMMENT '执行计数操作：1；取消计数操作：2',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间，Unix时间戳',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间，Unix时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid` (`uid`,`oid`,`biz_code`) COMMENT '唯一键',
  KEY `idx_ctime` (`ctime`),
  KEY `idx_mtime` (`mtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='计数记录表';


-- whimer.counter_summary definition

CREATE TABLE `counter_summary` (
  `biz_code` int NOT NULL DEFAULT '0' COMMENT '所属业务',
  `oid` bigint NOT NULL DEFAULT '0' COMMENT '对象id 哪个对象被操作',
  `cnt` bigint NOT NULL DEFAULT '0' COMMENT '数量',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间，Unix时间戳',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间，Unix时间戳',
  PRIMARY KEY (`oid`,`biz_code`) COMMENT '联合主键'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='计数统计表';


-- whimer.message definition

CREATE TABLE `message` (
  `id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '消息id 主键',
  `type` smallint NOT NULL DEFAULT '0' COMMENT '消息类型',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '消息状态 0-正常',
  `sender` bigint NOT NULL DEFAULT '0' COMMENT '消息发送者uid',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间',
  `content` text NOT NULL COMMENT '消息内容',
  `ext` tinyint NOT NULL DEFAULT '0' COMMENT '扩展相关',
  `cid` varchar(32) NOT NULL DEFAULT '' COMMENT '客户端侧消息id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_uk_cid` (`cid`),
  KEY `idx_sender` (`sender`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='消息表';


-- whimer.message_ext definition

CREATE TABLE `message_ext` (
  `msg_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '消息id',
  `recall` varbinary(255) NOT NULL DEFAULT '' COMMENT '消息撤回相关记录',
  PRIMARY KEY (`msg_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='消息扩展表';


-- whimer.note definition

CREATE TABLE `note` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` varchar(48) NOT NULL DEFAULT '' COMMENT '标题',
  `desc` varchar(2048) NOT NULL DEFAULT '' COMMENT '描述',
  `privacy` tinyint NOT NULL DEFAULT '0' COMMENT '公开类型',
  `owner` bigint NOT NULL DEFAULT '0' COMMENT '笔记作者',
  `create_at` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_at` bigint NOT NULL DEFAULT '0' COMMENT '更新时间',
  `ip` varbinary(16) NOT NULL DEFAULT '\0' COMMENT '发布/修改时ip',
  `note_type` tinyint NOT NULL DEFAULT '0' COMMENT '笔记类型',
  `state` tinyint NOT NULL DEFAULT '0' COMMENT '笔记状态',
  `modify_at` bigint NOT NULL DEFAULT '0' COMMENT '记录更新时间, 和笔记更新时间update_at区分开',
  PRIMARY KEY (`id`),
  KEY `idx_owner` (`owner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='笔记表';


-- whimer.note_asset definition

CREATE TABLE `note_asset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `asset_key` varchar(255) NOT NULL DEFAULT '' COMMENT '资源key',
  `asset_type` tinyint NOT NULL DEFAULT '0' COMMENT '资源类型',
  `note_id` bigint NOT NULL DEFAULT '0' COMMENT '所属笔记id',
  `create_at` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  `asset_meta` varbinary(512) NOT NULL DEFAULT '' COMMENT '资源的元数据',
  PRIMARY KEY (`id`),
  KEY `idx_note_id` (`note_id`),
  KEY `idx_asset_key` (`asset_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='笔记资源表';


-- whimer.note_ext definition

CREATE TABLE `note_ext` (
  `note_id` bigint NOT NULL DEFAULT '0' COMMENT '笔记id',
  `tags` varchar(255) NOT NULL DEFAULT '' COMMENT '标签id，比如1,2,3,10',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间 unix second',
  `utime` bigint NOT NULL DEFAULT '0' COMMENT '更新时间 unix second',
  `at_users` text NOT NULL COMMENT 'at用户 json格式存储 [{"nickname":"xxx","uid":yyy}]',
  PRIMARY KEY (`note_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='笔记额外信息表';


-- whimer.note_procedure_record definition

CREATE TABLE `note_procedure_record` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `note_id` bigint NOT NULL DEFAULT '0' COMMENT '笔记id',
  `protype` varchar(16) NOT NULL DEFAULT '' COMMENT '处理流程类型',
  `task_id` varchar(64) NOT NULL DEFAULT '' COMMENT '处理流程id',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '是否完成 0-处理中; 1-成功; 2-失败',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  `utime` bigint NOT NULL DEFAULT '0' COMMENT '更新时间',
  `cur_retry` int NOT NULL DEFAULT '0' COMMENT '当前重试次数',
  `max_retry_cnt` int NOT NULL DEFAULT '0' COMMENT '最大重试次数',
  `next_check_time` bigint NOT NULL DEFAULT '0' COMMENT '下次检查重试时间 单位unix sec',
  `params` varbinary(2048) NOT NULL DEFAULT '' COMMENT '调用参数',
  `expired_time` bigint NOT NULL DEFAULT '0' COMMENT '过期时间戳 unix sec',
  PRIMARY KEY (`id`),
  UNIQUE KEY `note_id` (`note_id`,`protype`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_next_check_time` (`next_check_time`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='笔记资源处理记录表';


-- whimer.relation definition

CREATE TABLE `relation` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `alpha` bigint NOT NULL DEFAULT '0' COMMENT '用户A的id',
  `beta` bigint NOT NULL DEFAULT '0' COMMENT '用户B的id',
  `link` tinyint NOT NULL DEFAULT '0' COMMENT '用户的关注关系 0：两者没有关系；2：A关注B；-2：B关注A；-4：AB相互关注',
  `actime` bigint NOT NULL DEFAULT '0' COMMENT 'A首次关注B的时间，Unix时间戳',
  `bctime` bigint NOT NULL DEFAULT '0' COMMENT 'B首次关注A的时间，Unix时间戳',
  `amtime` bigint NOT NULL DEFAULT '0' COMMENT 'A改变对B的关注状态的时间，Unix时间戳',
  `bmtime` bigint NOT NULL DEFAULT '0' COMMENT 'B改变对A的关注状态的时间，Unix时间戳',
  PRIMARY KEY (`id`) COMMENT '主键',
  UNIQUE KEY `uk_alpha_beta` (`alpha`,`beta`) COMMENT '唯一索引',
  KEY `idx_alpha` (`alpha`),
  KEY `idx_beta` (`beta`),
  KEY `idx_alpha_link_amtime` (`alpha`,`link`,`amtime`),
  KEY `idx_beta_link_bmtime` (`beta`,`link`,`bmtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户关系表';


-- whimer.relation_setting definition

CREATE TABLE `relation_setting` (
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '用户id',
  `settings` text NOT NULL COMMENT '配置json',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间，Unix时间戳',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间，Unix时间戳',
  PRIMARY KEY (`uid`) COMMENT '主键'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户关注设置表';


-- whimer.system_chat definition

CREATE TABLE `system_chat` (
  `id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '主键 uuidv7',
  `type` tinyint NOT NULL DEFAULT '0' COMMENT '系统会话类型：比如@我、收到的赞等',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '用户id',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间 unix ms',
  `last_msg_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '最后一条消息id',
  `last_read_msg_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '最后阅读的消息id',
  `unread_count` bigint NOT NULL DEFAULT '0' COMMENT '未读数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_type` (`uid`,`type`),
  KEY `idx_last_read_msg_id` (`last_read_msg_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT=' 系统消息会话表';


-- whimer.system_message definition

CREATE TABLE `system_message` (
  `id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '消息id 主键uuidv7',
  `system_chat_id` binary(16) NOT NULL DEFAULT '\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0' COMMENT '系统消息会话id',
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '触发系统消息的用户uid',
  `recv_uid` bigint NOT NULL DEFAULT '0' COMMENT '消息接收者 uid',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '消息状态',
  `msg_type` tinyint NOT NULL DEFAULT '0' COMMENT '消息内容类型',
  `content` text COMMENT '消息内容',
  `mtime` bigint NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_recv_uid` (`recv_uid`),
  KEY `idx_system_chat_id_recv_uid` (`system_chat_id`,`recv_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统消息表';


-- whimer.tag definition

CREATE TABLE `tag` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '标签名',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间 unix second',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='标签表';


-- whimer.`user` definition

CREATE TABLE `user` (
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '用户uid',
  `nickname` varchar(48) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户昵称',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '用户头像',
  `style_sign` varchar(128) NOT NULL DEFAULT '' COMMENT '用户个性签名',
  `gender` tinyint NOT NULL DEFAULT '0' COMMENT '性别',
  `tel` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '用户手机',
  `email` varchar(128) NOT NULL DEFAULT '' COMMENT '用户邮箱',
  `pass` varchar(256) NOT NULL DEFAULT '' COMMENT '加密后密码',
  `salt` varchar(256) NOT NULL DEFAULT '' COMMENT '盐',
  `create_at` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_at` bigint NOT NULL DEFAULT '0' COMMENT '更新时间',
  `status` tinyint DEFAULT '0' COMMENT '用户状态',
  PRIMARY KEY (`uid`),
  UNIQUE KEY `uk_nickname` (`nickname`),
  UNIQUE KEY `uk_tel` (`tel`),
  FULLTEXT KEY `ftidx_nickname` (`nickname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户基础信息表';


-- whimer.user_setting definition

CREATE TABLE `user_setting` (
  `uid` bigint NOT NULL DEFAULT '0' COMMENT '用户uid',
  `flags` bigint NOT NULL DEFAULT '0' COMMENT '用户设置, bitmap表示',
  `ext` text COMMENT '额外字段 json格式存储',
  `ctime` bigint NOT NULL DEFAULT '0' COMMENT '创建时间',
  `utime` bigint NOT NULL DEFAULT '0' COMMENT '修改时间',
  PRIMARY KEY (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户设置表';