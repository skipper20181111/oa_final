package coupon

import (
	"context"
	"encoding/json"
	"oa_final/cachemodel"
	"oa_final/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"oa_final/internal/svc"
)

type MissionUtilLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	phone       string
	MissionList []*types.Mission
}

func NewMissionUtilLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MissionUtilLogic {
	MissionList := make([]*types.Mission, 0)
	get, ok := svcCtx.LocalCache.Get(svc.MissionListKey)
	if ok {
		MissionList = get.([]*types.Mission)
	}
	return &MissionUtilLogic{
		Logger:      logx.WithContext(ctx),
		ctx:         ctx,
		svcCtx:      svcCtx,
		phone:       ctx.Value("phone").(string),
		MissionList: MissionList,
	}
}

func (l *MissionUtilLogic) Finishmission(missionid int64) (resp *types.GetMissionInfoResp, err error) {
	MissionInfos := make([]*types.MissionInfo, 0)
	phone, _ := l.svcCtx.UserMission.FindOneByPhone(l.ctx, l.phone)
	if phone != nil {
		json.Unmarshal([]byte(phone.MissionInfo), &MissionInfos)
		for _, missioninfo := range MissionInfos {
			if missioninfo.Mission.MissionId == missionid {
				missioninfo.Accomplished = true
			}
		}
		marshal, _ := json.Marshal(MissionInfos)
		phone.MissionInfo = string(marshal)
		l.svcCtx.UserMission.Update(l.ctx, phone)
	}
	return l.Getmissioninfo()
}

func (l *MissionUtilLogic) Getmissioninfo() (resp *types.GetMissionInfoResp, err error) {
	GetMissionInfoRp := &types.GetMissionInfoRp{}
	count, _ := l.svcCtx.UserOrder.CountByPhone(l.ctx, l.phone)
	GetMissionInfoRp.ConsumeTime = count
	MissionInfos := make([]*types.MissionInfo, 0)
	phone, _ := l.svcCtx.UserMission.FindOneByPhone(l.ctx, l.phone)
	if phone == nil {
		for _, mission := range l.MissionList {
			MissionInfos = append(MissionInfos, &types.MissionInfo{Mission: mission, Accomplished: false})
		}
		marshal, _ := json.Marshal(MissionInfos)
		l.svcCtx.UserMission.Insert(l.ctx, &cachemodel.UserMission{Finished: 0, Phone: l.phone, MissionInfo: string(marshal)})
	} else {
		json.Unmarshal([]byte(phone.MissionInfo), &MissionInfos)
		if len(MissionInfos) != len(l.MissionList) {
			for _, mission := range l.MissionList {
				MissionInfos = append(MissionInfos, &types.MissionInfo{Mission: mission, Accomplished: false})
			}
			marshal, _ := json.Marshal(MissionInfos)
			phone.MissionInfo = string(marshal)
			l.svcCtx.UserMission.Update(l.ctx, phone)
		}
	}
	GetMissionInfoRp.MissionInfoList = MissionInfos
	return &types.GetMissionInfoResp{Code: "10000", Msg: "success", Data: GetMissionInfoRp}, err
}
