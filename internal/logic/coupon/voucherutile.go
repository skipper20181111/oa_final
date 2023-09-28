package coupon

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	CRand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"math/rand"
	"oa_final/cachemodel"
	"oa_final/internal/logic/orderpay"

	"oa_final/internal/svc"
	"oa_final/internal/types"
	"strconv"
	"time"
)

type VoucherUtileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	lu     *orderpay.UtilLogic
	Phone  string
}

func NewVoucherUtileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VoucherUtileLogic {
	return &VoucherUtileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		lu:     orderpay.NewUtilLogic(ctx, svcCtx),
		Phone:  ctx.Value("phone").(string),
	}
}
func (l VoucherUtileLogic) CouponBindByCid(QrMsg *types.QrCode) (bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	overtime, _ := time.Parse("2006-01-02 15:04:05", QrMsg.Parameter2)
	if overtime.Before(time.Now()) {
		return false, "过期了"
	}
	get, ok := l.svcCtx.LocalCache.Get(svc.CouponInfoMapKey)
	if !ok {
		return false, "no"
	}
	couponinfomap := get.(map[int64]*types.CouponInfo)
	couponbyphone, _ := l.svcCtx.UserCoupon.FindOneByPhone(l.ctx, l.Phone)
	couponmap := make(map[int64]map[string]*types.CouponStoreInfo)
	Cid, _ := strconv.ParseInt(QrMsg.Parameter1, 10, 64)
	json.Unmarshal([]byte(couponbyphone.CouponIdMap), &couponmap)
	singleCouponInfo, ok := couponinfomap[Cid]
	if ok {
		_, ok := couponmap[Cid]
		if ok {
			return false, "不能重复扫码"
		} else {
			couponmap[Cid] = make(map[string]*types.CouponStoreInfo)
			couponmap[Cid][strconv.FormatInt(time.Now().UnixNano()+rand.Int63n(10000), 10)] = &types.CouponStoreInfo{CouponId: Cid, DisabledTime: time.Now().Add(time.Hour * time.Duration(24*singleCouponInfo.EfficientPeriod)).Format("2006-01-02 15:04:05")}
		}
	}
	couponmapbyte, err := json.Marshal(couponmap)
	if err != nil {
		return false, "nonono"
	} else {
		couponbyphone.CouponIdMap = string(couponmapbyte)
		l.svcCtx.UserCoupon.Update(l.ctx, couponbyphone)
	}
	return true, "yes"
}
func insertCoupon(couponmap map[int64]map[string]*types.CouponStoreInfo, couponinfomap map[int64]*types.CouponInfo) map[int64]map[string]*types.CouponStoreInfo {
	newusercouponid := []int64{10000}
	for _, cid := range newusercouponid {
		info, ok := couponinfomap[cid]
		if ok {
			couponmap[cid] = make(map[string]*types.CouponStoreInfo)
			couponmap[cid][strconv.FormatInt(time.Now().UnixNano()+rand.Int63n(10000), 10)] = &types.CouponStoreInfo{CouponId: cid, DisabledTime: time.Now().Add(time.Hour * time.Duration(24*info.EfficientPeriod)).Format("2006-01-02 15:04:05")}
		}
	}
	return couponmap
}
func (l VoucherUtileLogic) VoucherbindByVid(QrMsg *types.QrCode) (bool, string) {
	vid, _ := strconv.ParseInt(QrMsg.Parameter1, 10, 64)
	if vid != 0 {
		byVid, _ := l.svcCtx.Voucher.FindOneByVid(l.ctx, vid)
		if byVid != nil {
			return l.voucherbind(byVid.VoucherCode, "扫码")
		}
	}
	return false, "请检查二维码"
}
func (l *VoucherUtileLogic) voucherbind(voucherCode, path string) (bool, string) {
	voucher, _ := l.svcCtx.Voucher.FindOneByVoucherCode(l.ctx, voucherCode)
	if voucher == nil {
		return false, "无此兑换码"
	}

	lockmsglist := make([]*types.LockMsg, 0)
	lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.Phone, Field: "voucher"})
	lockmsglist = append(lockmsglist, &types.LockMsg{Phone: l.Phone, Field: "cash_account"})
	if l.lu.Getlocktry(lockmsglist) {
		if ok, msg := l.privatizationvoucher(voucher, path); ok {
			l.lu.Closelock(lockmsglist)
			return true, msg
		} else {
			l.lu.Closelock(lockmsglist)
			return false, msg
		}
	}
	return false, "未获取到锁，请1分钟后再试"
}
func (l *VoucherUtileLogic) privatizationvoucher(voucher *cachemodel.Voucher, path string) (bool, string) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	//开始更新voucher表
	if voucher.Phone != "" {
		return false, "此兑换码已使用"
	}
	voucher.Phone = l.Phone
	voucherlid := time.Now().UnixNano() + int64(rand.Intn(1024))
	l.lu.Oplog("voucher and cash_account", "扫码获取大额券并更新账户", "开始更新", voucherlid)
	l.lu.Oplog("voucher", voucher.VoucherCode, "开始更新", voucherlid)
	l.svcCtx.Voucher.Update(l.ctx, voucher)
	oneByVid, _ := l.svcCtx.Voucher.FindOneByVid(l.ctx, voucher.Vid)
	if oneByVid.Phone == l.Phone {
		l.lu.Oplog("voucher", voucher.VoucherCode, "结束更新", voucherlid)
	}
	//	开始更新账户表
	couponbyid, _ := l.svcCtx.Coupon.FindOneByCouponId(l.ctx, voucher.CouponId)
	if couponbyid == nil {
		return false, "无此类型兑换码"
	}
	phone, _ := l.svcCtx.CashAccount.FindOneByPhone(l.ctx, l.Phone)
	l.lu.Oplog("cash_account", l.Phone, "开始更新", voucherlid)
	if phone == nil {
		l.svcCtx.CashAccount.Insert(l.ctx, &cachemodel.CashAccount{Phone: l.Phone, Balance: couponbyid.AvailableAmount})
	} else {
		phone.Balance = phone.Balance + couponbyid.AvailableAmount
		l.svcCtx.CashAccount.Update(l.ctx, phone)
	}
	switch path {
	case "扫码":
		l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), OrderType: "扫码", OrderSn: voucher.VoucherCode, OrderDescribe: "扫大额代金券码获取现金", Behavior: "充值", Phone: l.Phone, Balance: phone.Balance, ChangeAmount: couponbyid.AvailableAmount})
	case "兑换码":
		l.svcCtx.CashLog.Insert(l.ctx, &cachemodel.CashLog{Date: time.Now(), OrderType: "兑换", OrderSn: voucher.VoucherCode, OrderDescribe: "输入兑换码获取现金", Behavior: "充值", Phone: l.Phone, Balance: phone.Balance, ChangeAmount: couponbyid.AvailableAmount})
	}
	l.lu.Oplog("cash_account", l.Phone, "结束更新", voucherlid)
	l.lu.Oplog("voucher and cash_account", "扫码获取大额券并更新账户", "结束更新", voucherlid)
	return true, "绑定成功,请查看现金账户"
}

