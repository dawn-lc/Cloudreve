package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/HFO4/cloudreve/pkg/auth"
	"github.com/HFO4/cloudreve/pkg/conf"
	"github.com/HFO4/cloudreve/pkg/serializer"
	"github.com/HFO4/cloudreve/pkg/util"
	"time"
)

// RemoteCallback 发送远程存储策略上传回调请求
func RemoteCallback(url string, body serializer.RemoteUploadCallback) error {
	callbackBody, err := json.Marshal(struct {
		Data serializer.RemoteUploadCallback `json:"data"`
	}{
		Data: body,
	})
	if err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "无法编码回调正文", err)
	}

	resp := GeneralClient.Request(
		"POST",
		url,
		bytes.NewReader(callbackBody),
		WithTimeout(time.Duration(conf.SlaveConfig.CallbackTimeout)*time.Second),
		WithCredential(auth.General, int64(conf.SlaveConfig.SignatureTTL)),
	)

	if resp.Err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "无法发起回调请求", resp.Err)
	}

	// 检查返回HTTP状态码
	rawResp, err := resp.CheckHTTPResponse(200).GetResponse()
	if err != nil {
		return serializer.NewError(serializer.CodeCallbackError, "服务器返回异常响应", err)
	}

	// 解析回调服务端响应
	var response serializer.Response
	err = json.Unmarshal([]byte(rawResp), &response)
	if err != nil {
		util.Log().Debug("无法解析回调服务端响应：%s", string(rawResp))
		return serializer.NewError(serializer.CodeCallbackError, "无法解析服务端返回的响应", err)
	}

	if response.Code != 0 {
		return serializer.NewError(response.Code, response.Msg, errors.New(response.Error))
	}

	return nil
}
