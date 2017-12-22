package toolkit

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	// ErrOk OK
	ErrOk = 0
	// ErrNotFound 404 route not found
	ErrNotFound = 1001
	// ErrException 500
	ErrException = 1002
	// ErrBadRequest 400 route params error
	ErrBadRequest = 1003
	// ErrMethodNotAllowed 405 不允许的请求方式
	ErrMethodNotAllowed = 1004
	// ErrParamsError 415 请求参数或格式错误 (路由参数或提交参数)
	ErrParamsError = 1005
	// ErrUnAuthorized 401 未登录
	ErrUnAuthorized = 1006
	// ErrDataNotFound 404 数据未找到
	ErrDataNotFound = 1007
	// ErrNotAllowed 405 没有访问权限
	ErrNotAllowed = 1008
	// ErrDataExists 400 数据已存在
	ErrDataExists = 1009
	// ErrDataValidate 403 数据验证错误
	ErrDataValidate = 1010

	// VarUserAuthorization 传递用户验证信息
	VarUserAuthorization = `access_token`

	// HTTPHeaderAuthorization HTTP header Authorization
	HTTPHeaderAuthorization = `Authorization`
)

var (
	statusMessage map[int]string
)

// ReplyData define API output data
type ReplyData struct {
	Status  int               `json:"status"`           // 状态码
	Message string            `json:"message"`          // 状态描述
	Errs    map[string]string `json:"errors,omitempty"` // 错误列表
	Total   int               `json:"total,omitempty"`  // 分页总数
	List    interface{}       `json:"rows,omitempty"`   // 数据列表
	Data    interface{}       `json:"data,omitempty"`   // 数据属性
}

// Pager define page request
type Pager struct {
	Page     int `json:"page" form:"page"`         // 当前页
	PageSize int `json:"pagesize" form:"pagesize"` // 每页数据条数
	Total    int `json:"total,omitempty"`          // 分页总数
}

func init() {
	statusMessage = make(map[int]string)
	statusMessage[ErrOk] = `ok`
	statusMessage[ErrNotFound] = `路由错误`
	statusMessage[ErrException] = `数据异常`
	statusMessage[ErrBadRequest] = `路由参数错误`
	statusMessage[ErrMethodNotAllowed] = `不允许的请求方式`
	statusMessage[ErrParamsError] = `参数或格式错误`
	statusMessage[ErrUnAuthorized] = `未登录或会话期已失效`
	statusMessage[ErrDataNotFound] = `数据未找到`
	statusMessage[ErrNotAllowed] = `没有访问权限`
	statusMessage[ErrDataExists] = `数据已存在`
	statusMessage[ErrDataValidate] = `数据验证错误`
}

// NewReplyData creates and return ReplyData with status and message
func NewReplyData(status int) *ReplyData {
	var (
		text   string
		exists bool
	)
	if text, exists = statusMessage[status]; !exists {
		text = `新错误类型`
	}
	return &ReplyData{
		Status:  status,
		Message: text,
	}
}

// OkReplyData creates and return ReplyData with ok
func OkReplyData() *ReplyData {
	message, _ := statusMessage[ErrOk]
	return &ReplyData{
		Status:  ErrOk,
		Message: message,
	}
}

// ErrReplyData creates and return ReplyData with error and message
func ErrReplyData(status int, message string) *ReplyData {
	text, _ := statusMessage[status]
	errs := map[string]string{
		"message": message,
	}
	return &ReplyData{
		Status:  status,
		Message: text,
		Errs:    errs,
	}
}

// ErrorsReplyData creates and return ReplyData with errors
func ErrorsReplyData(status int, errors map[string]string) *ReplyData {
	message, _ := statusMessage[status]
	return &ReplyData{
		Status:  status,
		Message: message,
		Errs:    errors,
	}
}

// RowsReplyData creates and return ReplyData with total and list
func RowsReplyData(total int, rows interface{}) *ReplyData {
	message, _ := statusMessage[ErrOk]
	return &ReplyData{
		Status:  ErrOk,
		Message: message,
		List:    rows,
		Total:   total,
	}
}

// RowReplyData creates and return ReplyData with attr row
func RowReplyData(row interface{}) *ReplyData {
	message, _ := statusMessage[ErrOk]
	return &ReplyData{
		Status:  ErrOk,
		Message: message,
		Data:    row,
	}
}

// HTTPWriteJSON response JSON data.
func HTTPWriteJSON(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Power", "csacred/0.2")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

// HTTPWriteString response string
func HTTPWriteString(w http.ResponseWriter, response interface{}) error {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response.([]byte))
	return nil
}

// HTTPWriteCtxJSON response JSON data.
func HTTPWriteCtxJSON(
	_ context.Context,
	w http.ResponseWriter,
	response interface{}) error {
	return HTTPWriteJSON(w, response)
}

// HTTPWriteCtxString response text data.
func HTTPWriteCtxString(
	_ context.Context,
	w http.ResponseWriter,
	response interface{}) error {
	return HTTPWriteString(w, response)
}

