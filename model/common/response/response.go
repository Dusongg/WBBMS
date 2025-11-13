package response

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func Ok(data interface{}, msg string) Response {
	return Response{
		Code: 200,
		Data: data,
		Msg:  msg,
	}
}

func OkWithMessage(msg string) Response {
	return Response{
		Code: 200,
		Data: nil,
		Msg:  msg,
	}
}

func OkWithData(data interface{}) Response {
	return Response{
		Code: 200,
		Data: data,
		Msg:  "操作成功",
	}
}

func OkWithDetailed(data interface{}, msg string) Response {
	return Response{
		Code: 200,
		Data: data,
		Msg:  msg,
	}
}

func Fail(msg string) Response {
	return Response{
		Code: 500,
		Data: nil,
		Msg:  msg,
	}
}

func FailWithMessage(msg string) Response {
	return Response{
		Code: 500,
		Data: nil,
		Msg:  msg,
	}
}

