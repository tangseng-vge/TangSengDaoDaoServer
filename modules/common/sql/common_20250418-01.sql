-- +migrate Up

ALTER TABLE `app_config` ADD COLUMN api_addr smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN api_addr_jw smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN web_addr smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN web_addr_jw smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN ws_addr smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN ws_addr_jw smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN wss_addr smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN wss_addr_jw smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN socket_addr smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN socket_addr_jw smallint not null DEFAULT 0 COMMENT '是否能修改服务器地址';
