package orderpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpc"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"oa_final/cachemodel"
	"oa_final/internal/svc"
	"oa_final/internal/types"
	"strconv"
	"strings"
	"time"
)

type SfUtilLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSfUtilLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SfUtilLogic {
	return &SfUtilLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func GetRoutesList(SfSn string) *types.RouteList {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	Timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	MsgDataStruct := &types.RouteMsgData{
		TrackingType:   1,
		TrackingNumber: SfSn,
	}
	MsgDataByte, _ := json.Marshal(MsgDataStruct)
	ToVerifyText := string(MsgDataByte) + Timestamp + svc.CheckCodeSbox
	ToVerifyText = url.QueryEscape(ToVerifyText)
	MsgDigest := md5V(ToVerifyText)
	params := url.Values{}
	params.Add("serviceCode", svc.GetRoutesServiceCode)
	params.Add("partnerID", svc.ParterID)
	params.Add("requestID", strconv.FormatInt(time.Now().UnixNano(), 10))
	params.Add("timestamp", Timestamp)
	params.Add("msgDigest", MsgDigest)
	params.Add("msgData", string(MsgDataByte))
	urlPath := "https://sfapi-sbox.sf-express.com/std/service"
	urlPath = urlPath + "?" + params.Encode()
	resp, err := httpc.Do(context.Background(), http.MethodPost, urlPath, nil)
	if err != nil {
		fmt.Println(err)
	}
	res := &types.MotherResponse{}
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, res)
	ApiResultDatastruct := &types.RouteResponse{}
	err = json.Unmarshal([]byte(res.ApiResultData), ApiResultDatastruct)
	resp.Body.Close()
	return ApiResultDatastruct.MsgData.RouteResps[0]
}
func (l SfUtilLogic) GetPDF(order *cachemodel.Order, sfsn string) {
	order.DeliverySn = sfsn
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	Timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	MsgDataStruct := &types.WayBillsSheetMsgData{
		TemplateCode: svc.TemplateCode,
		Documents:    []*types.WayBillInfo{{MasterWaybillNo: order.DeliverySn, Remark: order.ProductInfo}},
		Version:      "2.0",
		Sync:         true,
	}
	MsgDataByte, _ := json.Marshal(MsgDataStruct)
	ToVerifyText := string(MsgDataByte) + Timestamp + svc.CheckCodeSbox
	ToVerifyText = url.QueryEscape(ToVerifyText)
	MsgDigest := md5V(ToVerifyText)
	params := url.Values{}
	params.Add("serviceCode", svc.DownPDFServiceCode)
	params.Add("partnerID", svc.ParterID)
	params.Add("requestID", strconv.FormatInt(time.Now().UnixNano(), 10))
	params.Add("timestamp", Timestamp)
	params.Add("msgDigest", MsgDigest)
	params.Add("msgData", string(MsgDataByte))
	urlPath := "https://sfapi-sbox.sf-express.com/std/service"
	urlPath = urlPath + "?" + params.Encode()
	resp, err := httpc.Do(context.Background(), http.MethodPost, urlPath, nil)
	if err != nil {
		fmt.Println(err)
	}
	res := &types.MotherResponse{}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, res)
	ApiResultDatastruct := &types.WayBillSheetResponse{}
	json.Unmarshal([]byte(res.ApiResultData), ApiResultDatastruct)
	fmt.Println(resp.Body.Close())
	data := &types.UpdateArticle{Authorization: ApiResultDatastruct.Obj.Files[0].Token}
	downpdf, _ := httpc.Do(context.Background(), http.MethodGet, ApiResultDatastruct.Obj.Files[0].Url, data)
	if downpdf != nil {
		pdf, _ := ioutil.ReadAll(downpdf.Body)
		//os.WriteFile("sftest/test150.pdf", pdf, 0666)
		reader := bytes.NewReader(pdf)
		key := "img/sfpdf/" + order.DeliverySn + ".pdf"
		l.svcCtx.FileClient.Object.Put(context.Background(), key, reader, nil)
		//realPictureUrl := "https://img.waterflowfit.top/" + key
	}
}
func RefundSfOrder(order cachemodel.Order) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	Timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	MsgDataStruct := &types.RefundMsgData{
		OrderId:  GetSha256(order.OrderSn + order.Address),
		DealType: 2,
	}
	MsgDataByte, _ := json.Marshal(MsgDataStruct)
	ToVerifyText := string(MsgDataByte) + Timestamp + svc.CheckCodeSbox
	ToVerifyText = url.QueryEscape(ToVerifyText)
	MsgDigest := md5V(ToVerifyText)
	params := url.Values{}
	params.Add("serviceCode", svc.RefundServiceCode)
	params.Add("partnerID", svc.ParterID)
	params.Add("requestID", strconv.FormatInt(time.Now().UnixNano(), 10))
	params.Add("timestamp", Timestamp)
	params.Add("msgDigest", MsgDigest)
	params.Add("msgData", string(MsgDataByte))
	urlPath := "https://sfapi-sbox.sf-express.com/std/service"
	urlPath = urlPath + "?" + params.Encode()
	resp, _ := httpc.Do(context.Background(), http.MethodPost, urlPath, nil)
	resp.Body.Close()
}
func (l SfUtilLogic) GetSfSn(order *cachemodel.Order) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	_, sfsn := CreateOrder(order)
	if sfsn != "" {
		l.svcCtx.Order.UpdateDeliver(l.ctx, sfsn, "顺丰", order.OrderSn)
		l.GetPDF(order, sfsn)
	} else {
		ok, qsfsn := QuerySfSn(order)
		if ok {
			sfsn = qsfsn
			l.svcCtx.Order.UpdateDeliver(l.ctx, sfsn, "顺丰", order.OrderSn)
			l.GetPDF(order, sfsn)
		} else {
			l.svcCtx.ErrLog.Insert(l.ctx, &cachemodel.ErrLog{
				Interface: "顺丰单号无法获取，请手动查看",
				Info:      order.OrderSn,
				Time:      time.Now(),
			})
		}
	}
}
func (l SfUtilLogic) IfDelivering(order *cachemodel.Order) {
	routelist := GetRoutesList(order.DeliverySn)
	for _, route := range routelist.Routes {
		if route.OpCode == "50" || route.OpCode == "30" || strings.Contains(route.Remark, "收件") || strings.Contains(route.Remark, "已收取") || strings.Contains(route.Remark, "揽收") || strings.Contains(route.Remark, "已揽收") {
			l.svcCtx.Order.UpdateStatusByOrderSn(l.ctx, 2, order.OrderSn)
		}
	}
}

