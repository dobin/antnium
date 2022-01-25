package campaign

type WingmanData struct {
	FileExtension string
	FilenameReq   string
	FilenameAns   string
}

func MakeWingmanData() WingmanData {
	wingmanData := WingmanData{
		"tmp",
		"5feb6bf4-46b3-40a0-9d08-c62307f10387",
		"6d47b22b-d9cc-41c9-b469-773e578b65a3",
	}
	return wingmanData
}

func (wd WingmanData) Req() string {
	return wd.FilenameReq + "." + wd.FileExtension
}

func (wd WingmanData) Ans() string {
	return wd.FilenameAns + "." + wd.FileExtension
}
