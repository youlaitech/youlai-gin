package handler

import (
	"github.com/gin-gonic/gin"

	"youlai-gin/pkg/response"
	"youlai-gin/pkg/storage"
	"youlai-gin/pkg/utils"
)

// UploadResult 上传结果
type UploadResult struct {
	Name string `json:"name"` // 原始文件名
	URL  string `json:"url"`  // 访问URL
	Path string `json:"path"` // 存储路径
	Size int64  `json:"size"` // 文件大小
}

// UploadFile 单文件上传
// @Summary 文件上传
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param path formData string false "存储路径前缀"
// @Success 200 {object} response.Result{data=handler.UploadResult}
// @Router /api/v1/files [post]
func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "未选择文件")
		return
	}

	// 严格验证文件（使用文档验证器）
	if err := utils.ValidateDocument(file); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 获取存储路径前缀
	pathPrefix := c.DefaultPostForm("path", "uploads")

	// 生成文件名和路径
	filename := utils.GenerateFileName(file.Filename)
	path := utils.GeneratePath(pathPrefix, filename)

	// 打开文件
	src, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "打开文件失败")
		return
	}
	defer src.Close()

	// 上传文件
	contentType := utils.GetContentType(file.Filename)
	url, err := storage.DefaultStorage.Upload(path, src, contentType)
	if err != nil {
		response.InternalServerError(c, "上传失败: "+err.Error())
		return
	}

	result := UploadResult{
		Name: file.Filename,
		URL:  url,
		Path: path,
		Size: file.Size,
	}

	response.Ok(c, result)
}

// UploadFiles 批量文件上传
// @Summary 批量文件上传
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "文件列表"
// @Param path formData string false "存储路径前缀"
// @Success 200 {object} response.Result{data=[]handler.UploadResult}
// @Router /api/v1/files/batch [post]
func UploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "未选择文件")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.BadRequest(c, "未选择文件")
		return
	}

	pathPrefix := c.DefaultPostForm("path", "uploads")
	results := make([]UploadResult, 0, len(files))

	for _, file := range files {
		// 严格验证文件
		if err := utils.ValidateDocument(file); err != nil {
			continue // 跳过验证失败的文件
		}

		// 生成文件名和路径
		filename := utils.GenerateFileName(file.Filename)
		path := utils.GeneratePath(pathPrefix, filename)

		// 打开文件
		src, err := file.Open()
		if err != nil {
			continue
		}

		// 上传文件
		contentType := utils.GetContentType(file.Filename)
		url, err := storage.DefaultStorage.Upload(path, src, contentType)
		src.Close()

		if err != nil {
			continue
		}

		results = append(results, UploadResult{
			Name: file.Filename,
			URL:  url,
			Path: path,
			Size: file.Size,
		})
	}

	response.Ok(c, results)
}

// UploadImage 图片上传（带限制）
// @Summary 图片上传
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} response.Result{data=handler.UploadResult}
// @Router /api/v1/files/image [post]
func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "未选择文件")
		return
	}

	// 严格验证图片文件
	if err := utils.ValidateImage(file); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 生成文件名和路径
	filename := utils.GenerateFileName(file.Filename)
	path := utils.GeneratePath("images", filename)

	// 打开文件
	src, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "打开文件失败")
		return
	}
	defer src.Close()

	// 上传文件
	contentType := utils.GetContentType(file.Filename)
	url, err := storage.DefaultStorage.Upload(path, src, contentType)
	if err != nil {
		response.InternalServerError(c, "上传失败: "+err.Error())
		return
	}

	result := UploadResult{
		Name: file.Filename,
		URL:  url,
		Path: path,
		Size: file.Size,
	}

	response.Ok(c, result)
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Tags 文件管理
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} response.Result
// @Router /api/v1/files [delete]
func DeleteFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.BadRequest(c, "文件路径不能为空")
		return
	}

	err := storage.DefaultStorage.Delete(path)
	if err != nil {
		response.InternalServerError(c, "删除失败: "+err.Error())
		return
	}

	response.Ok(c, nil)
}
