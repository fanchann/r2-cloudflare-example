package controller

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zakirkun/dy"

	"r2_example/dto"
	"r2_example/services/r2"
	"r2_example/utils"
)

const maxUploadSize = 5 * 1024 * 1024 // 5 MB

type IR2Controller interface {
	UploadFile(c echo.Context) error
	GetListsFile(c echo.Context) error
	MakeFilePublic(c echo.Context) error
	GetFileByID(c echo.Context) error
}

type r2Controller struct {
	log *dy.Logger
	r2  r2.IR2Services
}

func NewR2Controller(log *dy.Logger, r2 r2.IR2Services) IR2Controller {
	return &r2Controller{log: log, r2: r2}
}

func (r *r2Controller) UploadFile(c echo.Context) error {
	var response dto.WebAPIResponse

	// limit upload file size
	c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, maxUploadSize)

	if err := c.Request().ParseMultipartForm(maxUploadSize); err != nil {
		r.log.Debug(err.Error())
		response.Success = false
		response.Message = err.Error()
		return c.JSON(http.StatusBadRequest, response)
	}

	// check total file uploaded
	if len(c.Request().MultipartForm.File["file"]) > 1 {
		r.log.Debug("multiple file uploads are not allowed")
		response.Success = false
		response.Message = "only one file can be uploaded at a time"
		return c.JSON(http.StatusBadRequest, response)
	}

	file, err := c.FormFile("file")
	if err != nil {
		r.log.Debug(err.Error())
		response.Success = false
		response.Message = err.Error()
		return c.JSON(http.StatusBadRequest, response)
	}

	fileSrc, err := file.Open()
	if err != nil {
		r.log.Debug(err.Error())
		response.Success = false
		response.Message = err.Error()
		return c.JSON(http.StatusInternalServerError, response)
	}
	defer fileSrc.Close()

	fileExt := filepath.Ext(file.Filename)
	fileId := fmt.Sprintf("%s%s", utils.NewIDGen(), fileExt)

	singleUploadResponse, err := r.r2.UploadSingleFile(c.Request().Context(), fileSrc, fileId)
	if err != nil {
		r.log.Debug(err.Error())
		response.Success = false
		response.Message = err.Error()
		return c.JSON(http.StatusRequestTimeout, response)
	}

	response.Success = true
	response.Data = singleUploadResponse

	return c.JSON(http.StatusCreated, response)
}

func (r *r2Controller) GetListsFile(c echo.Context) error {
	var response dto.WebAPIResponse
	listsFile, err := r.r2.GetListFile(context.Background())
	if err != nil {
		r.log.Debug(err.Error())
		response.Success = true
		response.Message = "file Not Found"
		return c.JSON(http.StatusNotFound, response)
	}

	response.Success = true
	response.Data = listsFile
	return c.JSON(http.StatusFound, response)
}

func (r *r2Controller) MakeFilePublic(c echo.Context) error {
	var makePubFile dto.MakeFilePublicRequest
	var response dto.WebAPIResponse

	if errBind := c.Bind(&makePubFile); errBind != nil {
		response.Success = false
		response.Message = errBind.Error()
		return c.JSON(http.StatusBadRequest, response)
	}

	_, err := r.r2.GetFileByKey(c.Request().Context(), makePubFile.FiileID)
	if err != nil {
		response.Success = false
		response.Message = err.Error()
		return c.JSON(http.StatusNotFound, response)
	}

	duration, err := time.ParseDuration(makePubFile.Duration)
	if err != nil {
		response.Success = false
		response.Message = err.Error()
		return c.JSON(http.StatusBadRequest, response)
	}

	r2Response, err := r.r2.GenSignedURL(context.Background(), makePubFile.FiileID, duration)
	if err != nil {
		response.Success = false
		response.Message = err.Error()
		return c.JSON(http.StatusNotFound, response)
	}

	response.Success = true
	response.Data = r2Response
	return c.JSON(http.StatusAccepted, response)

}

func (r *r2Controller) GetFileByID(c echo.Context) error {
	var response dto.WebAPIResponse

	id := c.Param("id")

	foundFile, err := r.r2.GetFileByKey(c.Request().Context(), id)
	if err != nil {
		response.Success = false
		response.Message = err.Error()
		return c.JSON(http.StatusNotFound, response)
	}

	response.Success = true
	response.Data = foundFile
	return c.JSON(http.StatusFound, response)
}