func Encrypt(msg string, keystr string) (string, error) {
	key := []byte(keystr)
	src := []byte(msg)
	//生成cipher.Block 数据块
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	} else if len(src) == 0 {
		return "", errors.New("src is empty")
	}

	//填充内容，如果不足16位字符
	blockSize := block.BlockSize()
	originData := pad(src, blockSize)

	//加密，输出到[]byte数组
	crypted := make([]byte, aes.BlockSize+len(originData))
	iv := crypted[:aes.BlockSize]
	if _, err := io.ReadFull(CRand.Reader, iv); err != nil {
		return "", nil
	}
	//加密方式
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(crypted[aes.BlockSize:], originData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func pad(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func Decrypt(src, keystr string) (string, error) {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	key := []byte(keystr)
	decode_data, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", nil
	}
	//生成密码数据块cipher.Block
	block, _ := aes.NewCipher(key)
	//解密模式
	blockMode := cipher.NewCBCDecrypter(block, decode_data[:aes.BlockSize])
	//输出到[]byte数组
	origin_data := make([]byte, len(decode_data)-aes.BlockSize)
	blockMode.CryptBlocks(origin_data, decode_data[aes.BlockSize:])
	//去除填充,并返回
	return string(unpad(origin_data)), nil
}

func unpad(ciphertext []byte) []byte {
	length := len(ciphertext)
	//去掉最后一次的padding
	unpadding := int(ciphertext[length-1])
	return ciphertext[:(length - unpadding)]
}
