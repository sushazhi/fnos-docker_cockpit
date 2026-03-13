package model

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	IDs       []string `json:"ids" binding:"required"`
	Operation string   `json:"operation" binding:"required"`
	Force     bool     `json:"force"`
	Timeout   int      `json:"timeout"`
}

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
	Success []string               `json:"success"`
	Failed  []BatchOperationError  `json:"failed"`
}

// BatchOperationError 批量操作错误
type BatchOperationError struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}
