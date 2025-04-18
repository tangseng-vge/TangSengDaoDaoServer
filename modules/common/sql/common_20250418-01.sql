-- +migrate Up

ALTER TABLE `app_config` ADD COLUMN api_addr varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN api_addr_jw varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN web_addr varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN web_addr_jw varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN ws_addr varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN ws_addr_jw varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN wss_addr varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN wss_addr_jw varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN socket_addr varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
ALTER TABLE `app_config` ADD COLUMN socket_addr_jw varchar(64) not null DEFAULT "" COMMENT '是否能修改服务器地址';
