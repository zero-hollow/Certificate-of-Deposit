package bstore

const (
	contractName          = "storage_contract"
	fileUploadCmd         = "fileUpload"
	fileDownloadCmd       = "fileDownload"
	bSAddressQueryCmd     = "getBsAddress"
	instanceIdQueryCmd    = "getInstanceId"
	fileHistoryQueryCmd   = "fileHistory"
	fileOperationQueryCmd = "fileOperation"
	uploadURI             = "/v1/blockstorage/user/upload"
	downloadURI           = "/v1/blockstorage/user/download"
	bsAddressPrefix       = "https://"
	bsAddressPortSuffix   = ":32117"
	contractNotFoundErr   = "NOT_FOUND"
)

type CommonParam struct {
	InstanceID string `json:"instance_id"`
	OrgID      string `json:"org_id"`
	Cert       string `json:"cert"`
	SignedData string `json:"signed_data"`
	Data       string `json:"data"`
}

type UploadParam struct {
	Filename string `json:"file_name"`
}

type DownloadParam struct {
	FileName string `json:"file_name"`
	Version  string `json:"version"`
}

type UploadFileResponse struct {
	FileName string
	FileHash string
	Version  string
}

type FileHistory struct {
	CreatedTime string
	UpdatedTime string
	Version     int
	Uploader    string
	HashCode    string
}

type FileInfo struct {
	CreatedTime     string `json:"CreatedTime"`
	UpdatedTime     string `json:"UpdatedTime"`
	Version         int    `json:"Version"`
	InternalVersion string `json:"InternalVersion"`
	Uploader        string `json:"Uploader"`
	HashCode        string `json:"HashCode"`
}

type StorageEvent struct {
	Operator  string `json:"Operator"`
	Time      string `json:"Time"`
	EventType string `json:"EventType"`
}
