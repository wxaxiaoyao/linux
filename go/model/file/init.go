package file

var (
	FILE_DB               = "dudb"
	FILE_QINIU_FILE_TABLE = "qiniu_file"
)

func init() {
	if err := QiniuInit(); err != nil {
		panic(err)
	}
}
