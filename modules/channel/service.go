package channel

import (
	chservice "github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/channel/service"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/config"
)

type service struct {
	ctx              *config.Context
	channelSettingDB *channelSettingDB
}

func NewService(ctx *config.Context) chservice.IService {
	return &service{
		ctx:              ctx,
		channelSettingDB: newChannelSettingDB(ctx),
	}
}

func (s *service) GetChannelSettings(channelIDs []string) ([]*chservice.ChannelSettingResp, error) {
	if len(channelIDs) == 0 {
		return nil, nil
	}
	channelSettingModels, err := s.channelSettingDB.queryWithChannelIDs(channelIDs)
	if err != nil {
		return nil, err
	}
	channelSettingResps := make([]*chservice.ChannelSettingResp, 0, len(channelSettingModels))
	if len(channelSettingModels) > 0 {
		for _, channelSettingM := range channelSettingModels {
			channelSettingResps = append(channelSettingResps, newChannelSettingResp(channelSettingM))
		}
	}
	return channelSettingResps, nil
}

func (s *service) CreateOrUpdateMsgAutoDelete(channelID string, channelType uint8, msgAutoDelete int64) error {
	return s.channelSettingDB.insertOrAddMsgAutoDelete(channelID, channelType, msgAutoDelete)
}

func newChannelSettingResp(m *channelSettingModel) *chservice.ChannelSettingResp {

	return &chservice.ChannelSettingResp{
		ChannelID:         m.ChannelID,
		ChannelType:       m.ChannelType,
		ParentChannelID:   m.ParentChannelID,
		ParentChannelType: m.ParentChannelType,
		OffsetMessageSeq:  m.OffsetMessageSeq,
	}
}
