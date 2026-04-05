package response

import "time"

const (
	MessageOK = "ok"
)

type MessageDetail struct {
	TitleEng string `json:"title_eng,omitempty"`
	DescEng  string `json:"desc_eng,omitempty"`
	TitleIdn string `json:"title_idn,omitempty"`
	DescIdn  string `json:"desc_idn,omitempty"`
}

type CustomError struct {
	Code       string
	Message    string
	StatusCode int
	Detail     MessageDetail
	// Data is optional JSON payload (e.g. validation issue list). Omitted in JSON when nil.
	Data any
}

func (m CustomError) Error() string {
	return m.Message
}

func (m CustomError) ToResponse() BaseResponse {
	msg := m.Code
	if msg == "" {
		msg = m.Message
	}

	detail := m.Detail
	if detail == (MessageDetail{}) && m.Message != "" {
		detail = MessageDetail{
			TitleEng: "Error",
			DescEng:  m.Message,
			TitleIdn: "Error",
			DescIdn:  m.Message,
		}
	}

	return BaseResponse{
		StatusCode:    m.StatusCode,
		Message:       msg,
		MessageDetail: detail,
		Data:          m.Data,
	}
}

type Meta struct {
	Page      int `json:"page"`
	PageSize  int `json:"page_size"`
	TotalData int `json:"total_data"`
}

type BaseResponse struct {
	StatusCode    int           `json:"-"`
	Message       string        `json:"message"`
	MessageDetail MessageDetail `json:"message_detail"`
	Data          any           `json:"data"`
	Meta          any           `json:"meta,omitempty"`
	TraceID       string        `json:"trace_id"`
	Timestamp     time.Time     `json:"timestamp"`
}

func NewResponseOK() *BaseResponse {
	return &BaseResponse{
		Message: MessageOK,
		MessageDetail: MessageDetail{
			TitleEng: "SUCCESS",
			DescEng:  "Operation completed successfully",
			TitleIdn: "SUKSES",
			DescIdn:  "Operasi berhasil diselesaikan",
		},
	}
}

type GetHealthCheckMemoryResp struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	HeapAlloc  uint64 `json:"heap_alloc"`
	HeapSys    uint64 `json:"heap_sys"`
}

type GetHealthCheckServiceStatusResp struct {
	Name string `json:"name"`
	IsUp bool   `json:"is_up"`
}

type GetHealthCheckResp struct {
	Status          string                            `json:"status"`
	Environtment    string                            `json:"environtment"`
	Version         string                            `json:"version"`
	GoVersion       string                            `json:"go_version"`
	GoRoutine       int                               `json:"go_routine"`
	Memory          GetHealthCheckMemoryResp          `json:"memory"`
	ServiceStatuses []GetHealthCheckServiceStatusResp `json:"service_statuses"`
}
