package utils

import (
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"r2_example/dto"
)

func R2TypesObjectToListsFileDto(r []types.Object) []dto.R2ListFile {
	var fileListsResponse []dto.R2ListFile
	fileList := new(dto.R2ListFile)
	for _, v := range r {
		fileList.FileName = *v.Key
		fileList.Size = *v.Size
		fileListsResponse = append(fileListsResponse, *fileList)
	}
	return fileListsResponse
}