// HTTPEncodeError request encode error response
func HTTPEncodeError(_ context.Context, err error, w http.ResponseWriter) {
	HTTPWriteJSON(w, ErrReplyData(ErrParamsError, err.Error()))
}

// HTTPDecodeResponse decode client
func HTTPDecodeResponse(
	ctx context.Context,
	r *http.Response) (interface{}, error) {
	var response ReplyData
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return ErrReplyData(ErrException, `服务端数据格式错误`), err
	}
	return response, nil
}

// HTTPDecodeStringResponse decode client
func HTTPDecodeStringResponse(
	ctx context.Context,
	r *http.Response) (interface{}, error) {

	return ioutil.ReadAll(r.Body)
}

// PopulateRequestContext is a RequestFunc that populates several values into
// the context from the HTTP request. Those values may be extracted using the
// corresponding ContextKey type in this package.
func PopulateRequestContext(
	ctx context.Context,
	r *http.Request) context.Context {
	var accessToken string
	accessToken = r.URL.Query().Get(VarUserAuthorization)
	if accessToken == "" {
		if cookie, err := r.Cookie(VarUserAuthorization); err == nil {
			accessToken, _ = url.QueryUnescape(cookie.Value)
		}
	}

	token := r.Header.Get(HTTPHeaderAuthorization)
	if accessToken == "" {
		if len(token) > 6 && strings.ToUpper(token[0:7]) == "BEARER " {
			accessToken = token[7:]
		}
	}

	for k, v := range map[contextKey]string{
		ContextKeyRequestMethod:          r.Method,
		ContextKeyRequestURI:             r.RequestURI,
		ContextKeyRequestPath:            r.URL.Path,
		ContextKeyRequestProto:           r.Proto,
		ContextKeyRequestHost:            r.Host,
		ContextKeyRequestRemoteAddr:      r.RemoteAddr,
		ContextKeyRequestXForwardedFor:   r.Header.Get("X-Forwarded-For"),
		ContextKeyRequestXForwardedProto: r.Header.Get("X-Forwarded-Proto"),
		ContextKeyRequestAuthorization:   token,
		ContextKeyRequestReferer:         r.Header.Get("Referer"),
		ContextKeyRequestUserAgent:       r.Header.Get("User-Agent"),
		ContextKeyRequestXRequestID:      r.Header.Get("X-Request-Id"),
		ContextKeyRequestAccept:          r.Header.Get("Accept"),
		ContextKeyAccessToken:            accessToken,
	} {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx
}

type contextKey int

const (
	// ContextKeyRequestMethod is populated in the context by
	// PopulateRequestContext. Its value is r.Method.
	ContextKeyRequestMethod contextKey = iota

	// ContextKeyRequestURI is populated in the context by
	// PopulateRequestContext. Its value is r.RequestURI.
	ContextKeyRequestURI

	// ContextKeyRequestPath is populated in the context by
	// PopulateRequestContext. Its value is r.URL.Path.
	ContextKeyRequestPath

	// ContextKeyRequestProto is populated in the context by
	// PopulateRequestContext. Its value is r.Proto.
	ContextKeyRequestProto

	// ContextKeyRequestHost is populated in the context by
	// PopulateRequestContext. Its value is r.Host.
	ContextKeyRequestHost

	// ContextKeyRequestRemoteAddr is populated in the context by
	// PopulateRequestContext. Its value is r.RemoteAddr.
	ContextKeyRequestRemoteAddr

	// ContextKeyRequestXForwardedFor is populated in the context by
	// PopulateRequestContext. Its value is r.Header.Get("X-Forwarded-For").
	ContextKeyRequestXForwardedFor

	// ContextKeyRequestXForwardedProto is populated in the context by
	// PopulateRequestContext. Its value is r.Header.Get("X-Forwarded-Proto").
	ContextKeyRequestXForwardedProto

	// ContextKeyRequestAuthorization is populated in the context by
	// PopulateRequestContext. Its value is r.Header.Get("Authorization").
	ContextKeyRequestAuthorization

	// ContextKeyRequestReferer is populated in the context by
	// PopulateRequestContext. Its value is r.Header.Get("Referer").
	ContextKeyRequestReferer

	// ContextKeyRequestUserAgent is populated in the context by
	// PopulateRequestContext. Its value is r.Header.Get("User-Agent").
	ContextKeyRequestUserAgent

	// ContextKeyRequestXRequestID is populated in the context by
	// PopulateRequestContext. Its value is r.Header.Get("X-Request-Id").
	ContextKeyRequestXRequestID

	// ContextKeyRequestAccept is populated in the context by
	// PopulateRequestContext. Its value is r.Header.Get("Accept").
	ContextKeyRequestAccept

	// ContextKeyResponseHeaders is populated in the context whenever a
	// ServerFinalizerFunc is specified. Its value is of type http.Header, and
	// is captured only once the entire response has been written.
	ContextKeyResponseHeaders

	// ContextKeyResponseSize is populated in the context whenever a
	// ServerFinalizerFunc is specified. Its value is of type int64.
	ContextKeyResponseSize

	// ContextKeyAccessToken 登录验证信息
	ContextKeyAccessToken
)