func (l SfUtilLogic) IfReceived(order *cachemodel.Order) {
	routelist := GetRoutesList(order.DeliverySn)
	for _, route := range routelist.Routes {
		if route.OpCode == "80" || strings.Contains(route.Remark, "已签收") {
			l.svcCtx.Order.UpdateReceivedByOrderSn(l.ctx, order.OrderSn)
		}
	}
}
func orderdb2sfinfodb(order *cachemodel.Order, SfSn string) *cachemodel.SfInfo {
	return &cachemodel.SfInfo{
		OutTradeNo:  order.OutTradeNo,
		OrderSn:     order.OrderSn,
		DeliverySn:  SfSn,
		Phone:       order.Phone,
		OrderNote:   order.OrderNote,
		ProductInfo: order.ProductInfo,
	}
}

func QuerySfSn(order *cachemodel.Order) (bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	Timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	MsgDataStruct := &types.QueryMsgData{
		OrderId: GetSha256(order.OrderSn + order.Address),
	}
	MsgDataByte, _ := json.Marshal(MsgDataStruct)
	ToVerifyText := string(MsgDataByte) + Timestamp + svc.CheckCodeSbox
	ToVerifyText = url.QueryEscape(ToVerifyText)
	MsgDigest := md5V(ToVerifyText)
	params := url.Values{}
	params.Add("serviceCode", svc.QueryOrderServiceCode)
	params.Add("partnerID", svc.ParterID)
	params.Add("requestID", strconv.FormatInt(time.Now().UnixNano(), 10))
	params.Add("timestamp", Timestamp)
	params.Add("msgDigest", MsgDigest)
	params.Add("msgData", string(MsgDataByte))
	urlPath := svc.SfUrl
	urlPath = urlPath + "?" + params.Encode()
	resp, err := httpc.Do(context.Background(), http.MethodPost, urlPath, nil)
	if err != nil {
		fmt.Println(err)
	}
	res := &types.MotherResponse{}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, res)
	resp.Body.Close()
	ApiResultDatastruct := &types.QueryResultData{}
	err = json.Unmarshal([]byte(res.ApiResultData), ApiResultDatastruct)
	SfSn := ApiResultDatastruct.MsgData.WaybillNoInfoList[0].WaybillNo
	return true, SfSn
}
func CreateOrder(order *cachemodel.Order) (status int, SfSn string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	Timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	PostInfo := GetPostInfo(order)
	ReceiveInfo := GetReceiveInfo(order)
	contactinfolist := []*types.ContactInfo{PostInfo, ReceiveInfo}
	CargoDetailList := []*types.CargoDetail{{Name: "毅明生鲜"}}
	MsgDataStruct := &types.CreateOrderMsgData{
		PayMethod: 1, Language: "zh-CN",
		OrderId:            GetSha256(order.OrderSn + order.Address),
		ContactInfoList:    contactinfolist,
		MonthlyCard:        svc.MonthlyCard,
		ExpressTypeId:      1,
		IsReturnRoutelabel: 1,
		CargoDetails:       CargoDetailList,
	}
	MsgDataByte, _ := json.Marshal(MsgDataStruct)
	ToVerifyText := string(MsgDataByte) + Timestamp + svc.CheckCodeSbox
	ToVerifyText = url.QueryEscape(ToVerifyText)
	MsgDigest := md5V(ToVerifyText)
	params := url.Values{}
	params.Add("serviceCode", svc.CreateOrderServiceCode)
	params.Add("partnerID", svc.ParterID)
	params.Add("requestID", strconv.FormatInt(time.Now().UnixNano()+int64(rand.Intn(2000000000)), 10))
	params.Add("timestamp", Timestamp)
	params.Add("msgDigest", MsgDigest)
	params.Add("msgData", string(MsgDataByte))
	urlPath := svc.SfUrl
	urlPath = urlPath + "?" + params.Encode()
	response, _ := httpc.Do(context.Background(), http.MethodPost, urlPath, nil)
	if response != nil {
		motherResponse := &types.MotherResponse{}
		body, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(body, motherResponse)
		if strings.Contains(motherResponse.ApiResultData, "重复下单") {
			return 2, ""
		}
		ApiResultDatastruct := &types.ApiResultData{}
		json.Unmarshal([]byte(motherResponse.ApiResultData), ApiResultDatastruct)
		return 1, ApiResultDatastruct.MsgData.RouteLabelInfo[0].RouteLabelData.WaybillNo
		response.Body.Close()
	}
	return 0, ""

}
func GetPostInfo(order *cachemodel.Order) *types.ContactInfo {
	return &types.ContactInfo{
		Address:     "浦东新区创新中路199弄齐爱家园",
		Contact:     "宋睿",
		Mobile:      "17854230846",
		Province:    "上海市",
		City:        "上海市",
		ContactType: 1,
		Country:     "CN",
		PostCode:    "200000",
	}
}
func GetReceiveInfo(order *cachemodel.Order) *types.ContactInfo {
	orderaddress := &types.AddressInfo{}
	err := json.Unmarshal([]byte(order.Address), orderaddress)
	if err == nil {
		return &types.ContactInfo{
			Address:     "上海市浦东新区环庆南路508号绿地香港菁舍公寓",
			Contact:     "宋睿",
			Mobile:      "17854230846",
			Province:    "上海市",
			City:        "上海市",
			ContactType: 2,
			Country:     "CN",
			PostCode:    "200000",
		}
	}
	ReceiveInfo := Address2SFAddress(orderaddress)
	ReceiveInfo.ContactType = 2
	return ReceiveInfo
}
func Address2SFAddress(addr *types.AddressInfo) *types.ContactInfo {
	sf := &types.ContactInfo{}
	sf.City = addr.City
	sf.Province = addr.Province
	sf.Address = addr.City + addr.Province + addr.DetailAddress + addr.DetailName + addr.RoomNumber
	sf.Contact = addr.Name
	sf.Mobile = addr.AddressPhone
	sf.PostCode = addr.PostCode
	return sf
}
