package model

type SecureDownloadArgs struct {
	Filename string `json:"filename"`
	Argline  string `json:"argline"`
	IsDotnet bool   `json:"isDotnet"`
}
